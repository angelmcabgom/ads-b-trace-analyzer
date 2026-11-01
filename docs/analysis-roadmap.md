# Analysis roadmap

Ideas for the analysis side of ads-b-trace-analyzer, in layers. A few
foundation pieces come first because every interesting analysis depends on
them; after that, most of the payoff comes from joining traces against the
`airports` and `runways` tables.

## 1. Foundation (build these first)

- **Split traces into legs.** One trace file often holds several flights, as in
  `mock/trace-big.mock.json`, which sits on the ground and then departs. Split
  on the "new leg" flag bit (index 6), on long time gaps, or on ground →
  airborne changes. Most of the current detector confusion comes from treating
  a whole day as one flight.
- **Label each point with a flight phase:** ground, takeoff roll, climb,
  cruise, descent, approach, landing, go-around. Use a small state machine with
  hysteresis, meaning a phase only changes after a condition holds for a few
  seconds. `DetectTakeOff`, `DetectLanding` and `DetectGoAround` become
  transitions in that one machine instead of three detectors that disagree.
- **Use time windows, not point counts.** readsb records points every second or
  two near the ground but much less often at cruise. So "the last 30 points" in
  `DetectLanding` means a different amount of time in every trace.
- **Measure height above the airport, not above sea level.** Once the nearby
  airport is known, compute height as altitude minus `airports.elevation_ft`.
  Two caveats:
  - Barometric altitude isn't corrected for local air pressure, which can be
    off by hundreds of feet on low-pressure days. It can be fixed with the
    altimeter setting from METAR weather reports (aviationweather.gov has a
    free API).
  - Geometric altitude (index 10) is measured above the WGS84 ellipsoid, which
    differs from sea level by up to about ±100 m depending on location. It
    needs a geoid correction.
- **Fix the track/ground-speed swap** in `trace/models.go` before anything
  else. In the readsb format, index 4 is ground speed and index 5 is track;
  the struct currently has them the other way round, so every result built on
  them is wrong.

A structure that scales as detectors are added:

```go
type Event struct {
    Kind  string // "takeoff", "landing", "go_around", "hold", ...
    Time  time.Time
    Pos   orb.Point
    Attrs map[string]any // runway, glide-path deviation, etc.
}

type Detector interface {
    Detect(f *Flight) []Event // f already cleaned, split into legs, phase-labelled
}
```

Store events in a `flight_events` table (flight, kind, time, location, runway,
JSONB attributes). Store each flight's track as a PostGIS `LineStringZM` with
the timestamp in M. The viewer (`../ads-b-viewer`) can then show tracks and
events using the same tile approach it uses for airports.

## 2. Airport and runway matching

- **Origin and destination.** Take the first or last low point of each leg and
  find the nearest airport with a single indexed PostGIS query:
  `ORDER BY location <-> $pt LIMIT 1`. This fills in `flights` with real routes.
- **Which runway was used.** On final approach or takeoff roll, pick the runway
  where:
  - the aircraft's track is within about 15° of `le_heading_deg_t` or
    `he_heading_deg_t`, and
  - its sideways distance from the runway centreline (`geom`) is under about
    150 m.

  That gives output like "landed LEMD 32L". Aggregated into runway usage by
  hour, it shows an airport switching between east and west operations when
  the wind changes.
- **Touchdown point and rollout.** Measure the distance from the threshold
  (allowing for the displaced threshold) to the first ground point, and from
  there to taxi speed. That flags long landings, for example touchdown beyond
  the first third of the runway.
- **Runway occupancy time** and **turnarounds.** Occupancy is the time from
  crossing the threshold to leaving the runway. A turnaround is the same
  aircraft landing and later departing the same airport.

## 3. Approach quality

- **Glide path deviation.** The ideal 3° glide path is about 318 ft of height
  per nautical mile from the threshold. Comparing actual height to that at each
  point shows whether an approach was high or low.
- **Unstable approaches.** Flag approaches that aren't "stable" by 1000 ft (or
  500 ft) above the airport: descending faster than about 1000 ft/min, speed
  still changing a lot, or not lined up with the runway.
- **Go-around vs touch-and-go vs missed approach.** Tell these apart properly:
  - A go-around descends low, lined up with the runway, then climbs away
    without touching down.
  - A touch-and-go has ground contact before the climb. The C172 mock
    (`mock/trace.mock.json`) is pattern work, so it's probably full of these.
- **Hard landing proxy.** Use the vertical rate on the last airborne sample:
  noisy, but useful across many flights.
- **Wind estimation.** Index 12 is indicated airspeed. The aircraft data object
  in index 8 can also carry heading, true airspeed and sometimes wind and
  temperature. Ground vector minus air vector gives the wind at that altitude,
  and many flights together build wind profiles.

## 4. Efficiency metrics (standard in air traffic management)

These are published Eurocontrol performance indicators:

- **Holding patterns.** Sum the unwrapped track change over a sliding window;
  ≥ 330° within a few minutes and a small radius is a hold. Report time in
  holds per airport and hour, which is effectively a delay measure. The same
  algorithm finds orbits (police, survey, helicopters).
- **Extra distance near the destination (ASMA).** Compare distance flown in the
  last 40 or 100 nm with the direct distance, which catches radar vectoring.
- **Horizontal flight efficiency.** Actual distance flown vs the great-circle
  distance between origin and destination.
- **Continuous descent operations.** Time spent in level flight during descent;
  a good airport keeps this low.
- **Climb and descent profiles by aircraft type.** Group by type (`A320`,
  `B772`, …) for time to 10,000 ft, climb gradient and top-of-descent distance.

## 5. Anomalies and data quality (what `is_anomaly` is for)

- **GNSS interference map.** Each point's metadata includes `nic` and `nac_p`,
  which report how precise the aircraft's GPS position is. Plot where many
  aircraft lose accuracy at once and you get a GPS jamming map, a current
  hot topic around the Baltic and eastern Mediterranean.
- **Impossible jumps.** The speed implied by two consecutive positions exceeds
  about 1000 kt, pointing to spoofing or bad data. Also flag altitude spikes
  and duplicate timestamps.
- **Emergency squawks** 7500, 7600 and 7700, plus the `emergency` field.
- **Filters by data source and category.** The `type` field distinguishes
  ADS-B from MLAT and TIS-B. `dbFlags` marks military aircraft (bit 1) and
  privacy-program aircraft (PIA, LADD).

## 6. Aggregate views (once the database fills up)

- Movements per airport per hour and busiest times of day.
- Go-around rate per airport or runway.
- Overflights of a chosen point with altitude, as a noise-exposure view.

adsb.lol publishes daily global trace archives in the same readsb format as the
mocks, an easy way to get volume once the pipeline is solid.

## Suggested order

1. Leg splitting and phase labelling, with tests against the three mocks.
2. Origin, destination and runway matching; it pays off immediately in the
   viewer.
3. Approach analysis, including go-around vs touch-and-go.
4. The GNSS interference map, which has the best effort-to-wow ratio.
5. Holds and the efficiency metrics.

For performance on large archives:

- Process files in parallel with an `errgroup` and a concurrency limit.
- Do per-point geometry in Go (haversine and distance to a line are enough).
- Use PostGIS only for the once-per-flight airport and runway lookups.

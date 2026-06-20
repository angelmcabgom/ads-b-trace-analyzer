package analysis

import (
	"ads-b-geom-parser/trace"
)

func DetectTakeOff(points []trace.TracePoint) bool {
    if len(points) < 5 {
        return false
    }

    // average first 5 points to smooth out baro noise
	// things do get pretty rocky out there
    var altSum int
    for _, p := range points[:5] {
        altSum += p.Altitude
    }
    groundAlt := altSum / 5

    consecutiveClimbs := 0
    for _, point := range points {
        climbing := point.Altitude > groundAlt+50 &&
            point.VerticalRate > 800 &&
            point.GroundSpeed > 50

        if climbing {
            consecutiveClimbs++
            if consecutiveClimbs >= 3 {
                return true
            }
        } else {
            consecutiveClimbs = 0
        }
    }

    return false
}
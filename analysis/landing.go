package analysis

import (
	"ads-b-geom-parser/trace"
)

func DetectLanding(points []trace.TracePoint) bool {

	// return false since i cant determine if theres a landing with so few points
	if len(points) < 10 {
		return false
	}

	// this approach assumes theres a landing present in the trace
	// aka its a full trace, some traces might not have this tho
	// get airport altitute
	var altSum int
	for _, p := range points[len(points)-10:] {
		altSum += p.Altitude
	}
	groundAlt := altSum / 10

	const groundProximityMeters = 500

	consecutiveDescents := 0
	for _, point := range points[len(points)-10:] {
		nearGround := point.Altitude-groundAlt < groundProximityMeters
		descending := point.VerticalRate < -100 && point.GroundSpeed < 100

		if nearGround && descending {
			consecutiveDescents++
			if consecutiveDescents >= 3 {
				return true
			}
		} else {
			consecutiveDescents = 0
		}
	}

	return false
}

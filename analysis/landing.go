package analysis

import (
	"ads-b-geom-parser/trace"
)

func DetectLanding(points []trace.TracePoint) bool {
	if len(points) < 30 {
		return false
	}
	tail := points[len(points)-30:]

	for i, p := range tail {
		if p.OnGround {
			// confirm it's a real touchdown, not a single noisy sample:
			// check the next few points are also OnGround
			confirmCount := 0
			for j := i; j < len(tail) && j < i+3; j++ {
				if tail[j].OnGround {
					confirmCount++
				}
			}
			if confirmCount >= 3 {
				return true
			}
		}
	}
	return false
}
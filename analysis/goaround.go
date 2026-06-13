package analysis

import "ads-b-geom-parser/trace"

func DetectGoAround(points []trace.TracePoint) bool {
	wasLow := false
	wasDescending := false

	for _, point := range points {
		if point.Altitude > 0 && point.Altitude < 2500 {
			wasLow = true
		}
		if point.VerticalRate < -300 {
			wasDescending = true
		}
		if wasLow && wasDescending && point.VerticalRate > 1000 && point.Altitude < 4000 {
			return true
		}
	}
	return false
}
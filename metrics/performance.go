package metrics

import (
	"runtime"
	"time"
	"fmt"
)


func TrackPerformance() func() {
	var memBefore runtime.MemStats
	runtime.ReadMemStats(&memBefore)
	startTime := time.Now()

	return func() {
		duration := time.Since(startTime)
		var memAfter runtime.MemStats
		runtime.ReadMemStats(&memAfter)
		totalMemoryAllocated := memAfter.TotalAlloc - memBefore.TotalAlloc

		fmt.Println("\n================ PERFORMANCE METRICS ================")
		fmt.Printf("Execution Time : %v\n", duration)
		fmt.Printf("Total Heap Churn: %.2f KB (%d bytes)\n", float64(totalMemoryAllocated)/1024, totalMemoryAllocated)
		fmt.Printf("Peak Virtual Mem: %.2f MB\n", float64(memAfter.Sys)/1024/1024)
		fmt.Println("=====================================================")
	}
}

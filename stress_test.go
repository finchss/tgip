package tgip

import (
	"runtime"
	"sync"
	"testing"
	"time"
)

func TestHighLoadNoLeaks(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping stress test in short mode")
	}

	runtime.GC()
	time.Sleep(100 * time.Millisecond)

	var m runtime.MemStats
	runtime.ReadMemStats(&m)
	baselineAlloc := m.Alloc
	baselineGoroutines := runtime.NumGoroutine()

	t.Logf("Baseline - Goroutines: %d, Memory: %d KB", baselineGoroutines, baselineAlloc/1024)

	// Run many concurrent calls
	const numCalls = 50
	var wg sync.WaitGroup

	for i := 0; i < numCalls; i++ {
		wg.Add(1)
		go func(n int) {
			defer wg.Done()
			_, err := GetMyIp()
			if err != nil {
				t.Logf("Call %d failed: %v", n, err)
			}
		}(i)
	}

	wg.Wait()

	// Give time for cleanup
	time.Sleep(1 * time.Second)
	runtime.GC()
	time.Sleep(100 * time.Millisecond)

	runtime.ReadMemStats(&m)
	finalAlloc := m.Alloc
	finalGoroutines := runtime.NumGoroutine()

	t.Logf("Final - Goroutines: %d, Memory: %d KB", finalGoroutines, finalAlloc/1024)
	t.Logf("Difference - Goroutines: %d, Memory: %d KB",
		finalGoroutines-baselineGoroutines,
		int(finalAlloc-baselineAlloc)/1024)

	if finalGoroutines-baselineGoroutines > 10 {
		t.Errorf("Goroutine leak: %d extra goroutines", finalGoroutines-baselineGoroutines)
	}

	// Check for excessive memory growth (allow some growth for caching)
	memGrowth := int(finalAlloc-baselineAlloc) / 1024
	if memGrowth > 5000 { // 5MB threshold
		t.Errorf("Potential memory leak: %d KB growth", memGrowth)
	}
}

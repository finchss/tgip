package tgip

import (
	"runtime"
	"testing"
	"time"
)

func TestNoGoroutineLeaks(t *testing.T) {
	// Get baseline goroutine count
	runtime.GC()
	time.Sleep(100 * time.Millisecond)
	baseline := runtime.NumGoroutine()

	// Run GetMyIp multiple times
	for i := 0; i < 10; i++ {
		_, err := GetMyIp()
		if err != nil {
			t.Logf("GetMyIp call %d failed: %v", i, err)
		}
	}

	// Give goroutines time to clean up
	time.Sleep(500 * time.Millisecond)
	runtime.GC()
	time.Sleep(100 * time.Millisecond)

	// Check goroutine count
	final := runtime.NumGoroutine()
	leaked := final - baseline

	t.Logf("Baseline goroutines: %d", baseline)
	t.Logf("Final goroutines: %d", final)
	t.Logf("Potential leaked goroutines: %d", leaked)

	// Allow some margin for system goroutines, but not too many
	if leaked > 5 {
		t.Errorf("Potential goroutine leak detected: %d extra goroutines", leaked)
	}
}

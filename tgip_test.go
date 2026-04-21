package tgip

import (
	"encoding/json"
	"sync"
	"testing"
	"time"
)

func TestGetMyIp(t *testing.T) {
	m, err := GetMyIp()
	if err != nil {
		t.Fatalf("First GetMyIp() failed: %v", err)
	}
	if m == "" {
		t.Error("First GetMyIp() returned empty result")
	}

	m, err = GetMyIp()
	if err != nil {
		t.Fatalf("Second GetMyIp() failed: %v", err)
	}
	if m == "" {
		t.Error("Second GetMyIp() returned empty result")
	}
}
func TestGetMyIpWithApiIpifyOrg(t *testing.T) {
	remoteIpServiceMutex.RLock()
	originalService := RemoteIpService
	remoteIpServiceMutex.RUnlock()

	defer func() {
		remoteIpServiceMutex.Lock()
		RemoteIpService = originalService
		remoteIpServiceMutex.Unlock()
	}()

	remoteIpServiceMutex.Lock()
	RemoteIpService = "api.ipify.org"
	remoteIpServiceMutex.Unlock()
	m, err := GetMyIp()
	if err != nil {
		t.Fatalf("First GetMyIp() failed: %v", err)
	}
	if m == "" {
		t.Error("First GetMyIp() returned empty result")
	}

	m, err = GetMyIp()
	if err != nil {
		t.Fatalf("Second GetMyIp() failed: %v", err)
	}
	if m == "" {
		t.Error("Second GetMyIp() returned empty result")
	}
}

func TestIfReplyFromApiTgipEuIsJson(t *testing.T) {

	response, err := GetMyIp()
	if err != nil {
		t.Errorf("Failed to get IP: %v", err)
		return
	}

	var result map[string]interface{}
	if err := json.Unmarshal([]byte(response), &result); err != nil {
		t.Errorf("Response is not valid JSON: %v", err)
		return
	}

	if _, ok := result["ip"]; !ok {
		t.Error("JSON response does not contain 'ip' field")
	}
}
func TestIfReplyFromApiIpifyOrg(t *testing.T) {
	remoteIpServiceMutex.RLock()
	originalService := RemoteIpService
	remoteIpServiceMutex.RUnlock()

	defer func() {
		remoteIpServiceMutex.Lock()
		RemoteIpService = originalService
		remoteIpServiceMutex.Unlock()
	}()

	remoteIpServiceMutex.Lock()
	RemoteIpService = "api.ipify.org"
	remoteIpServiceMutex.Unlock()
	response, err := GetMyIp()
	if err != nil {
		t.Errorf("Failed to get IP: %v", err)
		return
	}

	var result map[string]interface{}
	if err := json.Unmarshal([]byte(response), &result); err != nil {
		t.Errorf("Response is not valid JSON: %v", err)
		return
	}

	if _, ok := result["ip"]; !ok {
		t.Error("JSON response does not contain 'ip' field")
	}
}
func TestConcurrentGetMyIp(t *testing.T) {
	const goroutineCount = 100

	// Optional: Set a shorter timeout for test speed
	SetTimeOut(5 * time.Second)

	var wg sync.WaitGroup
	results := make([]string, goroutineCount)
	errors := make([]error, goroutineCount)

	for i := range goroutineCount {
		wg.Add(1)
		go func(index int) {
			defer wg.Done()
			ip, err := GetMyIp()
			results[index] = ip
			errors[index] = err
		}(i)
	}

	wg.Wait()

	for i := range goroutineCount {
		if errors[i] != nil {
			t.Errorf("goroutine %d failed: %v", i, errors[i])
		} else {
			t.Logf("goroutine %d result: %s", i, results[i])
		}
	}
}

func TestDebug(t *testing.T) {
	originalDebug := Debug
	defer SetDebug(originalDebug)

	SetDebug(true)
	if !Debug {
		t.Error("Debug should be true after SetDebug(true)")
	}

	SetDebug(false)
	if Debug {
		t.Error("Debug should be false after SetDebug(false)")
	}
}

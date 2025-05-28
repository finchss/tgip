package tgip

import (
	"encoding/json"
	"log"
	"sync"
	"testing"
	"time"
)

func TestGetMyIp(t *testing.T) {
	m, err := GetMyIp()
	log.Println(m, err)
	m, err = GetMyIp()
	log.Println(m, err)
}
func TestGetMyIpWithApiIpifyOrg(t *testing.T) {
	RemoteIpService = "api.ipify.org"
	m, err := GetMyIp()
	log.Println(m, err)
	m, err = GetMyIp()
	log.Println(m, err)
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

	RemoteIpService = "api.ipify.org"
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
	const goroutineCount = 50

	// Optional: Set a shorter timeout for test speed
	SetTimeOut(5 * time.Second)

	var wg sync.WaitGroup
	results := make([]string, goroutineCount)
	errors := make([]error, goroutineCount)

	for i := 0; i < goroutineCount; i++ {
		wg.Add(1)
		go func(index int) {
			defer wg.Done()
			ip, err := GetMyIp()
			results[index] = ip
			errors[index] = err
		}(i)
	}

	wg.Wait()

	for i := 0; i < goroutineCount; i++ {
		if errors[i] != nil {
			t.Errorf("goroutine %d failed: %v", i, errors[i])
		} else {
			t.Logf("goroutine %d result: %s", i, results[i])
		}
	}
}

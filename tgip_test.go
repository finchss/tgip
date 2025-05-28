package tgip

import (
	"encoding/json"
	"log"
	"testing"
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

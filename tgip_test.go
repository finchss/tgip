package tgip

import (
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

package tgip

import (
	"context"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"io"
	"math/rand"
	"net"
	"net/http"
	"sync"
	"time"
)

var RemoteIpService = "api.tgip.eu"

type Tgip struct {
	addrs   []string
	useHttp bool
	timeout time.Duration
	host    string
}

var (
	myip      *Tgip
	rng       = rand.New(rand.NewSource(time.Now().UnixNano()))
	initMutex sync.Mutex
)

type ipResponse struct {
	IP string `json:"ip"`
}

func initMyIp(tg **Tgip) {
	initMutex.Lock()
	defer initMutex.Unlock()

	if *tg == nil {
		*tg = &Tgip{
			useHttp: true,
			host:    RemoteIpService,
			timeout: 10 * time.Second,
		}
	}
}

func SetTimeOut(timeout time.Duration) {
	initMyIp(&myip)
	initMutex.Lock()
	defer initMutex.Unlock()
	myip.timeout = timeout
}

func GetMyIp() (string, error) {
	initMyIp(&myip)

	var timeout time.Duration
	initMutex.Lock()
	if len(myip.addrs) == 0 {
		addrs, lookupErr := net.LookupHost(myip.host)
		if lookupErr != nil {
			initMutex.Unlock()
			return "", lookupErr
		}
		myip.addrs = addrs
	}
	timeout = myip.timeout
	initMutex.Unlock()

	ips := GetRandomIps()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	resultChan := make(chan string, len(ips))
	var wg sync.WaitGroup

	client := &http.Client{
		Transport: &http.Transport{
			TLSHandshakeTimeout: timeout,
			DisableKeepAlives:   true,
			TLSClientConfig: &tls.Config{
				ServerName: myip.host,
			},
		},
		Timeout: timeout,
	}

	for _, ip := range ips {
		wg.Add(1)
		go func(ipAddr string) {
			defer wg.Done()

			req, err := http.NewRequestWithContext(ctx, "GET",
				fmt.Sprintf("https://%s/?format=json", ipAddr), nil)
			if err != nil {
				return
			}
			req.Host = myip.host

			resp, err := client.Do(req)
			if err != nil {
				return
			}
			defer resp.Body.Close()

			if resp.StatusCode == http.StatusOK {
				body, err := io.ReadAll(resp.Body)
				if err != nil {
					return
				}

				var ipResp ipResponse
				if err := json.Unmarshal(body, &ipResp); err != nil || ipResp.IP == "" {
					return
				}

				select {
				case resultChan <- ipResp.IP:
					cancel()
				case <-ctx.Done():
				}
			}
		}(ip)
	}

	go func() {
		wg.Wait()
		close(resultChan)
	}()

	result, ok := <-resultChan
	if !ok {
		return "", fmt.Errorf("no successful response from any IP")
	}

	return result, nil
}

// GetRandomIps returns up to 3 randomly selected IP addresses
func GetRandomIps() []string {
	initMyIp(&myip)
	initMutex.Lock()
	defer initMutex.Unlock()

	addrsCopy := append([]string(nil), myip.addrs...)
	if len(addrsCopy) > 3 {
		rng.Shuffle(len(addrsCopy), func(i, j int) {
			addrsCopy[i], addrsCopy[j] = addrsCopy[j], addrsCopy[i]
		})
		addrsCopy = addrsCopy[:3]
	}
	return addrsCopy
}

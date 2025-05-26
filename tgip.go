package tgip

import (
	"context"
	"crypto/tls"
	"fmt"
	"io"
	"math/rand"
	"net"
	"net/http"
	"sync"
	"time"
)

type Tgip struct {
	addrs   []string
	useHttp bool
	timeout time.Duration
}

var (
	myip      *Tgip
	rng       = rand.New(rand.NewSource(time.Now().UnixNano()))
	initMutex sync.Mutex
)

func initMyIp(tg **Tgip) {
	initMutex.Lock()
	defer initMutex.Unlock()

	if *tg == nil {
		*tg = &Tgip{
			useHttp: true,
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

	initMutex.Lock()
	hasAddrs := len(myip.addrs) > 0
	timeout := myip.timeout
	initMutex.Unlock()

	if !hasAddrs {
		addrs, lookupErr := net.LookupHost("api.tgip.eu")
		if lookupErr != nil {
			return "", lookupErr
		}

		initMutex.Lock()
		myip.addrs = addrs
		initMutex.Unlock()
	}

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
				ServerName: "api.tgip.eu",
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

			req.Host = "api.tgip.eu"

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

				select {
				case resultChan <- string(body):
					cancel()
				case <-ctx.Done():
					return
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

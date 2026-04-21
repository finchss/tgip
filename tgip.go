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

var (
	RemoteIpService      = "api.tgip.eu"
	remoteIpServiceMutex sync.RWMutex
)

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
	rngMutex  sync.Mutex
)

func initMyIp(tg **Tgip) {
	initMutex.Lock()
	defer initMutex.Unlock()

	if *tg == nil {
		remoteIpServiceMutex.RLock()
		service := RemoteIpService
		remoteIpServiceMutex.RUnlock()

		*tg = &Tgip{
			useHttp: true,
			host:    service,
			timeout: 10 * time.Second,
		}
	}
}

func SetTimeOut(timeout time.Duration) {
	initMutex.Lock()
	defer initMutex.Unlock()

	if myip == nil {
		remoteIpServiceMutex.RLock()
		service := RemoteIpService
		remoteIpServiceMutex.RUnlock()

		myip = &Tgip{
			useHttp: true,
			host:    service,
			timeout: timeout,
		}
	} else {
		myip.timeout = timeout
	}
}

func SetUseHttp(useHttp bool) {
	initMutex.Lock()
	defer initMutex.Unlock()

	if myip == nil {
		remoteIpServiceMutex.RLock()
		service := RemoteIpService
		remoteIpServiceMutex.RUnlock()

		myip = &Tgip{
			useHttp: useHttp,
			host:    service,
			timeout: 10 * time.Second,
		}
	} else {
		myip.useHttp = useHttp
	}
}

func GetMyIp() (string, error) {
	initMyIp(&myip)

	var timeout time.Duration
	var useHttp bool
	var host string
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
	useHttp = myip.useHttp
	host = myip.host
	initMutex.Unlock()

	ips := GetRandomIps()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	resultChan := make(chan string, len(ips))
	var wg sync.WaitGroup

	transport := &http.Transport{
		DisableKeepAlives: true,
	}

	if !useHttp {
		transport.TLSHandshakeTimeout = timeout
		transport.TLSClientConfig = &tls.Config{
			ServerName: host,
		}
	}

	client := &http.Client{
		Transport: transport,
		Timeout:   timeout,
	}

	for _, ip := range ips {
		wg.Add(1)
		go func(ipAddr string) {
			defer wg.Done()

			scheme := "https"
			if useHttp {
				scheme = "http"
			}

			req, err := http.NewRequestWithContext(ctx, "GET",
				fmt.Sprintf("%s://%s/?format=json", scheme, ipAddr), nil)
			if err != nil {
				return
			}
			req.Host = host

			resp, err := client.Do(req)
			if err != nil {
				return
			}
			defer func(Body io.ReadCloser) {
				err := Body.Close()
				if err != nil {
					return
				}
			}(resp.Body)

			if resp.StatusCode == http.StatusOK {
				body, err := io.ReadAll(resp.Body)
				if err != nil {
					return
				}

				select {
				case resultChan <- string(body):
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
	addrsCopy := append([]string(nil), myip.addrs...)
	initMutex.Unlock()

	if len(addrsCopy) > 3 {
		rngMutex.Lock()
		rng.Shuffle(len(addrsCopy), func(i, j int) {
			addrsCopy[i], addrsCopy[j] = addrsCopy[j], addrsCopy[i]
		})
		rngMutex.Unlock()
		addrsCopy = addrsCopy[:3]
	}
	return addrsCopy
}

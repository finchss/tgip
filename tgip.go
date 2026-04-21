package tgip

import (
	"context"
	"crypto/tls"
	"fmt"
	"io"
	"log"
	"math/rand"
	"net"
	"net/http"
	"os"
	"slices"
	"sync"
	"time"
)

var (
	RemoteIpService      = "api.tgip.eu"
	remoteIpServiceMutex sync.RWMutex
	Debug                bool
)

func init() {
	if os.Getenv("TGIP_DEBUG") != "" {
		Debug = true
	}
}

func SetDebug(enabled bool) {
	Debug = enabled
}

func debug(format string, v ...any) {
	if Debug {
		log.Printf("[TGIP] "+format, v...)
	}
}

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

		debug("Initializing Tgip with host: %s", service)
		*tg = &Tgip{
			useHttp: true,
			host:    service,
			timeout: 10 * time.Second,
		}
	}
}

func SetTimeOut(timeout time.Duration) {
	debug("Setting timeout to %v", timeout)
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
	debug("Setting useHttp to %v", useHttp)
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
		debug("Looking up host: %s", myip.host)
		addrs, lookupErr := net.LookupHost(myip.host)
		if lookupErr != nil {
			initMutex.Unlock()
			debug("Host lookup failed: %v", lookupErr)
			return "", lookupErr
		}
		myip.addrs = addrs
		debug("Found IPs for %s: %v", myip.host, addrs)
	}
	timeout = myip.timeout
	useHttp = myip.useHttp
	host = myip.host
	initMutex.Unlock()

	ips := GetRandomIps()
	debug("Trying IPs: %v", ips)

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

			url := fmt.Sprintf("%s://%s/?format=json", scheme, ipAddr)
			debug("Requesting: %s (Host: %s)", url, host)
			req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
			if err != nil {
				debug("Failed to create request for %s: %v", ipAddr, err)
				return
			}
			req.Host = host

			resp, err := client.Do(req)
			if err != nil {
				debug("Request to %s failed: %v", ipAddr, err)
				return
			}
			debug("Response from %s: %d", ipAddr, resp.StatusCode)
			defer func(Body io.ReadCloser) {
				err := Body.Close()
				if err != nil {
					return
				}
			}(resp.Body)

			if resp.StatusCode == http.StatusOK {
				body, err := io.ReadAll(resp.Body)
				if err != nil {
					debug("Failed to read body from %s: %v", ipAddr, err)
					return
				}

				select {
				case resultChan <- string(body):
					debug("Success from %s: %s", ipAddr, string(body))
					cancel()
				case <-ctx.Done():
					debug("Context done, ignoring result from %s", ipAddr)
				}
			} else {
				debug("Non-OK response from %s: %d", ipAddr, resp.StatusCode)
			}
		}(ip)
	}

	go func() {
		wg.Wait()
		debug("All requests finished")
		close(resultChan)
	}()

	result, ok := <-resultChan
	if !ok {
		debug("No successful response from any IP")
		return "", fmt.Errorf("no successful response from any IP")
	}

	return result, nil
}

// GetRandomIps returns up to 3 randomly selected IP addresses
func GetRandomIps() []string {
	initMyIp(&myip)
	initMutex.Lock()
	addrsCopy := slices.Clone(myip.addrs)
	initMutex.Unlock()

	debug("Available IPs: %v", addrsCopy)

	if len(addrsCopy) > 3 {
		rngMutex.Lock()
		rng.Shuffle(len(addrsCopy), func(i, j int) {
			addrsCopy[i], addrsCopy[j] = addrsCopy[j], addrsCopy[i]
		})
		rngMutex.Unlock()
		addrsCopy = addrsCopy[:3]
		debug("Selected 3 random IPs: %v", addrsCopy)
	}
	return addrsCopy
}

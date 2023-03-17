package adapter

import (
	"math/rand"
	"net/http"
	"sync"
	"time"

	"github.com/kislerdm/diagramastext/server/core/port"
)

type httpClient interface {
	Do(req *http.Request) (*http.Response, error)
}

// NewHTTPClient initialises the HTTP Client.
func NewHTTPClient(cfg HTTPClientConfig) port.HTTPClient {
	resolveConfig(&cfg)
	return &httpclient{
		httpClient: &http.Client{Timeout: cfg.Timeout},
		backoff: Backoff{
			MaxIterations:             cfg.MaxIterations,
			BackoffTimeMinMillisecond: cfg.BackoffTimeMinMillisecond,
			BackoffTimeMaxMillisecond: cfg.BackoffTimeMaxMillisecond,
		},
		backoffCounter: map[*http.Request]uint8{},
		mu:             &sync.RWMutex{},
	}
}

const (
	defaultTimeout                   = 2 * time.Minute
	defaultBackoffTimeMinMillisecond = 100
	defaultBackoffTimeMaxMillisecond = 500
)

func resolveConfig(cfg *HTTPClientConfig) {
	if cfg.Timeout < 0 {
		cfg.Timeout = defaultTimeout
	}
	if cfg.BackoffTimeMinMillisecond <= 0 {
		cfg.BackoffTimeMinMillisecond = defaultBackoffTimeMinMillisecond
	}
	if cfg.BackoffTimeMaxMillisecond <= 0 {
		cfg.BackoffTimeMaxMillisecond = defaultBackoffTimeMaxMillisecond
	}
	if cfg.BackoffTimeMaxMillisecond < cfg.BackoffTimeMinMillisecond {
		tmp := cfg.BackoffTimeMaxMillisecond
		cfg.BackoffTimeMaxMillisecond = cfg.BackoffTimeMinMillisecond
		cfg.BackoffTimeMinMillisecond = tmp
	}
}

type httpclient struct {
	httpClient     httpClient
	backoff        Backoff
	backoffCounter map[*http.Request]uint8
	mu             *sync.RWMutex
}

func (c *httpclient) Do(req *http.Request) (*port.HTTPResponse, error) {
	var (
		resp *http.Response
		err  error
	)

	for !c.maxIterations(req) {
		resp, err = c.httpClient.Do(req)
		c.requestCounterUp(req)
		if err != nil || resp.StatusCode > 209 {
			c.backoffDelay(req)
		} else {
			break
		}
	}

	c.requestCounterReset(req)

	return &port.HTTPResponse{
		Body:       resp.Body,
		StatusCode: resp.StatusCode,
	}, err
}

func (c *httpclient) generateRandomDelay() time.Duration {
	rand.Seed(time.Now().UnixNano())
	cnt := rand.Intn(c.backoff.BackoffTimeMaxMillisecond-c.backoff.BackoffTimeMinMillisecond+1) + c.backoff.BackoffTimeMaxMillisecond
	return time.Duration(cnt) * time.Millisecond
}

func (c *httpclient) requestCounterUp(req *http.Request) {
	c.mu.Lock()
	_, ok := c.backoffCounter[req]
	if !ok {
		c.backoffCounter[req] = 0
	}
	c.backoffCounter[req]++
	c.mu.Unlock()

}

func (c *httpclient) backoffDelay(req *http.Request) {
	time.Sleep(c.generateRandomDelay())
}

func (c *httpclient) requestCounterReset(req *http.Request) {
	c.mu.Lock()
	delete(c.backoffCounter, req)
	c.mu.Unlock()
}

func (c *httpclient) maxIterations(req *http.Request) bool {
	c.mu.RLock()
	cnt := c.backoffCounter[req]
	c.mu.RUnlock()
	return cnt >= c.backoff.MaxIterations
}

type Backoff struct {
	MaxIterations             uint8
	BackoffTimeMinMillisecond int
	BackoffTimeMaxMillisecond int
}

type HTTPClientConfig struct {
	Timeout time.Duration
	Backoff
}

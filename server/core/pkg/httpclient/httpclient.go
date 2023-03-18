package httpclient

import (
	"math/rand"
	"net/http"
	"sync"
	"time"
)

// NewHTTPClient initialises the HTTP Client.
func NewHTTPClient(cfg Config) *HTTPClient {
	resolveConfig(&cfg)
	return &HTTPClient{
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

// Config HTTP client configuration.
type Config struct {
	Timeout time.Duration
	Backoff
}

// Backoff retry configuration.
type Backoff struct {
	MaxIterations             uint8
	BackoffTimeMinMillisecond int
	BackoffTimeMaxMillisecond int
}

// HTTPClient defines the client object.
type HTTPClient struct {
	httpClient     httpClient
	backoff        Backoff
	backoffCounter map[*http.Request]uint8
	mu             *sync.RWMutex
}

func (c *HTTPClient) Do(req *http.Request) (*http.Response, error) {
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

	return resp, err
}

func (c *HTTPClient) generateRandomDelay() time.Duration {
	rand.Seed(time.Now().UnixNano())
	cnt := rand.Intn(c.backoff.BackoffTimeMaxMillisecond-c.backoff.BackoffTimeMinMillisecond+1) + c.backoff.BackoffTimeMaxMillisecond
	return time.Duration(cnt) * time.Millisecond
}

func (c *HTTPClient) requestCounterUp(req *http.Request) {
	c.mu.Lock()
	_, ok := c.backoffCounter[req]
	if !ok {
		c.backoffCounter[req] = 0
	}
	c.backoffCounter[req]++
	c.mu.Unlock()

}

func (c *HTTPClient) backoffDelay(req *http.Request) {
	time.Sleep(c.generateRandomDelay())
}

func (c *HTTPClient) requestCounterReset(req *http.Request) {
	c.mu.Lock()
	delete(c.backoffCounter, req)
	c.mu.Unlock()
}

func (c *HTTPClient) maxIterations(req *http.Request) bool {
	c.mu.RLock()
	cnt := c.backoffCounter[req]
	c.mu.RUnlock()
	return cnt >= c.backoff.MaxIterations
}

const (
	defaultTimeout                   = 2 * time.Minute
	defaultBackoffTimeMinMillisecond = 100
	defaultBackoffTimeMaxMillisecond = 500
)

func resolveConfig(cfg *Config) {
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

type httpClient interface {
	Do(req *http.Request) (*http.Response, error)
}

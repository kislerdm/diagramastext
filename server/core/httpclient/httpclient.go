package httpclient

import (
	"io"
	"math/rand"
	"net/http"
	"sync"
	"time"

	"github.com/kislerdm/diagramastext/server/core"
)

type client struct {
	httpClient     *http.Client
	backoff        Backoff
	backoffCounter map[*http.Request]uint8
	mu             *sync.RWMutex
}

func mustReadBody(v io.ReadCloser) []byte {
	buf, err := io.ReadAll(v)
	defer func() { _ = v.Close() }()
	if err != nil {
		panic(err)
	}
	return buf
}

func (c *client) Do(req *http.Request) (core.ClientHTTPResp, error) {
	c.backoffInit(req)

	resp, err := c.httpClient.Do(req)

	if err != nil || resp.StatusCode > 209 {
		c.backoffDelay(req)
		return c.Do(req)
	}

	c.backoffReset(req)

	return core.ClientHTTPResp{
		Body:       mustReadBody(resp.Body),
		StatusCode: resp.StatusCode,
	}, err
}

func (c *client) generateRandomDelay() time.Duration {
	rand.Seed(time.Now().UnixNano())
	cnt := rand.Intn(c.backoff.BackoffTimeMaxMillisecond-c.backoff.BackoffTimeMinMillisecond+1) + c.backoff.BackoffTimeMaxMillisecond
	return time.Duration(cnt) * time.Millisecond
}

func (c *client) backoffInit(req *http.Request) {
	c.mu.Lock()
	_, ok := c.backoffCounter[req]
	if !ok {
		c.backoffCounter[req] = 0
	}
	c.mu.Unlock()

}

func (c *client) backoffDelay(req *http.Request) {
	time.Sleep(c.generateRandomDelay())
	c.mu.Lock()
	c.backoffCounter[req]++
	c.mu.Unlock()
}

func (c *client) backoffReset(req *http.Request) {
	c.mu.Lock()
	delete(c.backoffCounter, req)
	c.mu.Unlock()
}

type Backoff struct {
	MaxIterations             uint8
	BackoffTimeMinMillisecond int
	BackoffTimeMaxMillisecond int
}

type Config struct {
	Timeout time.Duration
	Backoff
}

func NewHTTPClient(cfg Config) core.ClientHTTP {
	resolveConfig(&cfg)
	return &client{
		httpClient: &http.Client{Timeout: cfg.Timeout},
		backoff: Backoff{
			MaxIterations:             0,
			BackoffTimeMinMillisecond: 0,
			BackoffTimeMaxMillisecond: 0,
		},
		backoffCounter: map[*http.Request]uint8{},
		mu:             &sync.RWMutex{},
	}
}

const (
	defaultMaxIterations             = 2
	defaultTimeout                   = 2 * time.Minute
	defaultBackoffTimeMinMillisecond = 100
	defaultBackoffTimeMaxMillisecond = 500
)

func resolveConfig(cfg *Config) {
	if cfg.MaxIterations < 0 {
		cfg.MaxIterations = defaultMaxIterations
	}
	if cfg.Timeout < 0 {
		cfg.Timeout = defaultTimeout
	}
	if cfg.BackoffTimeMinMillisecond < 0 {
		cfg.BackoffTimeMinMillisecond = defaultBackoffTimeMinMillisecond
	}
	if cfg.BackoffTimeMaxMillisecond < 0 {
		cfg.BackoffTimeMaxMillisecond = defaultBackoffTimeMaxMillisecond
	}
}

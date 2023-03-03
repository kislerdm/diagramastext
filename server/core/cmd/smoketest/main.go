package main

import (
	"net/http"
	"time"
)

type httpClient struct {
	token      string
	maxRetries uint8
	backoff    time.Duration

	c http.Client
}

func (c httpClient) Post(url string, v []byte) ([]byte, error) {
	//resp, err := c.c.Post(url, "application/json", bytes.NewReader(v))
	//if err != nil {
	//	return nil, err
	//}
	panic("todo")
}

func coreC4Rendering() {
	panic("todo")
}

func main() {
	panic("todo")
}

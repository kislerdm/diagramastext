package core

import (
	"bytes"
	"errors"
	"io"
	"net/http"
	"strconv"
	"time"
	"unsafe"

	"github.com/kislerdm/diagramastext/core/go-zopfli/zopfli"
)

// HttpClient http base client.
type HttpClient interface {
	Get(url string) (resp *http.Response, err error)
}

type options struct {
	httpClient HttpClient
}

// PlantUMLClient client to communicate with the plantuml server.
type PlantUMLClient interface {
	// GenerateSVG generates the SVG diagram using the diagram as code input.
	GenerateSVG(code string) ([]byte, error)

	// GeneratePNG generates the PNG diagram using the diagram as code input.
	GeneratePNG(code string) ([]byte, error)
}

type client struct {
	options options

	baseURL string
}

func (c *client) GenerateSVG(code string) ([]byte, error) {
	p, err := code2Path(code)
	if err != nil {
		return nil, err
	}
	return c.requestHandler("svg/" + p)
}

func (c *client) GeneratePNG(code string) ([]byte, error) {
	p, err := code2Path(code)
	if err != nil {
		return nil, err
	}
	return c.requestHandler("png/" + p)
}

func (c *client) requestHandler(p string) ([]byte, error) {
	resp, err := c.options.httpClient.Get(c.baseURL + p)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode > 209 {
		return nil, errors.New("error status code: " + strconv.Itoa(resp.StatusCode))
	}

	buf, err := io.ReadAll(resp.Body)
	defer func() { _ = resp.Body.Close() }()
	if err != nil {
		return nil, err
	}

	return buf, nil
}

const (
	baseURL        = "https://www.plantuml.com/plantuml/"
	defaultTimeout = 1 * time.Minute
)

// NewPlantUMLClient initiates the client to communicate with the plantuml server.
func NewPlantUMLClient(optFns ...func(*options)) PlantUMLClient {
	o := options{
		httpClient: nil,
	}

	for _, fn := range optFns {
		fn(&o)
	}

	resolveHTTPClient(&o)

	return &client{
		options: o,
		baseURL: baseURL,
	}
}

func resolveHTTPClient(o *options) {
	if o.httpClient == nil {
		o.httpClient = &http.Client{Timeout: defaultTimeout}
	}
}

// code2Path converts the diagram as code to the 64Bytes encoded string to query plantuml
//
// Example: the diagram's code
// @startuml
//
//	a -> b
//
// @enduml
//
// will be converted to SoWkIImgAStDuL80WaG5NJk592w7rBmKe100
//
// The resulting string to be used to generate C4 diagram
// - as png: GET www.plantuml.com/plantuml/png/SoWkIImgAStDuL80WaG5NJk592w7rBmKe100
// - as svg: GET www.plantuml.com/plantuml/svg/SoWkIImgAStDuL80WaG5NJk592w7rBmKe100
func code2Path(s string) (string, error) {
	zb, err := compress(*(*[]byte)(unsafe.Pointer(&s)))
	if err != nil {
		return "", err
	}
	return encode64(zb), nil
}

func compress(v []byte) ([]byte, error) {
	var options = zopfli.DefaultOptions()
	var w bytes.Buffer
	if err := zopfli.Compress(&options, zopfli.FORMAT_DEFLATE, v, &w); err != nil {
		return nil, err
	}
	return w.Bytes(), nil
}

func encode64(e []byte) string {
	var r bytes.Buffer
	for i := 0; i < len(e); i += 3 {
		switch len(e) {
		case i + 2:
			r.Write(append3bytes(e[i], e[i+1], 0))
		case i + 1:
			r.Write(append3bytes(e[i], 0, 0))
		default:
			r.Write(append3bytes(e[i], e[i+1], e[i+2]))
		}
	}
	return r.String()
}

func append3bytes(e, n, t byte) []byte {
	c1 := e >> 2
	c2 := (3&e)<<4 | n>>4
	c3 := (15&n)<<2 | t>>6
	c4 := 63 & t

	var buf bytes.Buffer

	buf.WriteByte(encode6bit(c1 & 63))
	buf.WriteByte(encode6bit(c2 & 63))
	buf.WriteByte(encode6bit(c3 & 63))
	buf.WriteByte(encode6bit(c4 & 63))

	return buf.Bytes()
}

func encode6bit(e byte) byte {
	if e < 10 {
		return 48 + e
	}

	e -= 10
	if e < 26 {
		return 65 + e
	}

	e -= 26
	if e < 26 {
		return 97 + e
	}

	e -= 26
	switch e {
	case 0:
		return '-'
	case 1:
		return '_'
	default:
		return '?'
	}
}

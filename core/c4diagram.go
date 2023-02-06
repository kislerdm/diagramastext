// package defines diagram as code using the graph input
package core

import (
	"bytes"
	"errors"
	"io"
	"net/http"
	"strconv"
	"strings"
	"time"
	"unsafe"

	"github.com/kislerdm/diagramastext/core/compression"
)

type options struct {
	httpClient HttpClient
}

type clientPlantUML struct {
	options options

	baseURL string
}

func (c *clientPlantUML) Do(v DiagramGraph) ([]byte, error) {
	p, err := code2Path(diagramGraph2plantUMLCode(v))
	if err != nil {
		return nil, err
	}
	return c.requestHandler("svg/" + p)
}

func (c *clientPlantUML) requestHandler(p string) ([]byte, error) {
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

// NewPlantUMLClient initiates the clientPlantUML to communicate with the plantuml server.
func NewPlantUMLClient(optFns ...func(*options)) ClientGraphToDiagram {
	o := options{
		httpClient: nil,
	}

	for _, fn := range optFns {
		fn(&o)
	}

	resolveHTTPClient(&o)

	return &clientPlantUML{
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
	var options = compression.DefaultOptions()
	var w bytes.Buffer
	if err := compression.Compress(&options, compression.FORMAT_DEFLATE, v, &w); err != nil {
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

func trimQuotes(s string) string {
	if s[:1] == `"` {
		s = s[1:]
	}
	if s[len(s)-1:] == `"` {
		s = s[:len(s)-1]
	}
	return s
}

func diagramFooter2UML(graph DiagramGraph) string {
	if graph.Footer == "" {
		return `footer "generated by diagramastext.dev - %date('yyyy-MM-dd')"`
	}
	return `footer "` + stringCleaner(graph.Footer) + `"`
}

func diagramNode2UML(n Node) string {
	o := "Container"

	switch n.IsQueue && n.IsDatabase {
	case true:
	case false:
		if n.IsQueue {
			o += "Queue"
		}

		if n.IsDatabase {
			o += "Db"
		}
	}

	if n.External {
		o += "_Ext"
	}

	o += "(" + n.ID

	label := n.Label
	if label == "" {
		label = n.ID
	}
	o += `, "` + stringCleaner(label) + `"`

	if n.Technology != "" {
		o += `, "` + stringCleaner(n.Technology) + `"`
	}

	o += ")"

	return o
}

func diagramLink2UML(l Link) string {
	o := "Rel"

	if d := linkDirection(l.Direction); d != "" {
		o += "_" + d
	}

	o += "(" + l.From + ", " + l.To

	if l.Label != "" {
		o += ", " + l.Label
	}

	if l.Technology != "" {
		o += ", " + l.Technology
	}

	o += ")"

	return o
}

func linkDirection(s string) string {
	switch s := strings.ToUpper(s); s {
	case "LR":
		return "R"
	case "RL":
		return "L"
	case "TD":
		return "D"
	case "DT":
		return "U"
	default:
		return ""
	}
}

func stringCleaner(s string) string {
	s = strings.TrimSpace(s)
	s = trimQuotes(s)
	s = strings.ReplaceAll(s, "\n", "\\n")
	return s
}

// diagramGraph2plantUMLCode function to "transpile" the diagram definition graph to plantUML code as string.
func diagramGraph2plantUMLCode(graph DiagramGraph) string {
	o := `@startuml
!include https://raw.githubusercontent.com/plantuml-stdlib/C4-PlantUML/master/C4_Container.puml`

	o += "\n" + diagramFooter2UML(graph)

	if graph.Title != "" {
		o += "\n" + diagramTitle2UML(graph)
	}

	o += "\n@enduml"

	return o
}

func diagramTitle2UML(graph DiagramGraph) string {
	return `title "` + stringCleaner(graph.Title) + `"`
}

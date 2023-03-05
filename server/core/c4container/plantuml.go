// Package c4container defines client to convert the graph object
// to the PlantUMP DSL
// and to generate a C4 Container diagram artifact as SVG.
package c4container

import (
	"bytes"
	"context"
	"errors"
	"io"
	"net/http"
	"strconv"
	"strings"

	"github.com/kislerdm/diagramastext/server/core/c4container/compression"
	errs "github.com/kislerdm/diagramastext/server/core/errors"
)

func renderDiagram(ctx context.Context, httpClient HttpClient, v graph) ([]byte, error) {
	const baseURL = "https://www.plantuml.com/plantuml/"

	encodedDiagramString, err := convertToPlantUMLFormat(v)
	if err != nil {
		return nil, err
	}

	req, _ := http.NewRequestWithContext(ctx, http.MethodGet, baseURL+"svg/"+encodedDiagramString, nil)

	resp, err := httpClient.Do(req)
	if err != nil {
		return nil, errs.Error{
			Service: errs.ServiePlantUML,
			Message: err.Error(),
		}
	}

	if resp.StatusCode > 209 {
		return nil,
			errs.Error{
				Service:                   errs.ServiePlantUML,
				Message:                   "error status code: " + strconv.Itoa(resp.StatusCode),
				ServiceResponseStatusCode: resp.StatusCode,
			}
	}

	buf, err := io.ReadAll(resp.Body)
	defer func() { _ = resp.Body.Close() }()
	if err != nil {
		return nil, errs.Error{
			Service: errs.ServiePlantUML,
			Message: err.Error(),
		}
	}

	return buf, nil
}

func convertToPlantUMLFormat(v graph) (string, error) {
	code, err := defineDiagramPlantUMLDSL(v)
	if err != nil {
		return "", err
	}
	return plantUMLRequestPath(code)
}

// defineDiagramPlantUMLDSL function to "transpile" the diagram definition graph to plantUML code as string.
func defineDiagramPlantUMLDSL(graph graph) (string, error) {
	if len(graph.Nodes) == 0 {
		return "", errors.New("graph must contain at least one node")
	}

	// FIXME: user strings.Builder
	// see: https://github.com/kislerdm/diagramastext/pull/20/files/a8172589a31eda09b9a51c748d9b29b2fe985eb9#r1098011747
	o := `@startuml
!include https://raw.githubusercontent.com/plantuml-stdlib/C4-PlantUML/master/C4_Container.puml`

	o += "\n" + diagramFooter2UML(graph)

	if graph.Title != "" {
		o += "\n" + diagramTitle2UML(graph)
	}

	groups := map[string][]string{}
	for _, n := range graph.Nodes {
		containerStr, err := diagramNode2UML(n)
		if err != nil {
			return "", err
		}

		if _, ok := groups[n.Group]; !ok {
			groups[n.Group] = []string{}
		}
		groups[n.Group] = append(groups[n.Group], containerStr)
	}
	o += diagramUMLSystemBoundary(groups)

	for _, l := range graph.Links {
		linkStr, err := diagramLink2UML(l)
		if err != nil {
			return "", err
		}
		o += "\n" + linkStr
	}

	o += "\n@enduml"

	return o, nil
}

// FIXME: user strings.Builder
func diagramUMLSystemBoundary(v map[string][]string) string {
	o := ""
	if members, ok := v[""]; ok {
		o += "\n" + strings.Join(members, "\n")
		delete(v, "")
	}

	for groupName, members := range v {
		description := stringCleaner(groupName)
		id := strings.ReplaceAll(description, "\n", "_")
		o += "\nSystem_Boundary(" + id + `, "` + description + `") {
` + strings.Join(members, "\n") + "\n}"
	}
	return o
}

func diagramTitle2UML(graph graph) string {
	return `title "` + stringCleaner(graph.Title) + `"`
}

func diagramFooter2UML(graph graph) string {
	if graph.Footer == "" {
		return `footer "generated by diagramastext.dev - %date('yyyy-MM-dd')"`
	}
	return `footer "` + stringCleaner(graph.Footer) + `"`
}

func diagramNode2UML(n *node) (string, error) {
	if n.ID == "" {
		return "", errors.New("node must be identified: 'id' attribute")
	}

	// FIXME: user strings.Builder
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

	return o, nil
}

func diagramLink2UML(l *link) (string, error) {
	if l.From == "" || l.To == "" {
		return "", errors.New("link must specify the end nodes: 'from' and 'to' attributes")
	}

	// FIXME: user strings.Builder
	// see: https://github.com/kislerdm/diagramastext/pull/20/files/a8172589a31eda09b9a51c748d9b29b2fe985eb9#r1098011156
	o := "Rel"

	if d := linkDirection(l.Direction); d != "" {
		o += "_" + d
	}

	o += "(" + l.From + ", " + l.To

	if l.Label != "" {
		o += `, "` + stringCleaner(l.Label) + `"`
	}

	if l.Technology != "" {
		o += `, "` + stringCleaner(l.Technology) + `"`
	}

	o += ")"

	return o, nil
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
	s = strings.TrimLeft(s, `"`)
	s = strings.TrimRight(s, `"`)
	s = strings.ReplaceAll(s, "\n", "\\n")
	return s
}

// plantUMLRequestPath converts the diagram as code to the 64Bytes encoded string to query plantuml
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
func plantUMLRequestPath(s string) (string, error) {
	zb, err := compress([]byte(s))
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

// FIXME: replace with encode base64.Encoder (?)
// see: https://github.com/kislerdm/diagramastext/pull/20#discussion_r1098013688
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
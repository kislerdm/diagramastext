package utils

import (
	"bytes"
	"encoding/xml"
	"errors"
	"io"

	"golang.org/x/text/encoding/ianaindex"
)

type polygon struct {
	Fill   string `xml:"fill,attr"`
	Points string `xml:"points,attr"`
	Style  string `xml:"style,omitempty,attr"`
}

func (p polygon) IsValid() bool {
	return p.Points != ""
}

type line struct {
	X1 float64 `xml:"x1,attr"`
	Y1 float64 `xml:"y1,attr"`
	X2 float64 `xml:"x2,attr"`
	Y2 float64 `xml:"y2,attr"`
}

func (l line) IsValid() bool {
	return l.X1 != 0 && l.X2 != 0 && l.Y1 != 0 && l.Y2 != 0
}

type path struct {
	D    string `xml:"d,attr"`
	Fill string `xml:"fill,omitempty,attr"`
}

func (p path) IsValid() bool {
	return p.D != ""
}

type text struct {
	Fill     string  `xml:"fill,omitempty,attr"`
	FontSize string  `xml:"font-size,attr"`
	X        float64 `xml:"x,attr"`
	Y        float64 `xml:"y,attr"`
}

func (t text) IsValid() bool {
	return t.FontSize != ""
}

type rect struct {
	Fill  string  `xml:"fill,omitempty,attr"`
	Style string  `xml:"style,omitempty,attr"`
	X     float64 `xml:"x,attr"`
	Y     float64 `xml:"y,attr"`
	Rx    float64 `xml:"rx,attr"`
	Ry    float64 `xml:"ry,attr"`
	Width string  `xml:"width,attr"`
}

func (r rect) IsValid() bool {
	return r.Rx != 0 && r.Ry != 0 && r.Width != ""
}

type g struct {
	ID      string    `xml:"id,attr"`
	Rect    []rect    `xml:"rect,omitempty"`
	Text    []text    `xml:"text,omitempty"`
	Path    []path    `xml:"path,omitempty"`
	Line    []line    `xml:"line,omitempty"`
	Polygon []polygon `xml:"polygon,omitempty"`
}

type svg struct {
	SVG     xml.Name `xml:"svg"`
	ViewBox string   `xml:"viewBox,attr"`
	Height  string   `xml:"height,attr"`
	Width   string   `xml:"width,attr"`
	Defs    xml.Name `xml:"defs"`
	G       []g      `xml:"g>g"`
}

func (s svg) validSVGAttrs() error {
	if s.Width == "" {
		return errors.New("svg 'width' attr is missing")
	}

	if s.Height == "" {
		return errors.New("svg 'height' attr is missing")
	}

	if s.ViewBox == "" {
		return errors.New("svg 'viewBox' attr is missing")
	}

	return nil
}

func (s svg) validSVGGeometry() error {
	var cntGeometryElements int

	for _, geom := range s.G {
		cntGeometryElements += countValidSVGElements(geom.Rect)
		cntGeometryElements += countValidSVGElements(geom.Text)
		cntGeometryElements += countValidSVGElements(geom.Path)
		cntGeometryElements += countValidSVGElements(geom.Line)
		cntGeometryElements += countValidSVGElements(geom.Polygon)
	}

	if cntGeometryElements == 0 {
		return errors.New("svg does not contain any geometry")
	}

	return nil
}

type svgGeometries interface {
	rect | text | path | line | polygon
}

func countValidSVGElements[v svgGeometry](geom []v) int {
	var cnt int
	for _, g := range geom {
		if g.IsValid() {
			cnt++
		}
	}
	return cnt
}

type svgGeometry interface {
	IsValid() bool
}

// see: https://dzhg.dev/posts/2020/08/how-to-parse-xml-with-non-utf8-encoding-in-go/
func parseSVG(v []byte) (*svg, error) {
	decoder := xml.NewDecoder(bytes.NewBuffer(v))
	decoder.CharsetReader = func(charset string, reader io.Reader) (io.Reader, error) {
		enc, err := ianaindex.IANA.Encoding(charset)
		if err != nil {
			return nil, errors.New("charset " + charset + ":" + err.Error())
		}
		if enc == nil {
			// Assume it's compatible with (a subset of) UTF-8 encoding
			// Bug: https://github.com/golang/go/issues/19421
			return reader, nil
		}
		return enc.NewDecoder().Reader(reader), nil
	}

	var probe svg

	if err := decoder.Decode(&probe); err != nil {
		return nil, err
	}

	return &probe, nil
}

// ValidateSVG validates the SVG object content.
func ValidateSVG(v []byte) error {
	svg, err := parseSVG(v)
	if err != nil {
		return err
	}

	if err := svg.validSVGAttrs(); err != nil {
		return err
	}

	if err := svg.validSVGGeometry(); err != nil {
		return err
	}

	return nil
}

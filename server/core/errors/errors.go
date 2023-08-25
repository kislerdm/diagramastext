package errors

import (
	"bytes"
	"encoding/json"
	"net/http"
	"reflect"
	"runtime"
	"strconv"
	"strings"
	"sync"
)

type Error struct {
	mu  sync.Mutex
	buf []byte
}

func (e *Error) Error() string {
	return string(e.buf)
}

var std = &Error{}

// New defines the error pointing to the line where it occurred.
func New(s string) error {
	std.mu.Lock()
	defer std.mu.Unlock()

	std.mu.Unlock()
	_, file, line, ok := runtime.Caller(1)
	if !ok {
		file = "???"
		line = 0
	}
	file = cutPrefix(file)
	std.mu.Lock()

	std.buf = std.buf[:0]
	setBuffer(&std.buf, file, line)
	std.buf = append(std.buf, s...)

	return std
}

func cutPrefix(s string) string {
	_, prefix, _, ok := runtime.Caller(0)
	if !ok {
		return s
	}
	prefix = strings.TrimSuffix(prefix, "errors/errors.go")
	return strings.TrimPrefix(s, prefix)
}

func setBuffer(buf *[]byte, file string, line int) {
	*buf = append(*buf, file...)
	*buf = append(*buf, ':')
	itoa(buf, line)
	*buf = append(*buf, ": "...)
}

// Cheap integer to fixed-width decimal ASCII. Give a negative width to avoid zero-padding.
func itoa(buf *[]byte, i int) {
	// Assemble decimal in reverse order.
	var b [20]byte
	bp := len(b) - 1
	for i >= 10 {
		q := i / 10
		b[bp] = byte('0' + i - q*10)
		bp--
		i = q
	}
	// i < 10
	b[bp] = byte('0' + i)
	*buf = append(*buf, b[bp:]...)
}

func IsError(e error, text string) bool {
	return reflect.DeepEqual(&Error{buf: []byte(text)}, e)
}

// ModelPredictionError model prediction error.
type ModelPredictionError struct {
	RawJSON []byte
	msg     string
}

func (m ModelPredictionError) Error() string {
	return m.msg
}

// NewPredictionError create a response object with error response from the model.
func NewPredictionError(v []byte) error {
	if !bytes.Contains(v, []byte(`"error"`)) {
		return nil
	}
	var o struct {
		Error string `json:"error"`
	}
	if err := json.Unmarshal(v, &o); err != nil {
		return ModelPredictionError{RawJSON: v, msg: string(v)}
	}
	return ModelPredictionError{RawJSON: v, msg: o.Error}
}

type HTTPHandlerError struct {
	Msg      string
	Type     string
	HTTPCode int
}

func (e HTTPHandlerError) Error() string {
	var o strings.Builder
	writeStrings(&o, "[type:", e.Type, "][code:", strconv.Itoa(e.HTTPCode), "] ", e.Msg)
	return o.String()
}

func writeStrings(o *strings.Builder, text ...string) {
	for _, s := range text {
		_, _ = o.WriteString(s)
	}
}

func NewInputFormatValidationError(err error) error {
	msg := err.Error()

	switch err.(type) {
	case *json.SyntaxError:
		msg = "faulty JSON"
	}

	return HTTPHandlerError{
		Msg:      msg,
		Type:     ErrorInvalidRequest,
		HTTPCode: http.StatusBadRequest,
	}
}

func NewInputContentValidationError(err error) error {
	return HTTPHandlerError{
		Msg:      err.Error(),
		Type:     ErrorInvalidPrompt,
		HTTPCode: http.StatusUnprocessableEntity,
	}
}

const (
	ErrorInvalidPrompt  = "InputValidation:InvalidRequestContent"
	ErrorInvalidRequest = "InputValidation:InvalidRequestFormat"
	ErrorNotExists      = "Request:HandlerNotExists"
)

func NewHandlerNotExistsError(err error) error {
	return HTTPHandlerError{
		Msg:      err.Error(),
		Type:     ErrorNotExists,
		HTTPCode: http.StatusNotFound,
	}
}

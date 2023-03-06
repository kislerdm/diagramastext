package errors

import (
	"runtime"
	"sync"
)

type err struct {
	mu  sync.Mutex
	buf []byte
}

func (e *err) Error() string {
	return string(e.buf)
}

var std = &err{}

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
	std.mu.Lock()

	std.buf = std.buf[:0]
	setBuffer(&std.buf, file, line)
	std.buf = append(std.buf, s...)
	if len(s) == 0 || s[len(s)-1] != '\n' {
		std.buf = append(std.buf, '\n')
	}

	return std
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

package errors

import (
	"strings"
	"testing"
)

func TestNew(t *testing.T) {
	// GIVEN
	errorMessage := "foobar"

	// WHEN
	err := New(errorMessage)

	// THEN
	expectedLastElementsOfPath := "errors/errors_test.go"
	expectedLoC := "13"
	p := strings.Split(err.Error(), "/")
	if strings.Join(p[len(p)-2:], "/") != expectedLastElementsOfPath+":"+expectedLoC+": "+errorMessage+"\n" {
		t.Fatalf("wrong error message")
	}
}

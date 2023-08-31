package ciam

import (
	"net/http"
	"testing"
)

func Test_readAPIKey(t *testing.T) {
	type args struct {
		header http.Header
	}

	var newHeader = func(s string) http.Header {
		h := http.Header{}
		h.Add(s, "foo")
		return h
	}

	tests := []struct {
		name      string
		args      args
		wantKey   string
		wantFound bool
	}{
		{
			name: "found",
			args: args{
				header: newHeader("X-API-KEY"),
			},
			wantKey:   "foo",
			wantFound: true,
		},
		{
			name:      "not found: no header",
			args:      args{},
			wantKey:   "",
			wantFound: false,
		},
		{
			name: "wrong key",
			args: args{
				header: newHeader("X-API-KEY-wrong"),
			},
			wantKey:   "",
			wantFound: false,
		},
	}
	for _, tt := range tests {
		t.Run(
			tt.name, func(t *testing.T) {
				gotKey, gotFound := readAPIKey(tt.args.header)
				if gotKey != tt.wantKey {
					t.Errorf("readAPIKey() gotKey = %v, want %v", gotKey, tt.wantKey)
				}
				if gotFound != tt.wantFound {
					t.Errorf("readAPIKey() gotFound = %v, want %v", gotFound, tt.wantFound)
				}
			},
		)
	}
}

func Test_readAuthHeaderValue(t *testing.T) {
	var newHeader = func(s string, v string) http.Header {
		h := http.Header{}
		h.Add(s, v)
		return h
	}

	type args struct {
		header http.Header
	}
	tests := []struct {
		name      string
		args      args
		wantKey   string
		wantFound bool
	}{
		{
			name: "found",
			args: args{
				header: newHeader("Authorization", "Bearer foo"),
			},
			wantKey:   "foo",
			wantFound: true,
		},
		{
			name: "found small",
			args: args{
				header: newHeader("authorization", "Bearer foo"),
			},
			wantKey:   "foo",
			wantFound: true,
		},
		{
			name:      "not found: no header",
			args:      args{},
			wantKey:   "",
			wantFound: false,
		},
		{
			name: "wrong key",
			args: args{
				header: newHeader("Authorization-wrong", "Bearer foo"),
			},
			wantKey:   "",
			wantFound: false,
		},
		{
			name: "no Bearer separator",
			args: args{
				header: newHeader("Authorization", "foo"),
			},
			wantKey:   "",
			wantFound: false,
		},
	}
	for _, tt := range tests {
		t.Run(
			tt.name, func(t *testing.T) {
				gotKey, gotFound := readAuthHeaderValue(tt.args.header)
				if gotKey != tt.wantKey {
					t.Errorf("readAuthHeaderValue() gotKey = %v, want %v", gotKey, tt.wantKey)
				}
				if gotFound != tt.wantFound {
					t.Errorf("readAuthHeaderValue() gotFound = %v, want %v", gotFound, tt.wantFound)
				}
			},
		)
	}
}

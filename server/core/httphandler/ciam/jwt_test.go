package ciam

import (
	"errors"
	"reflect"
	"testing"
)

func Test_token_String(t *testing.T) {
	type fields struct {
		header    JWTHeader
		payload   JWTPayload
		signature string
	}
	tests := []struct {
		name    string
		fields  fields
		want    string
		wantErr error
	}{
		{
			name: "happy path: signed",
			fields: fields{
				header: JWTHeader{
					Alg: "foo",
					Typ: typ,
				},
				payload: JWTPayload{
					Sub: "bar",
				},
				signature: "qux",
			},
			want:    "eyJhbGciOiJmb28iLCJ0eXAiOiJKV1QifQ.eyJzdWIiOiJiYXIiLCJpc3MiOiIiLCJhdWQiOiIiLCJpYXQiOjAsImV4cCI6MH0.qux",
			wantErr: nil,
		},
		{
			name: "happy path: unsigned",
			fields: fields{
				header: JWTHeader{
					Alg: algNone,
					Typ: typ,
				},
				payload: JWTPayload{
					Sub: "bar",
				},
			},
			want: "eyJhbGciOiJub25lIiwidHlwIjoiSldUIn0.eyJzdWIiOiJiYXIiLCJpc3MiOiIiLCJhdWQiOiIiLCJpYXQiOjAsImV4cCI6MH0",
		},
		{
			name: "unhappy path: missing signature",
			fields: fields{
				header: JWTHeader{
					Alg: "foo",
					Typ: typ,
				},
				payload: JWTPayload{
					Sub: "bar",
				},
			},
			wantErr: errors.New("signature is missing"),
		},
		{
			name: "unhappy path: signature and alg mismatch - alg='none'",
			fields: fields{
				header: JWTHeader{
					Alg: algNone,
					Typ: typ,
				},
				payload: JWTPayload{
					Sub: "foo",
				},
				signature: "bar",
			},
			wantErr: errors.New("JWT header corrupt: alg value"),
		},
		{
			name: "unhappy path: signature and alg mismatch - alg=''",
			fields: fields{
				header: JWTHeader{
					Alg: "",
					Typ: typ,
				},
				payload: JWTPayload{
					Sub: "foo",
				},
				signature: "bar",
			},
			wantErr: errors.New("JWT header corrupt: alg value"),
		},
	}
	for _, tt := range tests {
		t.Run(
			tt.name, func(t *testing.T) {
				tkn := token{
					header:    tt.fields.header,
					payload:   tt.fields.payload,
					signature: tt.fields.signature,
				}
				got, err := tkn.String()
				if !reflect.DeepEqual(err, tt.wantErr) {
					t.Errorf("String() error = %v, wantErr %v", err, tt.wantErr)
					return
				}
				if got != tt.want {
					t.Errorf("String() got = %v, want %v", got, tt.want)
				}
			},
		)
	}
}

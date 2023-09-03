package ciam

import (
	"crypto/ed25519"
	"reflect"
	"testing"
)

func TestReadWritePrivateKey(t *testing.T) {
	generateKeyString := func(key ed25519.PrivateKey) string {
		v, err := MarshalKey(key)
		if err != nil {
			panic(err)
		}
		return string(v)
	}
	type args struct {
		s string
	}
	key := GenerateCertificate()
	tests := []struct {
		name    string
		args    args
		want    ed25519.PrivateKey
		wantErr bool
	}{
		{
			name: "hapy path",
			args: args{
				s: generateKeyString(key),
			},
			want:    key,
			wantErr: false,
		},
		{
			name: "unhappy path",
			args: args{
				s: "",
			},
			want:    nil,
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(
			tt.name, func(t *testing.T) {
				got, err := ReadPrivateKey(tt.args.s)
				if (err != nil) != tt.wantErr {
					t.Errorf("ReadPrivateKey() error = %v, wantErr %v", err, tt.wantErr)
					return
				}
				if !reflect.DeepEqual(got, tt.want) {
					t.Errorf("ReadPrivateKey() got = %v, want %v", got, tt.want)
				}
			},
		)
	}
}

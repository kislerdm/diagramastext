package errors

import (
	"log"
	"testing"
)

func TestNewError(t *testing.T) {
	type args struct {
		msg string
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			name: "foo",
			args: args{
				"foo",
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(
			tt.name, func(t *testing.T) {
				err := NewError(tt.args.msg)
				if (err != nil) != tt.wantErr {
					t.Errorf("NewError() error = %v, wantErr %v", err, tt.wantErr)
				}
				log.Println(err)
			},
		)
	}
}

package main

import "testing"

func Test_mustParseInt(t *testing.T) {
	type args struct {
		s string
	}
	tests := []struct {
		name string
		args args
		want int
	}{
		{
			name: "happy path",
			args: args{"10"},
			want: 10,
		},
		{
			name: "unhappy path",
			args: args{"foo"},
			want: 0,
		},
	}
	for _, tt := range tests {
		t.Run(
			tt.name, func(t *testing.T) {
				if got := mustParseInt(tt.args.s); got != tt.want {
					t.Errorf("mustParseInt() = %v, want %v", got, tt.want)
				}
			},
		)
	}
}

func Test_mustParseFloat32(t *testing.T) {
	type args struct {
		s string
	}
	tests := []struct {
		name string
		args args
		want float32
	}{
		{
			name: "happy path",
			args: args{"0.2"},
			want: 0.2,
		},
		{
			name: "unhappy path",
			args: args{"foo"},
			want: 0,
		},
	}
	for _, tt := range tests {
		t.Run(
			tt.name, func(t *testing.T) {
				if got := mustParseFloat32(tt.args.s); got != tt.want {
					t.Errorf("mustParseFloat32() = %v, want %v", got, tt.want)
				}
			},
		)
	}
}

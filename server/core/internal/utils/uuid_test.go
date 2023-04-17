package utils

import "testing"

// FIXME(?): employ gofuzz
func TestValidateUUID(t *testing.T) {
	type args struct {
		s string
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			name: "valid",
			args: args{
				"1a44e41a-49a7-49ce-a0a5-684f3d2dbe7d",
			},
			wantErr: false,
		},
		{
			name: "false",
			args: args{
				"11",
			},
			wantErr: true,
		},
		{
			name: "false",
			args: args{
				"6ba7b810-9dad-11d1-80b4-00c04fd430c",
			},
			wantErr: true,
		},
		{
			name: "false",
			args: args{
				"6ba7b8109dad11d180b400c04fd430c",
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(
			tt.name, func(t *testing.T) {
				if err := ValidateUUID(tt.args.s); (err != nil) != tt.wantErr {
					t.Errorf("ValidateUUID() error = %v, wantErr %v", err, tt.wantErr)
				}
			},
		)
	}
}

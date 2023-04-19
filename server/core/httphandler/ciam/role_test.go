package ciam

import "testing"

func TestRole_IsRegisteredUser(t *testing.T) {
	tests := []struct {
		name string
		r    Role
		want bool
	}{
		{
			name: "registered",
			r:    roleRegisteredUser,
			want: true,
		},
		{
			name: "not registered",
			r:    roleAnonymUser,
			want: false,
		},
	}
	for _, tt := range tests {
		t.Run(
			tt.name, func(t *testing.T) {
				if got := tt.r.IsRegisteredUser(); got != tt.want {
					t.Errorf("IsRegisteredUser() = %v, want %v", got, tt.want)
				}
			},
		)
	}
}

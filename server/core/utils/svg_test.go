package utils

import "testing"

func TestValidateSVG(t *testing.T) {
	type args struct {
		v []byte
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			name: "happy path",
			args: args{
				v: []byte(`<?xml version="1.0" encoding="us-ascii" standalone="no"?>
<svg xmlns="http://www.w3.org/2000/svg" contentstyletype="text/css" height="179px" preserveAspectRatio="none" version="1.1" viewBox="0 0 375 179" width="375px" zoomAndPan="magnify">
<defs></defs>
<g>
	<g id="elem_n0">
		<rect fill="#438DD5" height="52.5938" rx="2.5" ry="2.5" style="stroke:#3C7FC0;stroke-width:0.5;" width="125" x="7" y="11.8301"></rect>
		<text fill="#FFFFFF" font-family="sans-serif" font-size="16" font-weight="bold" lengthAdjust="spacing" textLength="40" x="17" y="36.6816">Web
		</text>
		<text fill="#FFFFFF" font-family="sans-serif" font-size="16" font-weight="bold" lengthAdjust="spacing" textLength="6" x="57" y="36.6816"></text>
		<text fill="#FFFFFF" font-family="sans-serif" font-size="16" font-weight="bold" lengthAdjust="spacing" textLength="59" x="63" y="36.6816">Server
		</text>
		<text fill="#FFFFFF" font-family="sans-serif" font-size="12" font-style="italic" lengthAdjust="spacing" textLength="26" x="56.5" y="51.5938">[Go]
		</text>
	</g>
	<g id="elem_n1">
		<path d="M251,17.3301 C251,7.3301 304.5,7.3301 304.5,7.3301 C304.5,7.3301 358,7.3301 358,17.3301 L358,58.9238 C358,68.9238 304.5,68.9238 304.5,68.9238 C304.5,68.9238 251,68.9238 251,58.9238 L251,17.3301 " fill="#B3B3B3" style="stroke:#A6A6A6;stroke-width:0.5;"></path>
		<path d="M251,17.3301 C251,27.3301 304.5,27.3301 304.5,27.3301 C304.5,27.3301 358,27.3301 358,17.3301 " fill="none" style="stroke:#A6A6A6;stroke-width:0.5;"></path>
		<text fill="#FFFFFF" font-family="sans-serif" font-size="16" font-weight="bold" lengthAdjust="spacing" textLength="87" x="261" y="46.1816">Database
		</text>
		<text fill="#FFFFFF" font-family="sans-serif" font-size="12" font-style="italic" lengthAdjust="spacing" textLength="61" x="274" y="61.0938">[Postgres]
		</text>
	</g>
</g>
</svg>`),
			},
			wantErr: false,
		},
		{
			name: "unhappy path: corrupt encoding",
			args: args{
				v: []byte(`<?xml version="1.0" encoding="foo"?><svg xmlns="http://www.w3.org/2000/svg" xmlns:xlink="http://www.w3.org/1999/xlink"></svg>`),
			},
			wantErr: true,
		},
		{
			name: "unhappy path: corrupt svg",
			args: args{
				v: []byte(`<?xml version="1.0" encoding="us-ascii"?><svg xmlns="http://www.w3.org/2000/svg" xmlns:xlink="http://www.w3.org/1999/xlink"`),
			},
			wantErr: true,
		},
		{
			name: "unhappy path: missing svg attributes",
			args: args{
				v: []byte(`<?xml version="1.0" encoding="us-ascii" standalone="no"?>
<svg xmlns="http://www.w3.org/2000/svg" contentstyletype="text/css" version="1.1" viewBox="0 0 375 179" width="375px" zoomAndPan="magnify">
<defs></defs><g></g></svg>`),
			},
			wantErr: true,
		},
		{
			name: "unhappy path: no geometries present",
			args: args{
				v: []byte(`<?xml version="1.0" encoding="us-ascii" standalone="no"?>
<svg xmlns="http://www.w3.org/2000/svg" contentstyletype="text/css" height="179px" preserveAspectRatio="none" version="1.1" viewBox="0 0 375 179" width="375px" zoomAndPan="magnify">
<defs></defs><g></g></svg>`),
			},
			wantErr: true,
		},
	}

	t.Parallel()
	for _, tt := range tests {
		t.Run(
			tt.name, func(t *testing.T) {
				if err := ValidateSVG(tt.args.v); (err != nil) != tt.wantErr {
					t.Errorf("ValidateSVG() error = %v, wantErr %v", err, tt.wantErr)
				}
			},
		)
	}
}

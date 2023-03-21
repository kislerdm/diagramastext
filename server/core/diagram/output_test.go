package diagram

import (
	"errors"
	"reflect"
	"testing"
)

func TestNewResultSVG(t *testing.T) {
	type args struct {
		v []byte
	}
	tests := []struct {
		name    string
		args    args
		want    Output
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
	</g>
</g>
</svg>`),
			},
			want: &responseSVG{
				SVG: `<?xml version="1.0" encoding="us-ascii" standalone="no"?>
<svg xmlns="http://www.w3.org/2000/svg" contentstyletype="text/css" height="179px" preserveAspectRatio="none" version="1.1" viewBox="0 0 375 179" width="375px" zoomAndPan="magnify">
<defs></defs>
<g>
	<g id="elem_n0">
		<rect fill="#438DD5" height="52.5938" rx="2.5" ry="2.5" style="stroke:#3C7FC0;stroke-width:0.5;" width="125" x="7" y="11.8301"></rect>
	</g>
</g>
</svg>`,
			},
			wantErr: false,
		},
		{
			name: "unhappy path: invalid svg",
			args: args{
				v: []byte{0},
			},
			want:    nil,
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(
			tt.name, func(t *testing.T) {
				got, err := NewResultSVG(tt.args.v)
				if (err != nil) != tt.wantErr {
					t.Errorf("NewResultSVG() error = %v, wantErr %v", err, tt.wantErr)
					return
				}
				if !reflect.DeepEqual(got, tt.want) {
					t.Errorf("NewResultSVG() got = %v, want %v", got, tt.want)
				}
			},
		)
	}
}

func Test_responseSVG_Serialize(t *testing.T) {
	type fields struct {
		SVG string
	}

	tests := []struct {
		name    string
		fields  fields
		want    []byte
		wantErr bool
	}{
		{
			name: "happy path",
			fields: fields{
				SVG: "foo",
			},
			want:    []byte(`{"svg":"foo"}`),
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(
			tt.name, func(t *testing.T) {
				r := responseSVG{
					SVG: tt.fields.SVG,
				}
				got, err := r.Serialize()
				if (err != nil) != tt.wantErr {
					t.Errorf("Serialize() error = %v, wantErr %v", err, tt.wantErr)
					return
				}
				if !reflect.DeepEqual(got, tt.want) {
					t.Errorf("Serialize() got = %v, want %v", got, tt.want)
				}
			},
		)
	}
}

func TestMockOutput_Serialize(t *testing.T) {
	type fields struct {
		V   []byte
		Err error
	}
	tests := []struct {
		name    string
		fields  fields
		want    []byte
		wantErr error
	}{
		{
			name: "happy path: error",
			fields: fields{
				Err: errors.New("foobar"),
			},
			wantErr: errors.New("foobar"),
		},
		{
			name: "happy path: results",
			fields: fields{
				V: []byte(`foo`),
			},
			want:    []byte(`foo`),
			wantErr: nil,
		},
	}
	for _, tt := range tests {
		t.Run(
			tt.name, func(t *testing.T) {
				m := MockOutput{
					V:   tt.fields.V,
					Err: tt.fields.Err,
				}
				got, err := m.Serialize()
				if !reflect.DeepEqual(tt.wantErr, err) {
					t.Errorf("unexpected error")
					return
				}

				if !reflect.DeepEqual(got, tt.want) {
					t.Errorf("unexpected result")
					return
				}
			},
		)
	}
}

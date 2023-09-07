package ciam

import (
	"net/smtp"
	"reflect"
	"testing"
)

func TestNewSMTClient(t *testing.T) {
	// GIVEN
	const (
		user        = "foo"
		password    = "bar"
		host        = "localhost"
		port        = "1025"
		senderEmail = "baz@qux.com"
	)

	// WHEN
	got := NewSMTPClient(user, password, host, port, senderEmail)
	var c *smtClient

	// THEN
	t.Parallel()
	t.Run(
		"shall initiate the client of *smtClient/SMTPClient type", func(t *testing.T) {
			var ok bool
			if c, ok = got.(*smtClient); !ok {
				t.Errorf("unexpected client")
			}
		},
	)
	t.Run(
		"shall initiate struct with the server address", func(t *testing.T) {
			if c.addr != host+":"+port {
				t.Errorf("unexpected addr: got = %s, want = %s", c.addr, host+":"+port)
			}
		},
	)
	t.Run(
		"shall initiate struct with the sender email", func(t *testing.T) {
			if c.sender != senderEmail {
				t.Errorf("unexpected sender email: got = %s, want = %s", c.sender, senderEmail)
			}
		},
	)
	t.Run(
		"shall initiate struct with plain authentication", func(t *testing.T) {
			if !reflect.DeepEqual(c.auth, smtp.PlainAuth("", user, password, host)) {
				t.Errorf("unexpected auth")
			}
		},
	)
}

func Test_generateMessage(t *testing.T) {
	t.Run(
		"happy path", func(t *testing.T) {
			// GIVEN
			const (
				recipient  = "foo@bar.baz"
				authSecret = "qux"
			)
			want := []byte(`To: foo@bar.baz
Subject: diagramastext.dev authentication code: qux
Content-Type: multipart/alternative; boundary="00"

--00
Content-Type: text/plain; charset="UTF-8";
Content-Transfer-Encoding: quoted-printable
Content-Disposition: inline

Complete authentication: copy the code qux and paste it in your browser with https://diagramastext.dev opened. 
Please ignore the email if you feel that it was received by mistake.

--00
Content-Type: text/html; charset="UTF-8";
Content-Transfer-Encoding: quoted-printable
Content-Disposition: inline

<!doctypehtml><html lang=en><title>diagramastext.dev authentication code: qux</title><meta content="width=device-width,initial-scale=1" name=viewport><style>*,:after,:before{box-sizing:border-box;border:0 solid #e5e7eb}html{line-height:1.5;-webkit-text-size-adjust:100%;tab-size:4;font-family:ui-sans-serif,system-ui,-apple-system,BlinkMacSystemFont,Segoe UI,Roboto,Helvetica Neue,Arial,sans-serif}body{display:flex;flex-direction:column;align-items:center;background-color:#e8e5e5;margin:0}main{width:600px}@media only screen and (max-width:600px){main{width:100%}}h1{font-size:30px;font-weight:700}a{color:#000}a:link{text-decoration:underline}a:active,a:hover,a:visited{text-decoration:none}.box{border-radius:1.2rem;padding:.8rem;border:#ccc9c9 solid 5px;box-shadow:0 0 5px 5px #ccc9c9;background:#263950;text-align:center;font-weight:700;font-size:25px;color:#fff}footer{margin-top:40px;align-content:center;text-align:center}p{font-size:14px}</style><main><h1>Complete authentication</h1><p>Copy the code below and paste it in your browser with <a href=https://diagramastext.dev target=_blank>diagramastext.dev</a> opened.<div class=box>qux</div><p>Please ignore the email if you feel that it was received by mistake.<footer><a href=https://diagramastext.dev target=_blank><svg fill=none preserveAspectRatio=true viewBox="0 0 128 93" width=80 xmlns="http://www.w3.org/2000/svg"><g filter=url(#a)><path d="M46.8 88.5 71.4 63l-12-63L128 88.5H46.8Z" fill=#B9CFE4 /></g><path d="M76 71.8v.4a7.3 7.3 0 0 1-14.5 0v-.4a7.3 7.3 0 0 1 14.5 0z" fill=#aaa stroke=#888 /><path d="m72 65 8.5-7.9-11-3.4L72 65zm7-26.3-5.3 17.4 1.9.6L81 39.3l-2-.6z" fill=#000 /><g filter=url(#b)><path d="M0 .6h59.5L72 63.1 46.7 88.5 0 .6z" fill=#084580 /><path d="M0 .6h59.5L72 63.1H0V.6Z" fill=#1168BD /></g><path d="M108 71.8v.4a7.3 7.3 0 0 1-14.5 0v-.4a7.3 7.3 0 0 1 14.5 0z" fill=#aaa stroke=#888 /><path d="m98 65 1.4-11.5L88.8 58l9.2 7zM86 39.4 93.7 57l1.8-.8L88 38.6l-1.8.8z" fill=#000 /><path d="M91 33.8v.4a7.3 7.3 0 0 1-14.5 0v-.4a7.3 7.3 0 0 1 14.5 0z" fill=#aaa stroke=#888 /><g filter=url(#c)><path d="m8 42.4 5.7-21.9h4.2l5.6 22h-3.3L19 36.8h-6.2l-1.3 5.5H8zm5.3-8.2h5L16.8 28a253 253 0 0 1-1-4.6 230.9 230.9 0 0 0-1 4.6l-1.5 6.3zm12.5 8.2V20.5h6.5c2 0 3.7.5 4.9 1.5s1.7 2.4 1.7 4.2a5 5 0 0 1-.6 2.6c-.4.7-1 1.3-1.8 1.7s-1.6.6-2.7.6v-.3a6 6 0 0 1 3 .6c.8.4 1.5 1 2 1.9s.7 1.8.7 3-.3 2.3-.9 3.3c-.5.9-1.3 1.6-2.3 2-1 .6-2.2.8-3.6.8h-6.9zm3.2-2.8h3.4c1.2 0 2.1-.3 2.8-.9s1-1.5 1-2.6c0-1-.3-2-1-2.7s-1.6-1-2.8-1H29v7.2zm0-9.9h3.3c1 0 1.9-.3 2.5-.9s1-1.3 1-2.3-.4-1.7-1-2.3c-.6-.6-1.5-.9-2.5-.9H29v6.4zm19.9 13a8 8 0 0 1-3.6-.7c-1-.5-1.8-1.3-2.3-2.2a7 7 0 0 1-.8-3.5v-9.7c0-1.3.2-2.4.8-3.4.5-1 1.3-1.7 2.3-2.2 1-.5 2.2-.8 3.6-.8s2.5.3 3.5.8 1.8 1.3 2.3 2.2c.6 1 .9 2.1.9 3.4h-3.3c0-1.1-.3-2-.9-2.6s-1.4-.9-2.5-.9-2 .3-2.6 1c-.6.5-.9 1.4-.9 2.5v9.7c0 1.2.3 2 1 2.7.5.6 1.4.9 2.5.9s2-.3 2.5-1c.6-.6 1-1.4 1-2.6h3.2c0 1.3-.3 2.5-.9 3.4-.5 1-1.3 1.7-2.3 2.3s-2.1.7-3.5.7z" fill=#fff /></g><defs><filter color-interpolation-filters=sRGB filterUnits=userSpaceOnUse height=96.5 id=a width=89.2 x=42.8 y=-4><feFlood flood-opacity=0 result=BackgroundImageFix /><feBlend in2=BackgroundImageFix result=shape in=SourceGraphic /><feGaussianBlur stdDeviation=2 result=effect1_foregroundBlur_10_107 /></filter><filter color-interpolation-filters=sRGB filterUnits=userSpaceOnUse height=95.9 id=b width=80.1 x=-4 y=-3.4><feFlood flood-opacity=0 result=BackgroundImageFix /><feBlend in2=BackgroundImageFix result=shape in=SourceGraphic /><feGaussianBlur stdDeviation=2 result=effect1_foregroundBlur_10_107 /></filter><filter color-interpolation-filters=sRGB filterUnits=userSpaceOnUse height=26.5 id=c width=47.5 x=8.1 y=20.2><feFlood flood-opacity=0 result=BackgroundImageFix /><feBlend in2=BackgroundImageFix result=shape in=SourceGraphic /><feColorMatrix values="0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 127 0" in=SourceAlpha result=hardAlpha /><feOffset dy=4 /><feGaussianBlur stdDeviation=2 /><feComposite in2=hardAlpha k2=-1 k3=1 operator=arithmetic /><feColorMatrix values="0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0.25 0"/><feBlend in2=shape result=effect1_innerShadow_10_107 /></filter></defs></svg></a><p style=margin-top:-5px><a href=https://diagramastext.dev target=_blank style=text-decoration:none>diagramastext.dev</a> &copy; 2023</footer></main>
--00--`)

			// WHEN
			got, err := generateMessage(recipient, authSecret)

			// THEN
			if err != nil {
				t.Errorf("unexpected error")
				return
			}
			if !reflect.DeepEqual(got, want) {
				t.Errorf("unexpected email template rendering result")
			}
		},
	)
}

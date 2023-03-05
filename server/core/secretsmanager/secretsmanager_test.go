package secretsmanager

import (
	"context"
	"reflect"
	"testing"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/secretsmanager"
)

func TestAWSSecretManager_ReadLatestSecret(t *testing.T) {
	type fields struct {
		Client awsSecretsManager
	}
	type args struct {
		ctx    context.Context
		uri    string
		output interface{}
	}

	type output struct {
		Foo string `json:"foo"`
	}

	tests := []struct {
		name    string
		fields  fields
		args    args
		want    *output
		wantErr bool
	}{
		{
			name: "happy path",
			fields: fields{
				Client: MockAWSSecretsmanagerClient{
					Output: &secretsmanager.GetSecretValueOutput{
						SecretString: aws.String(`{"foo":"bar"}`),
					},
				},
			},
			args: args{
				ctx:    context.TODO(),
				uri:    "arn:aws:secretsmanager:us-east-2:027889758114:secret:foo/bar/openai/lambda-C335bP",
				output: &output{},
			},
			want:    &output{Foo: "bar"},
			wantErr: false,
		},
		{
			name: "unhappy path: no secret found",
			fields: fields{
				Client: MockAWSSecretsmanagerClient{},
			},
			args: args{
				ctx:    context.TODO(),
				uri:    "arn:aws:secretsmanager:us-east-2:027889758114:secret:non-existing-A12Db1",
				output: &output{},
			},
			want:    nil,
			wantErr: true,
		},
		{
			name: "unhappy path: deserialization error",
			fields: fields{
				MockAWSSecretsmanagerClient{
					Output: &secretsmanager.GetSecretValueOutput{
						SecretString: aws.String(`foo`),
					},
				},
			},
			args: args{
				ctx:    context.TODO(),
				uri:    "arn:aws:secretsmanager:us-east-2:027889758114:secret:non-existing-A12Db1",
				output: &output{},
			},
			want:    nil,
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(
			tt.name, func(t *testing.T) {
				c := client{
					tt.fields.Client,
				}
				if err := c.ReadLatestSecret(tt.args.ctx, tt.args.uri, tt.args.output); (err != nil) != tt.wantErr {
					t.Errorf("ReadLatestSecret() error = %v, wantErr %v", err, tt.wantErr)
				}
				if !tt.wantErr && !reflect.DeepEqual(tt.want, tt.args.output) {
					t.Errorf("ReadLatestSecret() unexpected result of deserialisation")
				}
			},
		)
	}
}

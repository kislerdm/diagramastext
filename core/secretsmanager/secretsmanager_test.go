package secretsmanager

import (
	"context"
	"errors"
	"reflect"
	"testing"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/secretsmanager"
)

type mockAWSSecretsmanagerClient struct {
	output *secretsmanager.GetSecretValueOutput
}

func (m mockAWSSecretsmanagerClient) GetSecretValue(
	ctx context.Context, params *secretsmanager.GetSecretValueInput, optFns ...func(*secretsmanager.Options),
) (*secretsmanager.GetSecretValueOutput, error) {
	if m.output == nil {
		return nil, errors.New("no secret found")
	}
	return m.output, nil
}

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
				Client: mockAWSSecretsmanagerClient{
					output: &secretsmanager.GetSecretValueOutput{
						SecretString: aws.String(`{"foo":"bar"}`),
					},
				},
			},
			args: args{
				ctx:    context.TODO(),
				uri:    "arn:aws:secretsmanager:us-east-2:027889758114:secret:foo/bar/core/lambda-C335bP",
				output: &output{},
			},
			want:    &output{Foo: "bar"},
			wantErr: false,
		},
		{
			name: "unhappy path: no secret found",
			fields: fields{
				Client: mockAWSSecretsmanagerClient{},
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
				mockAWSSecretsmanagerClient{
					output: &secretsmanager.GetSecretValueOutput{
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
				c := AWSSecretManager{
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

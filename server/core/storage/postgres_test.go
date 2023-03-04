package storage

//func TestNewClient(t *testing.T) {
//	type args struct {
//		ctx      context.Context
//		host     string
//		dbname   string
//		user     string
//		password string
//	}
//	tests := []struct {
//		name    string
//		args    args
//		want    Client
//		wantErr bool
//	}{
//		{
//			name: "happy path",
//			args: args{
//				ctx:      context.TODO(),
//				host:     "mock",
//				dbname:   "postgres",
//				user:     "postgres",
//				password: "foo",
//			},
//			want:    pgClient{mockDbClient{}},
//			wantErr: false,
//		},
//		{
//			name: "unhappy path: host missing",
//			args: args{
//				ctx:      context.TODO(),
//				host:     "",
//				dbname:   "postgres",
//				user:     "postgres",
//				password: "",
//			},
//			wantErr: true,
//		},
//		{
//			name: "unhappy path: dbname missing",
//			args: args{
//				ctx:      context.TODO(),
//				host:     "localhost",
//				dbname:   "",
//				user:     "postgres",
//				password: "",
//			},
//			wantErr: true,
//		},
//		{
//			name: "unhappy path: user missing",
//			args: args{
//				ctx:      context.TODO(),
//				host:     "localhost",
//				dbname:   "postgres",
//				user:     "",
//				password: "",
//			},
//			wantErr: true,
//		},
//		{
//			name: "unhappy path: db not available",
//			args: args{
//				ctx:      context.TODO(),
//				host:     "localhost",
//				dbname:   "postgres",
//				user:     "postgres",
//				password: "",
//			},
//			wantErr: true,
//		},
//	}
//
//	t.Parallel()
//
//	for _, tt := range tests {
//		t.Run(
//			tt.name, func(t *testing.T) {
//				got, err := NewPgClient(tt.args.ctx, tt.args.host, tt.args.dbname, tt.args.user, tt.args.password)
//				if (err != nil) != tt.wantErr {
//					t.Errorf("NewPgClient() error = %v, wantErr %v", err, tt.wantErr)
//					return
//				}
//				if !reflect.DeepEqual(got, tt.want) {
//					t.Errorf("NewPgClient() got = %v, want %v", got, tt.want)
//				}
//			},
//		)
//	}
//}

//func Test_client_WritePrompt(t *testing.T) {
//	type fields struct {
//		c dbClient
//	}
//	type args struct {
//		ctx context.Context
//		v   UserInput
//	}
//	tests := []struct {
//		name    string
//		fields  fields
//		args    args
//		wantErr error
//	}{
//		{
//			name:   "happy path",
//			fields: fields{mockDbClient{}},
//			args: args{
//				ctx: context.TODO(),
//				v: UserInput{
//					CallID: CallID{
//						RequestID: "693a35ba-e42c-4168-8afc-5a7c359d1d05",
//						UserID:    "c40bad11-0822-4d84-9f61-44b9a97b0432",
//					},
//					Prompt:    "c4 diagram of four boxes",
//					Timestamp: time.Unix(0, 0),
//				},
//			},
//			wantErr: nil,
//		},
//		//{
//		//	name: "unhappy path: no table exists",
//		//	fields: fields{
//		//		mockDbClient{
//		//			err: errs.New(`relation "` + tableWritePrompt + `" does not exist (SQLSTATE 42P01)`),
//		//		},
//		//	},
//		//	args: args{
//		//		ctx: context.TODO(),
//		//		v: UserInput{
//		//			CallID: CallID{
//		//				RequestID: "693a35ba-e42c-4168-8afc-5a7c359d1d05",
//		//				UserID:    "c40bad11-0822-4d84-9f61-44b9a97b0432",
//		//			},
//		//			Prompt:    "c4 diagram of four boxes",
//		//			Timestamp: time.Unix(0, 0),
//		//		},
//		//	},
//		//	wantErr: errors.Error{
//		//		Service: errors.ServiceStorage,
//		//		Message: `relation "` + tableWritePrompt + `" does not exist (SQLSTATE 42P01)`,
//		//	},
//		//},
//		{
//			name:   "unhappy path: no request_id provided",
//			fields: fields{mockDbClient{}},
//			args: args{
//				ctx: context.TODO(),
//				v: UserInput{
//					CallID: CallID{
//						UserID: "c40bad11-0822-4d84-9f61-44b9a97b0432",
//					},
//					Prompt:    "c4 diagram of four boxes",
//					Timestamp: time.Unix(0, 0),
//				},
//			},
//			wantErr: errors.Error{
//				Service: errors.ServiceStorage,
//				Message: "request_id is required",
//			},
//		},
//		{
//			name:   "unhappy path: no prompt provided",
//			fields: fields{mockDbClient{}},
//			args: args{
//				ctx: context.TODO(),
//				v: UserInput{
//					CallID: CallID{
//						UserID:    "c40bad11-0822-4d84-9f61-44b9a97b0432",
//						RequestID: "693a35ba-e42c-4168-8afc-5a7c359d1d05",
//					},
//					Timestamp: time.Unix(0, 0),
//				},
//			},
//			wantErr: errors.Error{
//				Service: errors.ServiceStorage,
//				Message: "prompt is required",
//			},
//		},
//	}
//	for _, tt := range tests {
//		t.Run(
//			tt.name, func(t *testing.T) {
//				c := pgClient{c: tt.fields.c}
//				if err := c.WritePrompt(tt.args.ctx, tt.args.v); !reflect.DeepEqual(err, tt.wantErr) {
//					t.Errorf("WritePrompt() error = %v, wantErr %v", err, tt.wantErr)
//				}
//			},
//		)
//	}
//}

//func Test_client_WriteModelPrediction(t *testing.T) {
//	type fields struct {
//		c dbClient
//	}
//	type args struct {
//		ctx context.Context
//		v   ModelOutput
//	}
//	tests := []struct {
//		name    string
//		fields  fields
//		args    args
//		wantErr error
//	}{
//		{
//			name:   "happy path",
//			fields: fields{mockDbClient{}},
//			args: args{
//				ctx: context.TODO(),
//				v: ModelOutput{
//					CallID: CallID{
//						RequestID: "693a35ba-e42c-4168-8afc-5a7c359d1d05",
//						UserID:    "c40bad11-0822-4d84-9f61-44b9a97b0432",
//					},
//					Response:  `{"nodes:[{"id": "0"},{"id": "1"}],"links":[{"from":"0","to":"1"}]}`,
//					Timestamp: time.Unix(0, 0),
//				},
//			},
//			wantErr: nil,
//		},
//		{
//			name:   "unhappy path: no user_id provided",
//			fields: fields{mockDbClient{}},
//			args: args{
//				ctx: context.TODO(),
//				v: ModelOutput{
//					CallID: CallID{
//						RequestID: "693a35ba-e42c-4168-8afc-5a7c359d1d05",
//					},
//					Response:  `{"nodes:[{"id": "0"},{"id": "1"}],"links":[{"from":"0","to":"1"}]}`,
//					Timestamp: time.Unix(0, 0),
//				},
//			},
//			wantErr: errors.Error{
//				Service: errors.ServiceStorage,
//				Message: "user_id is required",
//			},
//		},
//		{
//			name:   "unhappy path: no request_id provided",
//			fields: fields{mockDbClient{}},
//			args: args{
//				ctx: context.TODO(),
//				v: ModelOutput{
//					CallID: CallID{
//						UserID: "c40bad11-0822-4d84-9f61-44b9a97b0432",
//					},
//					Response:  `{"nodes:[{"id": "0"},{"id": "1"}],"links":[{"from":"0","to":"1"}]}`,
//					Timestamp: time.Unix(0, 0),
//				},
//			},
//			wantErr: errors.Error{
//				Service: errors.ServiceStorage,
//				Message: "request_id is required",
//			},
//		},
//		{
//			name:   "unhappy path: no response provided",
//			fields: fields{mockDbClient{}},
//			args: args{
//				ctx: context.TODO(),
//				v: ModelOutput{
//					CallID: CallID{
//						RequestID: "693a35ba-e42c-4168-8afc-5a7c359d1d05",
//						UserID:    "c40bad11-0822-4d84-9f61-44b9a97b0432",
//					},
//					Timestamp: time.Unix(0, 0),
//				},
//			},
//			wantErr: errors.Error{
//				Service: errors.ServiceStorage,
//				Message: "response is required",
//			},
//		},
//	}
//	for _, tt := range tests {
//		t.Run(
//			tt.name, func(t *testing.T) {
//				c := pgClient{
//					c: tt.fields.c,
//				}
//				if err := c.WriteModelPrediction(tt.args.ctx, tt.args.v); !reflect.DeepEqual(err, tt.wantErr) {
//					t.Errorf("WriteModelPrediction() error = %v, wantErr %v", err, tt.wantErr)
//				}
//			},
//		)
//	}
//}

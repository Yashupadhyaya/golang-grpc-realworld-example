package handler

import (
	"context"
	"errors"
	"testing"
	"github.com/DATA-DOG/go-sqlmock"
	"github.com/raahii/golang-grpc-realworld-example/auth"
	"github.com/raahii/golang-grpc-realworld-example/proto"
	"github.com/raahii/golang-grpc-realworld-example/store"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"github.com/stretchr/testify/assert"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)





type ExpectedBegin struct {
	commonExpectation
	delay time.Duration
}
type ExpectedCommit struct {
	commonExpectation
}
type ExpectedExec struct {
	queryBasedExpectation
	result driver.Result
	delay  time.Duration
}
type ExpectedRollback struct {
	commonExpectation
}
type UserStore struct {
	db *gorm.DB
}
type T struct {
	common
	isEnvSet bool
	context  *testContext
}


/*
ROOST_METHOD_HASH=CreateUser_f2f8a1c84a
ROOST_METHOD_SIG_HASH=CreateUser_a3af3934da


 */
func TestHandlerCreateUser(t *testing.T) {
	type fields struct {
		logger *zerolog.Logger
		us     *store.UserStore
	}
	type args struct {
		ctx context.Context
		req *proto.CreateUserRequest
	}

	tests := []struct {
		name    string
		fields  fields
		args    args
		want    *proto.UserResponse
		wantErr error
		setup   func(mock sqlmock.Sqlmock, h *Handler)
	}{
		{
			name: "Scenario 1: Successful User Creation",
			fields: fields{
				logger: &log.Logger,
			},
			args: args{
				ctx: context.TODO(),
				req: &proto.CreateUserRequest{
					User: &proto.CreateUserRequest_User{
						Username: "validuser",
						Email:    "valid@example.com",
						Password: "validpassword",
					},
				},
			},
			want: &proto.UserResponse{
				User: &proto.User{
					Username: "validuser",
					Email:    "valid@example.com",
					Bio:      "",
					Image:    "",
				},
			},
			wantErr: nil,
			setup: func(mock sqlmock.Sqlmock, h *Handler) {
				mock.ExpectBegin()
				mock.ExpectExec("INSERT INTO `users`").WillReturnResult(sqlmock.NewResult(1, 1))
				mock.ExpectCommit()
				auth.GenerateToken = func(id uint) (string, error) {
					return "validtoken", nil
				}
			},
		},
		{
			name: "Scenario 2: Validation Error on Missing Username",
			fields: fields{
				logger: &log.Logger,
			},
			args: args{
				ctx: context.TODO(),
				req: &proto.CreateUserRequest{
					User: &proto.CreateUserRequest_User{
						Email:    "valid@example.com",
						Password: "validpassword",
					},
				},
			},
			want:    nil,
			wantErr: status.Error(codes.InvalidArgument, "validation error"),
			setup:   nil,
		},
		{
			name: "Scenario 3: Validation Error on Invalid Email",
			fields: fields{
				logger: &log.Logger,
			},
			args: args{
				ctx: context.TODO(),
				req: &proto.CreateUserRequest{
					User: &proto.CreateUserRequest_User{
						Username: "validuser",
						Email:    "invalidemail",
						Password: "validpassword",
					},
				},
			},
			want:    nil,
			wantErr: status.Error(codes.InvalidArgument, "validation error"),
			setup:   nil,
		},
		{
			name: "Scenario 4: Internal Error on Password Hashing Failure",
			fields: fields{
				logger: &log.Logger,
			},
			args: args{
				ctx: context.TODO(),
				req: &proto.CreateUserRequest{
					User: &proto.CreateUserRequest_User{
						Username: "validuser",
						Email:    "valid@example.com",
						Password: "",
					},
				},
			},
			want:    nil,
			wantErr: status.Error(codes.Aborted, "internal server error"),
			setup:   nil,
		},
		{
			name: "Scenario 5: Internal Error on User Store Failure",
			fields: fields{
				logger: &log.Logger,
			},
			args: args{
				ctx: context.TODO(),
				req: &proto.CreateUserRequest{
					User: &proto.CreateUserRequest_User{
						Username: "validuser",
						Email:    "valid@example.com",
						Password: "validpassword",
					},
				},
			},
			want:    nil,
			wantErr: status.Error(codes.Canceled, "internal server error"),
			setup: func(mock sqlmock.Sqlmock, h *Handler) {
				mock.ExpectBegin()
				mock.ExpectExec("INSERT INTO `users`").WillReturnError(errors.New("db error"))
				mock.ExpectRollback()
			},
		},
		{
			name: "Scenario 6: Internal Error on Token Generation Failure",
			fields: fields{
				logger: &log.Logger,
			},
			args: args{
				ctx: context.TODO(),
				req: &proto.CreateUserRequest{
					User: &proto.CreateUserRequest_User{
						Username: "validuser",
						Email:    "valid@example.com",
						Password: "validpassword",
					},
				},
			},
			want:    nil,
			wantErr: status.Error(codes.Aborted, "internal server error"),
			setup: func(mock sqlmock.Sqlmock, h *Handler) {
				mock.ExpectBegin()
				mock.ExpectExec("INSERT INTO `users`").WillReturnResult(sqlmock.NewResult(1, 1))
				mock.ExpectCommit()
				auth.GenerateToken = func(id uint) (string, error) {
					return "", errors.New("token error")
				}
			},
		},
		{
			name: "Scenario 7: Duplicate User Error",
			fields: fields{
				logger: &log.Logger,
			},
			args: args{
				ctx: context.TODO(),
				req: &proto.CreateUserRequest{
					User: &proto.CreateUserRequest_User{
						Username: "duplicateuser",
						Email:    "duplicate@example.com",
						Password: "validpassword",
					},
				},
			},
			want:    nil,
			wantErr: status.Error(codes.Canceled, "internal server error"),
			setup: func(mock sqlmock.Sqlmock, h *Handler) {
				mock.ExpectBegin()
				mock.ExpectExec("INSERT INTO `users`").WillReturnError(errors.New("duplicate error"))
				mock.ExpectRollback()
			},
		},
		{
			name: "Scenario 8: Successful User Creation with Optional Fields",
			fields: fields{
				logger: &log.Logger,
			},
			args: args{
				ctx: context.TODO(),
				req: &proto.CreateUserRequest{
					User: &proto.CreateUserRequest_User{
						Username: "validuser",
						Email:    "valid@example.com",
						Password: "validpassword",
					},
				},
			},
			want: &proto.UserResponse{
				User: &proto.User{
					Username: "validuser",
					Email:    "valid@example.com",
				},
			},
			wantErr: nil,
			setup: func(mock sqlmock.Sqlmock, h *Handler) {
				mock.ExpectBegin()
				mock.ExpectExec("INSERT INTO `users`").WillReturnResult(sqlmock.NewResult(1, 1))
				mock.ExpectCommit()
				auth.GenerateToken = func(id uint) (string, error) {
					return "validtoken", nil
				}
			},
		},
		{
			name: "Scenario 9: Handling Context Cancellation",
			fields: fields{
				logger: &log.Logger,
			},
			args: args{
				ctx: context.TODO(),
				req: &proto.CreateUserRequest{
					User: &proto.CreateUserRequest_User{
						Username: "validuser",
						Email:    "valid@example.com",
						Password: "validpassword",
					},
				},
			},
			want:    nil,
			wantErr: status.Error(codes.Canceled, "context canceled"),
			setup: func(mock sqlmock.Sqlmock, h *Handler) {
				h.CreateUser(context.Background(), &proto.CreateUserRequest{
					User: &proto.CreateUserRequest_User{
						Username: "validuser",
						Email:    "valid@example.com",
						Password: "validpassword",
					},
				})
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			db, mock, err := sqlmock.New()
			if err != nil {
				t.Fatalf("failed to open sqlmock database: %v", err)
			}
			defer db.Close()
			tt.fields.us = &store.UserStore{DB: db}

			h := &Handler{
				logger: tt.fields.logger,
				us:     tt.fields.us,
			}

			if tt.setup != nil {
				tt.setup(mock, h)
			}

			got, err := h.CreateUser(tt.args.ctx, tt.args.req)
			if tt.wantErr != nil {
				assert.Error(t, err)
				assert.Equal(t, tt.wantErr, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.want, got)
			}

			if err := mock.ExpectationsWereMet(); err != nil {
				t.Errorf("there were unfulfilled expectations: %s", err)
			}
		})
	}
}


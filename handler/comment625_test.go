package handler

import (
	"context"
	"errors"
	"testing"
	"github.com/DATA-DOG/go-sqlmock"
	"github.com/raahii/golang-grpc-realworld-example/auth"
	"github.com/raahii/golang-grpc-realworld-example/model"
	"github.com/raahii/golang-grpc-realworld-example/store"
	"github.com/raahii/golang-grpc-realworld-example/proto"
	"github.com/rs/zerolog"
	"github.com/stretchr/testify/assert"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)





type ExpectedExec struct {
	queryBasedExpectation
	result driver.Result
	delay  time.Duration
}
type ExpectedQuery struct {
	queryBasedExpectation
	rows             driver.Rows
	delay            time.Duration
	rowsMustBeClosed bool
	rowsWereClosed   bool
}
type ArticleStore struct {
	db *gorm.DB
}
type UserStore struct {
	db *gorm.DB
}
type Logger struct {
	w       LevelWriter
	level   Level
	sampler Sampler
	context []byte
	hooks   []Hook
}
type T struct {
	common
	isEnvSet bool
	context  *testContext
}


/*
ROOST_METHOD_HASH=CreateComment_c4ccd62dc5
ROOST_METHOD_SIG_HASH=CreateComment_19a3ee5a3b


 */
func TestHandlerCreateComment(t *testing.T) {

	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("failed to open sqlmock database: %v", err)
	}
	defer db.Close()

	userStore := &store.UserStore{DB: db}
	articleStore := &store.ArticleStore{DB: db}
	logger := zerolog.New(nil)

	handler := &Handler{
		logger: &logger,
		us:     userStore,
		as:     articleStore,
	}

	tests := []struct {
		name         string
		req          *pb.CreateCommentRequest
		mockSetup    func()
		expectError  bool
		expectedCode codes.Code
	}{
		{
			name: "Successful Comment Creation",
			req: &pb.CreateCommentRequest{
				Slug: "1",
				Comment: &pb.CreateCommentRequest_Comment{
					Body: "This is a test comment",
				},
			},
			mockSetup: func() {
				originalAuthGetUserID := auth.GetUserID
				auth.GetUserID = func(ctx context.Context) (uint, error) {
					return 1, nil
				}
				defer func() { auth.GetUserID = originalAuthGetUserID }()

				mock.ExpectQuery("^SELECT (.+) FROM `users` WHERE `users`.`id` = \\?").
					WithArgs(1).
					WillReturnRows(sqlmock.NewRows([]string{"id", "username"}).AddRow(1, "testuser"))
				mock.ExpectQuery("^SELECT (.+) FROM `articles` WHERE `articles`.`id` = \\?").
					WithArgs(1).
					WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(1))
				mock.ExpectExec("^INSERT INTO `comments`").
					WillReturnResult(sqlmock.NewResult(1, 1))
			},
			expectError: false,
		},
		{
			name: "Unauthenticated User",
			req: &pb.CreateCommentRequest{
				Slug: "1",
				Comment: &pb.CreateCommentRequest_Comment{
					Body: "This is a test comment",
				},
			},
			mockSetup: func() {
				originalAuthGetUserID := auth.GetUserID
				auth.GetUserID = func(ctx context.Context) (uint, error) {
					return 0, errors.New("unauthenticated")
				}
				defer func() { auth.GetUserID = originalAuthGetUserID }()
			},
			expectError:  true,
			expectedCode: codes.Unauthenticated,
		},
		{
			name: "User Not Found",
			req: &pb.CreateCommentRequest{
				Slug: "1",
				Comment: &pb.CreateCommentRequest_Comment{
					Body: "This is a test comment",
				},
			},
			mockSetup: func() {
				originalAuthGetUserID := auth.GetUserID
				auth.GetUserID = func(ctx context.Context) (uint, error) {
					return 1, nil
				}
				defer func() { auth.GetUserID = originalAuthGetUserID }()

				mock.ExpectQuery("^SELECT (.+) FROM `users` WHERE `users`.`id` = \\?").
					WithArgs(1).
					WillReturnError(errors.New("user not found"))
			},
			expectError:  true,
			expectedCode: codes.NotFound,
		},
		{
			name: "Invalid Article ID (Non-integer Slug)",
			req: &pb.CreateCommentRequest{
				Slug: "invalid-slug",
				Comment: &pb.CreateCommentRequest_Comment{
					Body: "This is a test comment",
				},
			},
			mockSetup:    func() {},
			expectError:  true,
			expectedCode: codes.InvalidArgument,
		},
		{
			name: "Article Not Found",
			req: &pb.CreateCommentRequest{
				Slug: "1",
				Comment: &pb.CreateCommentRequest_Comment{
					Body: "This is a test comment",
				},
			},
			mockSetup: func() {
				originalAuthGetUserID := auth.GetUserID
				auth.GetUserID = func(ctx context.Context) (uint, error) {
					return 1, nil
				}
				defer func() { auth.GetUserID = originalAuthGetUserID }()

				mock.ExpectQuery("^SELECT (.+) FROM `users` WHERE `users`.`id` = \\?").
					WithArgs(1).
					WillReturnRows(sqlmock.NewRows([]string{"id", "username"}).AddRow(1, "testuser"))
				mock.ExpectQuery("^SELECT (.+) FROM `articles` WHERE `articles`.`id` = \\?").
					WithArgs(1).
					WillReturnError(errors.New("article not found"))
			},
			expectError:  true,
			expectedCode: codes.InvalidArgument,
		},
		{
			name: "Comment Validation Failure",
			req: &pb.CreateCommentRequest{
				Slug: "1",
				Comment: &pb.CreateCommentRequest_Comment{
					Body: "",
				},
			},
			mockSetup: func() {
				originalAuthGetUserID := auth.GetUserID
				auth.GetUserID = func(ctx context.Context) (uint, error) {
					return 1, nil
				}
				defer func() { auth.GetUserID = originalAuthGetUserID }()

				mock.ExpectQuery("^SELECT (.+) FROM `users` WHERE `users`.`id` = \\?").
					WithArgs(1).
					WillReturnRows(sqlmock.NewRows([]string{"id", "username"}).AddRow(1, "testuser"))
				mock.ExpectQuery("^SELECT (.+) FROM `articles` WHERE `articles`.`id` = \\?").
					WithArgs(1).
					WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(1))
			},
			expectError:  true,
			expectedCode: codes.InvalidArgument,
		},
		{
			name: "Comment Creation Failure",
			req: &pb.CreateCommentRequest{
				Slug: "1",
				Comment: &pb.CreateCommentRequest_Comment{
					Body: "This is a test comment",
				},
			},
			mockSetup: func() {
				originalAuthGetUserID := auth.GetUserID
				auth.GetUserID = func(ctx context.Context) (uint, error) {
					return 1, nil
				}
				defer func() { auth.GetUserID = originalAuthGetUserID }()

				mock.ExpectQuery("^SELECT (.+) FROM `users` WHERE `users`.`id` = \\?").
					WithArgs(1).
					WillReturnRows(sqlmock.NewRows([]string{"id", "username"}).AddRow(1, "testuser"))
				mock.ExpectQuery("^SELECT (.+) FROM `articles` WHERE `articles`.`id` = \\?").
					WithArgs(1).
					WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(1))
				mock.ExpectExec("^INSERT INTO `comments`").
					WillReturnError(errors.New("comment creation failed"))
			},
			expectError:  true,
			expectedCode: codes.Aborted,
		},
		{
			name: "Missing Comment Body",
			req: &pb.CreateCommentRequest{
				Slug: "1",
				Comment: &pb.CreateCommentRequest_Comment{
					Body: "",
				},
			},
			mockSetup: func() {
				originalAuthGetUserID := auth.GetUserID
				auth.GetUserID = func(ctx context.Context) (uint, error) {
					return 1, nil
				}
				defer func() { auth.GetUserID = originalAuthGetUserID }()

				mock.ExpectQuery("^SELECT (.+) FROM `users` WHERE `users`.`id` = \\?").
					WithArgs(1).
					WillReturnRows(sqlmock.NewRows([]string{"id", "username"}).AddRow(1, "testuser"))
				mock.ExpectQuery("^SELECT (.+) FROM `articles` WHERE `articles`.`id` = \\?").
					WithArgs(1).
					WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(1))
			},
			expectError:  true,
			expectedCode: codes.InvalidArgument,
		},
		{
			name: "Missing Article Slug",
			req: &pb.CreateCommentRequest{
				Slug: "",
				Comment: &pb.CreateCommentRequest_Comment{
					Body: "This is a test comment",
				},
			},
			mockSetup: func() {
				originalAuthGetUserID := auth.GetUserID
				auth.GetUserID = func(ctx context.Context) (uint, error) {
					return 1, nil
				}
				defer func() { auth.GetUserID = originalAuthGetUserID }()

				mock.ExpectQuery("^SELECT (.+) FROM `users` WHERE `users`.`id` = \\?").
					WithArgs(1).
					WillReturnRows(sqlmock.NewRows([]string{"id", "username"}).AddRow(1, "testuser"))
			},
			expectError:  true,
			expectedCode: codes.InvalidArgument,
		},
		{
			name: "User Not Authorized",
			req: &pb.CreateCommentRequest{
				Slug: "1",
				Comment: &pb.CreateCommentRequest_Comment{
					Body: "This is a test comment",
				},
			},
			mockSetup: func() {
				originalAuthGetUserID := auth.GetUserID
				auth.GetUserID = func(ctx context.Context) (uint, error) {
					return 1, nil
				}
				defer func() { auth.GetUserID = originalAuthGetUserID }()

				mock.ExpectQuery("^SELECT (.+) FROM `users` WHERE `users`.`id` = \\?").
					WithArgs(1).
					WillReturnRows(sqlmock.NewRows([]string{"id", "username"}).AddRow(1, "testuser"))
				mock.ExpectQuery("^SELECT (.+) FROM `articles` WHERE `articles`.`id` = \\?").
					WithArgs(1).
					WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(1))
				mock.ExpectExec("^INSERT INTO `comments`").
					WillReturnError(errors.New("not authorized"))
			},
			expectError:  true,
			expectedCode: codes.PermissionDenied,
		},
	}

	for _, tc := range tests {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Logf("Running test case: %s", tc.name)
			if tc.mockSetup != nil {
				tc.mockSetup()
			}

			resp, err := handler.CreateComment(context.Background(), tc.req)

			if tc.expectError {
				assert.Error(t, err)
				st, ok := status.FromError(err)
				assert.True(t, ok)
				assert.Equal(t, tc.expectedCode, st.Code())
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, resp)
				assert.Equal(t, tc.req.GetComment().GetBody(), resp.Comment.Body)
			}

			if err := mock.ExpectationsWereMet(); err != nil {
				t.Errorf("there were unfulfilled expectations: %s", err)
			}
		})
	}
}


package handler

import (
	"context"
	"testing"
	"time"
	"github.com/DATA-DOG/go-sqlmock"
	"github.com/raahii/golang-grpc-realworld-example/auth"
	"github.com/raahii/golang-grpc-realworld-example/model"
	pb "github.com/raahii/golang-grpc-realworld-example/proto"
	"github.com/raahii/golang-grpc-realworld-example/store"
	"github.com/rs/zerolog"
	"github.com/stretchr/testify/assert"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)


/*
ROOST_METHOD_HASH=CreateArticle_64372fa1a8
ROOST_METHOD_SIG_HASH=CreateArticle_ce1c125740


 */
func TestHandlerCreateArticle(t *testing.T) {

	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("Failed to open sqlmock database: %v", err)
	}
	defer db.Close()
	logger := zerolog.Nop()

	userStore := &store.UserStore{DB: db}
	articleStore := &store.ArticleStore{DB: db}
	h := &Handler{logger: &logger, us: userStore, as: articleStore}

	type testCase struct {
		name          string
		ctx           context.Context
		req           *pb.CreateAritcleRequest
		mockSetup     func()
		expectedError error
		expectedResp  *pb.ArticleResponse
	}

	testCases := []testCase{
		{
			name: "Successfully Create an Article",
			ctx:  auth.ContextWithUserID(context.Background(), 1),
			req: &pb.CreateAritcleRequest{
				Article: &pb.CreateAritcleRequest_Article{
					Title:       "Test Article",
					Description: "Test Description",
					Body:        "Test Body",
					TagList:     []string{"tag1", "tag2"},
				},
			},
			mockSetup: func() {
				mock.ExpectQuery("SELECT \\* FROM `users` WHERE `id` = \\?").
					WithArgs(1).
					WillReturnRows(
						sqlmock.NewRows([]string{"id", "username"}).
							AddRow(1, "testuser"))

				mock.ExpectBegin()
				mock.ExpectExec("INSERT INTO `articles`").
					WithArgs(sqlmock.AnyArg(), "Test Article", "Test Description", "Test Body", 1, sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg()).
					WillReturnResult(sqlmock.NewResult(1, 1))
				mock.ExpectCommit()
			},
			expectedError: nil,
			expectedResp: &pb.ArticleResponse{
				Article: &pb.Article{
					Title:          "Test Article",
					Description:    "Test Description",
					Body:           "Test Body",
					TagList:        []string{"tag1", "tag2"},
					Favorited:      true,
					FavoritesCount: 0,
					Author: &pb.Profile{
						Username: "testuser",
					},
				},
			},
		},
		{
			name: "Unauthenticated User",
			ctx:  context.Background(),
			req: &pb.CreateAritcleRequest{
				Article: &pb.CreateAritcleRequest_Article{
					Title:       "Test Article",
					Description: "Test Description",
					Body:        "Test Body",
					TagList:     []string{"tag1", "tag2"},
				},
			},
			mockSetup:     func() {},
			expectedError: status.Errorf(codes.Unauthenticated, "unauthenticated"),
			expectedResp:  nil,
		},
		{
			name: "User Not Found",
			ctx:  auth.ContextWithUserID(context.Background(), 1),
			req: &pb.CreateAritcleRequest{
				Article: &pb.CreateAritcleRequest_Article{
					Title:       "Test Article",
					Description: "Test Description",
					Body:        "Test Body",
					TagList:     []string{"tag1", "tag2"},
				},
			},
			mockSetup: func() {
				mock.ExpectQuery("SELECT \\* FROM `users` WHERE `id` = \\?").
					WithArgs(1).
					WillReturnError(status.Error(codes.NotFound, "user not found"))
			},
			expectedError: status.Error(codes.NotFound, "user not found"),
			expectedResp:  nil,
		},
		{
			name: "Validation Error in Article Data",
			ctx:  auth.ContextWithUserID(context.Background(), 1),
			req: &pb.CreateAritcleRequest{
				Article: &pb.CreateAritcleRequest_Article{
					Title:       "",
					Description: "Test Description",
					Body:        "Test Body",
					TagList:     []string{"tag1", "tag2"},
				},
			},
			mockSetup: func() {
				mock.ExpectQuery("SELECT \\* FROM `users` WHERE `id` = \\?").
					WithArgs(1).
					WillReturnRows(
						sqlmock.NewRows([]string{"id", "username"}).
							AddRow(1, "testuser"))
			},
			expectedError: status.Error(codes.InvalidArgument, "validation error"),
			expectedResp:  nil,
		},
		{
			name: "Database Error When Creating Article",
			ctx:  auth.ContextWithUserID(context.Background(), 1),
			req: &pb.CreateAritcleRequest{
				Article: &pb.CreateAritcleRequest_Article{
					Title:       "Test Article",
					Description: "Test Description",
					Body:        "Test Body",
					TagList:     []string{"tag1", "tag2"},
				},
			},
			mockSetup: func() {
				mock.ExpectQuery("SELECT \\* FROM `users` WHERE `id` = \\?").
					WithArgs(1).
					WillReturnRows(
						sqlmock.NewRows([]string{"id", "username"}).
							AddRow(1, "testuser"))

				mock.ExpectBegin()
				mock.ExpectExec("INSERT INTO `articles`").
					WithArgs(sqlmock.AnyArg(), "Test Article", "Test Description", "Test Body", 1, sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg()).
					WillReturnError(status.Error(codes.Canceled, "Failed to create user."))
				mock.ExpectRollback()
			},
			expectedError: status.Error(codes.Canceled, "Failed to create user."),
			expectedResp:  nil,
		},
		{
			name: "Error Getting Following Status",
			ctx:  auth.ContextWithUserID(context.Background(), 1),
			req: &pb.CreateAritcleRequest{
				Article: &pb.CreateAritcleRequest_Article{
					Title:       "Test Article",
					Description: "Test Description",
					Body:        "Test Body",
					TagList:     []string{"tag1", "tag2"},
				},
			},
			mockSetup: func() {
				mock.ExpectQuery("SELECT \\* FROM `users` WHERE `id` = \\?").
					WithArgs(1).
					WillReturnRows(
						sqlmock.NewRows([]string{"id", "username"}).
							AddRow(1, "testuser"))

				mock.ExpectBegin()
				mock.ExpectExec("INSERT INTO `articles`").
					WithArgs(sqlmock.AnyArg(), "Test Article", "Test Description", "Test Body", 1, sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg()).
					WillReturnResult(sqlmock.NewResult(1, 1))
				mock.ExpectCommit()

				mock.ExpectQuery("SELECT \\* FROM `follows` WHERE `from_user_id` = \\? AND `to_user_id` = \\?").
					WithArgs(1, 1).
					WillReturnError(status.Error(codes.Internal, "internal server error"))
			},
			expectedError: status.Error(codes.Internal, "internal server error"),
			expectedResp:  nil,
		},
		{
			name: "Article Created with Tags",
			ctx:  auth.ContextWithUserID(context.Background(), 1),
			req: &pb.CreateAritcleRequest{
				Article: &pb.CreateAritcleRequest_Article{
					Title:       "Test Article",
					Description: "Test Description",
					Body:        "Test Body",
					TagList:     []string{"tag1", "tag2"},
				},
			},
			mockSetup: func() {
				mock.ExpectQuery("SELECT \\* FROM `users` WHERE `id` = \\?").
					WithArgs(1).
					WillReturnRows(
						sqlmock.NewRows([]string{"id", "username"}).
							AddRow(1, "testuser"))

				mock.ExpectBegin()
				mock.ExpectExec("INSERT INTO `articles`").
					WithArgs(sqlmock.AnyArg(), "Test Article", "Test Description", "Test Body", 1, sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg()).
					WillReturnResult(sqlmock.NewResult(1, 1))
				mock.ExpectCommit()
			},
			expectedError: nil,
			expectedResp: &pb.ArticleResponse{
				Article: &pb.Article{
					Title:          "Test Article",
					Description:    "Test Description",
					Body:           "Test Body",
					TagList:        []string{"tag1", "tag2"},
					Favorited:      true,
					FavoritesCount: 0,
					Author: &pb.Profile{
						Username: "testuser",
					},
				},
			},
		},
		{
			name: "Article Favorited by Default",
			ctx:  auth.ContextWithUserID(context.Background(), 1),
			req: &pb.CreateAritcleRequest{
				Article: &pb.CreateAritcleRequest_Article{
					Title:       "Test Article",
					Description: "Test Description",
					Body:        "Test Body",
					TagList:     []string{"tag1", "tag2"},
				},
			},
			mockSetup: func() {
				mock.ExpectQuery("SELECT \\* FROM `users` WHERE `id` = \\?").
					WithArgs(1).
					WillReturnRows(
						sqlmock.NewRows([]string{"id", "username"}).
							AddRow(1, "testuser"))

				mock.ExpectBegin()
				mock.ExpectExec("INSERT INTO `articles`").
					WithArgs(sqlmock.AnyArg(), "Test Article", "Test Description", "Test Body", 1, sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg()).
					WillReturnResult(sqlmock.NewResult(1, 1))
				mock.ExpectCommit()
			},
			expectedError: nil,
			expectedResp: &pb.ArticleResponse{
				Article: &pb.Article{
					Title:          "Test Article",
					Description:    "Test Description",
					Body:           "Test Body",
					TagList:        []string{"tag1", "tag2"},
					Favorited:      true,
					FavoritesCount: 0,
					Author: &pb.Profile{
						Username: "testuser",
					},
				},
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			tc.mockSetup()

			resp, err := h.CreateArticle(tc.ctx, tc.req)
			if tc.expectedError != nil {
				assert.Equal(t, tc.expectedError, err)
			} else {
				assert.NoError(t, err)
			}
			assert.Equal(t, tc.expectedResp, resp)

			if err := mock.ExpectationsWereMet(); err != nil {
				t.Errorf("there were unfulfilled expectations: %s", err)
			}
		})
	}
}


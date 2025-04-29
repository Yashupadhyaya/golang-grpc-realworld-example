package handler

import (
	"context"
	"errors"
	"testing"
	"github.com/DATA-DOG/go-sqlmock"
	"github.com/raahii/golang-grpc-realworld-example/auth"
	"github.com/raahii/golang-grpc-realworld-example/model"
	"github.com/raahii/golang-grpc-realworld-example/proto"
	"github.com/raahii/golang-grpc-realworld-example/store"
	"github.com/rs/zerolog"
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
type ExpectedQuery struct {
	queryBasedExpectation
	rows             driver.Rows
	delay            time.Duration
	rowsMustBeClosed bool
	rowsWereClosed   bool
}
type ExpectedRollback struct {
	commonExpectation
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
ROOST_METHOD_HASH=CreateArticle_64372fa1a8
ROOST_METHOD_SIG_HASH=CreateArticle_ce1c125740


 */
func TestHandlerCreateArticle(t *testing.T) {

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
		name             string
		ctx              context.Context
		req              *pb.CreateAritcleRequest
		mockUserStore    func()
		mockArticleStore func()
		expectedResp     *pb.ArticleResponse
		expectedErr      error
	}{
		{
			name: "Successful Article Creation",
			ctx:  context.WithValue(context.Background(), "user_id", uint(1)),
			req: &pb.CreateAritcleRequest{
				Article: &pb.CreateAritcleRequest_Article{
					Title:       "Test Article",
					Description: "Test Description",
					Body:        "Test Body",
					TagList:     []string{"tag1", "tag2"},
				},
			},
			mockUserStore: func() {
				mock.ExpectQuery("SELECT * FROM `users` WHERE `users`.`id` = ?").
					WithArgs(1).
					WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(1))
			},
			mockArticleStore: func() {
				mock.ExpectBegin()
				mock.ExpectExec("INSERT INTO `articles`").
					WithArgs(sqlmock.AnyArg(), "Test Article", "Test Description", "Test Body", uint(1), 0).
					WillReturnResult(sqlmock.NewResult(1, 1))
				mock.ExpectCommit()
			},
			expectedResp: &pb.ArticleResponse{
				Article: &pb.Article{
					Title:          "Test Article",
					Description:    "Test Description",
					Body:           "Test Body",
					TagList:        []string{"tag1", "tag2"},
					Favorited:      true,
					FavoritesCount: 0,
					Author: &pb.Profile{
						Username:  "TestUser",
						Bio:       "",
						Image:     "",
						Following: false,
					},
				},
			},
			expectedErr: nil,
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
			mockUserStore:    func() {},
			mockArticleStore: func() {},
			expectedResp:     nil,
			expectedErr:      status.Error(codes.Unauthenticated, "unauthenticated"),
		},
		{
			name: "User Not Found",
			ctx:  context.WithValue(context.Background(), "user_id", uint(1)),
			req: &pb.CreateAritcleRequest{
				Article: &pb.CreateAritcleRequest_Article{
					Title:       "Test Article",
					Description: "Test Description",
					Body:        "Test Body",
					TagList:     []string{"tag1", "tag2"},
				},
			},
			mockUserStore: func() {
				mock.ExpectQuery("SELECT * FROM `users` WHERE `users`.`id` = ?").
					WithArgs(1).
					WillReturnError(errors.New("user not found"))
			},
			mockArticleStore: func() {},
			expectedResp:     nil,
			expectedErr:      status.Error(codes.NotFound, "user not found"),
		},
		{
			name: "Validation Error in Article",
			ctx:  context.WithValue(context.Background(), "user_id", uint(1)),
			req: &pb.CreateAritcleRequest{
				Article: &pb.CreateAritcleRequest_Article{
					Title:       "",
					Description: "Test Description",
					Body:        "Test Body",
					TagList:     []string{"tag1", "tag2"},
				},
			},
			mockUserStore: func() {
				mock.ExpectQuery("SELECT * FROM `users` WHERE `users`.`id` = ?").
					WithArgs(1).
					WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(1))
			},
			mockArticleStore: func() {},
			expectedResp:     nil,
			expectedErr:      status.Error(codes.InvalidArgument, "validation error"),
		},
		{
			name: "Article Store Creation Failure",
			ctx:  context.WithValue(context.Background(), "user_id", uint(1)),
			req: &pb.CreateAritcleRequest{
				Article: &pb.CreateAritcleRequest_Article{
					Title:       "Test Article",
					Description: "Test Description",
					Body:        "Test Body",
					TagList:     []string{"tag1", "tag2"},
				},
			},
			mockUserStore: func() {
				mock.ExpectQuery("SELECT * FROM `users` WHERE `users`.`id` = ?").
					WithArgs(1).
					WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(1))
			},
			mockArticleStore: func() {
				mock.ExpectBegin()
				mock.ExpectExec("INSERT INTO `articles`").
					WithArgs(sqlmock.AnyArg(), "Test Article", "Test Description", "Test Body", uint(1), 0).
					WillReturnError(errors.New("creation failed"))
				mock.ExpectRollback()
			},
			expectedResp: nil,
			expectedErr:  status.Error(codes.Canceled, "Failed to create user."),
		},
		{
			name: "Following Status Retrieval Failure",
			ctx:  context.WithValue(context.Background(), "user_id", uint(1)),
			req: &pb.CreateAritcleRequest{
				Article: &pb.CreateAritcleRequest_Article{
					Title:       "Test Article",
					Description: "Test Description",
					Body:        "Test Body",
					TagList:     []string{"tag1", "tag2"},
				},
			},
			mockUserStore: func() {
				mock.ExpectQuery("SELECT * FROM `users` WHERE `users`.`id` = ?").
					WithArgs(1).
					WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(1))
				mock.ExpectQuery("SELECT * FROM `follows` WHERE `from_user_id` = ? AND `to_user_id` = ?").
					WithArgs(1, 1).
					WillReturnError(errors.New("following status retrieval failed"))
			},
			mockArticleStore: func() {
				mock.ExpectBegin()
				mock.ExpectExec("INSERT INTO `articles`").
					WithArgs(sqlmock.AnyArg(), "Test Article", "Test Description", "Test Body", uint(1), 0).
					WillReturnResult(sqlmock.NewResult(1, 1))
				mock.ExpectCommit()
			},
			expectedResp: nil,
			expectedErr:  status.Error(codes.NotFound, "internal server error"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			tt.mockUserStore()
			tt.mockArticleStore()

			resp, err := handler.CreateArticle(tt.ctx, tt.req)

			assert.Equal(t, tt.expectedResp, resp)
			assert.Equal(t, tt.expectedErr, err)

			if err := mock.ExpectationsWereMet(); err != nil {
				t.Errorf("there were unfulfilled expectations: %s", err)
			}
		})
	}
}


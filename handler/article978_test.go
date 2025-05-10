package handler

import (
	"context"
	"testing"
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
	us := &store.UserStore{DB: db}
	as := &store.ArticleStore{DB: db}
	handler := &Handler{logger: &logger, us: us, as: as}

	currentUser := &model.User{
		Model:    model.Model{ID: 1},
		Username: "testuser",
		Bio:      "test bio",
		Image:    "test image",
	}
	mock.ExpectQuery("^SELECT (.+) FROM `users` WHERE `users`.`id` = ? LIMIT 1$").
		WithArgs(currentUser.ID).
		WillReturnRows(sqlmock.NewRows([]string{"id", "username", "bio", "image"}).
			AddRow(currentUser.ID, currentUser.Username, currentUser.Bio, currentUser.Image))

	tests := []struct {
		name          string
		req           *pb.CreateAritcleRequest
		setupMocks    func()
		expectedError error
	}{
		{
			name: "Successfully Create an Article",
			req: &pb.CreateAritcleRequest{
				Article: &pb.CreateAritcleRequest_Article{
					Title:       "Test Title",
					Description: "Test Description",
					Body:        "Test Body",
					TagList:     []string{"tag1", "tag2"},
				},
			},
			setupMocks: func() {
				mock.ExpectBegin()
				mock.ExpectExec("^INSERT INTO `articles` (.+) VALUES (.+)$").
					WithArgs(sqlmock.AnyArg(), "Test Title", "Test Description", "Test Body", 1, sqlmock.AnyArg(), sqlmock.AnyArg()).
					WillReturnResult(sqlmock.NewResult(1, 1))
				mock.ExpectCommit()
				mock.ExpectQuery("^SELECT count(.+) FROM `follows` WHERE (.+)$").
					WithArgs(currentUser.ID, currentUser.ID).
					WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(0))
			},
			expectedError: nil,
		},
		{
			name: "Unauthenticated User",
			req: &pb.CreateAritcleRequest{
				Article: &pb.CreateAritcleRequest_Article{
					Title:       "Test Title",
					Description: "Test Description",
					Body:        "Test Body",
					TagList:     []string{"tag1", "tag2"},
				},
			},
			setupMocks: func() {
				auth.GetUserID = func(ctx context.Context) (uint, error) {
					return 0, status.Errorf(codes.Unauthenticated, "unauthenticated")
				}
			},
			expectedError: status.Errorf(codes.Unauthenticated, "unauthenticated"),
		},
		{
			name: "User Not Found",
			req: &pb.CreateAritcleRequest{
				Article: &pb.CreateAritcleRequest_Article{
					Title:       "Test Title",
					Description: "Test Description",
					Body:        "Test Body",
					TagList:     []string{"tag1", "tag2"},
				},
			},
			setupMocks: func() {
				mock.ExpectQuery("^SELECT (.+) FROM `users` WHERE `users`.`id` = ? LIMIT 1$").
					WithArgs(currentUser.ID).
					WillReturnError(status.Errorf(codes.NotFound, "user not found"))
			},
			expectedError: status.Errorf(codes.NotFound, "user not found"),
		},
		{
			name: "Validation Error in Article Data",
			req: &pb.CreateAritcleRequest{
				Article: &pb.CreateAritcleRequest_Article{
					Title:       "",
					Description: "Test Description",
					Body:        "Test Body",
					TagList:     []string{"tag1", "tag2"},
				},
			},
			setupMocks:    func() {},
			expectedError: status.Errorf(codes.InvalidArgument, "validation error"),
		},
		{
			name: "Database Error When Creating Article",
			req: &pb.CreateAritcleRequest{
				Article: &pb.CreateAritcleRequest_Article{
					Title:       "Test Title",
					Description: "Test Description",
					Body:        "Test Body",
					TagList:     []string{"tag1", "tag2"},
				},
			},
			setupMocks: func() {
				mock.ExpectBegin()
				mock.ExpectExec("^INSERT INTO `articles` (.+) VALUES (.+)$").
					WithArgs(sqlmock.AnyArg(), "Test Title", "Test Description", "Test Body", 1, sqlmock.AnyArg(), sqlmock.AnyArg()).
					WillReturnError(status.Errorf(codes.Canceled, "Failed to create user."))
				mock.ExpectRollback()
			},
			expectedError: status.Errorf(codes.Canceled, "Failed to create user."),
		},
		{
			name: "Error Getting Following Status",
			req: &pb.CreateAritcleRequest{
				Article: &pb.CreateAritcleRequest_Article{
					Title:       "Test Title",
					Description: "Test Description",
					Body:        "Test Body",
					TagList:     []string{"tag1", "tag2"},
				},
			},
			setupMocks: func() {
				mock.ExpectBegin()
				mock.ExpectExec("^INSERT INTO `articles` (.+) VALUES (.+)$").
					WithArgs(sqlmock.AnyArg(), "Test Title", "Test Description", "Test Body", 1, sqlmock.AnyArg(), sqlmock.AnyArg()).
					WillReturnResult(sqlmock.NewResult(1, 1))
				mock.ExpectCommit()
				mock.ExpectQuery("^SELECT count(.+) FROM `follows` WHERE (.+)$").
					WithArgs(currentUser.ID, currentUser.ID).
					WillReturnError(status.Errorf(codes.Internal, "failed to get following status"))
			},
			expectedError: status.Errorf(codes.Internal, "failed to get following status"),
		},
		{
			name: "Article Created with Tags",
			req: &pb.CreateAritcleRequest{
				Article: &pb.CreateAritcleRequest_Article{
					Title:       "Test Title",
					Description: "Test Description",
					Body:        "Test Body",
					TagList:     []string{"tag1", "tag2"},
				},
			},
			setupMocks: func() {
				mock.ExpectBegin()
				mock.ExpectExec("^INSERT INTO `articles` (.+) VALUES (.+)$").
					WithArgs(sqlmock.AnyArg(), "Test Title", "Test Description", "Test Body", 1, sqlmock.AnyArg(), sqlmock.AnyArg()).
					WillReturnResult(sqlmock.NewResult(1, 1))
				mock.ExpectCommit()
				mock.ExpectQuery("^SELECT count(.+) FROM `follows` WHERE (.+)$").
					WithArgs(currentUser.ID, currentUser.ID).
					WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(0))
			},
			expectedError: nil,
		},
		{
			name: "Article Favorited by Default",
			req: &pb.CreateAritcleRequest{
				Article: &pb.CreateAritcleRequest_Article{
					Title:       "Test Title",
					Description: "Test Description",
					Body:        "Test Body",
					TagList:     []string{"tag1", "tag2"},
				},
			},
			setupMocks: func() {
				mock.ExpectBegin()
				mock.ExpectExec("^INSERT INTO `articles` (.+) VALUES (.+)$").
					WithArgs(sqlmock.AnyArg(), "Test Title", "Test Description", "Test Body", 1, sqlmock.AnyArg(), sqlmock.AnyArg()).
					WillReturnResult(sqlmock.NewResult(1, 1))
				mock.ExpectCommit()
				mock.ExpectQuery("^SELECT count(.+) FROM `follows` WHERE (.+)$").
					WithArgs(currentUser.ID, currentUser.ID).
					WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(0))
			},
			expectedError: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.setupMocks()
			ctx := context.Background()

			resp, err := handler.CreateArticle(ctx, tt.req)
			if tt.expectedError != nil {
				assert.EqualError(t, err, tt.expectedError.Error())
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, resp)
				assert.Equal(t, tt.req.GetArticle().GetTitle(), resp.GetArticle().GetTitle())
				assert.Equal(t, tt.req.GetArticle().GetDescription(), resp.GetArticle().GetDescription())
				assert.Equal(t, tt.req.GetArticle().GetBody(), resp.GetArticle().GetBody())
				assert.Equal(t, tt.req.GetArticle().GetTagList(), resp.GetArticle().GetTagList())
				assert.True(t, resp.GetArticle().GetFavorited())
			}
		})
	}
}


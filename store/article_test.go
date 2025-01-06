package store

import (
	"errors"
	"testing"
	"github.com/DATA-DOG/go-sqlmock"
	"github.com/jinzhu/gorm"
	"github.com/jinzhu/gorm/dialects/sqlite"
	"github.com/raahii/golang-grpc-realworld-example/model"
	"github.com/stretchr/testify/assert"
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
type T struct {
	common
	isEnvSet bool
	context  *testContext
}


/*
ROOST_METHOD_HASH=Create_0a911e138d
ROOST_METHOD_SIG_HASH=Create_723c594377


 */
func TestArticleStoreCreate(t *testing.T) {

	newArticleStoreWithMockDB := func() (*ArticleStore, sqlmock.Sqlmock, func()) {
		db, mock, err := sqlmock.New()
		if err != nil {
			t.Fatalf("failed to create mock db: %v", err)
		}

		gormDB, err := gorm.Open("sqlite3", db)
		if err != nil {
			t.Fatalf("failed to open gorm db: %v", err)
		}
		return &ArticleStore{db: gormDB}, mock, func() { db.Close() }
	}

	t.Run("Successfully Create a New Article", func(t *testing.T) {
		store, mock, cleanup := newArticleStoreWithMockDB()
		defer cleanup()

		article := &model.Article{
			Title:       "Test Title",
			Description: "Test Description",
			Body:        "Test Body",
			UserID:      1,
		}

		mock.ExpectBegin()
		mock.ExpectExec(`INSERT INTO "articles"`).
			WithArgs(sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg()).
			WillReturnResult(sqlmock.NewResult(1, 1))
		mock.ExpectCommit()

		err := store.Create(article)
		assert.NoError(t, err)
		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("Fail to Create Article Due to Missing Required Fields", func(t *testing.T) {
		store, mock, cleanup := newArticleStoreWithMockDB()
		defer cleanup()

		article := &model.Article{
			Description: "Test Description",
			Body:        "Test Body",
			UserID:      1,
		}

		err := store.Create(article)
		assert.Error(t, err)
		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("Database Connection Error", func(t *testing.T) {
		store, mock, cleanup := newArticleStoreWithMockDB()
		defer cleanup()

		article := &model.Article{
			Title:       "Test Title",
			Description: "Test Description",
			Body:        "Test Body",
			UserID:      1,
		}

		mock.ExpectBegin().WillReturnError(errors.New("connection error"))

		err := store.Create(article)
		assert.Error(t, err)
		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("Create Article with Tags", func(t *testing.T) {
		store, mock, cleanup := newArticleStoreWithMockDB()
		defer cleanup()

		article := &model.Article{
			Title:       "Test Title",
			Description: "Test Description",
			Body:        "Test Body",
			UserID:      1,
			Tags:        []model.Tag{{Name: "Tag1"}, {Name: "Tag2"}},
		}

		mock.ExpectBegin()
		mock.ExpectExec(`INSERT INTO "articles"`).
			WithArgs(sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg()).
			WillReturnResult(sqlmock.NewResult(1, 1))
		mock.ExpectExec(`INSERT INTO "tags"`).
			WithArgs(sqlmock.AnyArg(), "Tag1").
			WillReturnResult(sqlmock.NewResult(1, 1))
		mock.ExpectExec(`INSERT INTO "tags"`).
			WithArgs(sqlmock.AnyArg(), "Tag2").
			WillReturnResult(sqlmock.NewResult(2, 1))
		mock.ExpectExec(`INSERT INTO "article_tags"`).
			WithArgs(sqlmock.AnyArg(), sqlmock.AnyArg()).
			WillReturnResult(sqlmock.NewResult(1, 1))
		mock.ExpectCommit()

		err := store.Create(article)
		assert.NoError(t, err)
		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("Create Article with Favorited Users", func(t *testing.T) {
		store, mock, cleanup := newArticleStoreWithMockDB()
		defer cleanup()

		article := &model.Article{
			Title:       "Test Title",
			Description: "Test Description",
			Body:        "Test Body",
			UserID:      1,
			FavoritedUsers: []model.User{
				{Username: "user1", Email: "user1@example.com"},
				{Username: "user2", Email: "user2@example.com"},
			},
		}

		mock.ExpectBegin()
		mock.ExpectExec(`INSERT INTO "articles"`).
			WithArgs(sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg()).
			WillReturnResult(sqlmock.NewResult(1, 1))
		mock.ExpectExec(`INSERT INTO "users"`).
			WithArgs(sqlmock.AnyArg(), "user1", "user1@example.com").
			WillReturnResult(sqlmock.NewResult(1, 1))
		mock.ExpectExec(`INSERT INTO "users"`).
			WithArgs(sqlmock.AnyArg(), "user2", "user2@example.com").
			WillReturnResult(sqlmock.NewResult(2, 1))
		mock.ExpectExec(`INSERT INTO "favorite_articles"`).
			WithArgs(sqlmock.AnyArg(), sqlmock.AnyArg()).
			WillReturnResult(sqlmock.NewResult(1, 1))
		mock.ExpectCommit()

		err := store.Create(article)
		assert.NoError(t, err)
		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("Create Article with Comments", func(t *testing.T) {
		store, mock, cleanup := newArticleStoreWithMockDB()
		defer cleanup()

		article := &model.Article{
			Title:       "Test Title",
			Description: "Test Description",
			Body:        "Test Body",
			UserID:      1,
			Comments: []model.Comment{
				{Body: "This is a comment", UserID: 1},
				{Body: "This is another comment", UserID: 2},
			},
		}

		mock.ExpectBegin()
		mock.ExpectExec(`INSERT INTO "articles"`).
			WithArgs(sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg()).
			WillReturnResult(sqlmock.NewResult(1, 1))
		mock.ExpectExec(`INSERT INTO "comments"`).
			WithArgs(sqlmock.AnyArg(), "This is a comment", 1).
			WillReturnResult(sqlmock.NewResult(1, 1))
		mock.ExpectExec(`INSERT INTO "comments"`).
			WithArgs(sqlmock.AnyArg(), "This is another comment", 2).
			WillReturnResult(sqlmock.NewResult(2, 1))
		mock.ExpectCommit()

		err := store.Create(article)
		assert.NoError(t, err)
		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("Fail to Create Article Due to Duplicate Title", func(t *testing.T) {
		store, mock, cleanup := newArticleStoreWithMockDB()
		defer cleanup()

		article := &model.Article{
			Title:       "Duplicate Title",
			Description: "Test Description",
			Body:        "Test Body",
			UserID:      1,
		}

		mock.ExpectBegin()
		mock.ExpectExec(`INSERT INTO "articles"`).
			WithArgs(sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg()).
			WillReturnError(errors.New("duplicate title"))
		mock.ExpectRollback()

		err := store.Create(article)
		assert.Error(t, err)
		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("Create Article with All Fields Populated", func(t *testing.T) {
		store, mock, cleanup := newArticleStoreWithMockDB()
		defer cleanup()

		article := &model.Article{
			Title:       "Complete Article",
			Description: "Complete Description",
			Body:        "Complete Body",
			UserID:      1,
			Tags:        []model.Tag{{Name: "Tag1"}, {Name: "Tag2"}},
			FavoritedUsers: []model.User{
				{Username: "user1", Email: "user1@example.com"},
				{Username: "user2", Email: "user2@example.com"},
			},
			Comments: []model.Comment{
				{Body: "This is a comment", UserID: 1},
				{Body: "This is another comment", UserID: 2},
			},
		}

		mock.ExpectBegin()
		mock.ExpectExec(`INSERT INTO "articles"`).
			WithArgs(sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg()).
			WillReturnResult(sqlmock.NewResult(1, 1))
		mock.ExpectExec(`INSERT INTO "tags"`).
			WithArgs(sqlmock.AnyArg(), "Tag1").
			WillReturnResult(sqlmock.NewResult(1, 1))
		mock.ExpectExec(`INSERT INTO "tags"`).
			WithArgs(sqlmock.AnyArg(), "Tag2").
			WillReturnResult(sqlmock.NewResult(2, 1))
		mock.ExpectExec(`INSERT INTO "article_tags"`).
			WithArgs(sqlmock.AnyArg(), sqlmock.AnyArg()).
			WillReturnResult(sqlmock.NewResult(1, 1))
		mock.ExpectExec(`INSERT INTO "users"`).
			WithArgs(sqlmock.AnyArg(), "user1", "user1@example.com").
			WillReturnResult(sqlmock.NewResult(1, 1))
		mock.ExpectExec(`INSERT INTO "users"`).
			WithArgs(sqlmock.AnyArg(), "user2", "user2@example.com").
			WillReturnResult(sqlmock.NewResult(2, 1))
		mock.ExpectExec(`INSERT INTO "favorite_articles"`).
			WithArgs(sqlmock.AnyArg(), sqlmock.AnyArg()).
			WillReturnResult(sqlmock.NewResult(1, 1))
		mock.ExpectExec(`INSERT INTO "comments"`).
			WithArgs(sqlmock.AnyArg(), "This is a comment", 1).
			WillReturnResult(sqlmock.NewResult(1, 1))
		mock.ExpectExec(`INSERT INTO "comments"`).
			WithArgs(sqlmock.AnyArg(), "This is another comment", 2).
			WillReturnResult(sqlmock.NewResult(2, 1))
		mock.ExpectCommit()

		err := store.Create(article)
		assert.NoError(t, err)
		assert.NoError(t, mock.ExpectationsWereMet())
	})
}


/*
ROOST_METHOD_HASH=CreateComment_58d394e2c6
ROOST_METHOD_SIG_HASH=CreateComment_28b95f60a6


 */
func TestArticleStoreCreateComment(t *testing.T) {
	type test struct {
		name      string
		comment   model.Comment
		mockSetup func(sqlmock.Sqlmock)
		wantErr   bool
		errMsg    string
	}

	tests := []test{
		{
			name: "Successfully Create a Comment",
			comment: model.Comment{
				Body:      "This is a valid comment",
				UserID:    1,
				ArticleID: 1,
			},
			mockSetup: func(mock sqlmock.Sqlmock) {
				mock.ExpectBegin()
				mock.ExpectExec("INSERT INTO \"comments\"").WithArgs(sqlmock.AnyArg(), "This is a valid comment", 1, 1, sqlmock.AnyArg(), sqlmock.AnyArg()).
					WillReturnResult(sqlmock.NewResult(1, 1))
				mock.ExpectCommit()
			},
			wantErr: false,
			errMsg:  "",
		},
		{
			name: "Fail to Create a Comment with Empty Body",
			comment: model.Comment{
				Body:      "",
				UserID:    1,
				ArticleID: 1,
			},
			mockSetup: func(mock sqlmock.Sqlmock) {
				mock.ExpectBegin()
				mock.ExpectExec("INSERT INTO \"comments\"").WithArgs(sqlmock.AnyArg(), "", 1, 1, sqlmock.AnyArg(), sqlmock.AnyArg()).
					WillReturnError(errors.New("Body cannot be empty"))
				mock.ExpectRollback()
			},
			wantErr: true,
			errMsg:  "Body cannot be empty",
		},
		{
			name: "Fail to Create a Comment with Missing UserID",
			comment: model.Comment{
				Body:      "This is a valid comment",
				UserID:    0,
				ArticleID: 1,
			},
			mockSetup: func(mock sqlmock.Sqlmock) {
				mock.ExpectBegin()
				mock.ExpectExec("INSERT INTO \"comments\"").WithArgs(sqlmock.AnyArg(), "This is a valid comment", 0, 1, sqlmock.AnyArg(), sqlmock.AnyArg()).
					WillReturnError(errors.New("UserID cannot be zero"))
				mock.ExpectRollback()
			},
			wantErr: true,
			errMsg:  "UserID cannot be zero",
		},
		{
			name: "Fail to Create a Comment with Missing ArticleID",
			comment: model.Comment{
				Body:      "This is a valid comment",
				UserID:    1,
				ArticleID: 0,
			},
			mockSetup: func(mock sqlmock.Sqlmock) {
				mock.ExpectBegin()
				mock.ExpectExec("INSERT INTO \"comments\"").WithArgs(sqlmock.AnyArg(), "This is a valid comment", 1, 0, sqlmock.AnyArg(), sqlmock.AnyArg()).
					WillReturnError(errors.New("ArticleID cannot be zero"))
				mock.ExpectRollback()
			},
			wantErr: true,
			errMsg:  "ArticleID cannot be zero",
		},
		{
			name: "Database Connection Failure",
			comment: model.Comment{
				Body:      "This is a valid comment",
				UserID:    1,
				ArticleID: 1,
			},
			mockSetup: func(mock sqlmock.Sqlmock) {
				mock.ExpectBegin().WillReturnError(errors.New("database connection error"))
			},
			wantErr: true,
			errMsg:  "database connection error",
		},
		{
			name: "Successfully Create Multiple Comments",
			comment: model.Comment{
				Body:      "This is a valid comment",
				UserID:    1,
				ArticleID: 1,
			},
			mockSetup: func(mock sqlmock.Sqlmock) {
				for i := 0; i < 5; i++ {
					mock.ExpectBegin()
					mock.ExpectExec("INSERT INTO \"comments\"").WithArgs(sqlmock.AnyArg(), "This is a valid comment", 1, 1, sqlmock.AnyArg(), sqlmock.AnyArg()).
						WillReturnResult(sqlmock.NewResult(1, 1))
					mock.ExpectCommit()
				}
			},
			wantErr: false,
			errMsg:  "",
		},
		{
			name: "Duplicate Comment Creation",
			comment: model.Comment{
				Body:      "This is a duplicate comment",
				UserID:    1,
				ArticleID: 1,
			},
			mockSetup: func(mock sqlmock.Sqlmock) {
				mock.ExpectBegin()
				mock.ExpectExec("INSERT INTO \"comments\"").WithArgs(sqlmock.AnyArg(), "This is a duplicate comment", 1, 1, sqlmock.AnyArg(), sqlmock.AnyArg()).
					WillReturnResult(sqlmock.NewResult(1, 1))
				mock.ExpectCommit()
				mock.ExpectBegin()
				mock.ExpectExec("INSERT INTO \"comments\"").WithArgs(sqlmock.AnyArg(), "This is a duplicate comment", 1, 1, sqlmock.AnyArg(), sqlmock.AnyArg()).
					WillReturnError(errors.New("duplicate comment"))
				mock.ExpectRollback()
			},
			wantErr: true,
			errMsg:  "duplicate comment",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			db, mock, err := sqlmock.New()
			if err != nil {
				t.Fatalf("failed to open sqlmock database: %v", err)
			}
			defer db.Close()

			gormDB, err := gorm.Open("postgres", db)
			if err != nil {
				t.Fatalf("failed to open gorm database: %v", err)
			}

			store := &ArticleStore{db: gormDB}
			tt.mockSetup(mock)

			err = store.CreateComment(&tt.comment)
			if tt.wantErr {
				assert.Error(t, err)
				assert.Equal(t, tt.errMsg, err.Error())
			} else {
				assert.NoError(t, err)
			}

			if err := mock.ExpectationsWereMet(); err != nil {
				t.Errorf("there were unfulfilled expectations: %v", err)
			}
		})
	}
}


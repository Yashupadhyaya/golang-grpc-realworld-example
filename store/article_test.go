package store

import (
	errors "errors"
	fmt "fmt"
	debug "runtime/debug"
	testing "testing"
	time "time"

	"github.com/DATA-DOG/go-sqlmock"
	gorm "github.com/jinzhu/gorm"
	model "github.com/raahii/golang-grpc-realworld-example/model"
)

type Comment struct {
	ID        uint   `gorm:"primaryKey"`
	ArticleID uint   `gorm:"not null"`
	Body      string `gorm:"type:text;not null"`
	CreatedAt time.Time
	UpdatedAt time.Time
}

/*
ROOST_METHOD_HASH=ArticleStore_CreateComment_b16d4a71d4
ROOST_METHOD_SIG_HASH=ArticleStore_CreateComment_7475736b06

FUNCTION_DEF=func (s *ArticleStore) CreateComment(m *model.Comment) error // CreateComment creates a comment of the article
*/
func TestArticleStoreCreateComment(t *testing.T) {

	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
	}
	defer db.Close()

	gormDB, err := gorm.Open("mysql", db)
	if err != nil {
		t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
	}
	defer gormDB.Close()

	articleStore := &ArticleStore{db: gormDB}

	tests := []struct {
		name    string
		comment *model.Comment
		mock    func()
		wantErr bool
	}{
		{
			name: "Valid Comment",
			comment: &model.Comment{
				ArticleID: 1,
				Body:      "This is a valid comment",
			},
			mock: func() {
				mock.ExpectBegin()
				mock.ExpectExec("INSERT INTO `comments` .*").WillReturnResult(sqlmock.NewResult(1, 1))
				mock.ExpectCommit()
			},
			wantErr: false,
		},
		{
			name: "Invalid Comment",
			comment: &model.Comment{
				ArticleID: 1,
				Body:      "",
			},
			mock: func() {
				mock.ExpectBegin()
				mock.ExpectExec("INSERT INTO `comments` .*").WillReturnError(errors.New("validation failed"))
				mock.ExpectRollback()
			},
			wantErr: true,
		},
		{
			name: "Database Error",
			comment: &model.Comment{
				ArticleID: 1,
				Body:      "This is a valid comment",
			},
			mock: func() {
				mock.ExpectBegin()
				mock.ExpectExec("INSERT INTO `comments` .*").WillReturnError(errors.New("database error"))
				mock.ExpectRollback()
			},
			wantErr: true,
		},
		{
			name: "Duplicate Comment",
			comment: &model.Comment{
				ArticleID: 1,
				Body:      "This is a duplicate comment",
			},
			mock: func() {
				mock.ExpectBegin()
				mock.ExpectExec("INSERT INTO `comments` .*").WillReturnError(errors.New("duplicate entry"))
				mock.ExpectRollback()
			},
			wantErr: true,
		},
		{
			name:    "Empty Comment",
			comment: &model.Comment{},
			mock: func() {
				mock.ExpectBegin()
				mock.ExpectExec("INSERT INTO `comments` .*").WillReturnError(errors.New("validation failed"))
				mock.ExpectRollback()
			},
			wantErr: true,
		},
		{
			name: "Maximum Length Comment",
			comment: &model.Comment{
				ArticleID: 1,
				Body:      string(make([]byte, 65535)),
			},
			mock: func() {
				mock.ExpectBegin()
				mock.ExpectExec("INSERT INTO `comments` .*").WillReturnResult(sqlmock.NewResult(1, 1))
				mock.ExpectCommit()
			},
			wantErr: false,
		},
		{
			name: "Minimum Length Comment",
			comment: &model.Comment{
				ArticleID: 1,
				Body:      "A",
			},
			mock: func() {
				mock.ExpectBegin()
				mock.ExpectExec("INSERT INTO `comments` .*").WillReturnResult(sqlmock.NewResult(1, 1))
				mock.ExpectCommit()
			},
			wantErr: false,
		},
		{
			name: "Special Characters in Comment",
			comment: &model.Comment{
				ArticleID: 1,
				Body:      "This is a comment with special characters: !@#$%^&*()",
			},
			mock: func() {
				mock.ExpectBegin()
				mock.ExpectExec("INSERT INTO `comments` .*").WillReturnResult(sqlmock.NewResult(1, 1))
				mock.ExpectCommit()
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			defer func() {
				if r := recover(); r != nil {
					t.Logf("Panic encountered so failing test. %v\n%s", r, string(debug.Stack()))
					t.Fail()
				}
			}()

			tt.mock()
			err := articleStore.CreateComment(tt.comment)
			if (err != nil) != tt.wantErr {
				t.Errorf("CreateComment() error = %v, wantErr %v", err, tt.wantErr)
			}

			if err := mock.ExpectationsWereMet(); err != nil {
				t.Errorf("there were unfulfilled expectations: %s", err)
			}
		})
	}

	t.Run("Concurrent Requests", func(t *testing.T) {
		defer func() {
			if r := recover(); r != nil {
				t.Logf("Panic encountered so failing test. %v\n%s", r, string(debug.Stack()))
				t.Fail()
			}
		}()

		numConcurrentRequests := 10
		errorsChan := make(chan error, numConcurrentRequests)

		for i := 0; i < numConcurrentRequests; i++ {
			comment := &model.Comment{
				ArticleID: uint(i + 1),
				Body:      fmt.Sprintf("This is comment %d", i+1),
			}
			go func() {
				mock.ExpectBegin()
				mock.ExpectExec("INSERT INTO `comments` .*").WillReturnResult(sqlmock.NewResult(1, 1))
				mock.ExpectCommit()
				err := articleStore.CreateComment(comment)
				errorsChan <- err
			}()
		}

		for i := 0; i < numConcurrentRequests; i++ {
			err := <-errorsChan
			if err != nil {
				t.Errorf("CreateComment() error = %v", err)
			}
		}

		if err := mock.ExpectationsWereMet(); err != nil {
			t.Errorf("there were unfulfilled expectations: %s", err)
		}
	})

	t.Run("Large Number of Comments", func(t *testing.T) {
		defer func() {
			if r := recover(); r != nil {
				t.Logf("Panic encountered so failing test. %v\n%s", r, string(debug.Stack()))
				t.Fail()
			}
		}()

		numComments := 1000
		for i := 0; i < numComments; i++ {
			comment := &model.Comment{
				ArticleID: uint(i + 1),
				Body:      fmt.Sprintf("This is comment %d", i+1),
			}
			mock.ExpectBegin()
			mock.ExpectExec("INSERT INTO `comments` .*").WillReturnResult(sqlmock.NewResult(1, 1))
			mock.ExpectCommit()
			err := articleStore.CreateComment(comment)
			if err != nil {
				t.Errorf("CreateComment() error = %v", err)
			}
		}

		if err := mock.ExpectationsWereMet(); err != nil {
			t.Errorf("there were unfulfilled expectations: %s", err)
		}
	})
}

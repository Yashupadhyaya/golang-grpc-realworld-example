package store

import (
	"errors"
	"testing"
	sqlmock "github.com/DATA-DOG/go-sqlmock"
	"github.com/jinzhu/gorm"
	"github.com/raahii/golang-grpc-realworld-example/model"
	"github.com/stretchr/testify/assert"
	_ "github.com/jinzhu/gorm/dialects/mysql"
)


/*
ROOST_METHOD_HASH=Create_0a911e138d
ROOST_METHOD_SIG_HASH=Create_723c594377


 */
func TestCreate(t *testing.T) {

	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("failed to open mock sql db, got error: %v", err)
	}
	defer db.Close()

	gormDB, err := gorm.Open("mysql", db)
	if err != nil {
		t.Fatalf("failed to open gorm db, got error: %v", err)
	}

	articleStore := &ArticleStore{db: gormDB}

	t.Run("Successfully Create a New Article", func(t *testing.T) {
		article := model.Article{Title: "New Article", Body: "This is a new article"}
		mock.ExpectBegin()
		mock.ExpectExec("INSERT INTO `articles`").WithArgs(article.Title, article.Body).WillReturnResult(sqlmock.NewResult(1, 1))
		mock.ExpectCommit()

		err := articleStore.Create(&article)
		assert.NoError(t, err, "expected no error, but got one")
		assert.Nil(t, mock.ExpectationsWereMet(), "there were unfulfilled expectations")
		t.Log("Successfully created a new article")
	})

	t.Run("Fail to Create an Article with Missing Required Fields", func(t *testing.T) {
		article := model.Article{Title: "", Body: ""}
		mock.ExpectBegin()
		mock.ExpectExec("INSERT INTO `articles`").WithArgs(article.Title, article.Body).WillReturnError(errors.New("missing required fields"))
		mock.ExpectRollback()

		err := articleStore.Create(&article)
		assert.Error(t, err, "expected an error, but got none")
		assert.Contains(t, err.Error(), "missing required fields", "expected error about missing fields")
		assert.Nil(t, mock.ExpectationsWereMet(), "there were unfulfilled expectations")
		t.Log("Failed to create an article due to missing required fields")
	})

	t.Run("Handle Database Connection Error", func(t *testing.T) {
		article := model.Article{Title: "New Article", Body: "This is a new article"}
		mock.ExpectBegin().WillReturnError(errors.New("database connection error"))

		err := articleStore.Create(&article)
		assert.Error(t, err, "expected an error, but got none")
		assert.Contains(t, err.Error(), "database connection error", "expected error about database connection")
		assert.Nil(t, mock.ExpectationsWereMet(), "there were unfulfilled expectations")
		t.Log("Handled database connection error gracefully")
	})

	t.Run("Create Article with Special Characters in Fields", func(t *testing.T) {
		article := model.Article{Title: "New @rt!cle", Body: "This is a new @rt!cle with special characters"}
		mock.ExpectBegin()
		mock.ExpectExec("INSERT INTO `articles`").WithArgs(article.Title, article.Body).WillReturnResult(sqlmock.NewResult(1, 1))
		mock.ExpectCommit()

		err := articleStore.Create(&article)
		assert.NoError(t, err, "expected no error, but got one")
		assert.Nil(t, mock.ExpectationsWereMet(), "there were unfulfilled expectations")
		t.Log("Successfully created an article with special characters in fields")
	})

	t.Run("Attempt to Create a Duplicate Article", func(t *testing.T) {
		article := model.Article{Title: "Duplicate Article", Body: "This is a duplicate article"}
		mock.ExpectBegin()
		mock.ExpectExec("INSERT INTO `articles`").WithArgs(article.Title, article.Body).WillReturnError(errors.New("duplicate entry"))
		mock.ExpectRollback()

		err := articleStore.Create(&article)
		assert.Error(t, err, "expected an error, but got none")
		assert.Contains(t, err.Error(), "duplicate entry", "expected error about duplicate entry")
		assert.Nil(t, mock.ExpectationsWereMet(), "there were unfulfilled expectations")
		t.Log("Failed to create a duplicate article")
	})

	t.Run("Create Article with Maximum Field Lengths", func(t *testing.T) {
		article := model.Article{
			Title: "This is a very long title that exceeds the usual length for titles but should be handled correctly by the function",
			Body:  "This is a very long body that also exceeds the usual length for bodies but should be handled correctly by the function",
		}
		mock.ExpectBegin()
		mock.ExpectExec("INSERT INTO `articles`").WithArgs(article.Title, article.Body).WillReturnResult(sqlmock.NewResult(1, 1))
		mock.ExpectCommit()

		err := articleStore.Create(&article)
		assert.NoError(t, err, "expected no error, but got one")
		assert.Nil(t, mock.ExpectationsWereMet(), "there were unfulfilled expectations")
		t.Log("Successfully created an article with maximum field lengths")
	})
}


package store

import (
	fmt "fmt"
	testing "testing"

	sqlmock "github.com/DATA-DOG/go-sqlmock"
	gorm "github.com/jinzhu/gorm"
	model "github.com/raahii/golang-grpc-realworld-example/model"
)

type MockArticleStore struct {
	db *gorm.DB
}

/*
ROOST_METHOD_HASH=ArticleStore_Create_1273475ade
ROOST_METHOD_SIG_HASH=ArticleStore_Create_a27282cad5

FUNCTION_DEF=func (s *ArticleStore) Create(m *model.Article) error // Create creates an article
*/
func TestArticleStoreCreate(t *testing.T) {
	testCases := []struct {
		testName    string
		article     model.Article
		mockError   error
		expectError bool
	}{
		{
			testName: "Successful creation of an article",
			article: model.Article{
				Title:       "Test Title",
				Description: "Test Description",
				Body:        "Test Body",
			},
			mockError:   nil,
			expectError: false,
		},
		{
			testName: "Failed creation of an article due to database error",
			article: model.Article{
				Title:       "Test Title",
				Description: "Test Description",
				Body:        "Test Body",
			},
			mockError:   fmt.Errorf("database error"),
			expectError: true,
		},
		{
			testName: "Attempt to create an article with invalid data",
			article: model.Article{
				Title: "",
			},
			mockError:   nil,
			expectError: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.testName, func(t *testing.T) {
			defer func() {
				if r := recover(); r != nil {
					t.Logf("Panic encountered so failing test. %v", r)
					t.Fail()
				}
			}()

			dbMock, mock, _ := sqlmock.New()
			gdb, _ := gorm.Open("postgres", dbMock)
			mock.ExpectBegin()
			mock.ExpectQuery(".").
				WillReturnError(tc.mockError)
			mock.ExpectCommit()

			store := &ArticleStore{
				db: gdb,
			}

			err := store.Create(&tc.article)

			if tc.expectError {
				if err == nil {
					t.Errorf("Expected error, got nil")
				} else {
					t.Logf("Expected error occurred: %v", err)
				}
			} else {
				if err != nil {
					t.Errorf("Unexpected error occurred: %v", err)
				} else {
					t.Logf("No error occurred as expected")
				}
			}

			gdb.Close()
		})
	}
}

/*
ROOST_METHOD_HASH=ArticleStore_DeleteFavorite_29c18a04a8
ROOST_METHOD_SIG_HASH=ArticleStore_DeleteFavorite_53deb5e792

FUNCTION_DEF=func (s *ArticleStore) DeleteFavorite(a *model.Article, u *model.User) error // DeleteFavorite unfavorite an article
*/
func (s *MockArticleStore) DeleteFavorite(a *model.Article, u *model.User) error {
	tx := s.db.Begin()
	err := tx.Model(a).Association("FavoritedUsers").Delete(u).Error
	if err != nil {
		tx.Rollback()
		return err
	}
	err = tx.Model(a).Update("favorites_count", gorm.Expr("favorites_count - ?", 1)).Error
	if err != nil {
		tx.Rollback()
		return err
	}
	tx.Commit()
	a.FavoritesCount--
	return nil
}

func TestArticleStoreDeleteFavorite(t *testing.T) {

	data := &model.Article{}

	userData := &model.User{}

	db, err := gorm.Open("sqlite3", ":memory:")
	if err != nil {
		t.Fail()
	}
	db.AutoMigrate(&model.Article{}, &model.User{})

	store := &MockArticleStore{db}

	err = store.DeleteFavorite(data, userData)

	if err != nil {
		t.Fail()
	}

	if data.FavoritesCount != -1 {
		t.Fail()
	}
}

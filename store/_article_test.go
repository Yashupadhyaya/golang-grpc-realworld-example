package store

import (
	errors "errors"
	fmt "fmt"
	testing "testing"
	gosqlmock "github.com/DATA-DOG/go-sqlmock"
	gorm "github.com/jinzhu/gorm"
	model "github.com/raahii/golang-grpc-realworld-example/model"
)








/*
ROOST_METHOD_HASH=ArticleStore_DeleteFavorite_29c18a04a8
ROOST_METHOD_SIG_HASH=ArticleStore_DeleteFavorite_53deb5e792

FUNCTION_DEF=func (s *ArticleStore) DeleteFavorite(a *model.Article, u *model.User) error // DeleteFavorite unfavorite an article


*/
func TestArticleStoreDeleteFavorite(t *testing.T) {

	var tests = []struct {
		name           string
		user           *model.User
		article        *model.Article
		expectedErr    error
		deleteErr      error
		updateErr      error
		expectedCounts int
	}{
		{
			"Unfavoriting an Article Successfully",
			&model.User{ID: 1},
			&model.Article{ID: 1, FavoritesCount: 1},
			nil,
			nil,
			nil,
			0,
		},
		{
			"Failed to Unfavorite due to Association Deletion Error",
			&model.User{ID: 1},
			&model.Article{ID: 1, FavoritesCount: 1},
			gorm.ErrRecordNotFound,
			gorm.ErrRecordNotFound,
			nil,
			1,
		},
		{
			"Failed to Unfavorite due to Update Error",
			&model.User{ID: 1},
			&model.Article{ID: 1, FavoritesCount: 1},
			gorm.ErrRecordNotFound,
			nil,
			gorm.ErrRecordNotFound,
			1,
		},
		{
			"Deleting Favorites when Article or User is nil",
			nil,
			nil,
			errors.New("Article or User should not be nil"),
			nil,
			nil,
			0,
		},
	}

	db, mock, _ := sqlmock.New(sqlmock.QueryMatcherOption(sqlmock.QueryMatcherRegexp))
	defer db.Close()

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			defer func() {
				if r := recover(); r != nil {
					t.Logf("Panic encountered so failing test. %v", r)
					t.Fail()
				}
			}()

			gdb, _ := gorm.Open("postgres", db)
			store := ArticleStore{db: gdb}

			mock.ExpectBegin()
			mock.ExpectExec("^DELETE FROM").WillReturnResult(sqlmock.NewResult(1, 1))
			if test.deleteErr != nil {
				mock.ExpectRollback()
			}

			if test.updateErr == nil && test.expectedErr == nil {
				mock.ExpectExec("^UPDATE").WillReturnResult(sqlmock.NewResult(1, 1))
				mock.ExpectCommit()
			} else {
				mock.ExpectRollback()
			}

			err := store.DeleteFavorite(test.article, test.user)

			if err != nil && test.expectedErr == nil {
				t.Errorf("unexpected error: %v", err)
			}

			if test.expectedErr != nil && err.Error() != test.expectedErr.Error() {
				t.Errorf("expected error: %v, but got: %v", test.expectedErr, err)
			}

			if test.article != nil && test.article.FavoritesCount != test.expectedCounts {
				t.Errorf("expected favorites count: %d, but got %d", test.expectedCounts, test.article.FavoritesCount)
			}
		})
	}
}


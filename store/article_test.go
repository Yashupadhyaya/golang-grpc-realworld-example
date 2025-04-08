package store

import (
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/jinzhu/gorm"
	"github.com/raahii/golang-grpc-realworld-example/model"
)

var cases = []struct {
	name                   string
	expectedFavoritesCount int32
	expectedError          error
}{
	{
		name:                   "Scenario 1: Normal operation - Deleting favorite article",
		expectedFavoritesCount: 5,
		expectedError:          nil,
	},
	{
		name:                   "Scenario 2: Edge case - Deleting non-existent user's favourite",
		expectedFavoritesCount: 0,
		expectedError:          gorm.ErrRecordNotFound,
	},
	{
		name:                   "Scenario 3: Edge case - Decreasing count of zero favourites",
		expectedFavoritesCount: 0,
		expectedError:          gorm.ErrRecordNotFound,
	},
}

type mockArticleStore struct {
	db *gorm.DB
}

/*
ROOST_METHOD_HASH=ArticleStore_DeleteFavorite_29c18a04a8
ROOST_METHOD_SIG_HASH=ArticleStore_DeleteFavorite_53deb5e792

FUNCTION_DEF=func (s *ArticleStore) DeleteFavorite(a *model.Article, u *model.User) error // DeleteFavorite unfavorite an article
*/
func TestArticleStoreDeleteFavorite(t *testing.T) {
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			defer func() {
				if r := recover(); r != nil {
					t.Logf("Panic encountered so failing test. %v", r)
					t.Fail()
				}
			}()

			mockSqlDB, _, _ := sqlmock.New()
			mockGormDB, _ := gorm.Open("postgres", mockSqlDB)
			mockGormDB.LogMode(false)

			mockArticleStore := &mockArticleStore{
				db: mockGormDB,
			}

			mockArticle := &model.Article{
				FavoritesCount: tc.expectedFavoritesCount,
			}
			mockUser := &model.User{}

			err := mockArticleStore.DeleteFavorite(mockArticle, mockUser)

			if err != tc.expectedError {
				t.Errorf("error got: %v, wanted: %v", err, tc.expectedError)
			}

			if mockArticle.FavoritesCount != tc.expectedFavoritesCount {
				t.Errorf("favourites count got: %d, wanted: %d", mockArticle.FavoritesCount, tc.expectedFavoritesCount)
			}
		})
	}
}

func (m *mockArticleStore) DeleteFavorite(a *model.Article, u *model.User) error {
	return nil
}

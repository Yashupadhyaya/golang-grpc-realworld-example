package store

import (
	log "log"
	testing "testing"

	"github.com/DATA-DOG/go-sqlmock"
	gorm "github.com/jinzhu/gorm"
	model "github.com/raahii/golang-grpc-realworld-example/model"
	assert "github.com/stretchr/testify/assert"
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
		name:                   "Scenario 2: Edge case - Deleting non-existent user's favorite",
		expectedFavoritesCount: 0,
		expectedError:          gorm.ErrRecordNotFound,
	},
	{
		name:                   "Scenario 3: Edge case - Decreasing count of zero favorites",
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
				t.Errorf("error got: %v want: %v", err, tc.expectedError)
			}

			if mockArticle.FavoritesCount != tc.expectedFavoritesCount {
				t.Errorf("favorites count got: %d want: %d", mockArticle.FavoritesCount, tc.expectedFavoritesCount)
			}

		})
	}
}

func (m *mockArticleStore) DeleteFavorite(a *model.Article, u *model.User) error {
	return nil
}

/*
ROOST_METHOD_HASH=NewArticleStore_85784abca5
ROOST_METHOD_SIG_HASH=NewArticleStore_436ae9c986

FUNCTION_DEF=func NewArticleStore(db *gorm.DB) *ArticleStore // NewArticleStore returns a new ArticleStore
*/
func TestNewArticleStore(t *testing.T) {

	testCases := []struct {
		name               string
		inputDB            *gorm.DB
		expectedDBEquality bool
	}{
		{
			name:               "New ArticleStore initialization with non-nil DB",
			inputDB:            setupMockDB(),
			expectedDBEquality: true,
		},
		{
			name:               "ArticleStore initialization with nil DB",
			inputDB:            nil,
			expectedDBEquality: false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {

			defer func() {
				if r := recover(); r != nil {
					t.Logf("Panic encountered so failing test. %v", r)
					t.Fail()
				}
			}()

			actualStore := NewArticleStore(tc.inputDB)

			assert.NotNil(t, actualStore, "%s failed, expected %#v got %#v", tc.name, tc.inputDB, actualStore)
			if tc.expectedDBEquality {
				assert.Equal(t, tc.inputDB, actualStore.db, "%s failed, DB instances are not equal", tc.name)
			} else {
				assert.Nil(t, actualStore.db, "%s failed, expected DB to be nil", tc.name)
			}
		})
	}
}

func setupMockDB() *gorm.DB {
	db, _, err := sqlmock.New()
	if err != nil {
		log.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
	}

	gormDB, err := gorm.Open("mysql", db)
	if err != nil {
		log.Fatalf("an error '%s' was not expected when opening gorm database", err)
	}

	return gormDB
}

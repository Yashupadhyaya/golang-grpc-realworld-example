package store

import (
	fmt "fmt"
	testing "testing"

	"github.com/DATA-DOG/go-sqlmock"
	gorm "github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/postgres"
	model "github.com/raahii/golang-grpc-realworld-example/model"
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

type mockDB struct {
	db      *gorm.DB
	connect bool
}
type mockArticleStore struct {
	db *gorm.DB
}

/*
ROOST_METHOD_HASH=Create_c9b61e3f60
ROOST_METHOD_SIG_HASH=Create_b9fba017bc

FUNCTION_DEF=func Create(m *model.Article) string
*/
func TestCreate(t *testing.T) {
	testCases := []struct {
		desc   string
		input  *model.Article
		db     *mockDB
		output string
	}{
		{
			desc:   "Normal Creation of an Article",
			input:  &model.Article{Title: "Test1", Description: "This is a test article."},
			db:     &mockDB{connect: true, db: &gorm.DB{}},
			output: "just for testing",
		},
		{
			desc:   "Invalid Article Data",
			input:  &model.Article{Title: "", Description: ""},
			db:     &mockDB{connect: true, db: &gorm.DB{}},
			output: "just for testing",
		},
		{
			desc:   "Database Connection Issue",
			input:  &model.Article{Title: "Test3", Description: "This is a test article."},
			db:     &mockDB{connect: false, db: &gorm.DB{}},
			output: "just for testing",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.desc, func(t *testing.T) {
			defer func() {
				if r := recover(); r != nil {
					t.Logf("Panic encountered so failing test. %v", r)
					t.Fail()
				}
			}()

			as := &ArticleStore{
				db: tc.db.db,
			}

			as.db = tc.db.Create(tc.input)

			if fmt.Sprint(as.db) != tc.output {
				t.Errorf("Failed: %s: expected %s, got %v", tc.desc, tc.output, as.db)
			} else {
				t.Logf("Success: %s", tc.desc)
			}
		})
	}
}

func (mdb *mockDB) Create(value interface{}) *gorm.DB {
	if mdb.connect {
		return mdb.db
	}
	return nil
}

/*
ROOST_METHOD_HASH=ArticleStore_CreateComment_b16d4a71d4
ROOST_METHOD_SIG_HASH=ArticleStore_CreateComment_7475736b06

FUNCTION_DEF=func (s *ArticleStore) CreateComment(m *model.Comment) error // CreateComment creates a comment of the article
*/
func TestArticleStoreCreateComment(t *testing.T) {
	tests := []struct {
		scenario string
		comment  *model.Comment
		create   func() (*gorm.DB, sqlmock.Sqlmock, error)
		expect   error
	}{}

	for _, test := range tests {
		t.Run(test.scenario, func(t *testing.T) {

		})
	}
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

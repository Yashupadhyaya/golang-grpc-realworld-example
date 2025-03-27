package store

import (
	gorm "github.com/jinzhu/gorm"
	model "github.com/raahii/golang-grpc-realworld-example/model"
	testing "testing"
	os "os"
	fmt "fmt"
	log "log"
	bytes "bytes"
	gosqlmock "github.com/DATA-DOG/go-sqlmock"
	require "github.com/stretchr/testify/require"
	errors "errors"
)








/*
ROOST_METHOD_HASH=Create_c9b61e3f60
ROOST_METHOD_SIG_HASH=Create_b9fba017bc

FUNCTION_DEF=func Create(m *model.Article) string 

*/
func TestCreate(t *testing.T) {

	testCases := []struct {
		desc           string
		inputArticle   *model.Article
		expectedString string
	}{
		{
			desc:           "Testing the return value of function Create with a valid Article object",
			inputArticle:   &model.Article{},
			expectedString: "just for testing",
		},
		{
			desc:           "Testing the function Create with an empty Article object",
			inputArticle:   &model.Article{},
			expectedString: "just for testing",
		},
		{
			desc:           "Testing the function Create with a null Article object",
			inputArticle:   nil,
			expectedString: "just for testing",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.desc, func(t *testing.T) {

			defer func() {
				if r := recover(); r != nil {
					t.Fatalf("test case: %s \nPanic encountered so failing test %v\n", tc.desc, r)
				}
			}()

			t.Log(tc.desc)

			retString := Create(tc.inputArticle)

			if retString != tc.expectedString {
				t.Errorf("Expected %s but got %s", tc.expectedString, retString)
			}
		})
	}
}


/*
ROOST_METHOD_HASH=ArticleStore_Create_1273475ade
ROOST_METHOD_SIG_HASH=ArticleStore_Create_a27282cad5

FUNCTION_DEF=func (s *ArticleStore) Create(m *model.Article) error // Create creates an article


*/
func TestArticleStoreCreate(t *testing.T) {

	type testData struct {
		name        string
		input       *model.Article
		mockDBError bool
		wantError   bool
	}

	testCases := []testData{
		{
			name:        "Successful Article Creation",
			input:       &model.Article{Slug: "slug1", Title: "title1", Description: "desc1", Body: "body1"},
			mockDBError: false,
			wantError:   false,
		},
		{
			name:        "Article Creation with Invalid Data",
			input:       &model.Article{Slug: "", Title: "title2", Description: "desc2", Body: "body2"},
			mockDBError: false,
			wantError:   true,
		},
		{
			name:        "Database Error during Article Creation",
			input:       &model.Article{Slug: "slug3", Title: "title3", Description: "desc3", Body: "body3"},
			mockDBError: true,
			wantError:   true,
		},
		{
			name:        "Article Creation with Nil Article",
			input:       nil,
			mockDBError: false,
			wantError:   true,
		},
	}

	for _, tt := range testCases {
		t.Run(tt.name, func(t *testing.T) {
			defer func() {
				if r := recover(); r != nil {
					t.Logf("Panic encountered so failing test. %v", r)
					t.Fail()
				}
			}()

			mockDB, mock, err := sqlmock.New()
			require.NoError(t, err)
			gormDB, err := gorm.Open("postgres", mockDB)
			require.NoError(t, err)

			if tt.mockDBError {
				mock.ExpectExec("INSERT INTO").
					WithArgs(tt.input.Slug, tt.input.Title, tt.input.Description, tt.input.Body).
					WillReturnError(fmt.Errorf("mock DB error"))
			} else {
				mock.ExpectExec("INSERT INTO").
					WithArgs(tt.input.Slug, tt.input.Title, tt.input.Description, tt.input.Body).
					WillReturnResult(sqlmock.NewResult(1, 1))
			}

			articleStore := &ArticleStore{db: gormDB}

			err = articleStore.Create(tt.input)

			if tt.wantError {
				require.Error(t, err, "Expected error but got none")
			} else {
				require.NoError(t, err, "Unexpected error encountered")
			}
		})
	}
}


/*
ROOST_METHOD_HASH=ArticleStore_CreateComment_b16d4a71d4
ROOST_METHOD_SIG_HASH=ArticleStore_CreateComment_7475736b06

FUNCTION_DEF=func (s *ArticleStore) CreateComment(m *model.Comment) error // CreateComment creates a comment of the article


*/
func TestArticleStoreCreateComment(t *testing.T) {
	t.Parallel()

	tables := []struct {
		name      string
		setupMock func(mock sqlmock.Sqlmock)
		comment   *model.Comment
		expectErr error
	}{
		{
			name: "Successful Comment Creation",
			setupMock: func(mock sqlmock.Sqlmock) {

				mock.ExpectExec("INSERT INTO comments").WillReturnResult(sqlmock.NewResult(1, 1))
			},
			comment:   &model.Comment{Body: "A comment"},
			expectErr: nil,
		},
		{
			name:      "Comment Creation with Invalid Arguments",
			setupMock: nil,
			comment:   &model.Comment{},
			expectErr: errors.New("comment validation failed: empty comment"),
		},
		{
			name: "Comment Creation with a Failed Connection to the Database",
			setupMock: func(mock sqlmock.Sqlmock) {

				mock.ExpectExec("INSERT INTO comments").
					WillReturnError(errors.New("db connection failed"))
			},
			comment:   &model.Comment{Body: "A comment"},
			expectErr: errors.New("db connection failed"),
		},
	}

	for _, table := range tables {
		func(tt struct {
			name      string
			setupMock func(mock sqlmock.Sqlmock)
			comment   *model.Comment
			expectErr error
		}) {
			t.Run(tt.name, func(t *testing.T) {
				defer func() {
					if r := recover(); r != nil {
						t.Logf("Panic encountered so failing test. %v\n%s", r, string(debug.Stack()))
						t.Fail()
					}
				}()

				db, mock, err := sqlmock.New()
				if err != nil {
					t.Fatalf("Failed to create mock database: %s", err.Error())
				}
				defer db.Close()

				gormDB, err := gorm.Open("sqlmock", db)
				if err != nil {
					t.Fatalf("Failed to create gorm db instance: %s", err.Error())
				}

				if tt.setupMock != nil {
					tt.setupMock(mock)
				}

				store := &ArticleStore{gormDB}
				err = store.CreateComment(tt.comment)

				if tt.expectErr != nil {
					if err == nil {
						t.Errorf("Expected error but got none")
					} else if err.Error() != tt.expectErr.Error() {
						t.Errorf("Expected error: %v, got: %v", tt.expectErr, err)
					}
				} else if err != nil {
					t.Errorf("did not expect error but got: %v", err)
				}
			})
		}(table)
	}
}


/*
ROOST_METHOD_HASH=ArticleStore_DeleteFavorite_29c18a04a8
ROOST_METHOD_SIG_HASH=ArticleStore_DeleteFavorite_53deb5e792

FUNCTION_DEF=func (s *ArticleStore) DeleteFavorite(a *model.Article, u *model.User) error // DeleteFavorite unfavorite an article


*/
func TestArticleStoreDeleteFavorite(t *testing.T) {
	var err error
	db, mock, _ := sqlmock.New()
	defer db.Close()
	mock.ExpectBegin()

	s := &ArticleStore{db: db}

	favoritedArticle := &model.Article{
		FavoritesCount: 2,
	}
	unfavoritedArticle := &model.Article{
		FavoritesCount: 1,
	}
	nonExistentArticle := &model.Article{
		FavoritesCount: 5,
	}
	favoriteUser := &model.User{}
	nonExistentUser := &model.User{}

	scenarios := []struct {
		desc          string
		article       *model.Article
		user          *model.User
		expectedFails bool
	}{
		{
			desc:          "Successful Removal of a Favorite Article",
			article:       favoritedArticle,
			user:          favoriteUser,
			expectedFails: false,
		},
		{
			desc:          "Attempt to Remove an Unfavorited Article",
			article:       unfavoritedArticle,
			user:          favoriteUser,
			expectedFails: true,
		},
		{
			desc:          "Attempt to Remove a Favorite from a Nonexistent User or Article",
			article:       nonExistentArticle,
			user:          nonExistentUser,
			expectedFails: true,
		},
		{
			desc:          "Unsuccessful Removal due to Database issues",
			article:       favoritedArticle,
			user:          favoriteUser,
			expectedFails: true,
		},
	}

	for _, _scenario := range scenarios {
		t.Run(_scenario.desc, func(t *testing.T) {
			defer func() {
				if r := recover(); r != nil {
					t.Logf("Panic encountered so failing test. %v", r)
					t.Fail()
				}
			}()

			err = s.DeleteFavorite(_scenario.article, _scenario.user)

			if _scenario.expectedFails {
				if err == nil {
					t.Errorf("Expected error but got nil for scenario: %v\n", _scenario.desc)
				} else {
					t.Logf("Expected error and got error: %v\n for scenario: %v\n", err, _scenario.desc)
				}
			} else {
				if err != nil {
					t.Errorf("Did not expect error but got error: %v\n for scenario: %v\n", err, _scenario.desc)
				} else {
					t.Logf("No error received as expected for scenario: %v\n", _scenario.desc)
				}
			}
		})
	}
}


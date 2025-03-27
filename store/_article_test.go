package store

import (
	fmt "fmt"
	testing "testing"
	gorm "github.com/jinzhu/gorm"
	model "github.com/raahii/golang-grpc-realworld-example/model"
	errors "errors"
	mock "github.com/stretchr/testify/mock"
	assert "github.com/stretchr/testify/assert"
	gosqlmock "github.com/DATA-DOG/go-sqlmock"
)



var cases = []struct {
	name                   string
	expectedFavoritesCount int
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
	connect bool
}
type MockedDB struct {
	mock.Mock
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
		db     *gorm.DB
		output string
	}{
		{
			desc:   "Normal Creation of an Article",
			input:  &model.Article{Title: "Test1", Description: "This is a test article."},
			db:     &gorm.DB{},
			output: "just for testing",
		},
		{
			desc:   "Invalid Article Data",
			input:  &model.Article{Title: "", Description: ""},
			db:     &gorm.DB{},
			output: "just for testing",
		},
		{
			desc:   "Database Connection Issue",
			input:  &model.Article{Title: "Test3", Description: "This is a test article."},
			db:     &mockDB{connect: false}.Create(nil),
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
				db: tc.db,
			}

			res := as.Create(tc.input)

			if res != tc.output {
				t.Errorf("Failed: %s: expected %s, got %s", tc.desc, tc.output, res)
			} else {
				t.Logf("Success: %s", tc.desc)
			}
		})
	}
}

func (mdb *mockDB) Create(value interface{}) *gorm.DB {
	if mdb.connect {
		return &gorm.DB{}
	}
	return nil
}


/*
ROOST_METHOD_HASH=ArticleStore_Create_1273475ade
ROOST_METHOD_SIG_HASH=ArticleStore_Create_a27282cad5

FUNCTION_DEF=func (s *ArticleStore) Create(m *model.Article) error // Create creates an article


*/
func (m *MockedDB) Create(value interface{}) *gorm.DB {
	args := m.Called(value)
	return args.Get(0).(*gorm.DB)
}

func (m *MockedDB) Error() error {
	args := m.Called()
	return args.Error(0)
}

func TestArticleStoreCreate(t *testing.T) {
	defer func() {
		if r := recover(); r != nil {
			t.Logf("Panic encountered so failing test. %v\n", r)
			t.Fail()
		}
	}()

	tests := []struct {
		name           string
		article        *model.Article
		dbError        error
		expectedResult error
	}{
		{
			name:           "Successful creation of an article",
			article:        &model.Article{Slug: "test", Title: "testTitle"},
			dbError:        nil,
			expectedResult: nil,
		},
		{
			name:           "Failed creation of an article due to database error",
			article:        &model.Article{Slug: "test", Title: "test"},
			dbError:        gorm.ErrRecordNotFound,
			expectedResult: gorm.ErrRecordNotFound,
		},
		{
			name:           "Attempt to create article with invalid data",
			article:        &model.Article{Slug: "test"},
			dbError:        nil,
			expectedResult: errors.New("required fields are missing"),
		},
		{
			name:           "Attempt to create article with null data",
			article:        nil,
			dbError:        nil,
			expectedResult: errors.New("article data is null"),
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			db := new(MockedDB)
			db.On("Create", mock.Anything).Return(db)
			db.On("Error").Return(test.dbError)

			store := ArticleStore{db: db}
			err := store.Create(test.article)

			if err != nil {
				t.Logf("Failed test. %s\n", test.name)
			} else {
				t.Logf("Successfull test. %s\n", test.name)
			}

			assert.Equal(t, test.expectedResult, err)
			db.AssertExpectations(t)
		})
	}
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
		create   func() (*gorm.DB, sqlmock.Sqlmock)
		expect   error
	}{
		{

			scenario: "Successful Creation of Comments",
			comment:  &model.Comment{ID: 1, Body: "test comment", UserID: 1, ArticleID: 2},
			create: func() (*gorm.DB, sqlmock.Sqlmock) {
				db, mock, _ := sqlmock.New()
				defer db.Close()

				mock.ExpectBegin()
				mock.ExpectExec("INSERT INTO \"comments\"").WillReturnResult(sqlmock.NewResult(1, 1))
				mock.ExpectCommit()

				return gorm.Open("postgres", db)
			},
			expect: nil,
		},
		{

			scenario: "Error Handling when Database Unreachable",
			comment:  &model.Comment{ID: 1, Body: "test comment", UserID: 1, ArticleID: 2},
			create: func() (*gorm.DB, sqlmock.Sqlmock) {
				db, mock, _ := sqlmock.New()
				defer db.Close()

				mock.ExpectBegin()
				mock.ExpectExec("INSERT INTO \"comments\"").WillReturnError(errors.New("database unreachable"))
				mock.ExpectRollback()

				return gorm.Open("postgres", db)
			},
			expect: errors.New("database unreachable"),
		},
		{

			scenario: "Error Handling when provided Comment is Nil",
			comment:  nil,
			create: func() (*gorm.DB, sqlmock.Sqlmock) {
				db, mock, _ := sqlmock.New()
				defer db.Close()

				mock.ExpectBegin()
				mock.ExpectExec("INSERT INTO \"comments\"").WillReturnError(errors.New("comment is nil"))
				mock.ExpectRollback()

				return gorm.Open("postgres", db)
			},
			expect: errors.New("comment is nil"),
		},
		{

			scenario: "Error handling when provided Comment is malformed",
			comment:  &model.Comment{ID: 1, Body: "", UserID: 1, ArticleID: 1},
			create: func() (*gorm.DB, sqlmock.Sqlmock) {
				db, mock, _ := sqlmock.New()
				defer db.Close()

				mock.ExpectBegin()
				mock.ExpectExec("INSERT INTO \"comments\"").WillReturnError(errors.New("malformed comment"))
				mock.ExpectRollback()

				return gorm.Open("postgres", db)
			},
			expect: errors.New("malformed comment"),
		},
	}

	for _, test := range tests {
		t.Run(test.scenario, func(t *testing.T) {
			defer func() {
				if r := recover(); r != nil {
					t.Logf("Panic encountered so failing test. %v", r)
					t.Fail()
				}
			}()

			db, _ := test.create()
			articleStore := &ArticleStore{db: db}

			err := articleStore.CreateComment(test.comment)

			if test.expect != nil && err == nil {
				t.Errorf("expected an error %v, got nil", test.expect)
				return
			}

			if err != nil {
				if test.expect == nil {
					t.Errorf("expected no error, got %v", err)
				}
				if err.Error() != test.expect.Error() {
					t.Errorf("expected error %v, got %v", test.expect, err)
				}
			}
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

			mockSqlDB, mockSqlMocks, _ := sqlmock.New()
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


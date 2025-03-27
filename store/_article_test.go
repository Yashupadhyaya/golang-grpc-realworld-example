// ********RoostGPT********
/*

roost_feedback [3/27/2025, 12:41:32 PM]:The import statements must be these:\r\n```\r\nimport (\r\n\terrors errors\r\n\ttesting testing\r\n\r\n\tsqlmock github.com/DATA-DOG/go-sqlmock\r\n\tgorm github.com/jinzhu/gorm\r\n\tmodel github.com/raahii/golang-grpc-realworld-example/model\r\n\tassert github.com/stretchr/testify/assert\r\n\tmock github.com/stretchr/testify/mock\r\n)\r\n```\r\n\r\nThe testing table in the TestCreate function must be this:\r\n```\r\ntestCases := []struct {\r\n\t\tdesc   string\r\n\t\tinput  *model.Article\r\n\t\tdb     *gorm.DB\r\n\t\toutput string\r\n\t}{\r\n\t\t{\r\n\t\t\tdesc:   Normal Creation of an Article,\r\n\t\t\tinput:  &model.Article{Title: Test1, Description: This is a test article.},\r\n\t\t\tdb:     &gorm.DB{},\r\n\t\t\toutput: just for testing,\r\n\t\t},\r\n\t\t{\r\n\t\t\tdesc:   Invalid Article Data,\r\n\t\t\tinput:  &model.Article{Title: , Description: },\r\n\t\t\tdb:     &gorm.DB{},\r\n\t\t\toutput: just for testing,\r\n\t\t},\r\n\t\t{\r\n\t\t\tdesc:  Database Connection Issue,\r\n\t\t\tinput: &model.Article{Title: Test3, Description: This is a test article.},\r\n\t\t\t// db:     &mockDB{connect: false}.Create(nil),\r\n\t\t\toutput: just for testing,\r\n\t\t},\r\n\t}\r\n```\r\n\r\nIn TestCreate function, compare res to tc.output in this way:\r\n```\r\nif res.Error() != tc.output {\r\n\t\t\t\tt.Errorf(Failed: %s: expected %s, got %s, tc.desc, tc.output, res)\r\n\t\t\t} else {\r\n\t\t\t\tt.Logf(Success: %s, tc.desc)\r\n\t\t\t}\r\n```\r\n\r\nThe testing table in TestArticleStoreCreate function must be this:\r\n```\r\ntests := []struct {\r\n\t\tname           string\r\n\t\tarticle        *model.Article\r\n\t\tdbError        error\r\n\t\texpectedResult error\r\n\t}{\r\n\t\t{\r\n\t\t\tname:           Successful creation of an article,\r\n\t\t\tarticle:        &model.Article{Title: testTitle},\r\n\t\t\tdbError:        nil,\r\n\t\t\texpectedResult: nil,\r\n\t\t},\r\n\t\t{\r\n\t\t\tname:           Failed creation of an article due to database error,\r\n\t\t\tarticle:        &model.Article{Title: test},\r\n\t\t\tdbError:        gorm.ErrRecordNotFound,\r\n\t\t\texpectedResult: gorm.ErrRecordNotFound,\r\n\t\t},\r\n\t\t{\r\n\t\t\tname:           Attempt to create article with invalid data,\r\n\t\t\tarticle:        &model.Article{},\r\n\t\t\tdbError:        nil,\r\n\t\t\texpectedResult: errors.New(required fields are missing),\r\n\t\t},\r\n\t\t{\r\n\t\t\tname:           Attempt to create article with null data,\r\n\t\t\tarticle:        nil,\r\n\t\t\tdbError:        nil,\r\n\t\t\texpectedResult: errors.New(article data is null),\r\n\t\t},\r\n\t}\r\n```\r\n\r\nThe testing table in TestArticleStoreCreateComment function must be this:\r\n\r\n```\r\ntests := []struct {\r\n\t\tscenario string\r\n\t\tcomment  *model.Comment\r\n\t\tcreate   func() (*gorm.DB, sqlmock.Sqlmock)\r\n\t\texpect   error\r\n\t}{\r\n\t\t{\r\n\r\n\t\t\tscenario: Successful Creation of Comments,\r\n\t\t\tcomment:  &model.Comment{Body: test comment, UserID: 1, ArticleID: 2},\r\n\t\t\tcreate: func() (*gorm.DB, sqlmock.Sqlmock) {\r\n\t\t\t\tdb, mock, _ := sqlmock.New()\r\n\t\t\t\tdefer db.Close()\r\n\r\n\t\t\t\tmock.ExpectBegin()\r\n\t\t\t\tmock.ExpectExec(INSERT INTO \comments\).WillReturnResult(sqlmock.NewResult(1, 1))\r\n\t\t\t\tmock.ExpectCommit()\r\n\t\t\t\tgormDB, _ := gorm.Open(postgres, db)\r\n\t\t\t\treturn gormDB, mock\r\n\t\t\t},\r\n\t\t\texpect: nil,\r\n\t\t},\r\n\t\t{\r\n\r\n\t\t\tscenario: Error Handling when Database Unreachable,\r\n\t\t\tcomment:  &model.Comment{Body: test comment, UserID: 1, ArticleID: 2},\r\n\t\t\tcreate: func() (*gorm.DB, sqlmock.Sqlmock) {\r\n\t\t\t\tdb, mock, _ := sqlmock.New()\r\n\t\t\t\tdefer db.Close()\r\n\r\n\t\t\t\tmock.ExpectBegin()\r\n\t\t\t\tmock.ExpectExec(INSERT INTO \comments\).WillReturnError(errors.New(database unreachable))\r\n\t\t\t\tmock.ExpectRollback()\r\n\r\n\t\t\t\tgormDB, _ := gorm.Open(postgres, db)\r\n\t\t\t\treturn gormDB, mock\r\n\t\t\t},\r\n\t\t\texpect: errors.New(database unreachable),\r\n\t\t},\r\n\t\t{\r\n\r\n\t\t\tscenario: Error Handling when provided Comment is Nil,\r\n\t\t\tcomment:  nil,\r\n\t\t\tcreate: func() (*gorm.DB, sqlmock.Sqlmock) {\r\n\t\t\t\tdb, mock, _ := sqlmock.New()\r\n\t\t\t\tdefer db.Close()\r\n\r\n\t\t\t\tmock.ExpectBegin()\r\n\t\t\t\tmock.ExpectExec(INSERT INTO \comments\).WillReturnError(errors.New(comment is nil))\r\n\t\t\t\tmock.ExpectRollback()\r\n\r\n\t\t\t\tgormDB, _ := gorm.Open(postgres, db)\r\n\t\t\t\treturn gormDB, mock\r\n\t\t\t},\r\n\t\t\texpect: errors.New(comment is nil),\r\n\t\t},\r\n\t\t{\r\n\r\n\t\t\tscenario: Error handling when provided Comment is malformed,\r\n\t\t\tcomment:  &model.Comment{Body: , UserID: 1, ArticleID: 1},\r\n\t\t\tcreate: func() (*gorm.DB, sqlmock.Sqlmock) {\r\n\t\t\t\tdb, mock, _ := sqlmock.New()\r\n\t\t\t\tdefer db.Close()\r\n\r\n\t\t\t\tmock.ExpectBegin()\r\n\t\t\t\tmock.ExpectExec(INSERT INTO \comments\).WillReturnError(errors.New(malformed comment))\r\n\t\t\t\tmock.ExpectRollback()\r\n\r\n\t\t\t\tgormDB, _ := gorm.Open(postgres, db)\r\n\t\t\t\treturn gormDB, mock\r\n\t\t\t},\r\n\t\t\texpect: errors.New(malformed comment),\r\n\t\t},\r\n\t}\r\n```\r\n\r\nThe TestArticleStoreDeleteFavorite function must be this:\r\n```\r\nfunc TestArticleStoreDeleteFavorite(t *testing.T) {\r\n\tfor _, tc := range cases {\r\n\t\tt.Run(tc.name, func(t *testing.T) {\r\n\t\t\tdefer func() {\r\n\t\t\t\tif r := recover(); r != nil {\r\n\t\t\t\t\tt.Logf(Panic encountered so failing test. %v, r)\r\n\t\t\t\t\tt.Fail()\r\n\t\t\t\t}\r\n\t\t\t}()\r\n\r\n\t\t\tmockSqlDB, _, _ := sqlmock.New()\r\n\t\t\tmockGormDB, _ := gorm.Open(postgres, mockSqlDB)\r\n\t\t\tmockGormDB.LogMode(false)\r\n\r\n\t\t\tmockArticleStore := &mockArticleStore{\r\n\t\t\t\tdb: mockGormDB,\r\n\t\t\t}\r\n\r\n\t\t\tmockArticle := &model.Article{\r\n\t\t\t\tFavoritesCount: int32(tc.expectedFavoritesCount),\r\n\t\t\t}\r\n\t\t\tmockUser := &model.User{}\r\n\r\n\t\t\terr := mockArticleStore.DeleteFavorite(mockArticle, mockUser)\r\n\r\n\t\t\tif err != tc.expectedError {\r\n\t\t\t\tt.Errorf(error got: %v want: %v, err, tc.expectedError)\r\n\t\t\t}\r\n\r\n\t\t\tif mockArticle.FavoritesCount != int32(tc.expectedFavoritesCount) {\r\n\t\t\t\tt.Errorf(favorites count got: %d want: %d, mockArticle.FavoritesCount, tc.expectedFavoritesCount)\r\n\t\t\t}\r\n\r\n\t\t})\r\n\t}\r\n}\r\n```
*/

// ********RoostGPT********

package store

import (
	"errors"
	"testing"

	sqlmock "github.com/DATA-DOG/go-sqlmock"
	gorm "github.com/jinzhu/gorm"
	model "github.com/raahii/golang-grpc-realworld-example/model"
	assert "github.com/stretchr/testify/assert"
	mock "github.com/stretchr/testify/mock"
)

type mockDB struct {
	connect bool
}

func (mdb *mockDB) Create(value interface{}) *gorm.DB {
	if mdb.connect {
		return &gorm.DB{}
	}
	return nil
}

type MockedDB struct {
	mock.Mock
}

func (m *MockedDB) Create(value interface{}) *gorm.DB {
	args := m.Called(value)
	return args.Get(0).(*gorm.DB)
}

func (m *MockedDB) Error() error {
	args := m.Called()
	return args.Error(0)
}

type mockArticleStore struct {
	db *gorm.DB
}

func (m *mockArticleStore) DeleteFavorite(a *model.Article, u *model.User) error {
	return nil
}

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
			desc:  "Database Connection Issue",
			input: &model.Article{Title: "Test3", Description: "This is a test article."},
			// db:     &mockDB{connect: false}.Create(nil),
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

			if res.Error() != tc.output {
				t.Errorf("Failed: %s: expected %s, got %s", tc.desc, tc.output, res)
			} else {
				t.Logf("Success: %s", tc.desc)
			}
		})
	}
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
			article:        &model.Article{Title: "testTitle"},
			dbError:        nil,
			expectedResult: nil,
		},
		{
			name:           "Failed creation of an article due to database error",
			article:        &model.Article{Title: "test"},
			dbError:        gorm.ErrRecordNotFound,
			expectedResult: gorm.ErrRecordNotFound,
		},
		{
			name:           "Attempt to create article with invalid data",
			article:        &model.Article{},
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
				t.Logf("Successful test. %s\n", test.name)
			}

			assert.Equal(t, test.expectedResult, err)
			db.AssertExpectations(t)
		})
	}
}

func TestArticleStoreCreateComment(t *testing.T) {

	tests := []struct {
		scenario string
		comment  *model.Comment
		create   func() (*gorm.DB, sqlmock.Sqlmock)
		expect   error
	}{
		{

			scenario: "Successful Creation of Comments",
			comment:  &model.Comment{Body: "test comment", UserID: 1, ArticleID: 2},
			create: func() (*gorm.DB, sqlmock.Sqlmock) {
				db, mock, _ := sqlmock.New()
				defer db.Close()

				mock.ExpectBegin()
				mock.ExpectExec("INSERT INTO \"comments\"").WillReturnResult(sqlmock.NewResult(1, 1))
				mock.ExpectCommit()
				gormDB, _ := gorm.Open("postgres", db)
				return gormDB, mock
			},
			expect: nil,
		},
		{

			scenario: "Error Handling when Database Unreachable",
			comment:  &model.Comment{Body: "test comment", UserID: 1, ArticleID: 2},
			create: func() (*gorm.DB, sqlmock.Sqlmock) {
				db, mock, _ := sqlmock.New()
				defer db.Close()

				mock.ExpectBegin()
				mock.ExpectExec("INSERT INTO \"comments\"").WillReturnError(errors.New("database unreachable"))
				mock.ExpectRollback()

				gormDB, _ := gorm.Open("postgres", db)
				return gormDB, mock
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

				gormDB, _ := gorm.Open("postgres", db)
				return gormDB, mock
			},
			expect: errors.New("comment is nil"),
		},
		{

			scenario: "Error handling when provided Comment is malformed",
			comment:  &model.Comment{Body: "", UserID: 1, ArticleID: 1},
			create: func() (*gorm.DB, sqlmock.Sqlmock) {
				db, mock, _ := sqlmock.New()
				defer db.Close()

				mock.ExpectBegin()
				mock.ExpectExec("INSERT INTO \"comments\"").WillReturnError(errors.New("malformed comment"))
				mock.ExpectRollback()

				gormDB, _ := gorm.Open("postgres", db)
				return gormDB, mock
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

func TestArticleStoreDeleteFavorite(t *testing.T) {
	cases := []struct {
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
				FavoritesCount: int32(tc.expectedFavoritesCount),
			}
			mockUser := &model.User{}

			err := mockArticleStore.DeleteFavorite(mockArticle, mockUser)

			if err != tc.expectedError {
				t.Errorf("error got: %v want: %v", err, tc.expectedError)
			}

			if mockArticle.FavoritesCount != int32(tc.expectedFavoritesCount) {
				t.Errorf("favorites count got: %d want: %d", mockArticle.FavoritesCount, tc.expectedFavoritesCount)
			}
		})
	}
}

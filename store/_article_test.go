package store

import (
	fmt "fmt"
	testing "testing"
	gorm "github.com/jinzhu/gorm"
	model "github.com/raahii/golang-grpc-realworld-example/model"
	sqlmock "github.com/DATA-DOG/go-sqlmock"
	assert "github.com/stretchr/testify/assert"
)





type mockDB struct {
	connect bool
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
			mock.ExpectQuery("^INSERT INTO \"articles\"*").
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
		})
	}
}


/*
ROOST_METHOD_HASH=ArticleStore_CreateComment_b16d4a71d4
ROOST_METHOD_SIG_HASH=ArticleStore_CreateComment_7475736b06

FUNCTION_DEF=func (s *ArticleStore) CreateComment(m *model.Comment) error // CreateComment creates a comment of the article


*/
func TestArticleStoreCreateComment(t *testing.T) {

	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("failed to open mock sql DB: %v", err)
	}
	defer db.Close()

	store := &ArticleStore{
		db: (*gorm.DB)(db),
	}

	tests := []struct {
		name          string
		comment       *model.Comment
		dbMock        func()
		errorExpected bool
	}{
		{
			name: "Successful Creation of Comments",
			comment: &model.Comment{
				ID:       123,
				Body:     "This is a test comment.",
				AuthorID: 456,
			},
			dbMock: func() {
				mock.ExpectBegin()
				mock.ExpectExec("INSERT INTO").WillReturnResult(sqlmock.NewResult(1, 1))
				mock.ExpectCommit()
			},
			errorExpected: false,
		},
		{
			name: "Error Handling when Database Unreachable",
			comment: &model.Comment{
				ID:       123,
				Body:     "This is a test comment.",
				AuthorID: 456,
			},
			dbMock: func() {
				mock.ExpectBegin()
				mock.ExpectExec("INSERT INTO").WillReturnError(fmt.Errorf("DB unreachable"))
			},
			errorExpected: true,
		},
		{
			name:          "Error Handling when provided Comment is Nil",
			comment:       nil,
			dbMock:        func() {},
			errorExpected: true,
		},
		{
			name: "Error handling when provided Comment is malformed",
			comment: &model.Comment{
				ID:       123,
				Body:     "",
				AuthorID: 456,
			},
			dbMock: func() {
				mock.ExpectBegin()
				mock.ExpectExec("INSERT INTO").WillReturnError(fmt.Errorf("Malformed comment"))
			},
			errorExpected: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			defer func() {
				if r := recover(); r != nil {
					t.Logf("Panic encountered so failing test. %v", r)
					t.Fail()
				}
			}()

			tt.dbMock()

			err := store.CreateComment(tt.comment)

			if tt.errorExpected {
				assert.Error(t, err, "Error was expected")
			} else {
				assert.NoError(t, err, "Error was not expected")
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
	ax1 := &model.Article{FavoritesCount: 1}
	ux1 := &model.User{Email: "user1@test.com"}
	ax2 := &model.Article{FavoritesCount: 1}
	ux2 := &model.User{Email: "user2@test.com"}
	ax3 := &model.Article{FavoritesCount: 0}
	ux3 := &model.User{Email: "user3@test.com"}

	tests := []struct {
		name        string
		a           *model.Article
		u           *model.User
		wantErr     bool
		errStr      string
		favCountDec bool
	}{
		{
			name:        "DeleteFavorite Successfully Removes Favorite User and Reduces Favorites Count",
			a:           ax1,
			u:           ux1,
			wantErr:     false,
			favCountDec: true,
		},
		{
			name:    "DeleteFavorite Handles Errors when User is not in 'FavoritedUsers'",
			a:       ax2,
			u:       ux2,
			wantErr: true,
			errStr:  fmt.Sprintf("User %v not found in article's FavoritedUsers", ux2.Email),
		},
		{
			name:    "DeleteFavorite Handles Error when Trying to Decrement 'favorites_count' Below 0",
			a:       ax3,
			u:       ux3,
			wantErr: true,
			errStr:  "Cannot decrement 'favorites_count' to a negative value",
		},
	}

	s := &ArticleStore{}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			defer func() {
				if r := recover(); r != nil {
					t.Logf("Panic encountered so failing test. %v", r)
					t.Fail()
				}
			}()

			err := s.DeleteFavorite(tt.a, tt.u)
			if (err != nil) != tt.wantErr {
				t.Errorf("DeleteFavorite() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if tt.wantErr && err != nil && err.Error() != tt.errStr {
				t.Errorf("DeleteFavorite() unexpected error string = %v, wantErrStr %v", err.Error(), tt.errStr)
				return
			}

			if !tt.wantErr && tt.a.FavoritesCount != 0 && tt.favCountDec {
				t.Errorf("DeleteFavorite() did not decrement FavoritesCount. Got = %v, want = %v", tt.a.FavoritesCount, tt.a.FavoritesCount-1)
				return
			}
		})
	}
}


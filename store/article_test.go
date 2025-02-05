package store

import (
	"testing"
	"github.com/jinzhu/gorm"
	"github.com/raahii/golang-grpc-realworld-example/model"
	"errors"
)





type ArticleStore struct {
	db *gorm.DB
}


/*
ROOST_METHOD_HASH=Create_c9b61e3f60
ROOST_METHOD_SIG_HASH=Create_b9fba017bc

FUNCTION_DEF=func Create(m *model.Article) string 

*/
func TestCreate(t *testing.T) {
	tests := []struct {
		name     string
		article  *model.Article
		expected string
	}{
		{
			name: "Successful Article Creation",
			article: &model.Article{
				Title:       "Test Article",
				Description: "This is a test article",
				Body:        "Article body",
			},
			expected: "just for testing",
		},
		{
			name:     "Handling Nil Input",
			article:  nil,
			expected: "just for testing",
		},
		{
			name:     "Article with Empty Fields",
			article:  &model.Article{},
			expected: "just for testing",
		},
		{
			name: "Article with Maximum Field Values",
			article: &model.Article{
				Title:       string(make([]byte, 1000)),
				Description: string(make([]byte, 1000)),
				Body:        string(make([]byte, 10000)),
			},
			expected: "just for testing",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := Create(tt.article)
			if result != tt.expected {
				t.Errorf("Create() = %v, want %v", result, tt.expected)
			}
		})
	}
}

func TestCreateConcurrent(t *testing.T) {
	concurrentTests := 100
	done := make(chan bool)

	for i := 0; i < concurrentTests; i++ {
		go func() {
			article := &model.Article{
				Title: "Concurrent Test Article",
			}
			result := Create(article)
			if result != "just for testing" {
				t.Errorf("Concurrent Create() = %v, want %v", result, "just for testing")
			}
			done <- true
		}()
	}

	for i := 0; i < concurrentTests; i++ {
		<-done
	}
}


/*
ROOST_METHOD_HASH=CreateComment_b16d4a71d4
ROOST_METHOD_SIG_HASH=CreateComment_7475736b06

FUNCTION_DEF=func (s *ArticleStore) CreateComment(m *model.Comment) error // CreateComment creates a comment of the article


*/
func (m *mockDB) Create(value interface{}) *gorm.DB {
	return m.createFunc(value)
}

func TestArticleStoreCreateComment(t *testing.T) {
	tests := []struct {
		name    string
		comment *model.Comment
		dbError error
		wantErr bool
	}{
		{
			name: "Successfully Create a Comment",
			comment: &model.Comment{
				Body:      "Test comment",
				ArticleID: 1,
				UserID:    1,
			},
			dbError: nil,
			wantErr: false,
		},
		{
			name: "Handling Database Error During Comment Creation",
			comment: &model.Comment{
				Body:      "Test comment",
				ArticleID: 1,
				UserID:    1,
			},
			dbError: errors.New("database error"),
			wantErr: true,
		},
		{
			name: "Creating a Comment with Empty Fields",
			comment: &model.Comment{
				Body:      "",
				ArticleID: 1,
				UserID:    1,
			},
			dbError: nil,
			wantErr: false,
		},
		{
			name: "Creating a Comment with Maximum Length Content",
			comment: &model.Comment{
				Body:      string(make([]byte, 1000)),
				ArticleID: 1,
				UserID:    1,
			},
			dbError: nil,
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockDB := &mockDB{
				createFunc: func(value interface{}) *gorm.DB {
					return &gorm.DB{Error: tt.dbError}
				},
			}

			store := &ArticleStore{db: mockDB}

			err := store.CreateComment(tt.comment)

			if (err != nil) != tt.wantErr {
				t.Errorf("CreateComment() error = %v, wantErr %v", err, tt.wantErr)
			}

			if tt.wantErr && err != tt.dbError {
				t.Errorf("CreateComment() expected error %v, got %v", tt.dbError, err)
			}
		})
	}
}


/*
ROOST_METHOD_HASH=DeleteFavorite_29c18a04a8
ROOST_METHOD_SIG_HASH=DeleteFavorite_53deb5e792

FUNCTION_DEF=func (s *ArticleStore) DeleteFavorite(a *model.Article, u *model.User) error // DeleteFavorite unfavorite an article


*/
func TestArticleStoreDeleteFavorite(t *testing.T) {
	tests := []struct {
		name           string
		article        *model.Article
		user           *model.User
		mockDB         *mockDB
		expectedError  error
		expectedCount  int
		expectRollback bool
		expectCommit   bool
	}{
		{
			name:           "Successfully Delete a Favorite",
			article:        &model.Article{FavoritesCount: 1},
			user:           &model.User{},
			mockDB:         &mockDB{},
			expectedError:  nil,
			expectedCount:  0,
			expectRollback: false,
			expectCommit:   true,
		},
		{
			name:           "Delete Favorite for Non-Existent Association",
			article:        &model.Article{FavoritesCount: 0},
			user:           &model.User{},
			mockDB:         &mockDB{},
			expectedError:  nil,
			expectedCount:  0,
			expectRollback: false,
			expectCommit:   true,
		},
		{
			name:           "Database Error During Association Deletion",
			article:        &model.Article{FavoritesCount: 1},
			user:           &model.User{},
			mockDB:         &mockDB{deleteAssocError: errors.New("association deletion error")},
			expectedError:  errors.New("association deletion error"),
			expectedCount:  1,
			expectRollback: true,
			expectCommit:   false,
		},
		{
			name:           "Database Error During Favorites Count Update",
			article:        &model.Article{FavoritesCount: 1},
			user:           &model.User{},
			mockDB:         &mockDB{updateCountError: errors.New("update count error")},
			expectedError:  errors.New("update count error"),
			expectedCount:  1,
			expectRollback: true,
			expectCommit:   false,
		},
		{
			name:           "Delete Favorite When Favorites Count is Already Zero",
			article:        &model.Article{FavoritesCount: 0},
			user:           &model.User{},
			mockDB:         &mockDB{},
			expectedError:  nil,
			expectedCount:  0,
			expectRollback: false,
			expectCommit:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			store := &ArticleStore{db: tt.mockDB}
			err := store.DeleteFavorite(tt.article, tt.user)

			if (err != nil) != (tt.expectedError != nil) {
				t.Errorf("DeleteFavorite() error = %v, expectedError %v", err, tt.expectedError)
			}

			if err != nil && err.Error() != tt.expectedError.Error() {
				t.Errorf("DeleteFavorite() error = %v, expectedError %v", err, tt.expectedError)
			}

			if tt.article.FavoritesCount != tt.expectedCount {
				t.Errorf("DeleteFavorite() FavoritesCount = %v, expected %v", tt.article.FavoritesCount, tt.expectedCount)
			}

			if tt.mockDB.rollbackCalled != tt.expectRollback {
				t.Errorf("DeleteFavorite() rollback called = %v, expected %v", tt.mockDB.rollbackCalled, tt.expectRollback)
			}

			if tt.mockDB.commitCalled != tt.expectCommit {
				t.Errorf("DeleteFavorite() commit called = %v, expected %v", tt.mockDB.commitCalled, tt.expectCommit)
			}

			if !tt.mockDB.beginCalled {
				t.Error("DeleteFavorite() begin not called")
			}

			if !tt.mockDB.associationCalled {
				t.Error("DeleteFavorite() association not called")
			}

			if !tt.mockDB.updateCalled {
				t.Error("DeleteFavorite() update not called")
			}
		})
	}
}


package store

import (
	"testing"
	"github.com/jinzhu/gorm"
	"github.com/raahii/golang-grpc-realworld-example/model"
	"errors"
	"time"
	"sync"
)





type mockDB struct {
	createFunc func(interface{}) *gorm.DB
}
type mockArticleStore struct {
	db *mockDB
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
			name: "Basic Article Creation",
			article: &model.Article{
				Title:       "Test Article",
				Description: "This is a test article",
				Body:        "Lorem ipsum dolor sit amet",
			},
			expected: "just for testing",
		},
		{
			name:     "Null Article Input",
			article:  nil,
			expected: "just for testing",
		},
		{
			name:     "Article with Empty Fields",
			article:  &model.Article{},
			expected: "just for testing",
		},
		{
			name: "Article with Maximum Field Lengths",
			article: &model.Article{
				Title:       string(make([]byte, 255)),
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
	article := &model.Article{
		Title:       "Concurrent Test Article",
		Description: "This is a concurrent test article",
		Body:        "Lorem ipsum dolor sit amet",
	}

	concurrency := 100
	done := make(chan bool)

	for i := 0; i < concurrency; i++ {
		go func() {
			result := Create(article)
			if result != "just for testing" {
				t.Errorf("Concurrent Create() = %v, want %v", result, "just for testing")
			}
			done <- true
		}()
	}

	for i := 0; i < concurrency; i++ {
		<-done
	}
}

func TestCreatePerformance(t *testing.T) {
	article := &model.Article{
		Title:       "Performance Test Article",
		Description: "This is a performance test article",
		Body:        "Lorem ipsum dolor sit amet",
	}

	iterations := 10000

	for i := 0; i < iterations; i++ {
		result := Create(article)
		if result != "just for testing" {
			t.Errorf("Performance Create() = %v, want %v", result, "just for testing")
		}
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
		dbFunc  func(interface{}) *gorm.DB
		wantErr bool
	}{
		{
			name: "Successfully Create a Comment",
			comment: &model.Comment{
				Body:      "Test comment",
				ArticleID: 1,
				UserID:    1,
				CreatedAt: time.Now(),
				UpdatedAt: time.Now(),
			},
			dbFunc: func(interface{}) *gorm.DB {
				return &gorm.DB{Error: nil}
			},
			wantErr: false,
		},
		{
			name:    "Attempt to Create a Comment with Invalid Data",
			comment: &model.Comment{},
			dbFunc: func(interface{}) *gorm.DB {
				return &gorm.DB{Error: errors.New("invalid data")}
			},
			wantErr: true,
		},
		{
			name: "Database Connection Failure",
			comment: &model.Comment{
				Body:      "Test comment",
				ArticleID: 1,
				UserID:    1,
				CreatedAt: time.Now(),
				UpdatedAt: time.Now(),
			},
			dbFunc: func(interface{}) *gorm.DB {
				return &gorm.DB{Error: errors.New("database connection failed")}
			},
			wantErr: true,
		},
		{
			name: "Duplicate Comment Creation",
			comment: &model.Comment{
				Body:      "Duplicate comment",
				ArticleID: 1,
				UserID:    1,
				CreatedAt: time.Now(),
				UpdatedAt: time.Now(),
			},
			dbFunc: func(interface{}) *gorm.DB {
				return &gorm.DB{Error: errors.New("duplicate entry")}
			},
			wantErr: true,
		},
		{
			name: "Create Comment with Maximum Allowed Length",
			comment: &model.Comment{
				Body:      string(make([]byte, 1000)),
				ArticleID: 1,
				UserID:    1,
				CreatedAt: time.Now(),
				UpdatedAt: time.Now(),
			},
			dbFunc: func(interface{}) *gorm.DB {
				return &gorm.DB{Error: nil}
			},
			wantErr: false,
		},
		{
			name: "Create Comment with Special Characters",
			comment: &model.Comment{
				Body:      "Comment with special characters: !@#$%^&*()_+{}[]|\\:;\"'<>,.?/~`",
				ArticleID: 1,
				UserID:    1,
				CreatedAt: time.Now(),
				UpdatedAt: time.Now(),
			},
			dbFunc: func(interface{}) *gorm.DB {
				return &gorm.DB{Error: nil}
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockDB := &mockDB{
				createFunc: tt.dbFunc,
			}
			s := &ArticleStore{
				db: mockDB,
			}
			err := s.CreateComment(tt.comment)
			if (err != nil) != tt.wantErr {
				t.Errorf("ArticleStore.CreateComment() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}


/*
ROOST_METHOD_HASH=DeleteFavorite_29c18a04a8
ROOST_METHOD_SIG_HASH=DeleteFavorite_53deb5e792

FUNCTION_DEF=func (s *ArticleStore) DeleteFavorite(a *model.Article, u *model.User) error // DeleteFavorite unfavorite an article


*/
func (s *mockArticleStore) DeleteFavorite(a *model.Article, u *model.User) error {
	if a == nil {
		return errors.New("article is nil")
	}
	if u == nil {
		return errors.New("user is nil")
	}

	tx := s.db.Begin()
	err := tx.Model(a).Association("FavoritedUsers").Delete(u).Error()
	if err != nil {
		tx.Rollback()
		return err
	}
	err = tx.Model(a).Update("favorites_count", tx.Expr("favorites_count - ?", 1)).Error()
	if err != nil {
		tx.Rollback()
		return err
	}
	tx.Commit()
	a.FavoritesCount--
	return nil
}

func TestArticleStoreDeleteFavorite(t *testing.T) {
	tests := []struct {
		name            string
		article         *model.Article
		user            *model.User
		setupMockDB     func(*mockDB)
		expectedError   error
		expectedCount   int
		concurrentCalls int
	}{
		{
			name:    "Successfully Delete a Favorite",
			article: &model.Article{FavoritesCount: 1},
			user:    &model.User{},
			setupMockDB: func(m *mockDB) {
				m.favoritesCount = 1
			},
			expectedError: nil,
			expectedCount: 0,
		},
		{
			name:    "Delete Favorite When User Hasn't Favorited the Article",
			article: &model.Article{FavoritesCount: 0},
			user:    &model.User{},
			setupMockDB: func(m *mockDB) {
				m.favoritesCount = 0
			},
			expectedError: nil,
			expectedCount: 0,
		},
		{
			name:    "Database Error During Association Deletion",
			article: &model.Article{FavoritesCount: 1},
			user:    &model.User{},
			setupMockDB: func(m *mockDB) {
				m.associationError = errors.New("association deletion error")
			},
			expectedError: errors.New("association deletion error"),
			expectedCount: 1,
		},
		{
			name:    "Database Error During Favorites Count Update",
			article: &model.Article{FavoritesCount: 1},
			user:    &model.User{},
			setupMockDB: func(m *mockDB) {
				m.updateError = errors.New("update error")
			},
			expectedError: errors.New("update error"),
			expectedCount: 1,
		},
		{
			name:    "Delete Favorite for Article with Zero Favorites",
			article: &model.Article{FavoritesCount: 0},
			user:    &model.User{},
			setupMockDB: func(m *mockDB) {
				m.favoritesCount = 0
			},
			expectedError: nil,
			expectedCount: 0,
		},
		{
			name:    "Concurrent Deletion of Favorites",
			article: &model.Article{FavoritesCount: 5},
			user:    &model.User{},
			setupMockDB: func(m *mockDB) {
				m.favoritesCount = 5
			},
			expectedError:   nil,
			expectedCount:   0,
			concurrentCalls: 5,
		},
		{
			name:          "Delete Favorite with Nil Article",
			article:       nil,
			user:          &model.User{},
			setupMockDB:   func(m *mockDB) {},
			expectedError: errors.New("article is nil"),
			expectedCount: 0,
		},
		{
			name:          "Delete Favorite with Nil User",
			article:       &model.Article{FavoritesCount: 1},
			user:          nil,
			setupMockDB:   func(m *mockDB) {},
			expectedError: errors.New("user is nil"),
			expectedCount: 1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockDB := &mockDB{}
			tt.setupMockDB(mockDB)

			store := &mockArticleStore{db: mockDB}

			if tt.concurrentCalls > 0 {
				var wg sync.WaitGroup
				for i := 0; i < tt.concurrentCalls; i++ {
					wg.Add(1)
					go func() {
						defer wg.Done()
						err := store.DeleteFavorite(tt.article, tt.user)
						if err != tt.expectedError {
							t.Errorf("DeleteFavorite() error = %v, expectedError %v", err, tt.expectedError)
						}
					}()
				}
				wg.Wait()
			} else {
				err := store.DeleteFavorite(tt.article, tt.user)
				if err != tt.expectedError && (err == nil || tt.expectedError == nil || err.Error() != tt.expectedError.Error()) {
					t.Errorf("DeleteFavorite() error = %v, expectedError %v", err, tt.expectedError)
				}
			}

			if tt.article != nil && tt.article.FavoritesCount != tt.expectedCount {
				t.Errorf("DeleteFavorite() favoritesCount = %v, expected %v", tt.article.FavoritesCount, tt.expectedCount)
			}

			if mockDB.favoritesCount != tt.expectedCount {
				t.Errorf("DeleteFavorite() db favoritesCount = %v, expected %v", mockDB.favoritesCount, tt.expectedCount)
			}
		})
	}
}


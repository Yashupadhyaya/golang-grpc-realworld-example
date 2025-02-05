package store

import (
	"testing"
	"github.com/jinzhu/gorm"
	"github.com/raahii/golang-grpc-realworld-example/model"
	"errors"
	"sync"
)





github.com/raahii/golang-grpc-realworld-example/store.mockDB


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
			name: "Valid Article",
			article: &model.Article{
				Title:       "Test Article",
				Description: "This is a test article",
				Body:        "Article body",
				Tags:        []model.Tag{{Name: "test"}},
				Author:      model.User{Username: "testuser"},
				UserID:      1,
			},
			expected: "just for testing",
		},
		{
			name:     "Nil Article",
			article:  nil,
			expected: "just for testing",
		},
		{
			name: "Empty Fields",
			article: &model.Article{
				Title:       "",
				Description: "",
				Body:        "",
			},
			expected: "just for testing",
		},
		{
			name: "Very Long Content",
			article: &model.Article{
				Title:       string(make([]byte, 10000)),
				Description: string(make([]byte, 10000)),
				Body:        string(make([]byte, 10000)),
			},
			expected: "just for testing",
		},
		{
			name: "Special Characters",
			article: &model.Article{
				Title:       "Special 🚀 Characters ©",
				Description: "<html>Description</html>",
				Body:        "Body with Unicode: こんにちは",
			},
			expected: "just for testing",
		},
		{
			name: "Maximum Tags",
			article: &model.Article{
				Title:       "Max Tags",
				Description: "Article with many tags",
				Body:        "Body",
				Tags:        make([]model.Tag, 100),
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
		mockDB  func(interface{}) *gorm.DB
		wantErr bool
	}{
		{
			name: "Successfully Create a Comment",
			comment: &model.Comment{
				Body:      "Test comment",
				UserID:    1,
				ArticleID: 1,
			},
			mockDB: func(value interface{}) *gorm.DB {
				return &gorm.DB{Error: nil}
			},
			wantErr: false,
		},
		{
			name: "Attempt to Create a Comment with Missing Required Fields",
			comment: &model.Comment{
				Body: "",
			},
			mockDB: func(value interface{}) *gorm.DB {
				return &gorm.DB{Error: errors.New("missing required fields")}
			},
			wantErr: true,
		},
		{
			name: "Create Comment with Very Long Body Text",
			comment: &model.Comment{
				Body:      string(make([]byte, 10000)),
				UserID:    1,
				ArticleID: 1,
			},
			mockDB: func(value interface{}) *gorm.DB {
				return &gorm.DB{Error: nil}
			},
			wantErr: false,
		},
		{
			name: "Create Comment with Non-Existent Article ID",
			comment: &model.Comment{
				Body:      "Test comment",
				UserID:    1,
				ArticleID: 9999,
			},
			mockDB: func(value interface{}) *gorm.DB {
				return &gorm.DB{Error: errors.New("foreign key constraint violation")}
			},
			wantErr: true,
		},
		{
			name: "Create Comment When Database is Down",
			comment: &model.Comment{
				Body:      "Test comment",
				UserID:    1,
				ArticleID: 1,
			},
			mockDB: func(value interface{}) *gorm.DB {
				return &gorm.DB{Error: errors.New("database connection lost")}
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockDB := &mockDB{
				createFunc: tt.mockDB,
			}

			store := &ArticleStore{
				db: mockDB,
			}

			err := store.CreateComment(tt.comment)
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
func TestArticleStoreDeleteFavorite(t *testing.T) {
	tests := []struct {
		name           string
		article        *model.Article
		user           *model.User
		mockDB         *mockDB
		expectedError  error
		expectedCount  int32
		expectedCommit bool
	}{
		{
			name: "Successfully Unfavorite an Article",
			article: &model.Article{
				FavoritesCount: 1,
				FavoritedUsers: []model.User{{Model: gorm.Model{ID: 1}}},
			},
			user:           &model.User{Model: gorm.Model{ID: 1}},
			mockDB:         &mockDB{},
			expectedError:  nil,
			expectedCount:  0,
			expectedCommit: true,
		},
		{
			name: "Attempt to Unfavorite an Article That Wasn't Favorited",
			article: &model.Article{
				FavoritesCount: 0,
				FavoritedUsers: []model.User{},
			},
			user:           &model.User{Model: gorm.Model{ID: 1}},
			mockDB:         &mockDB{},
			expectedError:  nil,
			expectedCount:  0,
			expectedCommit: true,
		},
		{
			name: "Database Error During Association Deletion",
			article: &model.Article{
				FavoritesCount: 1,
				FavoritedUsers: []model.User{{Model: gorm.Model{ID: 1}}},
			},
			user: &model.User{Model: gorm.Model{ID: 1}},
			mockDB: &mockDB{
				associationError: errors.New("association deletion error"),
			},
			expectedError:  errors.New("association deletion error"),
			expectedCount:  1,
			expectedCommit: false,
		},
		{
			name: "Database Error During FavoritesCount Update",
			article: &model.Article{
				FavoritesCount: 1,
				FavoritedUsers: []model.User{{Model: gorm.Model{ID: 1}}},
			},
			user: &model.User{Model: gorm.Model{ID: 1}},
			mockDB: &mockDB{
				updateError: errors.New("update error"),
			},
			expectedError:  errors.New("update error"),
			expectedCount:  1,
			expectedCommit: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			store := &ArticleStore{
				db: tt.mockDB,
			}

			err := store.DeleteFavorite(tt.article, tt.user)

			if (err != nil && tt.expectedError == nil) || (err == nil && tt.expectedError != nil) || (err != nil && tt.expectedError != nil && err.Error() != tt.expectedError.Error()) {
				t.Errorf("DeleteFavorite() error = %v, expectedError %v", err, tt.expectedError)
			}

			if tt.article.FavoritesCount != tt.expectedCount {
				t.Errorf("DeleteFavorite() FavoritesCount = %v, expected %v", tt.article.FavoritesCount, tt.expectedCount)
			}

			if tt.mockDB.commitCalled != tt.expectedCommit {
				t.Errorf("DeleteFavorite() commit called = %v, expected %v", tt.mockDB.commitCalled, tt.expectedCommit)
			}

			if tt.expectedError != nil && !tt.mockDB.rollbackCalled {
				t.Error("DeleteFavorite() rollback not called when error occurred")
			}
		})
	}
}

func TestArticleStoreDeleteFavoriteConcurrent(t *testing.T) {
	article := &model.Article{
		FavoritesCount: 10,
		FavoritedUsers: make([]model.User, 10),
	}
	for i := range article.FavoritedUsers {
		article.FavoritedUsers[i] = model.User{Model: gorm.Model{ID: uint(i + 1)}}
	}

	mockDB := &mockDB{}
	store := &ArticleStore{
		db: mockDB,
	}

	var wg sync.WaitGroup
	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			user := &model.User{Model: gorm.Model{ID: uint(i + 1)}}
			err := store.DeleteFavorite(article, user)
			if err != nil {
				t.Errorf("DeleteFavorite() concurrent error = %v", err)
			}
		}(i)
	}
	wg.Wait()

	if article.FavoritesCount != 0 {
		t.Errorf("DeleteFavorite() concurrent FavoritesCount = %v, expected 0", article.FavoritesCount)
	}

	if len(article.FavoritedUsers) != 0 {
		t.Errorf("DeleteFavorite() concurrent FavoritedUsers length = %v, expected 0", len(article.FavoritedUsers))
	}
}


package store

import (
	"errors"
	"testing"
	"github.com/jinzhu/gorm"
	"github.com/raahii/golang-grpc-realworld-example/model"
	"sync"
)





type mockDB struct {
	createFunc func(interface{}) *gorm.DB
}
type mockDB struct {
	*gorm.DB
	mockError     error
	mockAssocErr  error
	mockUpdateErr error
	mu            sync.Mutex
}


/*
ROOST_METHOD_HASH=Create_0a911e138d
ROOST_METHOD_SIG_HASH=Create_723c594377

FUNCTION_DEF=func (s *ArticleStore) Create(m *model.Article) error 

*/
func (m *mockDB) Create(value interface{}) *gorm.DB {
	return m.createFunc(value)
}

func TestCreate(t *testing.T) {
	tests := []struct {
		name    string
		article *model.Article
		dbError error
		wantErr bool
	}{
		{
			name: "Successfully Create a New Article",
			article: &model.Article{
				Title:       "Test Article",
				Description: "Test Description",
				Body:        "Test Body",
				UserID:      1,
			},
			dbError: nil,
			wantErr: false,
		},
		{
			name: "Attempt to Create an Article with Missing Required Fields",
			article: &model.Article{
				Description: "Test Description",
				Body:        "Test Body",
				UserID:      1,
			},
			dbError: errors.New("Title cannot be empty"),
			wantErr: true,
		},
		{
			name: "Create an Article with Associated Tags",
			article: &model.Article{
				Title:       "Test Article with Tags",
				Description: "Test Description",
				Body:        "Test Body",
				UserID:      1,
				Tags: []model.Tag{
					{Name: "Tag1"},
					{Name: "Tag2"},
				},
			},
			dbError: nil,
			wantErr: false,
		},
		{
			name: "Create an Article with an Author",
			article: &model.Article{
				Title:       "Test Article with Author",
				Description: "Test Description",
				Body:        "Test Body",
				UserID:      1,
				Author: model.User{
					Model:    gorm.Model{ID: 1},
					Username: "testuser",
				},
			},
			dbError: nil,
			wantErr: false,
		},
		{
			name: "Handle Database Connection Error",
			article: &model.Article{
				Title:       "Test Article",
				Description: "Test Description",
				Body:        "Test Body",
				UserID:      1,
			},
			dbError: errors.New("database connection error"),
			wantErr: true,
		},
		{
			name: "Create an Article with Maximum Length Content",
			article: &model.Article{
				Title:       string(make([]byte, 255)),
				Description: string(make([]byte, 1000)),
				Body:        string(make([]byte, 10000)),
				UserID:      1,
			},
			dbError: nil,
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mock := &mockDB{
				createFunc: func(value interface{}) *gorm.DB {
					return &gorm.DB{Error: tt.dbError}
				},
			}

			s := &ArticleStore{db: mock}

			err := s.Create(tt.article)

			if (err != nil) != tt.wantErr {
				t.Errorf("Create() error = %v, wantErr %v", err, tt.wantErr)
			}

			if tt.wantErr && err != tt.dbError {
				t.Errorf("Create() error = %v, want %v", err, tt.dbError)
			}
		})
	}
}


/*
ROOST_METHOD_HASH=CreateComment_58d394e2c6
ROOST_METHOD_SIG_HASH=CreateComment_28b95f60a6

FUNCTION_DEF=func (s *ArticleStore) CreateComment(m *model.Comment) error 

*/
func TestArticleStoreCreateComment(t *testing.T) {
	tests := []struct {
		name    string
		comment *model.Comment
		mockDB  func(comment *model.Comment) *mockDB
		wantErr bool
	}{
		{
			name: "Successfully Create a New Comment",
			comment: &model.Comment{
				Body:      "Test comment",
				UserID:    1,
				ArticleID: 1,
			},
			mockDB: func(comment *model.Comment) *mockDB {
				return &mockDB{
					createFunc: func(value interface{}) *gorm.DB {
						return &gorm.DB{Error: nil}
					},
				}
			},
			wantErr: false,
		},
		{
			name: "Fail to Create Comment Due to Database Error",
			comment: &model.Comment{
				Body:      "Test comment",
				UserID:    1,
				ArticleID: 1,
			},
			mockDB: func(comment *model.Comment) *mockDB {
				return &mockDB{
					createFunc: func(value interface{}) *gorm.DB {
						return &gorm.DB{Error: errors.New("database error")}
					},
				}
			},
			wantErr: true,
		},
		{
			name: "Create Comment with Minimum Required Fields",
			comment: &model.Comment{
				Body:      "Minimal comment",
				UserID:    1,
				ArticleID: 1,
			},
			mockDB: func(comment *model.Comment) *mockDB {
				return &mockDB{
					createFunc: func(value interface{}) *gorm.DB {
						return &gorm.DB{Error: nil}
					},
				}
			},
			wantErr: false,
		},
		{
			name: "Attempt to Create Comment with Invalid User ID",
			comment: &model.Comment{
				Body:      "Invalid user comment",
				UserID:    999,
				ArticleID: 1,
			},
			mockDB: func(comment *model.Comment) *mockDB {
				return &mockDB{
					createFunc: func(value interface{}) *gorm.DB {
						return &gorm.DB{Error: errors.New("foreign key constraint violation")}
					},
				}
			},
			wantErr: true,
		},
		{
			name: "Create Comment with Maximum Length Body",
			comment: &model.Comment{
				Body:      string(make([]byte, 1000)),
				UserID:    1,
				ArticleID: 1,
			},
			mockDB: func(comment *model.Comment) *mockDB {
				return &mockDB{
					createFunc: func(value interface{}) *gorm.DB {
						return &gorm.DB{Error: nil}
					},
				}
			},
			wantErr: false,
		},
		{
			name: "Attempt to Create Duplicate Comment",
			comment: &model.Comment{
				Body:      "Duplicate comment",
				UserID:    1,
				ArticleID: 1,
			},
			mockDB: func(comment *model.Comment) *mockDB {
				return &mockDB{
					createFunc: func(value interface{}) *gorm.DB {
						return &gorm.DB{Error: errors.New("unique constraint violation")}
					},
				}
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := &ArticleStore{
				db: tt.mockDB(tt.comment),
			}
			err := s.CreateComment(tt.comment)
			if (err != nil) != tt.wantErr {
				t.Errorf("ArticleStore.CreateComment() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}


/*
ROOST_METHOD_HASH=DeleteFavorite_a856bcbb70
ROOST_METHOD_SIG_HASH=DeleteFavorite_f7e5c0626f

FUNCTION_DEF=func (s *ArticleStore) DeleteFavorite(a *model.Article, u *model.User) error 

*/
func TestArticleStoreDeleteFavorite(t *testing.T) {
	tests := []struct {
		name           string
		article        *model.Article
		user           *model.User
		mockAssocErr   error
		mockUpdateErr  error
		expectedError  error
		expectedCount  int32
		concurrentOps  int
		setupMockDB    func(*mockDB)
		validateResult func(*testing.T, *model.Article, error)
	}{
		{
			name: "Successfully Delete a Favorite",
			article: &model.Article{
				FavoritesCount: 1,
				FavoritedUsers: []model.User{{Model: gorm.Model{ID: 1}}},
			},
			user:          &model.User{Model: gorm.Model{ID: 1}},
			expectedError: nil,
			expectedCount: 0,
			validateResult: func(t *testing.T, a *model.Article, err error) {
				if err != nil {
					t.Errorf("Expected no error, got %v", err)
				}
				if a.FavoritesCount != 0 {
					t.Errorf("Expected FavoritesCount to be 0, got %d", a.FavoritesCount)
				}
				if len(a.FavoritedUsers) != 0 {
					t.Errorf("Expected FavoritedUsers to be empty, got %v", a.FavoritedUsers)
				}
			},
		},
		{
			name: "Delete Favorite for Non-Existent Association",
			article: &model.Article{
				FavoritesCount: 0,
				FavoritedUsers: []model.User{},
			},
			user:          &model.User{Model: gorm.Model{ID: 1}},
			expectedError: nil,
			expectedCount: 0,
			validateResult: func(t *testing.T, a *model.Article, err error) {
				if err != nil {
					t.Errorf("Expected no error, got %v", err)
				}
				if a.FavoritesCount != 0 {
					t.Errorf("Expected FavoritesCount to remain 0, got %d", a.FavoritesCount)
				}
				if len(a.FavoritedUsers) != 0 {
					t.Errorf("Expected FavoritedUsers to remain empty, got %v", a.FavoritedUsers)
				}
			},
		},
		{
			name: "Database Error During Association Deletion",
			article: &model.Article{
				FavoritesCount: 1,
				FavoritedUsers: []model.User{{Model: gorm.Model{ID: 1}}},
			},
			user:          &model.User{Model: gorm.Model{ID: 1}},
			mockAssocErr:  errors.New("association deletion error"),
			expectedError: errors.New("association deletion error"),
			expectedCount: 1,
			setupMockDB: func(m *mockDB) {
				m.mockAssocErr = errors.New("association deletion error")
			},
			validateResult: func(t *testing.T, a *model.Article, err error) {
				if err == nil || err.Error() != "association deletion error" {
					t.Errorf("Expected association deletion error, got %v", err)
				}
				if a.FavoritesCount != 1 {
					t.Errorf("Expected FavoritesCount to remain 1, got %d", a.FavoritesCount)
				}
			},
		},
		{
			name: "Database Error During FavoritesCount Update",
			article: &model.Article{
				FavoritesCount: 1,
				FavoritedUsers: []model.User{{Model: gorm.Model{ID: 1}}},
			},
			user:          &model.User{Model: gorm.Model{ID: 1}},
			mockUpdateErr: errors.New("update error"),
			expectedError: errors.New("update error"),
			expectedCount: 1,
			setupMockDB: func(m *mockDB) {
				m.mockUpdateErr = errors.New("update error")
			},
			validateResult: func(t *testing.T, a *model.Article, err error) {
				if err == nil || err.Error() != "update error" {
					t.Errorf("Expected update error, got %v", err)
				}
				if a.FavoritesCount != 1 {
					t.Errorf("Expected FavoritesCount to remain 1, got %d", a.FavoritesCount)
				}
			},
		},
		{
			name: "Concurrent Deletion of Favorites",
			article: &model.Article{
				FavoritesCount: 5,
				FavoritedUsers: []model.User{
					{Model: gorm.Model{ID: 1}},
					{Model: gorm.Model{ID: 2}},
					{Model: gorm.Model{ID: 3}},
					{Model: gorm.Model{ID: 4}},
					{Model: gorm.Model{ID: 5}},
				},
			},
			user:          &model.User{Model: gorm.Model{ID: 1}},
			expectedError: nil,
			expectedCount: 0,
			concurrentOps: 5,
			validateResult: func(t *testing.T, a *model.Article, err error) {
				if err != nil {
					t.Errorf("Expected no error, got %v", err)
				}
				if a.FavoritesCount != 0 {
					t.Errorf("Expected FavoritesCount to be 0, got %d", a.FavoritesCount)
				}
				if len(a.FavoritedUsers) != 0 {
					t.Errorf("Expected FavoritedUsers to be empty, got %v", a.FavoritedUsers)
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockDB := &mockDB{
				mockAssocErr:  tt.mockAssocErr,
				mockUpdateErr: tt.mockUpdateErr,
			}
			if tt.setupMockDB != nil {
				tt.setupMockDB(mockDB)
			}

			store := &ArticleStore{db: mockDB}

			if tt.concurrentOps > 0 {
				var wg sync.WaitGroup
				for i := 0; i < tt.concurrentOps; i++ {
					wg.Add(1)
					go func(user *model.User) {
						defer wg.Done()
						_ = store.DeleteFavorite(tt.article, user)
					}(&model.User{Model: gorm.Model{ID: uint(i + 1)}})
				}
				wg.Wait()
			} else {
				err := store.DeleteFavorite(tt.article, tt.user)
				tt.validateResult(t, tt.article, err)
			}
		})
	}
}


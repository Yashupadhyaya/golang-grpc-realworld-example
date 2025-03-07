package store

import (
	testing "testing"
	gorm "github.com/jinzhu/gorm"
	model "github.com/raahii/golang-grpc-realworld-example/model"
	errors "errors"
	assert "github.com/stretchr/testify/assert"
	mock "github.com/stretchr/testify/mock"
)





type mockDB struct {
	mock.Mock
}
type mockAssociation struct {
	mock.Mock
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
			name: "Valid Article",
			article: &model.Article{
				Title:       "Test Title",
				Description: "Test Description",
				Body:        "Test Body",
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
				UserID:      1,
			},
			expected: "just for testing",
		},
		{
			name: "Maximum Length Fields",
			article: &model.Article{
				Title:       string(make([]byte, 1000)),
				Description: string(make([]byte, 1000)),
				Body:        string(make([]byte, 1000)),
				UserID:      1,
			},
			expected: "just for testing",
		},
		{
			name: "Article with Tags",
			article: &model.Article{
				Title:       "Test Title",
				Description: "Test Description",
				Body:        "Test Body",
				UserID:      1,
				Tags: []model.Tag{
					{Name: "Tag1"},
					{Name: "Tag2"},
				},
			},
			expected: "just for testing",
		},
		{
			name: "Article with Author",
			article: &model.Article{
				Title:       "Test Title",
				Description: "Test Description",
				Body:        "Test Body",
				UserID:      1,
				Author: model.User{
					Model:    gorm.Model{ID: 1},
					Username: "testuser",
					Email:    "test@example.com",
				},
			},
			expected: "just for testing",
		},
		{
			name: "Article with Pre-existing ID",
			article: &model.Article{
				Model:       gorm.Model{ID: 100},
				Title:       "Test Title",
				Description: "Test Description",
				Body:        "Test Body",
				UserID:      1,
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
ROOST_METHOD_HASH=ArticleStore_Create_1273475ade
ROOST_METHOD_SIG_HASH=ArticleStore_Create_a27282cad5

FUNCTION_DEF=func (s *ArticleStore) Create(m *model.Article) error // Create creates an article


*/
func TestArticleStoreCreate(t *testing.T) {
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
			},
			dbError: nil,
			wantErr: false,
		},
		{
			name: "Attempt to Create an Article with Invalid Data",
			article: &model.Article{
				Description: "Test Description",
				Body:        "Test Body",
			},
			dbError: errors.New("validation error"),
			wantErr: true,
		},
		{
			name: "Database Error During Article Creation",
			article: &model.Article{
				Title:       "Test Article",
				Description: "Test Description",
				Body:        "Test Body",
			},
			dbError: errors.New("database error"),
			wantErr: true,
		},
		{
			name: "Create Article with Associated Tags",
			article: &model.Article{
				Title:       "Test Article with Tags",
				Description: "Test Description",
				Body:        "Test Body",
				Tags: []model.Tag{
					{Name: "Tag1"},
					{Name: "Tag2"},
				},
			},
			dbError: nil,
			wantErr: false,
		},
		{
			name: "Create Article with Maximum Allowed Content",
			article: &model.Article{
				Title:       string(make([]byte, 255)),
				Description: string(make([]byte, 1000)),
				Body:        string(make([]byte, 10000)),
			},
			dbError: nil,
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockDB := new(mockDB)
			mockDB.On("Create", mock.AnythingOfType("*model.Article")).Return(&gorm.DB{Error: tt.dbError})

			store := &ArticleStore{
				db: mockDB,
			}

			err := store.Create(tt.article)

			if tt.wantErr {
				assert.Error(t, err)
				assert.Equal(t, tt.dbError, err)
			} else {
				assert.NoError(t, err)
			}

			mockDB.AssertCalled(t, "Create", tt.article)
		})
	}
}

func TestArticleStoreCreateMultiple(t *testing.T) {
	mockDB := new(mockDB)
	store := &ArticleStore{
		db: mockDB,
	}

	articles := []*model.Article{
		{Title: "Article 1", Description: "Desc 1", Body: "Body 1"},
		{Title: "Article 2", Description: "Desc 2", Body: "Body 2"},
		{Title: "Article 3", Description: "Desc 3", Body: "Body 3"},
	}

	for _, article := range articles {
		mockDB.On("Create", article).Return(&gorm.DB{Error: nil}).Once()
	}

	for _, article := range articles {
		err := store.Create(article)
		assert.NoError(t, err)
	}

	mockDB.AssertNumberOfCalls(t, "Create", len(articles))
}

func (m *mockDB) Create(value interface{}) *gorm.DB {
	args := m.Called(value)
	return args.Get(0).(*gorm.DB)
}


/*
ROOST_METHOD_HASH=ArticleStore_CreateComment_b16d4a71d4
ROOST_METHOD_SIG_HASH=ArticleStore_CreateComment_7475736b06

FUNCTION_DEF=func (s *ArticleStore) CreateComment(m *model.Comment) error // CreateComment creates a comment of the article


*/
func TestArticleStoreCreateComment(t *testing.T) {
	tests := []struct {
		name    string
		comment *model.Comment
		dbError error
		wantErr bool
	}{
		{
			name: "Successfully Create a New Comment",
			comment: &model.Comment{
				Body:      "Test comment",
				UserID:    1,
				ArticleID: 1,
			},
			dbError: nil,
			wantErr: false,
		},
		{
			name: "Attempt to Create a Comment with Invalid Data",
			comment: &model.Comment{
				Body:      "",
				UserID:    1,
				ArticleID: 1,
			},
			dbError: errors.New("validation error"),
			wantErr: true,
		},
		{
			name: "Database Error During Comment Creation",
			comment: &model.Comment{
				Body:      "Test comment",
				UserID:    1,
				ArticleID: 1,
			},
			dbError: errors.New("database error"),
			wantErr: true,
		},
		{
			name: "Create Comment with Maximum Length Content",
			comment: &model.Comment{
				Body:      string(make([]byte, 1000)),
				UserID:    1,
				ArticleID: 1,
			},
			dbError: nil,
			wantErr: false,
		},
		{
			name: "Create Comment for Non-Existent Article",
			comment: &model.Comment{
				Body:      "Test comment",
				UserID:    1,
				ArticleID: 999,
			},
			dbError: errors.New("foreign key constraint violation"),
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockDB := new(mockDB)
			mockDB.On("Create", mock.AnythingOfType("*model.Comment")).Return(&gorm.DB{Error: tt.dbError})

			store := &ArticleStore{
				db: &gorm.DB{},
			}
			store.db = mockDB

			err := store.CreateComment(tt.comment)

			if tt.wantErr {
				assert.Error(t, err)
				assert.Equal(t, tt.dbError, err)
			} else {
				assert.NoError(t, err)
			}

			mockDB.AssertExpectations(t)
		})
	}
}


/*
ROOST_METHOD_HASH=ArticleStore_DeleteFavorite_29c18a04a8
ROOST_METHOD_SIG_HASH=ArticleStore_DeleteFavorite_53deb5e792

FUNCTION_DEF=func (s *ArticleStore) DeleteFavorite(a *model.Article, u *model.User) error // DeleteFavorite unfavorite an article


*/
func TestArticleStoreDeleteFavorite(t *testing.T) {
	tests := []struct {
		name          string
		article       *model.Article
		user          *model.User
		setupMock     func(*mockDB)
		expectedError error
		expectedCount int32
	}{
		{
			name:    "Successfully Delete a Favorite",
			article: &model.Article{FavoritesCount: 2},
			user:    &model.User{},
			setupMock: func(m *mockDB) {
				tx := &gorm.DB{}
				m.On("Begin").Return(tx)
				m.On("Model", mock.Anything).Return(m)
				assoc := &mockAssociation{}
				assoc.On("Delete", mock.Anything).Return(&gorm.Association{})
				m.On("Association", "FavoritedUsers").Return(assoc)
				m.On("Update", "favorites_count", gorm.Expr("favorites_count - ?", 1)).Return(m)
				m.On("Commit").Return(tx)
			},
			expectedError: nil,
			expectedCount: 1,
		},
		{
			name:    "Attempt to Delete a Non-existent Favorite",
			article: &model.Article{FavoritesCount: 1},
			user:    &model.User{},
			setupMock: func(m *mockDB) {
				tx := &gorm.DB{}
				m.On("Begin").Return(tx)
				m.On("Model", mock.Anything).Return(m)
				assoc := &mockAssociation{}
				assoc.On("Delete", mock.Anything).Return(&gorm.Association{})
				m.On("Association", "FavoritedUsers").Return(assoc)
				m.On("Update", "favorites_count", gorm.Expr("favorites_count - ?", 1)).Return(m)
				m.On("Commit").Return(tx)
			},
			expectedError: nil,
			expectedCount: 0,
		},
		{
			name:    "Database Error During Association Deletion",
			article: &model.Article{FavoritesCount: 2},
			user:    &model.User{},
			setupMock: func(m *mockDB) {
				tx := &gorm.DB{}
				m.On("Begin").Return(tx)
				m.On("Model", mock.Anything).Return(m)
				assoc := &mockAssociation{}
				assoc.On("Delete", mock.Anything).Return(&gorm.Association{}).Error = errors.New("association deletion error")
				m.On("Association", "FavoritedUsers").Return(assoc)
				m.On("Rollback").Return(tx)
			},
			expectedError: errors.New("association deletion error"),
			expectedCount: 2,
		},
		{
			name:    "Database Error During FavoritesCount Update",
			article: &model.Article{FavoritesCount: 2},
			user:    &model.User{},
			setupMock: func(m *mockDB) {
				tx := &gorm.DB{}
				m.On("Begin").Return(tx)
				m.On("Model", mock.Anything).Return(m)
				assoc := &mockAssociation{}
				assoc.On("Delete", mock.Anything).Return(&gorm.Association{})
				m.On("Association", "FavoritedUsers").Return(assoc)
				m.On("Update", "favorites_count", gorm.Expr("favorites_count - ?", 1)).Return(m).Error = errors.New("update error")
				m.On("Rollback").Return(tx)
			},
			expectedError: errors.New("update error"),
			expectedCount: 2,
		},
		{
			name:    "Decrementing FavoritesCount to Zero",
			article: &model.Article{FavoritesCount: 1},
			user:    &model.User{},
			setupMock: func(m *mockDB) {
				tx := &gorm.DB{}
				m.On("Begin").Return(tx)
				m.On("Model", mock.Anything).Return(m)
				assoc := &mockAssociation{}
				assoc.On("Delete", mock.Anything).Return(&gorm.Association{})
				m.On("Association", "FavoritedUsers").Return(assoc)
				m.On("Update", "favorites_count", gorm.Expr("favorites_count - ?", 1)).Return(m)
				m.On("Commit").Return(tx)
			},
			expectedError: nil,
			expectedCount: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockDB := new(mockDB)
			tt.setupMock(mockDB)

			store := &ArticleStore{db: mockDB}

			err := store.DeleteFavorite(tt.article, tt.user)

			assert.Equal(t, tt.expectedError, err)
			assert.Equal(t, tt.expectedCount, tt.article.FavoritesCount)

			mockDB.AssertExpectations(t)
		})
	}
}

func (m *mockAssociation) Delete(values ...interface{}) *gorm.Association {
	args := m.Called(values...)
	return args.Get(0).(*gorm.Association)
}

func (m *mockDB) Association(column string) *gorm.Association {
	args := m.Called(column)
	return args.Get(0).(*gorm.Association)
}

func (m *mockDB) Begin() *gorm.DB {
	args := m.Called()
	return args.Get(0).(*gorm.DB)
}

func (m *mockDB) Commit() *gorm.DB {
	args := m.Called()
	return args.Get(0).(*gorm.DB)
}

func (m *mockDB) Model(value interface{}) *gorm.DB {
	args := m.Called(value)
	return args.Get(0).(*gorm.DB)
}

func (m *mockDB) Rollback() *gorm.DB {
	args := m.Called()
	return args.Get(0).(*gorm.DB)
}

func (m *mockDB) Update(column string, value interface{}) *gorm.DB {
	args := m.Called(column, value)
	return args.Get(0).(*gorm.DB)
}


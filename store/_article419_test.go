package store

import (
	"testing"
	"github.com/jinzhu/gorm"
	"github.com/raahii/golang-grpc-realworld-example/model"
	"errors"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"sync"
)





type mockDB struct {
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
			name: "Create Article with Valid Data",
			article: &model.Article{
				Title:       "Valid Article",
				Description: "This is a valid article",
				Body:        "Article body with valid content",
				UserID:      1,
			},
			expected: "just for testing",
		},
		{
			name: "Create Article with Missing Title",
			article: &model.Article{
				Description: "Article with missing title",
				Body:        "Body of the article",
				UserID:      1,
			},
			expected: "just for testing",
		},
		{
			name: "Create Article with Very Long Content",
			article: &model.Article{
				Title:       "Long Content Article",
				Description: "Article with very long body",
				Body:        string(make([]byte, 100000)),
				UserID:      1,
			},
			expected: "just for testing",
		},
		{
			name: "Create Article with Associated Tags",
			article: &model.Article{
				Title:       "Tagged Article",
				Description: "Article with associated tags",
				Body:        "Body of the tagged article",
				UserID:      1,
				Tags: []model.Tag{
					{Name: "tag1"},
					{Name: "tag2"},
				},
			},
			expected: "just for testing",
		},
		{
			name:     "Create Article with Nil Pointer",
			article:  nil,
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

func TestCreateMultipleArticles(t *testing.T) {
	articles := []*model.Article{
		{
			Title:       "First Article",
			Description: "Description of the first article",
			Body:        "Body of the first article",
			UserID:      1,
		},
		{
			Title:       "Second Article",
			Description: "Description of the second article",
			Body:        "Body of the second article",
			UserID:      2,
		},
		{
			Title:       "Third Article",
			Description: "Description of the third article",
			Body:        "Body of the third article",
			UserID:      3,
		},
	}

	for i, article := range articles {
		result := Create(article)
		expected := "just for testing"
		if result != expected {
			t.Errorf("Create() for article %d = %v, want %v", i+1, result, expected)
		}
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
			name: "Successfully Create a Valid Article",
			article: &model.Article{
				Title:       "Test Article",
				Description: "This is a test article",
				Body:        "Article body",
				UserID:      1,
			},
			dbError: nil,
			wantErr: false,
		},
		{
			name: "Attempt to Create an Article with Missing Required Fields",
			article: &model.Article{
				Description: "This is a test article",
				Body:        "Article body",
				UserID:      1,
			},
			dbError: errors.New("validation error: Title cannot be empty"),
			wantErr: true,
		},
		{
			name: "Handle Database Connection Error During Article Creation",
			article: &model.Article{
				Title:       "Test Article",
				Description: "This is a test article",
				Body:        "Article body",
				UserID:      1,
			},
			dbError: errors.New("database connection error"),
			wantErr: true,
		},
		{
			name: "Create an Article with Associated Tags",
			article: &model.Article{
				Title:       "Test Article with Tags",
				Description: "This is a test article with tags",
				Body:        "Article body",
				UserID:      1,
				Tags:        []model.Tag{{Name: "tag1"}, {Name: "tag2"}},
			},
			dbError: nil,
			wantErr: false,
		},
		{
			name: "Attempt to Create a Duplicate Article",
			article: &model.Article{
				Title:       "Duplicate Article",
				Description: "This is a duplicate article",
				Body:        "Article body",
				UserID:      1,
			},
			dbError: errors.New("duplicate entry"),
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockDB := new(mockDB)
			store := &ArticleStore{db: mockDB}

			mockDB.On("Create", mock.AnythingOfType("*model.Article")).Return(&gorm.DB{Error: tt.dbError})

			err := store.Create(tt.article)

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
				Body: "",
			},
			dbError: errors.New("invalid comment data"),
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
			name: "Create Comment with Maximum Allowed Length",
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
				ArticleID: 9999,
			},
			dbError: errors.New("foreign key constraint failed"),
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockDB := new(mockDB)
			store := &ArticleStore{
				db: mockDB,
			}

			mockDB.On("Create", mock.AnythingOfType("*model.Comment")).Return(&gorm.DB{Error: tt.dbError})

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
		setupMock     func(*mockDB, *mockAssociation)
		expectedError error
		expectedCount int32
	}{
		{
			name:    "Successfully Delete a Favorite Article",
			article: &model.Article{FavoritesCount: 1},
			user:    &model.User{},
			setupMock: func(db *mockDB, assoc *mockAssociation) {
				tx := &gorm.DB{}
				db.On("Begin").Return(tx)
				db.On("Model", mock.Anything).Return(tx)
				db.On("Association", "FavoritedUsers").Return(assoc)
				assoc.On("Delete", mock.Anything).Return(&gorm.Association{})
				db.On("Update", "favorites_count", gorm.Expr("favorites_count - ?", 1)).Return(tx)
				db.On("Commit").Return(tx)
			},
			expectedError: nil,
			expectedCount: 0,
		},
		{
			name:    "Attempt to Delete Favorite from an Article with Zero Favorites",
			article: &model.Article{FavoritesCount: 0},
			user:    &model.User{},
			setupMock: func(db *mockDB, assoc *mockAssociation) {
				tx := &gorm.DB{}
				db.On("Begin").Return(tx)
				db.On("Model", mock.Anything).Return(tx)
				db.On("Association", "FavoritedUsers").Return(assoc)
				assoc.On("Delete", mock.Anything).Return(&gorm.Association{})
				db.On("Update", "favorites_count", gorm.Expr("favorites_count - ?", 1)).Return(tx)
				db.On("Commit").Return(tx)
			},
			expectedError: nil,
			expectedCount: 0,
		},
		{
			name:    "Database Error During Association Deletion",
			article: &model.Article{FavoritesCount: 1},
			user:    &model.User{},
			setupMock: func(db *mockDB, assoc *mockAssociation) {
				tx := &gorm.DB{}
				db.On("Begin").Return(tx)
				db.On("Model", mock.Anything).Return(tx)
				db.On("Association", "FavoritedUsers").Return(assoc)
				assoc.On("Delete", mock.Anything).Return(&gorm.Association{Error: errors.New("association error")})
				db.On("Rollback").Return(tx)
			},
			expectedError: errors.New("association error"),
			expectedCount: 1,
		},
		{
			name:    "Database Error During Favorites Count Update",
			article: &model.Article{FavoritesCount: 1},
			user:    &model.User{},
			setupMock: func(db *mockDB, assoc *mockAssociation) {
				tx := &gorm.DB{}
				db.On("Begin").Return(tx)
				db.On("Model", mock.Anything).Return(tx)
				db.On("Association", "FavoritedUsers").Return(assoc)
				assoc.On("Delete", mock.Anything).Return(&gorm.Association{})
				db.On("Update", "favorites_count", gorm.Expr("favorites_count - ?", 1)).Return(tx.AddError(errors.New("update error")))
				db.On("Rollback").Return(tx)
			},
			expectedError: errors.New("update error"),
			expectedCount: 1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockDB := new(mockDB)
			mockAssoc := new(mockAssociation)
			tt.setupMock(mockDB, mockAssoc)
			db := &gorm.DB{Value: mockDB}
			store := &ArticleStore{db: db}
			err := store.DeleteFavorite(tt.article, tt.user)
			assert.Equal(t, tt.expectedError, err)
			assert.Equal(t, tt.expectedCount, tt.article.FavoritesCount)
			mockDB.AssertExpectations(t)
			mockAssoc.AssertExpectations(t)
		})
	}
}

func TestArticleStoreDeleteFavoriteConcurrent(t *testing.T) {
	article := &model.Article{FavoritesCount: 100}
	users := make([]*model.User, 10)
	for i := range users {
		users[i] = &model.User{}
	}

	mockDB := new(mockDB)
	mockAssoc := new(mockAssociation)

	tx := &gorm.DB{}
	mockDB.On("Begin").Return(tx)
	mockDB.On("Model", mock.Anything).Return(tx)
	mockDB.On("Association", "FavoritedUsers").Return(mockAssoc)
	mockAssoc.On("Delete", mock.Anything).Return(&gorm.Association{})
	mockDB.On("Update", "favorites_count", mock.Anything).Return(tx)
	mockDB.On("Commit").Return(tx)

	db := &gorm.DB{Value: mockDB}
	store := &ArticleStore{db: db}

	var wg sync.WaitGroup
	for _, user := range users {
		wg.Add(1)
		go func(u *model.User) {
			defer wg.Done()
			err := store.DeleteFavorite(article, u)
			assert.NoError(t, err)
		}(user)
	}
	wg.Wait()

	assert.Equal(t, int32(90), article.FavoritesCount)
	mockDB.AssertExpectations(t)
	mockAssoc.AssertExpectations(t)
}


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
				UserID:      1,
			},
			expected: "just for testing",
		},
		{
			name: "Article with Empty Title",
			article: &model.Article{
				Description: "This is a test article",
				Body:        "Article body",
				UserID:      1,
			},
			expected: "just for testing",
		},
		{
			name: "Article with Very Long Body",
			article: &model.Article{
				Title:       "Test Article",
				Description: "This is a test article",
				Body:        string(make([]byte, 100000)),
				UserID:      1,
			},
			expected: "just for testing",
		},
		{
			name: "Article with Associated Tags",
			article: &model.Article{
				Title:       "Test Article",
				Description: "This is a test article",
				Body:        "Article body",
				UserID:      1,
				Tags: []model.Tag{
					{Name: "tag1"},
					{Name: "tag2"},
				},
			},
			expected: "just for testing",
		},
		{
			name:     "Nil Article Pointer",
			article:  nil,
			expected: "just for testing",
		},
		{
			name: "Article with Existing Author",
			article: &model.Article{
				Title:       "Test Article",
				Description: "This is a test article",
				Body:        "Article body",
				UserID:      1,
				Author: model.User{
					Model:    gorm.Model{ID: 1},
					Username: "testuser",
					Email:    "test@example.com",
				},
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

func TestCreateMultipleArticles(t *testing.T) {
	articles := []*model.Article{
		{Title: "Article 1", Description: "Description 1", Body: "Body 1", UserID: 1},
		{Title: "Article 2", Description: "Description 2", Body: "Body 2", UserID: 2},
		{Title: "Article 3", Description: "Description 3", Body: "Body 3", UserID: 3},
		{Title: "Article 4", Description: "Description 4", Body: "Body 4", UserID: 4},
		{Title: "Article 5", Description: "Description 5", Body: "Body 5", UserID: 5},
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
ROOST_METHOD_HASH=CreateComment_b16d4a71d4
ROOST_METHOD_SIG_HASH=CreateComment_7475736b06

FUNCTION_DEF=func (s *ArticleStore) CreateComment(m *model.Comment) error // CreateComment creates a comment of the article


*/
func (m *mockDB) Create(value interface{}) *gorm.DB {
	args := m.Called(value)
	return args.Get(0).(*gorm.DB)
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
				UserID:    1,
				ArticleID: 1,
			},
			dbError: nil,
			wantErr: false,
		},
		{
			name: "Fail to Create Comment Due to Database Error",
			comment: &model.Comment{
				Body:      "Test comment",
				UserID:    1,
				ArticleID: 1,
			},
			dbError: errors.New("database error"),
			wantErr: true,
		},
		{
			name: "Create Comment with Minimum Required Fields",
			comment: &model.Comment{
				Body:      "Minimal comment",
				UserID:    1,
				ArticleID: 1,
			},
			dbError: nil,
			wantErr: false,
		},
		{
			name: "Attempt to Create Comment with Invalid Data",
			comment: &model.Comment{
				Body:      "",
				UserID:    1,
				ArticleID: 1,
			},
			dbError: errors.New("validation error"),
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
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockDB := new(mockDB)
			store := &ArticleStore{db: mockDB}

			mockDB.On("Create", mock.AnythingOfType("*model.Comment")).Return(&gorm.DB{Error: tt.dbError})

			err := store.CreateComment(tt.comment)

			if tt.wantErr {
				assert.Error(t, err)
				assert.Equal(t, tt.dbError, err)
			} else {
				assert.NoError(t, err)
			}

			mockDB.AssertCalled(t, "Create", tt.comment)
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
			name:    "Failure in Removing User from FavoritedUsers",
			article: &model.Article{FavoritesCount: 1},
			user:    &model.User{},
			setupMock: func(db *mockDB, assoc *mockAssociation) {
				tx := &gorm.DB{}
				db.On("Begin").Return(tx)
				db.On("Model", mock.Anything).Return(tx)
				db.On("Association", "FavoritedUsers").Return(assoc)
				assoc.On("Delete", mock.Anything).Return(&gorm.Association{Error: errors.New("association delete error")})
				db.On("Rollback").Return(tx)
			},
			expectedError: errors.New("association delete error"),
			expectedCount: 1,
		},
		{
			name:    "Failure in Updating FavoritesCount",
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
		{
			name:    "DeleteFavorite with Zero FavoritesCount",
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
	article := &model.Article{FavoritesCount: 5}
	users := []*model.User{{}, {}, {}, {}, {}}

	mockDB := new(mockDB)
	mockAssoc := new(mockAssociation)

	tx := &gorm.DB{}
	mockDB.On("Begin").Return(tx).Times(5)
	mockDB.On("Model", mock.Anything).Return(tx).Times(10)
	mockDB.On("Association", "FavoritedUsers").Return(mockAssoc).Times(5)
	mockAssoc.On("Delete", mock.Anything).Return(&gorm.Association{}).Times(5)
	mockDB.On("Update", "favorites_count", gorm.Expr("favorites_count - ?", 1)).Return(tx).Times(5)
	mockDB.On("Commit").Return(tx).Times(5)

	db := &gorm.DB{Value: mockDB}
	store := &ArticleStore{db: db}

	var wg sync.WaitGroup
	wg.Add(5)

	for i := 0; i < 5; i++ {
		go func(index int) {
			defer wg.Done()
			err := store.DeleteFavorite(article, users[index])
			assert.NoError(t, err)
		}(i)
	}

	wg.Wait()

	assert.Equal(t, int32(0), article.FavoritesCount)
	mockDB.AssertExpectations(t)
	mockAssoc.AssertExpectations(t)
}


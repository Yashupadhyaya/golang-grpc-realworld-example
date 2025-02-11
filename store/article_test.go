package store

import (
	"testing"
	"github.com/jinzhu/gorm"
	"github.com/raahii/golang-grpc-realworld-example/model"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)





type mockAssociation struct {
	mock.Mock
}
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
			name: "Article with Very Long Description",
			article: &model.Article{
				Title:       "Test Article",
				Description: string(make([]byte, 10000)),
				Body:        "Article body",
				UserID:      1,
			},
			expected: "just for testing",
		},
		{
			name:     "Nil Article Pointer",
			article:  nil,
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
		{
			Title:       "Article 1",
			Description: "Description 1",
			Body:        "Body 1",
			UserID:      1,
		},
		{
			Title:       "Article 2",
			Description: "Description 2",
			Body:        "Body 2",
			UserID:      2,
		},
		{
			Title:       "Article 3",
			Description: "Description 3",
			Body:        "Body 3",
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
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockDB := new(mockDB)
			mockAssoc := new(mockAssociation)
			tt.setupMock(mockDB, mockAssoc)

			db := &gorm.DB{
				Value: mockDB,
			}

			store := &ArticleStore{db: db}
			err := store.DeleteFavorite(tt.article, tt.user)

			assert.Equal(t, tt.expectedError, err)
			assert.Equal(t, tt.expectedCount, tt.article.FavoritesCount)

			mockDB.AssertExpectations(t)
			mockAssoc.AssertExpectations(t)
		})
	}
}


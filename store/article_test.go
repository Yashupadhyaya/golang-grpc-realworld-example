package store

import (
	"errors"
	"testing"
	"time"
	"github.com/jinzhu/gorm"
	"github.com/raahii/golang-grpc-realworld-example/model"
	"github.com/stretchr/testify/assert"
)









/*
ROOST_METHOD_HASH=GetArticles_6382a4fe7a
ROOST_METHOD_SIG_HASH=GetArticles_1a0b3b0e8b

FUNCTION_DEF=func (s *ArticleStore) GetArticles(tagName, username string, favoritedBy *model.User, limit, offset int64) ([]model.Article, error) 

 */
func TestArticleStoreGetArticles(t *testing.T) {

	mockDB := &gorm.DB{}

	tests := []struct {
		name        string
		tagName     string
		username    string
		favoritedBy *model.User
		limit       int64
		offset      int64
		mockSetup   func(*gorm.DB)
		expected    []model.Article
		expectedErr error
	}{
		{
			name:        "Retrieve Articles Without Any Filters",
			tagName:     "",
			username:    "",
			favoritedBy: nil,
			limit:       10,
			offset:      0,
			mockSetup: func(db *gorm.DB) {
				db.AddError(nil)

			},
			expected: []model.Article{
				{Model: gorm.Model{ID: 1}, Title: "Article 1"},
				{Model: gorm.Model{ID: 2}, Title: "Article 2"},
			},
			expectedErr: nil,
		},
		{
			name:        "Filter Articles by Tag Name",
			tagName:     "golang",
			username:    "",
			favoritedBy: nil,
			limit:       10,
			offset:      0,
			mockSetup: func(db *gorm.DB) {
				db.AddError(nil)

			},
			expected: []model.Article{
				{Model: gorm.Model{ID: 1}, Title: "Golang Article"},
			},
			expectedErr: nil,
		},
		{
			name:        "Filter Articles by Author Username",
			tagName:     "",
			username:    "johndoe",
			favoritedBy: nil,
			limit:       10,
			offset:      0,
			mockSetup: func(db *gorm.DB) {
				db.AddError(nil)

			},
			expected: []model.Article{
				{Model: gorm.Model{ID: 2}, Title: "John's Article"},
			},
			expectedErr: nil,
		},
		{
			name:        "Retrieve Favorited Articles",
			tagName:     "",
			username:    "",
			favoritedBy: &model.User{Model: gorm.Model{ID: 1}},
			limit:       10,
			offset:      0,
			mockSetup: func(db *gorm.DB) {
				db.AddError(nil)

			},
			expected: []model.Article{
				{Model: gorm.Model{ID: 3}, Title: "Favorited Article"},
			},
			expectedErr: nil,
		},
		{
			name:        "Test Pagination",
			tagName:     "",
			username:    "",
			favoritedBy: nil,
			limit:       2,
			offset:      2,
			mockSetup: func(db *gorm.DB) {
				db.AddError(nil)

			},
			expected: []model.Article{
				{Model: gorm.Model{ID: 3}, Title: "Article 3"},
				{Model: gorm.Model{ID: 4}, Title: "Article 4"},
			},
			expectedErr: nil,
		},
		{
			name:        "Combine Multiple Filters",
			tagName:     "golang",
			username:    "johndoe",
			favoritedBy: &model.User{Model: gorm.Model{ID: 1}},
			limit:       10,
			offset:      0,
			mockSetup: func(db *gorm.DB) {
				db.AddError(nil)

			},
			expected: []model.Article{
				{Model: gorm.Model{ID: 5}, Title: "John's Favorited Golang Article"},
			},
			expectedErr: nil,
		},
		{
			name:        "Handle Empty Result Set",
			tagName:     "nonexistent",
			username:    "",
			favoritedBy: nil,
			limit:       10,
			offset:      0,
			mockSetup: func(db *gorm.DB) {
				db.AddError(nil)

			},
			expected:    []model.Article{},
			expectedErr: nil,
		},
		{
			name:        "Error Handling for Database Issues",
			tagName:     "",
			username:    "",
			favoritedBy: nil,
			limit:       10,
			offset:      0,
			mockSetup: func(db *gorm.DB) {
				db.AddError(errors.New("database error"))
			},
			expected:    []model.Article{},
			expectedErr: errors.New("database error"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			mockDB = &gorm.DB{}
			tt.mockSetup(mockDB)

			store := &ArticleStore{db: mockDB}
			articles, err := store.GetArticles(tt.tagName, tt.username, tt.favoritedBy, tt.limit, tt.offset)

			assert.Equal(t, tt.expected, articles)
			assert.Equal(t, tt.expectedErr, err)
		})
	}
}


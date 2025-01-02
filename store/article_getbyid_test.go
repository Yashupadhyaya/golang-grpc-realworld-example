package store

import (
	"errors"
	"testing"
	"github.com/jinzhu/gorm"
	"github.com/raahii/golang-grpc-realworld-example/model"
	"github.com/stretchr/testify/assert"
)




type T struct {
	common
	isEnvSet bool
	context  *testContext // For running tests and subtests.
}
func TestArticleStoreGetByID(t *testing.T) {
	tests := []struct {
		name            string
		setupDB         func(*gorm.DB)
		inputID         uint
		expectedError   error
		expectedArticle *model.Article
	}{
		{
			name: "Successfully retrieve an existing article by ID",
			setupDB: func(db *gorm.DB) {
				author := model.User{ID: 1, Username: "testuser"}
				db.Create(&author)
				article := model.Article{
					ID:          1,
					Title:       "Test Article",
					Description: "Test Description",
					Body:        "Test Body",
					AuthorID:    1,
					Tags: []model.Tag{
						{ID: 1, Name: "tag1"},
						{ID: 2, Name: "tag2"},
					},
				}
				db.Create(&article)
			},
			inputID:       1,
			expectedError: nil,
			expectedArticle: &model.Article{
				ID:          1,
				Title:       "Test Article",
				Description: "Test Description",
				Body:        "Test Body",
				AuthorID:    1,
				Author:      model.User{ID: 1, Username: "testuser"},
				Tags: []model.Tag{
					{ID: 1, Name: "tag1"},
					{ID: 2, Name: "tag2"},
				},
			},
		},
		{
			name: "Attempt to retrieve a non-existent article",
			setupDB: func(db *gorm.DB) {

			},
			inputID:         999,
			expectedError:   gorm.ErrRecordNotFound,
			expectedArticle: nil,
		},
		{
			name: "Retrieve an article with no associated tags",
			setupDB: func(db *gorm.DB) {
				author := model.User{ID: 2, Username: "taglessuser"}
				db.Create(&author)
				article := model.Article{
					ID:          2,
					Title:       "Tagless Article",
					Description: "No Tags",
					Body:        "This article has no tags",
					AuthorID:    2,
				}
				db.Create(&article)
			},
			inputID:       2,
			expectedError: nil,
			expectedArticle: &model.Article{
				ID:          2,
				Title:       "Tagless Article",
				Description: "No Tags",
				Body:        "This article has no tags",
				AuthorID:    2,
				Author:      model.User{ID: 2, Username: "taglessuser"},
				Tags:        []model.Tag{},
			},
		},
		{
			name: "Retrieve an article with multiple tags",
			setupDB: func(db *gorm.DB) {
				author := model.User{ID: 3, Username: "multitaguser"}
				db.Create(&author)
				article := model.Article{
					ID:          3,
					Title:       "Multi-tag Article",
					Description: "Many Tags",
					Body:        "This article has many tags",
					AuthorID:    3,
					Tags: []model.Tag{
						{ID: 3, Name: "tag3"},
						{ID: 4, Name: "tag4"},
						{ID: 5, Name: "tag5"},
					},
				}
				db.Create(&article)
			},
			inputID:       3,
			expectedError: nil,
			expectedArticle: &model.Article{
				ID:          3,
				Title:       "Multi-tag Article",
				Description: "Many Tags",
				Body:        "This article has many tags",
				AuthorID:    3,
				Author:      model.User{ID: 3, Username: "multitaguser"},
				Tags: []model.Tag{
					{ID: 3, Name: "tag3"},
					{ID: 4, Name: "tag4"},
					{ID: 5, Name: "tag5"},
				},
			},
		},
		{
			name: "Database connection error",
			setupDB: func(db *gorm.DB) {

				db.Close()
			},
			inputID:         1,
			expectedError:   errors.New("sql: database is closed"),
			expectedArticle: nil,
		},
		{
			name: "Retrieve an article with a high ID value",
			setupDB: func(db *gorm.DB) {
				author := model.User{ID: 4, Username: "highiduser"}
				db.Create(&author)
				article := model.Article{
					ID:          ^uint(0),
					Title:       "High ID Article",
					Description: "Article with high ID",
					Body:        "This article has a very high ID",
					AuthorID:    4,
				}
				db.Create(&article)
			},
			inputID:       ^uint(0),
			expectedError: nil,
			expectedArticle: &model.Article{
				ID:          ^uint(0),
				Title:       "High ID Article",
				Description: "Article with high ID",
				Body:        "This article has a very high ID",
				AuthorID:    4,
				Author:      model.User{ID: 4, Username: "highiduser"},
				Tags:        []model.Tag{},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			db, err := gorm.Open("sqlite3", ":memory:")
			assert.NoError(t, err)
			defer db.Close()

			db.AutoMigrate(&model.User{}, &model.Article{}, &model.Tag{})

			tt.setupDB(db)

			store := &ArticleStore{db: db}

			article, err := store.GetByID(tt.inputID)

			if tt.expectedError != nil {
				assert.Error(t, err)
				assert.Equal(t, tt.expectedError.Error(), err.Error())
			} else {
				assert.NoError(t, err)
			}

			assert.Equal(t, tt.expectedArticle, article)
		})
	}
}

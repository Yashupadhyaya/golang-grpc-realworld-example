package store

import (
	"errors"
	"testing"
	"github.com/jinzhu/gorm"
	"github.com/raahii/golang-grpc-realworld-example/model"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)





type MockDB struct {
	mock.Mock
}


/*
ROOST_METHOD_HASH=GetByID_36e92ad6eb
ROOST_METHOD_SIG_HASH=GetByID_9616e43e52


 */
func (m *MockDB) Find(out interface{}, where ...interface{}) *gorm.DB {
	args := m.Called(out, where)
	return args.Get(0).(*gorm.DB)
}

func (m *MockDB) Preload(column string, conditions ...interface{}) *gorm.DB {
	args := m.Called(column, conditions)
	return args.Get(0).(*gorm.DB)
}

func TestArticleStoreGetByID(t *testing.T) {
	tests := []struct {
		name            string
		id              uint
		setupMock       func(*MockDB)
		expectedError   error
		expectedArticle *model.Article
	}{
		{
			name: "Successfully retrieve an existing article by ID",
			id:   1,
			setupMock: func(mockDB *MockDB) {
				mockDB.On("Preload", "Tags").Return(mockDB)
				mockDB.On("Preload", "Author").Return(mockDB)
				mockDB.On("Find", mock.AnythingOfType("*model.Article"), uint(1)).Run(func(args mock.Arguments) {
					arg := args.Get(0).(*model.Article)
					*arg = model.Article{
						Model:       gorm.Model{ID: 1},
						Title:       "Test Article",
						Description: "Test Description",
						Body:        "Test Body",
						Tags:        []model.Tag{{Name: "test"}},
						Author:      model.User{Model: gorm.Model{ID: 1}, Username: "testuser"},
					}
				}).Return(&gorm.DB{Error: nil})
			},
			expectedError: nil,
			expectedArticle: &model.Article{
				Model:       gorm.Model{ID: 1},
				Title:       "Test Article",
				Description: "Test Description",
				Body:        "Test Body",
				Tags:        []model.Tag{{Name: "test"}},
				Author:      model.User{Model: gorm.Model{ID: 1}, Username: "testuser"},
			},
		},
		{
			name: "Attempt to retrieve a non-existent article",
			id:   999,
			setupMock: func(mockDB *MockDB) {
				mockDB.On("Preload", "Tags").Return(mockDB)
				mockDB.On("Preload", "Author").Return(mockDB)
				mockDB.On("Find", mock.AnythingOfType("*model.Article"), uint(999)).Return(&gorm.DB{Error: gorm.ErrRecordNotFound})
			},
			expectedError:   gorm.ErrRecordNotFound,
			expectedArticle: nil,
		},
		{
			name: "Retrieve an article with no associated tags",
			id:   2,
			setupMock: func(mockDB *MockDB) {
				mockDB.On("Preload", "Tags").Return(mockDB)
				mockDB.On("Preload", "Author").Return(mockDB)
				mockDB.On("Find", mock.AnythingOfType("*model.Article"), uint(2)).Run(func(args mock.Arguments) {
					arg := args.Get(0).(*model.Article)
					*arg = model.Article{
						Model:       gorm.Model{ID: 2},
						Title:       "Tagless Article",
						Description: "No Tags",
						Body:        "This article has no tags",
						Tags:        []model.Tag{},
						Author:      model.User{Model: gorm.Model{ID: 1}, Username: "testuser"},
					}
				}).Return(&gorm.DB{Error: nil})
			},
			expectedError: nil,
			expectedArticle: &model.Article{
				Model:       gorm.Model{ID: 2},
				Title:       "Tagless Article",
				Description: "No Tags",
				Body:        "This article has no tags",
				Tags:        []model.Tag{},
				Author:      model.User{Model: gorm.Model{ID: 1}, Username: "testuser"},
			},
		},
		{
			name: "Retrieve an article with multiple tags",
			id:   3,
			setupMock: func(mockDB *MockDB) {
				mockDB.On("Preload", "Tags").Return(mockDB)
				mockDB.On("Preload", "Author").Return(mockDB)
				mockDB.On("Find", mock.AnythingOfType("*model.Article"), uint(3)).Run(func(args mock.Arguments) {
					arg := args.Get(0).(*model.Article)
					*arg = model.Article{
						Model:       gorm.Model{ID: 3},
						Title:       "Multi-tag Article",
						Description: "Multiple Tags",
						Body:        "This article has multiple tags",
						Tags:        []model.Tag{{Name: "tag1"}, {Name: "tag2"}, {Name: "tag3"}},
						Author:      model.User{Model: gorm.Model{ID: 1}, Username: "testuser"},
					}
				}).Return(&gorm.DB{Error: nil})
			},
			expectedError: nil,
			expectedArticle: &model.Article{
				Model:       gorm.Model{ID: 3},
				Title:       "Multi-tag Article",
				Description: "Multiple Tags",
				Body:        "This article has multiple tags",
				Tags:        []model.Tag{{Name: "tag1"}, {Name: "tag2"}, {Name: "tag3"}},
				Author:      model.User{Model: gorm.Model{ID: 1}, Username: "testuser"},
			},
		},
		{
			name: "Handle database connection error",
			id:   4,
			setupMock: func(mockDB *MockDB) {
				mockDB.On("Preload", "Tags").Return(mockDB)
				mockDB.On("Preload", "Author").Return(mockDB)
				mockDB.On("Find", mock.AnythingOfType("*model.Article"), uint(4)).Return(&gorm.DB{Error: errors.New("database connection error")})
			},
			expectedError:   errors.New("database connection error"),
			expectedArticle: nil,
		},
		{
			name: "Retrieve an article with high favorites count",
			id:   5,
			setupMock: func(mockDB *MockDB) {
				mockDB.On("Preload", "Tags").Return(mockDB)
				mockDB.On("Preload", "Author").Return(mockDB)
				mockDB.On("Find", mock.AnythingOfType("*model.Article"), uint(5)).Run(func(args mock.Arguments) {
					arg := args.Get(0).(*model.Article)
					*arg = model.Article{
						Model:          gorm.Model{ID: 5},
						Title:          "Popular Article",
						Description:    "Highly Favorited",
						Body:           "This article has many favorites",
						Tags:           []model.Tag{{Name: "popular"}},
						Author:         model.User{Model: gorm.Model{ID: 1}, Username: "testuser"},
						FavoritesCount: 1000000,
					}
				}).Return(&gorm.DB{Error: nil})
			},
			expectedError: nil,
			expectedArticle: &model.Article{
				Model:          gorm.Model{ID: 5},
				Title:          "Popular Article",
				Description:    "Highly Favorited",
				Body:           "This article has many favorites",
				Tags:           []model.Tag{{Name: "popular"}},
				Author:         model.User{Model: gorm.Model{ID: 1}, Username: "testuser"},
				FavoritesCount: 1000000,
			},
		},
		{
			name: "Retrieve an article with a deleted author",
			id:   6,
			setupMock: func(mockDB *MockDB) {
				mockDB.On("Preload", "Tags").Return(mockDB)
				mockDB.On("Preload", "Author").Return(mockDB)
				mockDB.On("Find", mock.AnythingOfType("*model.Article"), uint(6)).Run(func(args mock.Arguments) {
					arg := args.Get(0).(*model.Article)
					*arg = model.Article{
						Model:       gorm.Model{ID: 6},
						Title:       "Orphan Article",
						Description: "Article with deleted author",
						Body:        "This article's author has been deleted",
						Tags:        []model.Tag{{Name: "orphan"}},
						Author:      model.User{},
					}
				}).Return(&gorm.DB{Error: nil})
			},
			expectedError: nil,
			expectedArticle: &model.Article{
				Model:       gorm.Model{ID: 6},
				Title:       "Orphan Article",
				Description: "Article with deleted author",
				Body:        "This article's author has been deleted",
				Tags:        []model.Tag{{Name: "orphan"}},
				Author:      model.User{},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockDB := new(MockDB)
			tt.setupMock(mockDB)

			store := &ArticleStore{db: mockDB}
			article, err := store.GetByID(tt.id)

			if tt.expectedError != nil {
				assert.Error(t, err)
				assert.Equal(t, tt.expectedError.Error(), err.Error())
			} else {
				assert.NoError(t, err)
			}

			assert.Equal(t, tt.expectedArticle, article)
			mockDB.AssertExpectations(t)
		})
	}
}


package store

import (
		"errors"
		"testing"
		"time"
		"github.com/jinzhu/gorm"
		"github.com/raahii/golang-grpc-realworld-example/model"
		"github.com/stretchr/testify/assert"
		"github.com/stretchr/testify/mock"
)

type DBInterface interface {
	Preload(column string, conditions ...interface{}) DBInterface
	Find(out interface{}, where ...interface{}) DBInterface
}
type MockDB struct {
	mock.Mock
}
type Article struct {
	gorm.Model
	Title          string `gorm:"not null"`
	Description    string `gorm:"not null"`
	Body           string `gorm:"not null"`
	Tags           []Tag  `gorm:"many2many:article_tags"`
	Author         User   `gorm:"foreignkey:UserID"`
	UserID         uint   `gorm:"not null"`
	FavoritesCount int32  `gorm:"not null;default=0"`
	FavoritedUsers []User `gorm:"many2many:favorite_articles"`
	Comments       []Comment
}
type ArticleStore struct {
	db *gorm.DB
}
type Call struct {
	Parent *Mock

	// The name of the method that was or will be called.
	Method string

	// Holds the arguments of the method.
	Arguments Arguments

	// Holds the arguments that should be returned when
	// this method is called.
	ReturnArguments Arguments

	// Holds the caller info for the On() call
	callerInfo []string

	// The number of times to return the return arguments when setting
	// expectations. 0 means to always return the value.
	Repeatability int

	// Amount of times this call has been called
	totalCalls int

	// Call to this method can be optional
	optional bool

	// Holds a channel that will be used to block the Return until it either
	// receives a message or is closed. nil means it returns immediately.
	WaitFor <-chan time.Time

	waitTime time.Duration

	// Holds a handler used to manipulate arguments content that are passed by
	// reference. It's useful when mocking methods such as unmarshalers or
	// decoders.
	RunFn func(Arguments)

	// PanicMsg holds msg to be used to mock panic on the function call
	//  if the PanicMsg is set to a non nil string the function call will panic
	// irrespective of other settings
	PanicMsg *string

	// Calls which must be satisfied before this call can be
	requires []*Call
}
type T struct {
	common
	isEnvSet bool
	context  *testContext // For running tests and subtests.
}
/*
ROOST_METHOD_HASH=GetByID_36e92ad6eb
ROOST_METHOD_SIG_HASH=GetByID_9616e43e52


 */
func (m *MockDB) Find(out interface{}, where ...interface{}) DBInterface {
	args := m.Called(out, where)
	return args.Get(0).(DBInterface)
}

func (s *ArticleStore) GetByID(id uint) (*model.Article, error) {
	var m model.Article
	err := s.db.Preload("Tags").Preload("Author").Find(&m, id).(*MockDB).Error
	if err != nil {
		return nil, err
	}
	return &m, nil
}

func (m *MockDB) Preload(column string, conditions ...interface{}) DBInterface {
	args := m.Called(column, conditions)
	return args.Get(0).(DBInterface)
}

func TestArticleStoreGetByID(t *testing.T) {
	tests := []struct {
		name            string
		id              uint
		mockSetup       func(*MockDB)
		expectedError   error
		expectedArticle *model.Article
	}{
		{
			name: "Successfully retrieve an existing article by ID",
			id:   1,
			mockSetup: func(mockDB *MockDB) {
				mockDB.On("Preload", "Tags").Return(mockDB)
				mockDB.On("Preload", "Author").Return(mockDB)
				mockDB.On("Find", mock.AnythingOfType("*model.Article"), uint(1)).Run(func(args mock.Arguments) {
					arg := args.Get(0).(*model.Article)
					*arg = model.Article{
						Model:          gorm.Model{ID: 1, CreatedAt: time.Now(), UpdatedAt: time.Now()},
						Title:          "Test Article",
						Description:    "Test Description",
						Body:           "Test Body",
						Tags:           []model.Tag{{Model: gorm.Model{ID: 1}, Name: "test"}},
						Author:         model.User{Model: gorm.Model{ID: 1}, Username: "testuser"},
						UserID:         1,
						FavoritesCount: 5,
						Comments:       []model.Comment{{Model: gorm.Model{ID: 1}, Body: "Test Comment"}},
					}
				}).Return(mockDB)
			},
			expectedError: nil,
			expectedArticle: &model.Article{
				Model:          gorm.Model{ID: 1},
				Title:          "Test Article",
				Description:    "Test Description",
				Body:           "Test Body",
				Tags:           []model.Tag{{Model: gorm.Model{ID: 1}, Name: "test"}},
				Author:         model.User{Model: gorm.Model{ID: 1}, Username: "testuser"},
				UserID:         1,
				FavoritesCount: 5,
				Comments:       []model.Comment{{Model: gorm.Model{ID: 1}, Body: "Test Comment"}},
			},
		},
		{
			name: "Attempt to retrieve a non-existent article",
			id:   999,
			mockSetup: func(mockDB *MockDB) {
				mockDB.On("Preload", "Tags").Return(mockDB)
				mockDB.On("Preload", "Author").Return(mockDB)
				mockDB.On("Find", mock.AnythingOfType("*model.Article"), uint(999)).Return(mockDB).Run(func(args mock.Arguments) {
					mockDB.Error = gorm.ErrRecordNotFound
				})
			},
			expectedError:   gorm.ErrRecordNotFound,
			expectedArticle: nil,
		},
		{
			name: "Handle database connection error",
			id:   1,
			mockSetup: func(mockDB *MockDB) {
				mockDB.On("Preload", "Tags").Return(mockDB)
				mockDB.On("Preload", "Author").Return(mockDB)
				mockDB.On("Find", mock.AnythingOfType("*model.Article"), uint(1)).Return(mockDB).Run(func(args mock.Arguments) {
					mockDB.Error = errors.New("database connection error")
				})
			},
			expectedError:   errors.New("database connection error"),
			expectedArticle: nil,
		},
		{
			name: "Retrieve article with no associated tags",
			id:   2,
			mockSetup: func(mockDB *MockDB) {
				mockDB.On("Preload", "Tags").Return(mockDB)
				mockDB.On("Preload", "Author").Return(mockDB)
				mockDB.On("Find", mock.AnythingOfType("*model.Article"), uint(2)).Run(func(args mock.Arguments) {
					arg := args.Get(0).(*model.Article)
					*arg = model.Article{
						Model:          gorm.Model{ID: 2, CreatedAt: time.Now(), UpdatedAt: time.Now()},
						Title:          "Article without Tags",
						Description:    "No Tags",
						Body:           "This article has no tags",
						Tags:           []model.Tag{},
						Author:         model.User{Model: gorm.Model{ID: 1}, Username: "testuser"},
						UserID:         1,
						FavoritesCount: 0,
					}
				}).Return(mockDB)
			},
			expectedError: nil,
			expectedArticle: &model.Article{
				Model:          gorm.Model{ID: 2},
				Title:          "Article without Tags",
				Description:    "No Tags",
				Body:           "This article has no tags",
				Tags:           []model.Tag{},
				Author:         model.User{Model: gorm.Model{ID: 1}, Username: "testuser"},
				UserID:         1,
				FavoritesCount: 0,
			},
		},
		{
			name: "Retrieve article with multiple tags",
			id:   3,
			mockSetup: func(mockDB *MockDB) {
				mockDB.On("Preload", "Tags").Return(mockDB)
				mockDB.On("Preload", "Author").Return(mockDB)
				mockDB.On("Find", mock.AnythingOfType("*model.Article"), uint(3)).Run(func(args mock.Arguments) {
					arg := args.Get(0).(*model.Article)
					*arg = model.Article{
						Model:          gorm.Model{ID: 3, CreatedAt: time.Now(), UpdatedAt: time.Now()},
						Title:          "Multi-tag Article",
						Description:    "Article with multiple tags",
						Body:           "This article has multiple tags",
						Tags:           []model.Tag{{Model: gorm.Model{ID: 1}, Name: "tag1"}, {Model: gorm.Model{ID: 2}, Name: "tag2"}, {Model: gorm.Model{ID: 3}, Name: "tag3"}},
						Author:         model.User{Model: gorm.Model{ID: 1}, Username: "testuser"},
						UserID:         1,
						FavoritesCount: 10,
					}
				}).Return(mockDB)
			},
			expectedError: nil,
			expectedArticle: &model.Article{
				Model:          gorm.Model{ID: 3},
				Title:          "Multi-tag Article",
				Description:    "Article with multiple tags",
				Body:           "This article has multiple tags",
				Tags:           []model.Tag{{Model: gorm.Model{ID: 1}, Name: "tag1"}, {Model: gorm.Model{ID: 2}, Name: "tag2"}, {Model: gorm.Model{ID: 3}, Name: "tag3"}},
				Author:         model.User{Model: gorm.Model{ID: 1}, Username: "testuser"},
				UserID:         1,
				FavoritesCount: 10,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockDB := new(MockDB)
			tt.mockSetup(mockDB)

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


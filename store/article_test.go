package store

import (
		"errors"
		"testing"
		"github.com/jinzhu/gorm"
		"github.com/raahii/golang-grpc-realworld-example/model"
		"github.com/stretchr/testify/assert"
		"github.com/stretchr/testify/mock"
)

type DB struct {
	sync.RWMutex
	Value        interface{}
	Error        error
	RowsAffected int64

	// single db
	db                SQLCommon
	blockGlobalUpdate bool
	logMode           logModeValue
	logger            logger
	search            *search
	values            sync.Map

	// global db
	parent        *DB
	callbacks     *Callback
	dialect       Dialect
	singularTable bool

	// function to be used to override the creating of a new timestamp
	nowFuncOverride func() time.Time
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
type mockDB struct {
	mock.Mock
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
func (m *mockDB) Find(out interface{}, where ...interface{}) *gorm.DB {
	args := m.Called(out, where)
	return args.Get(0).(*gorm.DB)
}

func (m *mockDB) Preload(column string, conditions ...interface{}) *gorm.DB {
	args := m.Called(column, conditions)
	return args.Get(0).(*gorm.DB)
}

func TestArticleStoreGetByID(t *testing.T) {
	tests := []struct {
		name            string
		id              uint
		setupMock       func(*mockDB)
		expectedError   error
		expectedArticle *model.Article
	}{
		{
			name: "Successfully retrieve an existing article",
			id:   1,
			setupMock: func(m *mockDB) {
				m.On("Preload", "Tags").Return(m)
				m.On("Preload", "Author").Return(m)
				m.On("Find", mock.AnythingOfType("*model.Article"), uint(1)).Run(func(args mock.Arguments) {
					arg := args.Get(0).(*model.Article)
					*arg = model.Article{
						Model:       gorm.Model{ID: 1},
						Title:       "Test Article",
						Description: "Test Description",
						Body:        "Test Body",
						Tags:        []model.Tag{{Model: gorm.Model{ID: 1}, Name: "test"}},
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
				Tags:        []model.Tag{{Model: gorm.Model{ID: 1}, Name: "test"}},
				Author:      model.User{Model: gorm.Model{ID: 1}, Username: "testuser"},
			},
		},
		{
			name: "Attempt to retrieve a non-existent article",
			id:   9999,
			setupMock: func(m *mockDB) {
				m.On("Preload", "Tags").Return(m)
				m.On("Preload", "Author").Return(m)
				m.On("Find", mock.AnythingOfType("*model.Article"), uint(9999)).Return(&gorm.DB{Error: gorm.ErrRecordNotFound})
			},
			expectedError:   gorm.ErrRecordNotFound,
			expectedArticle: nil,
		},
		{
			name: "Retrieve an article with associated tags and author",
			id:   2,
			setupMock: func(m *mockDB) {
				m.On("Preload", "Tags").Return(m)
				m.On("Preload", "Author").Return(m)
				m.On("Find", mock.AnythingOfType("*model.Article"), uint(2)).Run(func(args mock.Arguments) {
					arg := args.Get(0).(*model.Article)
					*arg = model.Article{
						Model:       gorm.Model{ID: 2},
						Title:       "Article with Tags",
						Description: "Description",
						Body:        "Body",
						Tags:        []model.Tag{{Model: gorm.Model{ID: 1}, Name: "tag1"}, {Model: gorm.Model{ID: 2}, Name: "tag2"}},
						Author:      model.User{Model: gorm.Model{ID: 2}, Username: "author"},
					}
				}).Return(&gorm.DB{Error: nil})
			},
			expectedError: nil,
			expectedArticle: &model.Article{
				Model:       gorm.Model{ID: 2},
				Title:       "Article with Tags",
				Description: "Description",
				Body:        "Body",
				Tags:        []model.Tag{{Model: gorm.Model{ID: 1}, Name: "tag1"}, {Model: gorm.Model{ID: 2}, Name: "tag2"}},
				Author:      model.User{Model: gorm.Model{ID: 2}, Username: "author"},
			},
		},
		{
			name: "Handle database connection error",
			id:   3,
			setupMock: func(m *mockDB) {
				m.On("Preload", "Tags").Return(m)
				m.On("Preload", "Author").Return(m)
				m.On("Find", mock.AnythingOfType("*model.Article"), uint(3)).Return(&gorm.DB{Error: errors.New("database connection error")})
			},
			expectedError:   errors.New("database connection error"),
			expectedArticle: nil,
		},
		{
			name: "Retrieve an article with no associated tags",
			id:   4,
			setupMock: func(m *mockDB) {
				m.On("Preload", "Tags").Return(m)
				m.On("Preload", "Author").Return(m)
				m.On("Find", mock.AnythingOfType("*model.Article"), uint(4)).Run(func(args mock.Arguments) {
					arg := args.Get(0).(*model.Article)
					*arg = model.Article{
						Model:       gorm.Model{ID: 4},
						Title:       "Article without Tags",
						Description: "Description",
						Body:        "Body",
						Tags:        []model.Tag{},
						Author:      model.User{Model: gorm.Model{ID: 3}, Username: "author"},
					}
				}).Return(&gorm.DB{Error: nil})
			},
			expectedError: nil,
			expectedArticle: &model.Article{
				Model:       gorm.Model{ID: 4},
				Title:       "Article without Tags",
				Description: "Description",
				Body:        "Body",
				Tags:        []model.Tag{},
				Author:      model.User{Model: gorm.Model{ID: 3}, Username: "author"},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockDB := new(mockDB)
			tt.setupMock(mockDB)

			store := &ArticleStore{
				db: mockDB,
			}

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


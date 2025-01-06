package store

import (
	"errors"
	"testing"
	"github.com/jinzhu/gorm"
	"github.com/raahii/golang-grpc-realworld-example/model"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)





type Call struct {
	Parent *Mock

	Method string

	Arguments Arguments

	ReturnArguments Arguments

	callerInfo []string

	Repeatability int

	totalCalls int

	optional bool

	WaitFor <-chan time.Time

	waitTime time.Duration

	RunFn func(Arguments)
}
type mockDB struct {
	mock.Mock
}
type T struct {
	common
	isEnvSet bool
	context  *testContext
}


/*
ROOST_METHOD_HASH=GetByID_36e92ad6eb
ROOST_METHOD_SIG_HASH=GetByID_9616e43e52


 */
func (m *mockDB) Find(out interface{}, where ...interface{}) *gorm.DB {
	args := m.Called(out, where)
	return args.Get(0).(*gorm.DB)
}

func (s *MockArticleStore) GetByID(id uint) (*model.Article, error) {
	var m model.Article
	err := s.db.Preload("Tags").Preload("Author").Find(&m, id).Error
	if err != nil {
		return nil, err
	}
	return &m, nil
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
			name: "Successfully retrieve an existing article by ID",
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
						Tags:        []model.Tag{{Model: gorm.Model{ID: 1}, Name: "test-tag"}},
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
				Tags:        []model.Tag{{Model: gorm.Model{ID: 1}, Name: "test-tag"}},
				Author:      model.User{Model: gorm.Model{ID: 1}, Username: "testuser"},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockDB := new(mockDB)
			tt.setupMock(mockDB)

			store := &MockArticleStore{db: mockDB}
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


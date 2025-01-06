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





type MockDB struct {
	mock.Mock
}


/*
ROOST_METHOD_HASH=GetByID_bbf946112e
ROOST_METHOD_SIG_HASH=GetByID_728dd55ed1


 */
func (m *MockDB) Find(out interface{}, where ...interface{}) *gorm.DB {
	args := m.Called(out, where)
	return args.Get(0).(*gorm.DB)
}

func TestUserStoreGetByID(t *testing.T) {
	tests := []struct {
		name         string
		id           uint
		mockSetup    func(*MockDB)
		expectedUser *model.User
		expectedErr  error
	}{
		{
			name: "Successfully retrieve a user by ID",
			id:   1,
			mockSetup: func(m *MockDB) {
				m.On("Find", mock.AnythingOfType("*model.User"), uint(1)).Return(&gorm.DB{Error: nil}).Run(func(args mock.Arguments) {
					arg := args.Get(0).(*model.User)
					*arg = model.User{
						Model:    gorm.Model{ID: 1, CreatedAt: time.Now(), UpdatedAt: time.Now()},
						Username: "testuser",
						Email:    "test@example.com",
						Password: "password",
						Bio:      "Test bio",
						Image:    "test.jpg",
					}
				})
			},
			expectedUser: &model.User{
				Model:    gorm.Model{ID: 1},
				Username: "testuser",
				Email:    "test@example.com",
				Password: "password",
				Bio:      "Test bio",
				Image:    "test.jpg",
			},
			expectedErr: nil,
		},
		{
			name: "Attempt to retrieve a non-existent user",
			id:   999,
			mockSetup: func(m *MockDB) {
				m.On("Find", mock.AnythingOfType("*model.User"), uint(999)).Return(&gorm.DB{Error: gorm.ErrRecordNotFound})
			},
			expectedUser: nil,
			expectedErr:  gorm.ErrRecordNotFound,
		},
		{
			name: "Handle database connection error",
			id:   2,
			mockSetup: func(m *MockDB) {
				m.On("Find", mock.AnythingOfType("*model.User"), uint(2)).Return(&gorm.DB{Error: errors.New("database connection error")})
			},
			expectedUser: nil,
			expectedErr:  errors.New("database connection error"),
		},
		{
			name: "Retrieve a user with minimum fields populated",
			id:   3,
			mockSetup: func(m *MockDB) {
				m.On("Find", mock.AnythingOfType("*model.User"), uint(3)).Return(&gorm.DB{Error: nil}).Run(func(args mock.Arguments) {
					arg := args.Get(0).(*model.User)
					*arg = model.User{
						Model:    gorm.Model{ID: 3, CreatedAt: time.Now(), UpdatedAt: time.Now()},
						Username: "minuser",
						Email:    "min@example.com",
						Password: "minpass",
					}
				})
			},
			expectedUser: &model.User{
				Model:    gorm.Model{ID: 3},
				Username: "minuser",
				Email:    "min@example.com",
				Password: "minpass",
			},
			expectedErr: nil,
		},
		{
			name: "Retrieve a user with all fields populated",
			id:   4,
			mockSetup: func(m *MockDB) {
				m.On("Find", mock.AnythingOfType("*model.User"), uint(4)).Return(&gorm.DB{Error: nil}).Run(func(args mock.Arguments) {
					arg := args.Get(0).(*model.User)
					*arg = model.User{
						Model:            gorm.Model{ID: 4, CreatedAt: time.Now(), UpdatedAt: time.Now()},
						Username:         "fulluser",
						Email:            "full@example.com",
						Password:         "fullpass",
						Bio:              "Full bio",
						Image:            "full.jpg",
						Follows:          []model.User{{Model: gorm.Model{ID: 5}}},
						FavoriteArticles: []model.Article{{Model: gorm.Model{ID: 1}}},
					}
				})
			},
			expectedUser: &model.User{
				Model:            gorm.Model{ID: 4},
				Username:         "fulluser",
				Email:            "full@example.com",
				Password:         "fullpass",
				Bio:              "Full bio",
				Image:            "full.jpg",
				Follows:          []model.User{{Model: gorm.Model{ID: 5}}},
				FavoriteArticles: []model.Article{{Model: gorm.Model{ID: 1}}},
			},
			expectedErr: nil,
		},
		{
			name: "Handle zero ID input",
			id:   0,
			mockSetup: func(m *MockDB) {
				m.On("Find", mock.AnythingOfType("*model.User"), uint(0)).Return(&gorm.DB{Error: gorm.ErrRecordNotFound})
			},
			expectedUser: nil,
			expectedErr:  gorm.ErrRecordNotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockDB := new(MockDB)
			tt.mockSetup(mockDB)

			store := &UserStore{db: mockDB}

			user, err := store.GetByID(tt.id)

			if tt.expectedErr != nil {
				assert.Error(t, err)
				assert.Equal(t, tt.expectedErr.Error(), err.Error())
			} else {
				assert.NoError(t, err)
			}

			assert.Equal(t, tt.expectedUser, user)

			mockDB.AssertExpectations(t)
		})
	}
}


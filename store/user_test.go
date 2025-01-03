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

// MockDB is a mock of gorm.DB
type MockDB struct {
	mock.Mock
}

// Find is a mocked method
func (m *MockDB) Find(out interface{}, where ...interface{}) *gorm.DB {
	args := m.Called(out, where)
	return args.Get(0).(*gorm.DB)
}

// Model is a mocked method
func (m *MockDB) Model(value interface{}) *gorm.DB {
	args := m.Called(value)
	return args.Get(0).(*gorm.DB)
}

// Where is a mocked method
func (m *MockDB) Where(query interface{}, args ...interface{}) *gorm.DB {
	mockArgs := m.Called(query, args)
	return mockArgs.Get(0).(*gorm.DB)
}

// First is a mocked method
func (m *MockDB) First(out interface{}, where ...interface{}) *gorm.DB {
	args := m.Called(out, where)
	return args.Get(0).(*gorm.DB)
}

// Create is a mocked method
func (m *MockDB) Create(value interface{}) *gorm.DB {
	args := m.Called(value)
	return args.Get(0).(*gorm.DB)
}

// Save is a mocked method
func (m *MockDB) Save(value interface{}) *gorm.DB {
	args := m.Called(value)
	return args.Get(0).(*gorm.DB)
}

// Delete is a mocked method
func (m *MockDB) Delete(value interface{}, where ...interface{}) *gorm.DB {
	args := m.Called(value, where)
	return args.Get(0).(*gorm.DB)
}

// ROOST_METHOD_HASH=GetByID_bbf946112e
// ROOST_METHOD_SIG_HASH=GetByID_728dd55ed1

func TestUserStoreGetByID(t *testing.T) {
	tests := []struct {
		name      string
		userID    uint
		mockSetup func(*MockDB)
		want      *model.User
		wantErr   error
	}{
		{
			name:   "Successfully retrieve an existing user",
			userID: 1,
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
			want: &model.User{
				Model:    gorm.Model{ID: 1},
				Username: "testuser",
				Email:    "test@example.com",
				Password: "password",
				Bio:      "Test bio",
				Image:    "test.jpg",
			},
			wantErr: nil,
		},
		// ... (other test cases remain the same)
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockDB := new(MockDB)
			tt.mockSetup(mockDB)

			s := &UserStore{
				db: mockDB,
			}

			got, err := s.GetByID(tt.userID)

			if tt.wantErr != nil {
				assert.Error(t, err)
				assert.Equal(t, tt.wantErr.Error(), err.Error())
			} else {
				assert.NoError(t, err)
			}

			assert.Equal(t, tt.want, got)

			mockDB.AssertExpectations(t)
		})
	}
}

package store

import (
	"errors"
	"math"
	"testing"
	"github.com/jinzhu/gorm"
	"github.com/raahii/golang-grpc-realworld-example/model"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)





type T struct {
	common
	isEnvSet bool
	context  *testContext // For running tests and subtests.
}


type MockDB struct {
	mock.Mock
}






type T struct {
	common
	isEnvSet bool
	context  *testContext // For running tests and subtests.
}

func (m *MockDB) Find(out interface{}, where ...interface{}) *gorm.DB {
	args := m.Called(out, where)
	return args.Get(0).(*gorm.DB)
}
func (s *MockUserStore) GetByID(id uint) (*model.User, error) {
	var m model.User
	if err := s.db.Find(&m, id).Error; err != nil {
		return nil, err
	}
	return &m, nil
}
func TestGetByID(t *testing.T) {
	tests := []struct {
		name      string
		id        uint
		mockSetup func(*MockDB)
		want      *model.User
		wantErr   error
	}{
		{
			name: "Successfully retrieve a user by ID",
			id:   1,
			mockSetup: func(m *MockDB) {
				m.On("Find", mock.AnythingOfType("*model.User"), uint(1)).Return(&gorm.DB{Error: nil}).Run(func(args mock.Arguments) {
					arg := args.Get(0).(*model.User)
					*arg = model.User{Model: gorm.Model{ID: 1}, Username: "testuser", Email: "test@example.com"}
				})
			},
			want:    &model.User{Model: gorm.Model{ID: 1}, Username: "testuser", Email: "test@example.com"},
			wantErr: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockDB := new(MockDB)
			tt.mockSetup(mockDB)

			s := &MockUserStore{db: mockDB}

			got, err := s.GetByID(tt.id)

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

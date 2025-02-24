package store

import (
	errors "errors"
	testing "testing"

	gorm "github.com/jinzhu/gorm"
	model "github.com/raahii/golang-grpc-realworld-example/model"
	assert "github.com/stretchr/testify/assert"
)

type MockUserStore struct {
	db *mockDB
}

/*
ROOST_METHOD_HASH=UserStore_Create_9495ddb29d
ROOST_METHOD_SIG_HASH=UserStore_Create_18451817fe

FUNCTION_DEF=func (s *UserStore) Create(m *model.User) error // Create create a user
*/
func (s *MockUserStore) Create(m *model.User) error {
	return s.db.Create(m).Error
}

func NewMockUserStore(db *mockDB) *MockUserStore {
	return &MockUserStore{db: db}
}

func TestUserStoreCreate(t *testing.T) {
	tests := []struct {
		name    string
		user    *model.User
		dbError error
		wantErr bool
	}{
		{
			name: "Successfully Create a New User",
			user: &model.User{
				Username: "testuser",
				Email:    "test@example.com",
				Password: "password123",
			},
			dbError: nil,
			wantErr: false,
		},
		{
			name: "Attempt to Create a User with Invalid Data",
			user: &model.User{
				Username: "",
				Email:    "test@example.com",
				Password: "password123",
			},
			dbError: errors.New("validation error"),
			wantErr: true,
		},
		{
			name: "Handle Database Connection Error",
			user: &model.User{
				Username: "testuser",
				Email:    "test@example.com",
				Password: "password123",
			},
			dbError: errors.New("database connection error"),
			wantErr: true,
		},
		{
			name: "Create User with Maximum Allowed Field Lengths",
			user: &model.User{
				Username: "testuser_with_maximum_length_username",
				Email:    "very_long_email_address@very_long_domain_name.com",
				Password: "very_long_password_that_meets_maximum_length_requirements",
			},
			dbError: nil,
			wantErr: false,
		},
		{
			name: "Attempt to Create a Duplicate User",
			user: &model.User{
				Username: "existinguser",
				Email:    "existing@example.com",
				Password: "password123",
			},
			dbError: errors.New("unique constraint violation"),
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockDB := new(mockDB)
			userStore := NewMockUserStore(mockDB)

			mockDB.On("Create", tt.user).Return(&gorm.DB{Error: tt.dbError})

			err := userStore.Create(tt.user)

			if tt.wantErr {
				assert.Error(t, err)
				assert.Equal(t, tt.dbError, err)
			} else {
				assert.NoError(t, err)
			}

			mockDB.AssertExpectations(t)
		})
	}
}

func (m *mockDB) AddError(err error) error {
	args := m.Called(err)
	return args.Error(0)
}

func (m *mockDB) Association(column string) *gorm.Association {
	args := m.Called(column)
	return args.Get(0).(*gorm.Association)
}

func (m *mockDB) Attrs(attrs ...interface{}) *gorm.DB {
	args := m.Called(attrs)
	return args.Get(0).(*gorm.DB)
}

func (m *mockDB) Create(value interface{}) *gorm.DB {
	args := m.Called(value)
	return args.Get(0).(*gorm.DB)
}

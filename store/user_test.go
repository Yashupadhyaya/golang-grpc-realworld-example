package store

import (
	errors "errors"
	testing "testing"

	gorm "github.com/jinzhu/gorm"
	model "github.com/raahii/golang-grpc-realworld-example/model"
)

type mockUserStore struct {
	expectedUser  *model.User
	expectedError error
}

/*
ROOST_METHOD_HASH=UserStore_GetByEmail_fda09af5c4
ROOST_METHOD_SIG_HASH=UserStore_GetByEmail_9e84f3286b

FUNCTION_DEF=func (s *UserStore) GetByEmail(email string) (*model.User, error) // GetByEmail finds a user from email
*/
func TestUserStoreGetByEmail(t *testing.T) {
	testCases := []struct {
		name          string
		email         string
		expectedUser  *model.User
		expectedError error
	}{
		{
			name:  "Valid Email Exist",
			email: "test@example.com",
			expectedUser: &model.User{
				Email: "test@example.com",
			},
			expectedError: nil,
		},
		{
			name:          "Email Does Not Exist",
			email:         "unknown@example.com",
			expectedUser:  nil,
			expectedError: gorm.ErrRecordNotFound,
		},
		{
			name:          "Empty Email String",
			email:         "",
			expectedUser:  nil,
			expectedError: errors.New("invalid email"),
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			defer func() {
				if r := recover(); r != nil {
					t.Logf("Panic encountered. Fail test. %v", r)
					t.Fail()
				}
			}()

			userStore := &mockUserStore{
				expectedUser:  tc.expectedUser,
				expectedError: tc.expectedError,
			}

			user, err := userStore.GetByEmail(tc.email)
			if err != tc.expectedError {
				t.Errorf("Unexpected error. Want '%v'. Got '%v'", tc.expectedError, err)
				return
			}

			if tc.expectedUser != nil && user.Email != tc.expectedUser.Email {
				t.Errorf("Unexpected user returned. Want '%v'. Got '%v'", tc.expectedUser, user)
			}
		})
	}
}

func (store *mockUserStore) GetByEmail(email string) (*model.User, error) {
	return store.expectedUser, store.expectedError
}

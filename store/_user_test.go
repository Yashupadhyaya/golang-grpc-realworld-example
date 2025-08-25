package store

import (
	testing "testing"
	os "os"
	bytes "bytes"
	fmt "fmt"
	debug "runtime/debug"
	sqlmock "github.com/DATA-DOG/go-sqlmock"
	gorm "github.com/jinzhu/gorm"
	model "github.com/raahii/golang-grpc-realworld-example/model"
)








/*
ROOST_METHOD_HASH=UserStore.GetByUsername_622b1b9e41
ROOST_METHOD_SIG_HASH=UserStore.GetByUsername_992f00baec

FUNCTION_DEF=func (s *UserStore) GetByUsername(username string) (*model.User, error) // GetByUsername finds a user from username


*/
func TestUserStoreGetByUsername(t *testing.T) {

	type mockSetup struct {
		username string
		user     model.User
		err      error
	}

	tests := []struct {
		name       string
		mockInput  mockSetup
		expectUser *model.User
		expectErr  bool
	}{

		{
			name: "Valid Username Exists",
			mockInput: mockSetup{
				username: "testuser",
				user: model.User{
					ID:       1,
					Username: "testuser",
					Email:    "testuser@example.com",
				},
				err: nil,
			},
			expectUser: &model.User{
				ID:       1,
				Username: "testuser",
				Email:    "testuser@example.com",
			},
			expectErr: false,
		},

		{
			name: "Username Does Not Exist",
			mockInput: mockSetup{
				username: "nonexistentuser",
				user:     model.User{},
				err:      gorm.ErrRecordNotFound,
			},
			expectUser: nil,
			expectErr:  true,
		},

		{
			name: "Empty Username Input",
			mockInput: mockSetup{
				username: "",
				user:     model.User{},
				err:      gorm.ErrRecordNotFound,
			},
			expectUser: nil,
			expectErr:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			defer func() {
				if r := recover(); r != nil {
					t.Logf("Panic encountered: %v\n%s", r, string(debug.Stack()))
					t.Fail()
				}
			}()

			mockDB, mock, err := sqlmock.New()
			if err != nil {
				t.Fatalf("Failed to create sqlmock: %v", err)
			}
			defer mockDB.Close()

			gormDB, err := gorm.Open("sqlite3", mockDB)
			if err != nil {
				t.Fatalf("Failed to create GORM DB: %v", err)
			}
			defer gormDB.Close()

			userStore := &UserStore{
				db: gormDB,
			}

			mock.ExpectQuery(`SELECT * FROM "users" WHERE username = ?`).
				WithArgs(tt.mockInput.username).
				WillReturnRows(sqlmock.NewRows([]string{"id", "username", "email"}).
					AddRow(tt.mockInput.user.ID, tt.mockInput.user.Username, tt.mockInput.user.Email)).
				WillReturnError(tt.mockInput.err)

			var buf bytes.Buffer
			stdout := os.Stdout
			os.Stdout = &buf
			defer func() { os.Stdout = stdout }()

			result, err := userStore.GetByUsername(tt.mockInput.username)

			if (err != nil) != tt.expectErr {
				t.Errorf("Expected error: %v, got: %v", tt.expectErr, err)
			}

			if err == nil {
				if result.ID != tt.expectUser.ID || result.Username != tt.expectUser.Username || result.Email != tt.expectUser.Email {
					t.Errorf("Expected user: %+v, got: %+v", tt.expectUser, result)
				}
			}

			if err == nil {
				t.Logf("Successfully fetched user: %+v", result)
			} else {
				t.Logf("Error while fetching user: %v", err)
			}
		})
	}
}


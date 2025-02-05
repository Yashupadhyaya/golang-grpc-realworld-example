package store


import (
	"errors"
	"testing"
	"github.com/jinzhu/gorm"
	"github.com/raahii/golang-grpc-realworld-example/model"
)








/*
ROOST_METHOD_HASH=Create_9495ddb29d
ROOST_METHOD_SIG_HASH=Create_18451817fe

FUNCTION_DEF=func (s *UserStore) Create(m *model.User) error // Create create a user


*/
func (m *mockDB) Create(value interface{}) *gorm.DB {
	return m.createFunc(value)
}

func TestUserStoreCreate(t *testing.T) {
	tests := []struct {
		name    string
		user    *model.User
		mockErr error
		wantErr bool
	}{
		{
			name: "Successfully Create a New User",
			user: &model.User{
				Username: "testuser",
				Email:    "test@example.com",
				Password: "password123",
			},
			mockErr: nil,
			wantErr: false,
		},
		{
			name: "Attempt to Create a User with Duplicate Username",
			user: &model.User{
				Username: "existinguser",
				Email:    "new@example.com",
				Password: "password123",
			},
			mockErr: errors.New("ERROR: duplicate key value violates unique constraint \"users_username_key\" (SQLSTATE 23505)"),
			wantErr: true,
		},
		{
			name: "Attempt to Create a User with Duplicate Email",
			user: &model.User{
				Username: "newuser",
				Email:    "existing@example.com",
				Password: "password123",
			},
			mockErr: errors.New("ERROR: duplicate key value violates unique constraint \"users_email_key\" (SQLSTATE 23505)"),
			wantErr: true,
		},
		{
			name: "Create User with Minimum Required Fields",
			user: &model.User{
				Username: "minimaluser",
				Email:    "minimal@example.com",
				Password: "password123",
			},
			mockErr: nil,
			wantErr: false,
		},
		{
			name: "Attempt to Create User with Invalid Email Format",
			user: &model.User{
				Username: "invaliduser",
				Email:    "invalid-email",
				Password: "password123",
			},
			mockErr: errors.New("ERROR: invalid email format"),
			wantErr: true,
		},
		{
			name: "Database Connection Error During User Creation",
			user: &model.User{
				Username: "erroruser",
				Email:    "error@example.com",
				Password: "password123",
			},
			mockErr: errors.New("ERROR: database connection failed"),
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mock := &mockDB{
				createFunc: func(value interface{}) *gorm.DB {
					return &gorm.DB{Error: tt.mockErr}
				},
			}

			store := &UserStore{
				db: mock,
			}

			err := store.Create(tt.user)

			if (err != nil) != tt.wantErr {
				t.Errorf("UserStore.Create() error = %v, wantErr %v", err, tt.wantErr)
			}

			if tt.wantErr && err != tt.mockErr {
				t.Errorf("UserStore.Create() error = %v, want %v", err, tt.mockErr)
			}
		})
	}
}


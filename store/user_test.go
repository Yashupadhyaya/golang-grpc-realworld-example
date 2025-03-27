package store

import (
	errors "errors"
	fmt "fmt"
	strings "strings"
	testing "testing"

	"github.com/DATA-DOG/go-sqlmock"
	gosqlmock "github.com/DATA-DOG/go-sqlmock"
	gorm "github.com/jinzhu/gorm"
	model "github.com/raahii/golang-grpc-realworld-example/model"
)

type mockUserStore struct {
	db *gorm.DB
}

/*
ROOST_METHOD_HASH=UserStore_Create_9495ddb29d
ROOST_METHOD_SIG_HASH=UserStore_Create_18451817fe

FUNCTION_DEF=func (s *UserStore) Create(m *model.User) error // Create create a user
*/
func TestUserStoreCreate(t *testing.T) {

	t.Run("Database function successfully creates user", func(t *testing.T) {
		defer func() {
			if r := recover(); r != nil {
				t.Logf("Panic encountered so failing test. %v", r)
				t.Fail()
			}
		}()

		mockDb, _, _ := sqlmock.New()
		db, _ := gorm.Open("postgres", mockDb)
		store := &UserStore{db: db}
		user := &model.User{Username: "testuser", Email: "test@test.com"}

		err := store.Create(user)

		if err != nil {
			t.Fatalf("Expected no error from the Create function but got: %v", err)
		}
	})

	t.Run("UserStore.Create function error handling", func(t *testing.T) {
		defer func() {
			if r := recover(); r != nil {
				t.Logf("Panic encountered so failing test. %v", r)
				t.Fail()
			}
		}()

		mockDb, _, _ := sqlmock.New()
		db, _ := gorm.Open("postgres", mockDb)
		store := &mockUserStore{db: db}
		user := &model.User{Username: "testuser", Email: "test@test.com"}

		err := store.Create(user)

		if err == nil {
			t.Fatalf("Expected an error but got nil")
		}
	})

	t.Run("Database function fails to create a user due to null or invalid details", func(t *testing.T) {
		defer func() {
			if r := recover(); r != nil {
				t.Logf("Panic encountered so failing test. %v", r)
				t.Fail()
			}
		}()

		mockDb, _, _ := sqlmock.New()
		db, _ := gorm.Open("postgres", mockDb)
		store := &UserStore{db: db}
		user := &model.User{Username: "", Email: "test@test.com"}

		err := store.Create(user)

		if err == nil {
			t.Fatalf("Expected an error but got nil")
		}
	})
}

func (s *mockUserStore) Create(m *model.User) error {

	if s.db.Error != nil {
		return errors.New("db error")
	}
	return nil
}

/*
ROOST_METHOD_HASH=UserStore_GetByEmail_fda09af5c4
ROOST_METHOD_SIG_HASH=UserStore_GetByEmail_9e84f3286b

FUNCTION_DEF=func (s *UserStore) GetByEmail(email string) (*model.User, error) // GetByEmail finds a user from email
*/
func TestUserStoreGetByEmail(t *testing.T) {

	testCases := []struct {
		desc        string
		email       string
		mock        func(gosqlmock.Sqlmock, string)
		expectedErr error
	}{
		{
			desc:  "Valid Email ID",
			email: "test@example.com",
			mock: func(mock gosqlmock.Sqlmock, email string) {
				rows := gosqlmock.NewRows([]string{"id", "email"}).
					AddRow(1, email)
				mock.ExpectQuery("^SELECT (.+) FROM `users` WHERE (.+)").
					WithArgs(email).
					WillReturnRows(rows)
			},
			expectedErr: nil,
		},
		{
			desc:  "Non-existent Email ID",
			email: "non-existent@example.com",
			mock: func(mock gosqlmock.Sqlmock, email string) {
				mock.ExpectQuery("^SELECT (.+) FROM `users` WHERE (.+)").
					WithArgs(email).
					WillReturnError(gorm.ErrRecordNotFound)
			},
			expectedErr: gorm.ErrRecordNotFound,
		},
		{
			desc:  "Invalid Email ID Format",
			email: "invalidemail",
			mock: func(mock gosqlmock.Sqlmock, email string) {

			},
			expectedErr: errors.New("invalid email format"),
		},
		{
			desc:  "Empty Email ID",
			email: "",
			mock: func(mock gosqlmock.Sqlmock, email string) {

			},
			expectedErr: errors.New("email id is required"),
		},
	}

	for _, tc := range testCases {
		t.Run(tc.desc, func(t *testing.T) {
			defer func() {
				if r := recover(); r != nil {
					t.Logf("Panic encountered so failing test. %v", r)
					t.Fail()
				}
			}()

			db, mock, err := gosqlmock.New()
			if err != nil {
				t.Fatalf("Failed to setup database mock: %v", err)
			}
			defer db.Close()

			gormDb, err := gorm.Open("postgres", db)
			if err != nil {
				t.Fatalf("Failed to open gorm db: %v", err)
			}

			store := &UserStore{db: gormDb}

			tc.mock(mock, tc.email)

			_, err = store.GetByEmail(tc.email)
			if err != tc.expectedErr {
				t.Errorf("Error: %v, Expected Error: %v", err, tc.expectedErr)
			}

			if err := mock.ExpectationsWereMet(); err != nil {
				t.Errorf("Unfulfilled mock expectations: %v", err)
			}
		})
	}
}

/*
ROOST_METHOD_HASH=UserStore_GetByUsername_622b1b9e41
ROOST_METHOD_SIG_HASH=UserStore_GetByUsername_992f00baec

FUNCTION_DEF=func (s *UserStore) GetByUsername(username string) (*model.User, error) // GetByUsername finds a user from username
*/
func TestUserStoreGetByUsername(t *testing.T) {
	type args struct {
		username string
	}
	type want struct {
		user *model.User
		err  string
	}
	tests := []struct {
		name    string
		mock    func(mock sqlmock.Sqlmock, args args)
		args    args
		want    want
		wantErr bool
	}{
		{
			name: "Valid User Test",
			mock: func(mock sqlmock.Sqlmock, args args) {
				rows := sqlmock.NewRows([]string{"username"}).AddRow(args.username)
				mock.ExpectQuery("^SELECT (.+) FROM \"users\" WHERE (.+)").WithArgs(args.username).WillReturnRows(rows)
			},
			args: args{username: "testuser"},
			want: want{
				user: &model.User{Username: "testuser"},
				err:  "",
			},
			wantErr: false,
		},
		{
			name: "User Does Not Exist Test",
			mock: func(mock sqlmock.Sqlmock, args args) {
				mock.ExpectQuery("^SELECT (.+) FROM \"users\" WHERE (.+)").WithArgs(args.username).WillReturnError(gorm.ErrRecordNotFound)
			},
			args:    args{username: "nonexistinguser"},
			wantErr: true,
		},
		{
			name: "Empty Username Test",
			mock: func(mock sqlmock.Sqlmock, args args) {
				mock.ExpectQuery("^SELECT (.+) FROM \"users\" WHERE (.+)").WithArgs(args.username).WillReturnError(fmt.Errorf("username cannot be an empty string"))
			},
			args:    args{username: ""},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			defer func() {
				if r := recover(); r != nil {
					t.Logf("Panic encountered so failing test. Panic: %v", r)
					t.Fail()
				}
			}()

			db, mock, err := sqlmock.New()
			if err != nil {
				t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
			}
			defer db.Close()

			gormDB, err := gorm.Open("postgres", db)
			if err != nil {
				t.Fatalf("an error '%s' was not expected when opening gorm database from sql.DB", err)
			}

			tt.mock(mock, tt.args)

			s := &UserStore{
				db: gormDB,
			}

			user, err := s.GetByUsername(tt.args.username)
			if (err != nil) != tt.wantErr {
				t.Errorf("UserStore.GetByUsername() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if err != nil {
				if tt.wantErr && !strings.Contains(err.Error(), tt.want.err) {
					t.Errorf("UserStore.GetByUsername() = %v, want %v", err, tt.want.err)
				}
				return
			}
			if user.Username != tt.want.user.Username {
				t.Errorf("UserStore.GetByUsername() = %v, want %v", user, tt.want.user)
			}
		})
	}
}

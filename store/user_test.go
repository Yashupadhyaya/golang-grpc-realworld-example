package store

import (
	testing "testing"

	gosqlmock "github.com/DATA-DOG/go-sqlmock"
	sqlmock "github.com/DATA-DOG/go-sqlmock"
	gorm "github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/sqlite"
	model "github.com/raahii/golang-grpc-realworld-example/model"
	assert "github.com/stretchr/testify/assert"
)

type UserStoreMock struct {
	db *gorm.DB
}

/*
ROOST_METHOD_HASH=UserStore_Create_9495ddb29d
ROOST_METHOD_SIG_HASH=UserStore_Create_18451817fe

FUNCTION_DEF=func (s *UserStore) Create(m *model.User) error // Create create a user
*/
func TestUserStoreCreate(t *testing.T) {

	var tests = []struct {
		name       string
		user       *model.User
		dbFunction func(*gorm.DB)
		expectErr  bool
	}{

		{
			"Successful User Creation",
			&model.User{
				Username: "testUser1",
				Email:    "testUser1@example.com",
			},
			func(db *gorm.DB) {

			},
			false,
		},

		{
			"Unsuccessful User creation due to existing User",
			&model.User{
				Username: "testUser2",
				Email:    "testUser2@example.com",
			},
			func(db *gorm.DB) {

				_ = db.Create(&model.User{
					Username: "testUser2",
					Email:    "testUser2@example.com",
				})
			},
			true,
		},

		{
			"Unsuccessful User creation due to invalid User details",
			&model.User{},
			func(db *gorm.DB) {

			},
			true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			defer func() {
				if r := recover(); r != nil {
					t.Logf("Panic encountered during test. %v\n", r)
					t.Fail()
				}
			}()

			mockSql, _, _ := sqlmock.New()
			mockDB, _ := gorm.Open("sqlmock", mockSql)

			s := &UserStoreMock{db: mockDB}
			tt.dbFunction(mockDB)

			err := s.Create(tt.user)

			if tt.expectErr {
				assert.Error(t, err, "Expected error but got none")
			} else {
				assert.NoError(t, err, "Unexpected error, expected no error")
			}
		})
	}
}

func (us *UserStoreMock) Create(m *model.User) error {
	return us.db.Create(m).Error
}

/*
ROOST_METHOD_HASH=UserStore_GetByEmail_fda09af5c4
ROOST_METHOD_SIG_HASH=UserStore_GetByEmail_9e84f3286b

FUNCTION_DEF=func (s *UserStore) GetByEmail(email string) (*model.User, error) // GetByEmail finds a user from email
*/
func TestUserStoreGetByEmail(t *testing.T) {

	mockEmail := "test@example.com"
	expectedUser := &model.User{
		Email: mockEmail,
	}

	scenarios := []struct {
		desc    string
		email   string
		mock    func(gosqlmock.Sqlmock)
		wantErr bool
		want    *model.User
	}{
		{
			desc:  "Scenario 1: Email Exists in Database",
			email: mockEmail,
			mock: func(m gosqlmock.Sqlmock) {

				rows := gosqlmock.NewRows([]string{"email"}).AddRow(mockEmail)
				m.ExpectQuery("^SELECT").WithArgs(mockEmail).WillReturnRows(rows)
			},
			wantErr: false,
			want:    expectedUser,
		},
		{
			desc:  "Scenario 2: Email Does Not Exist in Database",
			email: "nonexistent@example.com",
			mock: func(m gosqlmock.Sqlmock) {

				rows := gosqlmock.NewRows([]string{"email"})
				m.ExpectQuery("^SELECT").WithArgs("nonexistent@example.com").WillReturnRows(rows)
			},
			wantErr: true,
			want:    nil,
		},
		{
			desc:  "Scenario 3: Database Error",
			email: mockEmail,
			mock: func(m gosqlmock.Sqlmock) {

				m.ExpectQuery("^SELECT").WithArgs(mockEmail).WillReturnError(gorm.ErrInvalidSQL)
			},
			wantErr: true,
			want:    nil,
		},
	}

	db, mock, _ := gosqlmock.New()
	defer db.Close()
	gdb, _ := gorm.Open("postgres", db)

	userStore := UserStore{db: gdb}

	for _, tt := range scenarios {
		t.Run(tt.desc, func(t *testing.T) {
			defer func() {
				if r := recover(); r != nil {
					t.Logf("Panic encountered so failing test. %v", r)
					t.Fail()
				}
			}()

			tt.mock(mock)

			got, err := userStore.GetByEmail(tt.email)

			if (err != nil) != tt.wantErr {
				t.Errorf("UserStore.GetByEmail() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("UserStore.GetByEmail() = %v, want %v", got, tt.want)
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
	tests := []struct {
		name    string
		args    args
		want    *model.User
		wantErr bool
		setMock func(mockSql gosqlmock.Sqlmock, args args)
	}{
		{
			name: "Scenario 1: Retrieving a Valid User",
			args: args{"exampleUsername"},
			want: &model.User{Username: "exampleUsername"},
			setMock: func(mock gosqlmock.Sqlmock, args args) {
				mock.ExpectQuery("^SELECT (.+) FROM (.+) WHERE (.+)$").
					WithArgs(args.username).
					WillReturnRows(gosqlmock.NewRows([]string{"username"}).AddRow(args.username))
			},
		},
		{
			name:    "Scenario 2: Given a Non-Existent User",
			args:    args{"exampleUsername"},
			wantErr: true,
			setMock: func(mock gosqlmock.Sqlmock, args args) {
				mock.ExpectQuery("^SELECT (.+) FROM (.+) WHERE (.+)$").
					WithArgs(args.username).
					WillReturnError(gorm.ErrRecordNotFound)
			},
		},
		{
			name:    "Scenario 3: Database Connection Error",
			args:    args{"exampleUsername"},
			wantErr: true,
			setMock: func(mock gosqlmock.Sqlmock, args args) {
				mock.ExpectQuery("^SELECT (.+) FROM (.+) WHERE (.+)$").
					WithArgs(args.username).
					WillReturnError(gorm.ErrInvalidSQL)
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			defer func() {
				if r := recover(); r != nil {
					t.Logf("Panic encountered. Failing test. %v", r)
					t.Fail()
				}
			}()

			db, mock, err := gosqlmock.New()
			if err != nil {
				t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
			}

			tt.setMock(mock, tt.args)

			gDb, err := gorm.Open("mysql", db)
			if err != nil {
				t.Fatalf("Failed to open the stub database connection: %v", err)
			}
			defer db.Close()

			s := &UserStore{db: gDb}

			got, err := s.GetByUsername(tt.args.username)
			if (err != nil) != tt.wantErr {
				t.Errorf("GetByUsername() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			assert.Equal(t, tt.want, got)
		})
	}
}

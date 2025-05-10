package store

import (
	"reflect"
	"sync"
	"testing"
	"time"
	"github.com/jinzhu/gorm"
	"errors"
	"github.com/raahii/golang-grpc-realworld-example/model"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"math"
	"github.com/stretchr/testify/require"
	"github.com/jinzhu/gorm/dialects/sqlite"
	"github.com/DATA-DOG/go-sqlmock"
)

type T struct {
	common
	isEnvSet bool
	context  *testContext
}
type Time struct {
	wall uint64
	ext  int64

	loc *Location
}
type MockDB struct {
	mock.Mock
}
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
type Association struct {
	Error  error
	scope  *Scope
	column string
	field  *Field
}
type mockDB struct {
	*gorm.DB
	mock.Mock
}
type ExpectedQuery struct {
	queryBasedExpectation
	rows             driver.Rows
	delay            time.Duration
	rowsMustBeClosed bool
	rowsWereClosed   bool
}
type Rows struct {
	converter driver.ValueConverter
	cols      []string
	def       []*Column
	rows      [][]driver.Value
	pos       int
	nextErr   map[int]error
	closeErr  error
}
/*
ROOST_METHOD_HASH=NewUserStore_6a331dd890
ROOST_METHOD_SIG_HASH=NewUserStore_4f0c2dfca9


 */
func TestNewUserStore(t *testing.T) {
	tests := []struct {
		name string
		db   *gorm.DB
		want *UserStore
	}{
		{
			name: "Create UserStore with valid gorm.DB",
			db:   &gorm.DB{},
			want: &UserStore{
				db: &gorm.DB{},
			},
		},
		{
			name: "Create UserStore with nil gorm.DB",
			db:   nil,
			want: &UserStore{db: nil},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := NewUserStore(tt.db)
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("NewUserStore() = %v, want %v", got, tt.want)
			}
		})
	}

	t.Run("Verify NewUserStore doesn't modify input gorm.DB", func(t *testing.T) {
		originalDB := &gorm.DB{}
		inputDB := *originalDB
		NewUserStore(&inputDB)
		if !reflect.DeepEqual(originalDB, &inputDB) {
			t.Errorf("NewUserStore() modified the input gorm.DB")
		}
	})

	t.Run("Create multiple UserStore instances with the same gorm.DB", func(t *testing.T) {
		db := &gorm.DB{}
		us1 := NewUserStore(db)
		us2 := NewUserStore(db)
		if us1 == us2 {
			t.Errorf("NewUserStore() returned the same instance for multiple calls")
		}
		if us1.db != us2.db {
			t.Errorf("NewUserStore() did not use the same gorm.DB instance for multiple calls")
		}
	})

	t.Run("Verify thread safety of NewUserStore", func(t *testing.T) {
		db := &gorm.DB{}
		var wg sync.WaitGroup
		userStores := make([]*UserStore, 100)
		for i := 0; i < 100; i++ {
			wg.Add(1)
			go func(index int) {
				defer wg.Done()
				userStores[index] = NewUserStore(db)
			}(i)
		}
		wg.Wait()
		for _, us := range userStores {
			if us == nil {
				t.Errorf("NewUserStore() failed in concurrent execution")
			}
			if us.db != db {
				t.Errorf("NewUserStore() did not use the correct gorm.DB instance in concurrent execution")
			}
		}
	})

	t.Run("Performance test for NewUserStore", func(t *testing.T) {
		db := &gorm.DB{}
		iterations := 10000
		start := time.Now()
		for i := 0; i < iterations; i++ {
			NewUserStore(db)
		}
		duration := time.Since(start)
		t.Logf("Time taken for %d iterations: %v", iterations, duration)

	})
}


/*
ROOST_METHOD_HASH=Create_889fc0fc45
ROOST_METHOD_SIG_HASH=Create_4c48ec3920


 */
func (m *MockDB) Create(value interface{}) *MockDB {
	args := m.Called(value)
	return args.Get(0).(*MockDB)
}

func (m *MockDB) Error() error {
	args := m.Called()
	return args.Error(0)
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
			name: "Attempt to Create a User with a Duplicate Username",
			user: &model.User{
				Username: "existinguser",
				Email:    "new@example.com",
				Password: "password123",
			},
			dbError: errors.New("duplicate username"),
			wantErr: true,
		},
		{
			name: "Attempt to Create a User with a Duplicate Email",
			user: &model.User{
				Username: "newuser",
				Email:    "existing@example.com",
				Password: "password123",
			},
			dbError: errors.New("duplicate email"),
			wantErr: true,
		},
		{
			name: "Create a User with Minimum Required Fields",
			user: &model.User{
				Username: "minuser",
				Email:    "min@example.com",
				Password: "password123",
			},
			dbError: nil,
			wantErr: false,
		},
		{
			name: "Attempt to Create a User with Invalid Data",
			user: &model.User{
				Username: "",
				Email:    "invalid@example.com",
				Password: "password123",
			},
			dbError: errors.New("invalid data"),
			wantErr: true,
		},
		{
			name: "Create a User with Maximum Length Data",
			user: &model.User{
				Username: "verylongusername",
				Email:    "verylongemail@example.com",
				Password: "password123",
				Bio:      "This is a very long bio that tests the maximum length of the bio field",
				Image:    "https://example.com/very/long/image/url.jpg",
			},
			dbError: nil,
			wantErr: false,
		},
		{
			name: "Verify Database Connection Error Handling",
			user: &model.User{
				Username: "connectionerror",
				Email:    "connection@example.com",
				Password: "password123",
			},
			dbError: errors.New("database connection error"),
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockDB := new(MockDB)
			mockDB.On("Create", tt.user).Return(mockDB)
			mockDB.On("Error").Return(tt.dbError)

			userStore := &UserStore{
				db: mockDB,
			}

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


/*
ROOST_METHOD_HASH=GetByID_bbf946112e
ROOST_METHOD_SIG_HASH=GetByID_728dd55ed1


 */
func TestUserStoreGetByID(t *testing.T) {
	tests := []struct {
		name     string
		id       uint
		mockDB   func() *gorm.DB
		expected *model.User
		wantErr  bool
	}{
		{
			name: "Successfully retrieve a user by ID",
			id:   1,
			mockDB: func() *gorm.DB {
				db, _ := gorm.Open("sqlite3", ":memory:")
				user := &model.User{
					Model:    gorm.Model{ID: 1},
					Username: "testuser",
					Email:    "test@example.com",
					Password: "password",
					Bio:      "Test bio",
					Image:    "test.jpg",
				}
				db.Create(user)
				return db
			},
			expected: &model.User{
				Model:    gorm.Model{ID: 1},
				Username: "testuser",
				Email:    "test@example.com",
				Password: "password",
				Bio:      "Test bio",
				Image:    "test.jpg",
			},
			wantErr: false,
		},
		{
			name: "Attempt to retrieve a non-existent user",
			id:   999,
			mockDB: func() *gorm.DB {
				db, _ := gorm.Open("sqlite3", ":memory:")
				return db
			},
			expected: nil,
			wantErr:  true,
		},
		{
			name: "Handle database connection error",
			id:   1,
			mockDB: func() *gorm.DB {
				db, _ := gorm.Open("sqlite3", ":memory:")
				db.Close()
				return db
			},
			expected: nil,
			wantErr:  true,
		},
		{
			name: "Retrieve user with minimum valid ID",
			id:   1,
			mockDB: func() *gorm.DB {
				db, _ := gorm.Open("sqlite3", ":memory:")
				user := &model.User{
					Model:    gorm.Model{ID: 1},
					Username: "minuser",
					Email:    "min@example.com",
				}
				db.Create(user)
				return db
			},
			expected: &model.User{
				Model:    gorm.Model{ID: 1},
				Username: "minuser",
				Email:    "min@example.com",
			},
			wantErr: false,
		},
		{
			name: "Attempt to retrieve user with ID 0",
			id:   0,
			mockDB: func() *gorm.DB {
				db, _ := gorm.Open("sqlite3", ":memory:")
				return db
			},
			expected: nil,
			wantErr:  true,
		},
		{
			name: "Verify all user fields are correctly populated",
			id:   1,
			mockDB: func() *gorm.DB {
				db, _ := gorm.Open("sqlite3", ":memory:")
				user := &model.User{
					Model:    gorm.Model{ID: 1},
					Username: "fulluser",
					Email:    "full@example.com",
					Password: "fullpass",
					Bio:      "Full bio",
					Image:    "full.jpg",
					Follows:  []model.User{{Model: gorm.Model{ID: 2}}},
				}
				db.Create(user)
				return db
			},
			expected: &model.User{
				Model:    gorm.Model{ID: 1},
				Username: "fulluser",
				Email:    "full@example.com",
				Password: "fullpass",
				Bio:      "Full bio",
				Image:    "full.jpg",
				Follows:  []model.User{{Model: gorm.Model{ID: 2}}},
			},
			wantErr: false,
		},
		{
			name: "Handle maximum uint ID value",
			id:   math.MaxUint32,
			mockDB: func() *gorm.DB {
				db, _ := gorm.Open("sqlite3", ":memory:")
				return db
			},
			expected: nil,
			wantErr:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			db := tt.mockDB()
			s := &UserStore{db: db}

			got, err := s.GetByID(tt.id)

			if tt.wantErr {
				assert.Error(t, err)
				assert.Nil(t, got)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expected, got)
			}

			if tt.expected != nil {
				assert.Equal(t, tt.expected.ID, got.ID)
				assert.Equal(t, tt.expected.Username, got.Username)
				assert.Equal(t, tt.expected.Email, got.Email)
				assert.Equal(t, tt.expected.Password, got.Password)
				assert.Equal(t, tt.expected.Bio, got.Bio)
				assert.Equal(t, tt.expected.Image, got.Image)
				assert.Equal(t, len(tt.expected.Follows), len(got.Follows))

			}
		})
	}
}


/*
ROOST_METHOD_HASH=GetByEmail_3574af40e5
ROOST_METHOD_SIG_HASH=GetByEmail_5731b833c1


 */
func TestUserStoreGetByEmail(t *testing.T) {
	tests := []struct {
		name          string
		setupDB       func() *gorm.DB
		email         string
		expectedUser  *model.User
		expectedError error
	}{
		{
			name: "Successfully retrieve a user by email",
			setupDB: func() *gorm.DB {
				db, _ := gorm.Open("sqlite3", ":memory:")
				db.AutoMigrate(&model.User{})
				db.Create(&model.User{Email: "test@example.com", Username: "testuser"})
				return db
			},
			email: "test@example.com",
			expectedUser: &model.User{
				Email:    "test@example.com",
				Username: "testuser",
			},
			expectedError: nil,
		},
		{
			name: "Attempt to retrieve a non-existent user",
			setupDB: func() *gorm.DB {
				db, _ := gorm.Open("sqlite3", ":memory:")
				db.AutoMigrate(&model.User{})
				return db
			},
			email:         "nonexistent@example.com",
			expectedUser:  nil,
			expectedError: gorm.ErrRecordNotFound,
		},
		{
			name: "Handle database connection error",
			setupDB: func() *gorm.DB {

				db, _ := gorm.Open("sqlite3", ":memory:")
				db.Close()
				return db
			},
			email:         "test@example.com",
			expectedUser:  nil,
			expectedError: errors.New("sql: database is closed"),
		},
		{
			name: "Retrieve user with empty email string",
			setupDB: func() *gorm.DB {
				db, _ := gorm.Open("sqlite3", ":memory:")
				db.AutoMigrate(&model.User{})
				return db
			},
			email:         "",
			expectedUser:  nil,
			expectedError: gorm.ErrRecordNotFound,
		},
		{
			name: "Case sensitivity in email lookup",
			setupDB: func() *gorm.DB {
				db, _ := gorm.Open("sqlite3", ":memory:")
				db.AutoMigrate(&model.User{})
				db.Create(&model.User{Email: "Test@Example.com", Username: "testuser"})
				return db
			},
			email:         "test@example.com",
			expectedUser:  nil,
			expectedError: gorm.ErrRecordNotFound,
		},
		{
			name: "Handling of special characters in email",
			setupDB: func() *gorm.DB {
				db, _ := gorm.Open("sqlite3", ":memory:")
				db.AutoMigrate(&model.User{})
				db.Create(&model.User{Email: "user+test@example.com", Username: "testuser"})
				return db
			},
			email: "user+test@example.com",
			expectedUser: &model.User{
				Email:    "user+test@example.com",
				Username: "testuser",
			},
			expectedError: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			db := tt.setupDB()
			store := &UserStore{db: db}

			user, err := store.GetByEmail(tt.email)

			if tt.expectedError != nil {
				assert.Error(t, err)
				assert.Equal(t, tt.expectedError.Error(), err.Error())
			} else {
				assert.NoError(t, err)
			}

			if tt.expectedUser != nil {
				assert.NotNil(t, user)
				assert.Equal(t, tt.expectedUser.Email, user.Email)
				assert.Equal(t, tt.expectedUser.Username, user.Username)
			} else {
				assert.Nil(t, user)
			}

			db.Close()
		})
	}

}


/*
ROOST_METHOD_HASH=GetByUsername_f11f114df2
ROOST_METHOD_SIG_HASH=GetByUsername_954d096e24


 */
func TestUserStoreGetByUsername(t *testing.T) {
	tests := []struct {
		name            string
		setupDB         func(*gorm.DB)
		username        string
		expectedUser    *model.User
		expectedError   error
		dbError         error
		performanceTest bool
	}{
		{
			name: "Successfully retrieve a user by username",
			setupDB: func(db *gorm.DB) {
				db.Create(&model.User{Username: "testuser", Email: "test@example.com"})
			},
			username: "testuser",
			expectedUser: &model.User{
				Username: "testuser",
				Email:    "test@example.com",
			},
			expectedError: nil,
		},
		{
			name:          "Attempt to retrieve a non-existent user",
			setupDB:       func(db *gorm.DB) {},
			username:      "nonexistent",
			expectedUser:  nil,
			expectedError: gorm.ErrRecordNotFound,
		},
		{
			name:          "Handle database connection error",
			setupDB:       func(db *gorm.DB) {},
			username:      "anyuser",
			expectedUser:  nil,
			expectedError: errors.New("database connection error"),
			dbError:       errors.New("database connection error"),
		},
		{
			name: "Retrieve user with empty username",
			setupDB: func(db *gorm.DB) {
				db.Create(&model.User{Username: "", Email: "empty@example.com"})
			},
			username: "",
			expectedUser: &model.User{
				Username: "",
				Email:    "empty@example.com",
			},
			expectedError: nil,
		},
		{
			name: "Case sensitivity in username retrieval",
			setupDB: func(db *gorm.DB) {
				db.Create(&model.User{Username: "TestUser", Email: "testuser@example.com"})
			},
			username:      "testuser",
			expectedUser:  nil,
			expectedError: gorm.ErrRecordNotFound,
		},
		{
			name: "Performance with large dataset",
			setupDB: func(db *gorm.DB) {
				for i := 0; i < 100000; i++ {
					db.Create(&model.User{Username: "user" + string(rune(i)), Email: "user" + string(rune(i)) + "@example.com"})
				}
				db.Create(&model.User{Username: "lastuser", Email: "lastuser@example.com"})
			},
			username: "lastuser",
			expectedUser: &model.User{
				Username: "lastuser",
				Email:    "lastuser@example.com",
			},
			expectedError:   nil,
			performanceTest: true,
		},
		{
			name: "Handling of special characters in username",
			setupDB: func(db *gorm.DB) {
				db.Create(&model.User{Username: "@user_name!", Email: "special@example.com"})
			},
			username: "@user_name!",
			expectedUser: &model.User{
				Username: "@user_name!",
				Email:    "special@example.com",
			},
			expectedError: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			db, _ := gorm.Open("sqlite3", ":memory:")
			defer db.Close()
			db.AutoMigrate(&model.User{})

			tt.setupDB(db)

			store := &UserStore{db: db}

			if tt.dbError != nil {

				store.db = &gorm.DB{Error: tt.dbError}
			}

			start := time.Now()
			user, err := store.GetByUsername(tt.username)
			duration := time.Since(start)

			if tt.expectedError != nil {
				assert.Error(t, err)
				assert.Equal(t, tt.expectedError, err)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, user)
				assert.Equal(t, tt.expectedUser.Username, user.Username)
				assert.Equal(t, tt.expectedUser.Email, user.Email)
			}

			if tt.performanceTest {
				assert.Less(t, duration, 100*time.Millisecond, "Query took too long")
			}
		})
	}
}


/*
ROOST_METHOD_HASH=Update_68f27dd78a
ROOST_METHOD_SIG_HASH=Update_87150d6435


 */
func TestUserStoreUpdate(t *testing.T) {
	tests := []struct {
		name        string
		setupDB     func(*gorm.DB)
		inputUser   *model.User
		expectedErr error
		validate    func(*testing.T, *gorm.DB, *model.User)
	}{
		{
			name: "Successfully Update User Information",
			setupDB: func(db *gorm.DB) {
				require.NoError(t, db.Create(&model.User{Model: gorm.Model{ID: 1}, Username: "olduser", Email: "old@example.com"}).Error)
			},
			inputUser:   &model.User{Model: gorm.Model{ID: 1}, Username: "newuser", Email: "new@example.com"},
			expectedErr: nil,
			validate: func(t *testing.T, db *gorm.DB, u *model.User) {
				var updatedUser model.User
				require.NoError(t, db.First(&updatedUser, 1).Error)
				assert.Equal(t, "newuser", updatedUser.Username)
				assert.Equal(t, "new@example.com", updatedUser.Email)
			},
		},
		{
			name:        "Attempt to Update Non-existent User",
			setupDB:     func(db *gorm.DB) {},
			inputUser:   &model.User{Model: gorm.Model{ID: 999}, Username: "nonexistent"},
			expectedErr: gorm.ErrRecordNotFound,
			validate:    func(t *testing.T, db *gorm.DB, u *model.User) {},
		},
		{
			name: "Update User with Duplicate Username",
			setupDB: func(db *gorm.DB) {
				require.NoError(t, db.Create(&model.User{Model: gorm.Model{ID: 1}, Username: "user1", Email: "user1@example.com"}).Error)
				require.NoError(t, db.Create(&model.User{Model: gorm.Model{ID: 2}, Username: "user2", Email: "user2@example.com"}).Error)
			},
			inputUser:   &model.User{Model: gorm.Model{ID: 2}, Username: "user1", Email: "user2@example.com"},
			expectedErr: errors.New("UNIQUE constraint failed: users.username"),
			validate: func(t *testing.T, db *gorm.DB, u *model.User) {
				var user model.User
				require.NoError(t, db.First(&user, 2).Error)
				assert.Equal(t, "user2", user.Username)
			},
		},
		{
			name: "Update User with Empty Fields",
			setupDB: func(db *gorm.DB) {
				require.NoError(t, db.Create(&model.User{Model: gorm.Model{ID: 1}, Username: "user", Email: "user@example.com"}).Error)
			},
			inputUser:   &model.User{Model: gorm.Model{ID: 1}, Username: "", Email: ""},
			expectedErr: errors.New("NOT NULL constraint failed: users.username"),
			validate: func(t *testing.T, db *gorm.DB, u *model.User) {
				var user model.User
				require.NoError(t, db.First(&user, 1).Error)
				assert.Equal(t, "user", user.Username)
				assert.Equal(t, "user@example.com", user.Email)
			},
		},
		{
			name: "Partial Update of User Information",
			setupDB: func(db *gorm.DB) {
				require.NoError(t, db.Create(&model.User{Model: gorm.Model{ID: 1}, Username: "user", Email: "user@example.com", Bio: "Old bio"}).Error)
			},
			inputUser:   &model.User{Model: gorm.Model{ID: 1}, Bio: "New bio"},
			expectedErr: nil,
			validate: func(t *testing.T, db *gorm.DB, u *model.User) {
				var updatedUser model.User
				require.NoError(t, db.First(&updatedUser, 1).Error)
				assert.Equal(t, "user", updatedUser.Username)
				assert.Equal(t, "user@example.com", updatedUser.Email)
				assert.Equal(t, "New bio", updatedUser.Bio)
			},
		},
		{
			name: "Update User with Very Long Input",
			setupDB: func(db *gorm.DB) {
				require.NoError(t, db.Create(&model.User{Model: gorm.Model{ID: 1}, Username: "user", Email: "user@example.com"}).Error)
			},
			inputUser:   &model.User{Model: gorm.Model{ID: 1}, Bio: string(make([]byte, 10000))},
			expectedErr: nil,
			validate: func(t *testing.T, db *gorm.DB, u *model.User) {
				var updatedUser model.User
				require.NoError(t, db.First(&updatedUser, 1).Error)
				assert.Equal(t, 10000, len(updatedUser.Bio))
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			db, err := gorm.Open("sqlite3", ":memory:")
			require.NoError(t, err)
			defer db.Close()

			require.NoError(t, db.AutoMigrate(&model.User{}).Error)

			tt.setupDB(db)

			userStore := &UserStore{db: db}

			err = userStore.Update(tt.inputUser)

			if tt.expectedErr != nil {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.expectedErr.Error())
			} else {
				assert.NoError(t, err)
			}

			tt.validate(t, db, tt.inputUser)
		})
	}
}


/*
ROOST_METHOD_HASH=Follow_48fdf1257b
ROOST_METHOD_SIG_HASH=Follow_8217e61c06


 */
func (m *mockAssociation) Append(values ...interface{}) error {
	args := m.Called(values...)
	return args.Error(0)
}

func (m *mockDB) Association(column string) *gorm.Association {
	args := m.Called(column)
	return args.Get(0).(*gorm.Association)
}

func (m *mockDB) Model(value interface{}) *gorm.DB {
	args := m.Called(value)
	return args.Get(0).(*gorm.DB)
}

func TestUserStoreFollow(t *testing.T) {
	tests := []struct {
		name     string
		follower *model.User
		followed *model.User
		dbError  error
		expected error
	}{
		{
			name:     "Successfully follow a user",
			follower: &model.User{Username: "userA"},
			followed: &model.User{Username: "userB"},
			dbError:  nil,
			expected: nil,
		},
		{
			name:     "Attempt to follow a user that is already being followed",
			follower: &model.User{Username: "userA", Follows: []model.User{{Username: "userB"}}},
			followed: &model.User{Username: "userB"},
			dbError:  nil,
			expected: nil,
		},
		{
			name:     "Follow operation with a nil follower user",
			follower: nil,
			followed: &model.User{Username: "userB"},
			dbError:  nil,
			expected: errors.New("follower user is nil"),
		},
		{
			name:     "Follow operation with a nil user to be followed",
			follower: &model.User{Username: "userA"},
			followed: nil,
			dbError:  nil,
			expected: errors.New("user to be followed is nil"),
		},
		{
			name:     "Follow operation with database error",
			follower: &model.User{Username: "userA"},
			followed: &model.User{Username: "userB"},
			dbError:  errors.New("database error"),
			expected: errors.New("database error"),
		},
		{
			name:     "Follow operation with large number of follows",
			follower: &model.User{Username: "userA", Follows: make([]model.User, 1000)},
			followed: &model.User{Username: "userB"},
			dbError:  nil,
			expected: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockDB := new(mockDB)
			mockAssoc := new(mockAssociation)

			userStore := &UserStore{
				db: mockDB,
			}

			if tt.follower == nil {
				mockDB.On("Model", (*model.User)(nil)).Return(mockDB)
			} else {
				mockDB.On("Model", tt.follower).Return(mockDB)
			}

			mockDB.On("Association", "Follows").Return(mockAssoc)
			mockAssoc.On("Append", tt.followed).Return(tt.dbError)

			err := userStore.Follow(tt.follower, tt.followed)

			if tt.expected == nil {
				assert.NoError(t, err)
			} else {
				assert.EqualError(t, err, tt.expected.Error())
			}

			mockDB.AssertExpectations(t)
			mockAssoc.AssertExpectations(t)
		})
	}
}


/*
ROOST_METHOD_HASH=IsFollowing_f53a5d9cef
ROOST_METHOD_SIG_HASH=IsFollowing_9eba5a0e9c


 */
func TestUserStoreIsFollowing(t *testing.T) {
	tests := []struct {
		name     string
		setup    func(*gorm.DB) (*model.User, *model.User)
		wantBool bool
		wantErr  error
	}{
		{
			name: "User A is following User B",
			setup: func(db *gorm.DB) (*model.User, *model.User) {
				userA := &model.User{Username: "UserA"}
				userB := &model.User{Username: "UserB"}
				db.Create(userA)
				db.Create(userB)
				db.Exec("INSERT INTO follows (from_user_id, to_user_id) VALUES (?, ?)", userA.ID, userB.ID)
				return userA, userB
			},
			wantBool: true,
			wantErr:  nil,
		},
		{
			name: "User A is not following User B",
			setup: func(db *gorm.DB) (*model.User, *model.User) {
				userA := &model.User{Username: "UserA"}
				userB := &model.User{Username: "UserB"}
				db.Create(userA)
				db.Create(userB)
				return userA, userB
			},
			wantBool: false,
			wantErr:  nil,
		},
		{
			name: "Null user arguments",
			setup: func(db *gorm.DB) (*model.User, *model.User) {
				return nil, nil
			},
			wantBool: false,
			wantErr:  nil,
		},
		{
			name: "Database error",
			setup: func(db *gorm.DB) (*model.User, *model.User) {
				userA := &model.User{Username: "UserA"}
				userB := &model.User{Username: "UserB"}
				db.Create(userA)
				db.Create(userB)

				db.Close()
				return userA, userB
			},
			wantBool: false,
			wantErr:  errors.New("database error"),
		},
		{
			name: "User following themselves",
			setup: func(db *gorm.DB) (*model.User, *model.User) {
				userA := &model.User{Username: "UserA"}
				db.Create(userA)
				db.Exec("INSERT INTO follows (from_user_id, to_user_id) VALUES (?, ?)", userA.ID, userA.ID)
				return userA, userA
			},
			wantBool: true,
			wantErr:  nil,
		},
		{
			name: "Large number of follows",
			setup: func(db *gorm.DB) (*model.User, *model.User) {
				userA := &model.User{Username: "UserA"}
				userB := &model.User{Username: "UserB"}
				db.Create(userA)
				db.Create(userB)

				for i := 0; i < 1000; i++ {
					db.Exec("INSERT INTO follows (from_user_id, to_user_id) VALUES (?, ?)", userA.ID, i)
				}
				db.Exec("INSERT INTO follows (from_user_id, to_user_id) VALUES (?, ?)", userA.ID, userB.ID)
				return userA, userB
			},
			wantBool: true,
			wantErr:  nil,
		},
		{
			name: "Recently added follow relationship",
			setup: func(db *gorm.DB) (*model.User, *model.User) {
				userA := &model.User{Username: "UserA"}
				userB := &model.User{Username: "UserB"}
				db.Create(userA)
				db.Create(userB)
				db.Exec("INSERT INTO follows (from_user_id, to_user_id) VALUES (?, ?)", userA.ID, userB.ID)
				return userA, userB
			},
			wantBool: true,
			wantErr:  nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			db, err := gorm.Open("sqlite3", ":memory:")
			require.NoError(t, err)
			defer db.Close()

			err = db.AutoMigrate(&model.User{}).Error
			require.NoError(t, err)
			err = db.Exec("CREATE TABLE IF NOT EXISTS follows (from_user_id INTEGER, to_user_id INTEGER)").Error
			require.NoError(t, err)

			userA, userB := tt.setup(db)

			userStore := &UserStore{db: db}

			gotBool, gotErr := userStore.IsFollowing(userA, userB)

			assert.Equal(t, tt.wantBool, gotBool)
			if tt.wantErr != nil {
				assert.Error(t, gotErr)
				assert.Contains(t, gotErr.Error(), tt.wantErr.Error())
			} else {
				assert.NoError(t, gotErr)
			}
		})
	}
}


/*
ROOST_METHOD_HASH=Unfollow_57959a8a53
ROOST_METHOD_SIG_HASH=Unfollow_8bd8e0bc55


 */
func TestUserStoreUnfollow(t *testing.T) {
	tests := []struct {
		name          string
		setupFunc     func(*gorm.DB) (*model.User, *model.User)
		expectedError error
		validateFunc  func(*testing.T, *gorm.DB, *model.User, *model.User)
	}{
		{
			name: "Successful Unfollow Operation",
			setupFunc: func(db *gorm.DB) (*model.User, *model.User) {
				userA := &model.User{Username: "userA", Email: "userA@example.com"}
				userB := &model.User{Username: "userB", Email: "userB@example.com"}
				db.Create(userA)
				db.Create(userB)
				db.Model(userA).Association("Follows").Append(userB)
				return userA, userB
			},
			expectedError: nil,
			validateFunc: func(t *testing.T, db *gorm.DB, userA *model.User, userB *model.User) {
				var follows []model.User
				db.Model(userA).Association("Follows").Find(&follows)
				assert.NotContains(t, follows, userB)
			},
		},
		{
			name: "Unfollow a User That Is Not Being Followed",
			setupFunc: func(db *gorm.DB) (*model.User, *model.User) {
				userA := &model.User{Username: "userA", Email: "userA@example.com"}
				userB := &model.User{Username: "userB", Email: "userB@example.com"}
				db.Create(userA)
				db.Create(userB)
				return userA, userB
			},
			expectedError: nil,
			validateFunc: func(t *testing.T, db *gorm.DB, userA *model.User, userB *model.User) {
				var follows []model.User
				db.Model(userA).Association("Follows").Find(&follows)
				assert.NotContains(t, follows, userB)
			},
		},
		{
			name: "Unfollow with Non-Existent User (Follower)",
			setupFunc: func(db *gorm.DB) (*model.User, *model.User) {
				userA := &model.User{Username: "userA", Email: "userA@example.com"}
				userB := &model.User{Username: "userB", Email: "userB@example.com"}
				db.Create(userB)
				return userA, userB
			},
			expectedError: gorm.ErrRecordNotFound,
			validateFunc:  nil,
		},
		{
			name: "Unfollow a Non-Existent User (Followee)",
			setupFunc: func(db *gorm.DB) (*model.User, *model.User) {
				userA := &model.User{Username: "userA", Email: "userA@example.com"}
				userB := &model.User{Username: "userB", Email: "userB@example.com"}
				db.Create(userA)
				return userA, userB
			},
			expectedError: gorm.ErrRecordNotFound,
			validateFunc:  nil,
		},
		{
			name: "Database Connection Error",
			setupFunc: func(db *gorm.DB) (*model.User, *model.User) {
				userA := &model.User{Username: "userA", Email: "userA@example.com"}
				userB := &model.User{Username: "userB", Email: "userB@example.com"}
				return userA, userB
			},
			expectedError: errors.New("database connection error"),
			validateFunc:  nil,
		},
		{
			name: "Unfollow Self",
			setupFunc: func(db *gorm.DB) (*model.User, *model.User) {
				userA := &model.User{Username: "userA", Email: "userA@example.com"}
				db.Create(userA)
				return userA, userA
			},
			expectedError: nil,
			validateFunc: func(t *testing.T, db *gorm.DB, userA *model.User, _ *model.User) {
				var follows []model.User
				db.Model(userA).Association("Follows").Find(&follows)
				assert.NotContains(t, follows, userA)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			db, err := gorm.Open("sqlite3", ":memory:")
			require.NoError(t, err)
			defer db.Close()

			db.AutoMigrate(&model.User{})

			userStore := &UserStore{db: db}

			userA, userB := tt.setupFunc(db)

			if tt.name == "Database Connection Error" {
				userStore.db = &gorm.DB{Error: tt.expectedError}
			}

			err = userStore.Unfollow(userA, userB)

			assert.Equal(t, tt.expectedError, err)

			if tt.validateFunc != nil {
				tt.validateFunc(t, db, userA, userB)
			}
		})
	}
}


/*
ROOST_METHOD_HASH=GetFollowingUserIDs_ba3670aa2c
ROOST_METHOD_SIG_HASH=GetFollowingUserIDs_55ccc2afd7


 */
func TestUserStoreGetFollowingUserIDs(t *testing.T) {
	tests := []struct {
		name    string
		user    *model.User
		mockDB  func(mock sqlmock.Sqlmock)
		want    []uint
		wantErr bool
	}{
		{
			name: "Successful retrieval of following user IDs",
			user: &model.User{Model: gorm.Model{ID: 1}},
			mockDB: func(mock sqlmock.Sqlmock) {
				rows := sqlmock.NewRows([]string{"to_user_id"}).
					AddRow(2).
					AddRow(3).
					AddRow(4)
				mock.ExpectQuery("SELECT to_user_id FROM follows WHERE from_user_id = ?").
					WithArgs(1).
					WillReturnRows(rows)
			},
			want:    []uint{2, 3, 4},
			wantErr: false,
		},
		{
			name: "User with no followers",
			user: &model.User{Model: gorm.Model{ID: 1}},
			mockDB: func(mock sqlmock.Sqlmock) {
				rows := sqlmock.NewRows([]string{"to_user_id"})
				mock.ExpectQuery("SELECT to_user_id FROM follows WHERE from_user_id = ?").
					WithArgs(1).
					WillReturnRows(rows)
			},
			want:    []uint{},
			wantErr: false,
		},
		{
			name: "Database error handling",
			user: &model.User{Model: gorm.Model{ID: 1}},
			mockDB: func(mock sqlmock.Sqlmock) {
				mock.ExpectQuery("SELECT to_user_id FROM follows WHERE from_user_id = ?").
					WithArgs(1).
					WillReturnError(errors.New("database error"))
			},
			want:    nil,
			wantErr: true,
		},
		{
			name: "Large number of followers",
			user: &model.User{Model: gorm.Model{ID: 1}},
			mockDB: func(mock sqlmock.Sqlmock) {
				rows := sqlmock.NewRows([]string{"to_user_id"})
				for i := 2; i <= 1001; i++ {
					rows.AddRow(uint(i))
				}
				mock.ExpectQuery("SELECT to_user_id FROM follows WHERE from_user_id = ?").
					WithArgs(1).
					WillReturnRows(rows)
			},
			want: func() []uint {
				ids := make([]uint, 1000)
				for i := range ids {
					ids[i] = uint(i + 2)
				}
				return ids
			}(),
			wantErr: false,
		},
		{
			name: "Deleted user handling",
			user: &model.User{Model: gorm.Model{ID: 1}},
			mockDB: func(mock sqlmock.Sqlmock) {
				rows := sqlmock.NewRows([]string{"to_user_id"}).
					AddRow(2).
					AddRow(4)
				mock.ExpectQuery("SELECT to_user_id FROM follows WHERE from_user_id = ?").
					WithArgs(1).
					WillReturnRows(rows)
			},
			want:    []uint{2, 4},
			wantErr: false,
		},
		{
			name: "User not found in database",
			user: &model.User{Model: gorm.Model{ID: 999}},
			mockDB: func(mock sqlmock.Sqlmock) {
				rows := sqlmock.NewRows([]string{"to_user_id"})
				mock.ExpectQuery("SELECT to_user_id FROM follows WHERE from_user_id = ?").
					WithArgs(999).
					WillReturnRows(rows)
			},
			want:    []uint{},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			db, mock, err := sqlmock.New()
			if err != nil {
				t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
			}
			defer db.Close()

			gdb, err := gorm.Open("mysql", db)
			if err != nil {
				t.Fatalf("an error '%s' was not expected when opening a gorm database", err)
			}
			defer gdb.Close()

			tt.mockDB(mock)

			s := &UserStore{db: gdb}

			got, err := s.GetFollowingUserIDs(tt.user)
			if (err != nil) != tt.wantErr {
				t.Errorf("UserStore.GetFollowingUserIDs() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("UserStore.GetFollowingUserIDs() = %v, want %v", got, tt.want)
			}

			if err := mock.ExpectationsWereMet(); err != nil {
				t.Errorf("there were unfulfilled expectations: %s", err)
			}
		})
	}
}


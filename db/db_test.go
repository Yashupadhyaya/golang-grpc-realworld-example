package db

import (
	"os"
	"testing"
	"database/sql"
	"errors"
	"sync"
	"time"
	"github.com/jinzhu/gorm"
	"github.com/raahii/golang-grpc-realworld-example/model"
	"io/ioutil"
	"github.com/BurntSushi/toml"
	"github.com/DATA-DOG/go-txdb"
	"github.com/joho/godotenv"
)




var autoMigrateCalled bool

type MockDB struct {
	AutoMigrateFunc func(...interface{}) *gorm.DB
}
type mockDB struct {
	closed bool
	err    error
}
type mockDB struct {
	createError  error
	createdUsers int
}
type mockDB struct {
	gorm.DB
	openError    error
	maxIdleConns int
	logMode      bool
}
type mockDB struct {
	*gorm.DB
}


/*
ROOST_METHOD_HASH=dsn_e202d1c4f9
ROOST_METHOD_SIG_HASH=dsn_b336e03d64

FUNCTION_DEF=func dsn() (string, error) 

 */
func TestDsn(t *testing.T) {
	tests := []struct {
		name     string
		envVars  map[string]string
		expected string
		wantErr  bool
		errMsg   string
	}{
		{
			name: "All Environment Variables Set Correctly",
			envVars: map[string]string{
				"DB_HOST":     "localhost",
				"DB_USER":     "user",
				"DB_PASSWORD": "password",
				"DB_NAME":     "testdb",
				"DB_PORT":     "3306",
			},
			expected: "user:password@(localhost:3306)/testdb?charset=utf8mb4&parseTime=True&loc=Local",
			wantErr:  false,
		},
		{
			name: "Missing DB_HOST Environment Variable",
			envVars: map[string]string{
				"DB_USER":     "user",
				"DB_PASSWORD": "password",
				"DB_NAME":     "testdb",
				"DB_PORT":     "3306",
			},
			wantErr: true,
			errMsg:  "$DB_HOST is not set",
		},
		{
			name: "Missing DB_USER Environment Variable",
			envVars: map[string]string{
				"DB_HOST":     "localhost",
				"DB_PASSWORD": "password",
				"DB_NAME":     "testdb",
				"DB_PORT":     "3306",
			},
			wantErr: true,
			errMsg:  "$DB_USER is not set",
		},
		{
			name: "Missing DB_PASSWORD Environment Variable",
			envVars: map[string]string{
				"DB_HOST": "localhost",
				"DB_USER": "user",
				"DB_NAME": "testdb",
				"DB_PORT": "3306",
			},
			wantErr: true,
			errMsg:  "$DB_PASSWORD is not set",
		},
		{
			name: "Missing DB_NAME Environment Variable",
			envVars: map[string]string{
				"DB_HOST":     "localhost",
				"DB_USER":     "user",
				"DB_PASSWORD": "password",
				"DB_PORT":     "3306",
			},
			wantErr: true,
			errMsg:  "$DB_NAME is not set",
		},
		{
			name: "Missing DB_PORT Environment Variable",
			envVars: map[string]string{
				"DB_HOST":     "localhost",
				"DB_USER":     "user",
				"DB_PASSWORD": "password",
				"DB_NAME":     "testdb",
			},
			wantErr: true,
			errMsg:  "$DB_PORT is not set",
		},
		{
			name: "All Environment Variables Set with Empty Values",
			envVars: map[string]string{
				"DB_HOST":     "",
				"DB_USER":     "",
				"DB_PASSWORD": "",
				"DB_NAME":     "",
				"DB_PORT":     "",
			},
			expected: ":@(:)/?charset=utf8mb4&parseTime=True&loc=Local",
			wantErr:  false,
		},
		{
			name: "Special Characters in Environment Variables",
			envVars: map[string]string{
				"DB_HOST":     "localhost",
				"DB_USER":     "user@123",
				"DB_PASSWORD": "p@ssw0rd!",
				"DB_NAME":     "test_db",
				"DB_PORT":     "3306",
			},
			expected: "user@123:p@ssw0rd!@(localhost:3306)/test_db?charset=utf8mb4&parseTime=True&loc=Local",
			wantErr:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			os.Clearenv()

			for key, value := range tt.envVars {
				os.Setenv(key, value)
			}

			got, err := dsn()

			if (err != nil) != tt.wantErr {
				t.Errorf("dsn() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if tt.wantErr && err.Error() != tt.errMsg {
				t.Errorf("dsn() error message = %v, want %v", err.Error(), tt.errMsg)
				return
			}

			if !tt.wantErr && got != tt.expected {
				t.Errorf("dsn() = %v, want %v", got, tt.expected)
			}
		})
	}
}


/*
ROOST_METHOD_HASH=AutoMigrate_94b22622a5
ROOST_METHOD_SIG_HASH=AutoMigrate_2cd152caa7

FUNCTION_DEF=func AutoMigrate(db *gorm.DB) error 

 */
func (m *MockDB) AutoMigrate(values ...interface{}) *gorm.DB {
	return m.AutoMigrateFunc(values...)
}

func TestAutoMigrate(t *testing.T) {
	tests := []struct {
		name    string
		dbSetup func() *gorm.DB
		wantErr bool
	}{
		{
			name: "Successful Auto-Migration",
			dbSetup: func() *gorm.DB {
				mockDB := &MockDB{
					AutoMigrateFunc: func(values ...interface{}) *gorm.DB {
						return &gorm.DB{Error: nil}
					},
				}
				return &gorm.DB{Value: mockDB}
			},
			wantErr: false,
		},
		{
			name: "Database Connection Error",
			dbSetup: func() *gorm.DB {
				mockDB := &MockDB{
					AutoMigrateFunc: func(values ...interface{}) *gorm.DB {
						return &gorm.DB{Error: errors.New("database connection error")}
					},
				}
				return &gorm.DB{Value: mockDB}
			},
			wantErr: true,
		},
		{
			name: "Partial Migration Failure",
			dbSetup: func() *gorm.DB {
				callCount := 0
				mockDB := &MockDB{
					AutoMigrateFunc: func(values ...interface{}) *gorm.DB {
						callCount++
						if callCount > 2 {
							return &gorm.DB{Error: errors.New("migration failed for some models")}
						}
						return &gorm.DB{Error: nil}
					},
				}
				return &gorm.DB{Value: mockDB}
			},
			wantErr: true,
		},
		{
			name: "Concurrent Auto-Migration Attempts",
			dbSetup: func() *gorm.DB {
				var mu sync.Mutex
				callCount := 0
				mockDB := &MockDB{
					AutoMigrateFunc: func(values ...interface{}) *gorm.DB {
						mu.Lock()
						defer mu.Unlock()
						callCount++
						time.Sleep(10 * time.Millisecond)
						return &gorm.DB{Error: nil}
					},
				}
				return &gorm.DB{Value: mockDB}
			},
			wantErr: false,
		},
		{
			name: "Auto-Migration with Existing Schema",
			dbSetup: func() *gorm.DB {
				mockDB := &MockDB{
					AutoMigrateFunc: func(values ...interface{}) *gorm.DB {
						return &gorm.DB{Error: nil}
					},
				}
				return &gorm.DB{Value: mockDB}
			},
			wantErr: false,
		},
		{
			name: "Performance Test for Large Schema",
			dbSetup: func() *gorm.DB {
				mockDB := &MockDB{
					AutoMigrateFunc: func(values ...interface{}) *gorm.DB {
						time.Sleep(100 * time.Millisecond)
						return &gorm.DB{Error: nil}
					},
				}
				return &gorm.DB{Value: mockDB}
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			db := tt.dbSetup()
			err := AutoMigrate(db)

			if (err != nil) != tt.wantErr {
				t.Errorf("AutoMigrate() error = %v, wantErr %v", err, tt.wantErr)
			}

			if tt.name == "Concurrent Auto-Migration Attempts" {
				var wg sync.WaitGroup
				for i := 0; i < 5; i++ {
					wg.Add(1)
					go func() {
						defer wg.Done()
						err := AutoMigrate(db)
						if err != nil {
							t.Errorf("Concurrent AutoMigrate() error = %v", err)
						}
					}()
				}
				wg.Wait()
			}

			if tt.name == "Performance Test for Large Schema" {
				start := time.Now()
				err := AutoMigrate(db)
				duration := time.Since(start)
				if err != nil {
					t.Errorf("Performance AutoMigrate() error = %v", err)
				}
				if duration > 500*time.Millisecond {
					t.Errorf("Performance AutoMigrate() took too long: %v", duration)
				}
			}
		})
	}
}


/*
ROOST_METHOD_HASH=DropTestDB_4c6b54d5e5
ROOST_METHOD_SIG_HASH=DropTestDB_69b51a825b

FUNCTION_DEF=func DropTestDB(d *gorm.DB) error 

 */
func (m *mockDB) Close() error {
	m.closed = true
	return m.err
}

func TestDropTestDb(t *testing.T) {
	tests := []struct {
		name    string
		db      *gorm.DB
		wantErr bool
	}{
		{
			name: "Successfully Close Database Connection",
			db: &gorm.DB{
				db: &mockDB{},
			},
			wantErr: false,
		},
		{
			name: "Handle Already Closed Database",
			db: &gorm.DB{
				db: &mockDB{closed: true},
			},
			wantErr: false,
		},
		{
			name:    "Handle Nil Database Pointer",
			db:      nil,
			wantErr: false,
		},
		{
			name: "Verify No Side Effects",
			db: &gorm.DB{
				db: &mockDB{},
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := DropTestDB(tt.db)
			if (err != nil) != tt.wantErr {
				t.Errorf("DropTestDB() error = %v, wantErr %v", err, tt.wantErr)
			}

			if tt.db != nil {
				mockDB := tt.db.db.(*mockDB)
				if !mockDB.closed {
					t.Errorf("DropTestDB() did not close the database connection")
				}
			}
		})
	}
}

func TestDropTestDbConcurrent(t *testing.T) {
	const numGoroutines = 10
	var wg sync.WaitGroup
	wg.Add(numGoroutines)

	for i := 0; i < numGoroutines; i++ {
		go func() {
			defer wg.Done()
			db := &gorm.DB{
				db: &mockDB{},
			}
			err := DropTestDB(db)
			if err != nil {
				t.Errorf("DropTestDB() error = %v", err)
			}
			mockDB := db.db.(*mockDB)
			if !mockDB.closed {
				t.Errorf("DropTestDB() did not close the database connection")
			}
		}()
	}

	wg.Wait()
}


/*
ROOST_METHOD_HASH=Seed_5ad31c3a6c
ROOST_METHOD_SIG_HASH=Seed_878933cebc

FUNCTION_DEF=func Seed(db *gorm.DB) error 

 */
func (m *mockDB) Create(value interface{}) *gorm.DB {
	if m.createError != nil {
		return &gorm.DB{Error: m.createError}
	}
	m.createdUsers++
	return &gorm.DB{}
}

func TestSeed(t *testing.T) {
	tests := []struct {
		name          string
		setupMock     func() *mockDB
		setupTOML     func() error
		expectedError error
		expectedUsers int
	}{
		{
			name: "Successful Seeding of Users",
			setupMock: func() *mockDB {
				return &mockDB{}
			},
			setupTOML: func() error {
				content := `
				[[Users]]
				username = "user1"
				email = "user1@example.com"
				password = "password1"

				[[Users]]
				username = "user2"
				email = "user2@example.com"
				password = "password2"
				`
				return ioutil.WriteFile("db/seed/users.toml", []byte(content), 0644)
			},
			expectedError: nil,
			expectedUsers: 2,
		},
		{
			name: "Handling Non-Existent TOML File",
			setupMock: func() *mockDB {
				return &mockDB{}
			},
			setupTOML: func() error {
				return os.Remove("db/seed/users.toml")
			},
			expectedError: errors.New("open db/seed/users.toml: no such file or directory"),
			expectedUsers: 0,
		},
		{
			name: "Handling Invalid TOML File Format",
			setupMock: func() *mockDB {
				return &mockDB{}
			},
			setupTOML: func() error {
				content := `
				[[Users]
				username = "user1"
				email = "user1@example.com"
				password = "password1"
				`
				return ioutil.WriteFile("db/seed/users.toml", []byte(content), 0644)
			},
			expectedError: errors.New("toml: line 2: expected '=', '.' or ']' after a key"),
			expectedUsers: 0,
		},
		{
			name: "Database Connection Failure",
			setupMock: func() *mockDB {
				return &mockDB{createError: errors.New("database connection failed")}
			},
			setupTOML: func() error {
				content := `
				[[Users]]
				username = "user1"
				email = "user1@example.com"
				password = "password1"
				`
				return ioutil.WriteFile("db/seed/users.toml", []byte(content), 0644)
			},
			expectedError: errors.New("database connection failed"),
			expectedUsers: 0,
		},
		{
			name: "Partial Seeding Due to Database Error",
			setupMock: func() *mockDB {
				mock := &mockDB{}
				mock.createError = errors.New("database error after partial seeding")
				return mock
			},
			setupTOML: func() error {
				content := `
				[[Users]]
				username = "user1"
				email = "user1@example.com"
				password = "password1"

				[[Users]]
				username = "user2"
				email = "user2@example.com"
				password = "password2"
				`
				return ioutil.WriteFile("db/seed/users.toml", []byte(content), 0644)
			},
			expectedError: errors.New("database error after partial seeding"),
			expectedUsers: 1,
		},
		{
			name: "Empty TOML File",
			setupMock: func() *mockDB {
				return &mockDB{}
			},
			setupTOML: func() error {
				return ioutil.WriteFile("db/seed/users.toml", []byte(""), 0644)
			},
			expectedError: nil,
			expectedUsers: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			mockDB := tt.setupMock()
			err := tt.setupTOML()
			if err != nil {
				t.Fatalf("Failed to setup TOML file: %v", err)
			}

			err = Seed(mockDB)

			if (err != nil && tt.expectedError == nil) || (err == nil && tt.expectedError != nil) || (err != nil && tt.expectedError != nil && err.Error() != tt.expectedError.Error()) {
				t.Errorf("Expected error: %v, got: %v", tt.expectedError, err)
			}

			if mockDB.createdUsers != tt.expectedUsers {
				t.Errorf("Expected %d users to be created, got %d", tt.expectedUsers, mockDB.createdUsers)
			}

			os.Remove("db/seed/users.toml")
		})
	}
}


/*
ROOST_METHOD_HASH=New_1d2840dc39
ROOST_METHOD_SIG_HASH=New_f9cc65f555

FUNCTION_DEF=func New() (*gorm.DB, error) 

 */
func (m *mockDB) DB() *sql.DB {
	return &sql.DB{}
}

func (m *mockDB) LogMode(enable bool) *gorm.DB {
	m.logMode = enable
	return &m.DB
}

func (m *mockDB) Open(dialect string, args ...interface{}) (db *gorm.DB, err error) {
	if m.openError != nil {
		return nil, m.openError
	}
	return &m.DB, nil
}

func (m *mockDB) SetMaxIdleConns(n int) {
	m.maxIdleConns = n
}

func TestNew(t *testing.T) {
	originalDSN := dsn
	defer func() { dsn = originalDSN }()

	tests := []struct {
		name            string
		mockDB          *mockDB
		mockDSN         func() (string, error)
		expectedDB      bool
		expectedError   bool
		setupEnv        func()
		cleanupEnv      func()
		concurrentCalls int
	}{
		{
			name:       "Successful Database Connection",
			mockDB:     &mockDB{},
			mockDSN:    mockDSN("valid_dsn", nil),
			expectedDB: true,
		},
		{
			name:       "Database Connection Retry",
			mockDB:     &mockDB{openError: errors.New("connection failed")},
			mockDSN:    mockDSN("valid_dsn", nil),
			expectedDB: true,
		},
		{
			name:          "Maximum Retry Limit Reached",
			mockDB:        &mockDB{openError: errors.New("connection failed")},
			mockDSN:       mockDSN("valid_dsn", nil),
			expectedError: true,
		},
		{
			name:          "Invalid DSN",
			mockDB:        &mockDB{},
			mockDSN:       mockDSN("", errors.New("invalid DSN")),
			expectedError: true,
		},
		{
			name:       "Correct Database Configuration",
			mockDB:     &mockDB{},
			mockDSN:    mockDSN("valid_dsn", nil),
			expectedDB: true,
		},
		{
			name:          "Environment Variable Dependency",
			mockDB:        &mockDB{},
			mockDSN:       mockDSN("", errors.New("missing environment variable")),
			expectedError: true,
			setupEnv: func() {
				os.Unsetenv("DB_HOST")
			},
			cleanupEnv: func() {
				os.Setenv("DB_HOST", "localhost")
			},
		},
		{
			name:            "Concurrent Access",
			mockDB:          &mockDB{},
			mockDSN:         mockDSN("valid_dsn", nil),
			expectedDB:      true,
			concurrentCalls: 10,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.setupEnv != nil {
				tt.setupEnv()
			}
			if tt.cleanupEnv != nil {
				defer tt.cleanupEnv()
			}

			dsn = tt.mockDSN
			gorm.Open = tt.mockDB.Open

			if tt.concurrentCalls > 0 {
				var wg sync.WaitGroup
				for i := 0; i < tt.concurrentCalls; i++ {
					wg.Add(1)
					go func() {
						defer wg.Done()
						db, err := New()
						if (err != nil) != tt.expectedError {
							t.Errorf("New() error = %v, expectedError %v", err, tt.expectedError)
						}
						if (db != nil) != tt.expectedDB {
							t.Errorf("New() db = %v, expectedDB %v", db, tt.expectedDB)
						}
					}()
				}
				wg.Wait()
			} else {
				db, err := New()
				if (err != nil) != tt.expectedError {
					t.Errorf("New() error = %v, expectedError %v", err, tt.expectedError)
				}
				if (db != nil) != tt.expectedDB {
					t.Errorf("New() db = %v, expectedDB %v", db, tt.expectedDB)
				}

				if db != nil {
					mockDB := tt.mockDB
					if mockDB.maxIdleConns != 3 {
						t.Errorf("Expected max idle connections to be 3, got %d", mockDB.maxIdleConns)
					}
					if mockDB.logMode != false {
						t.Errorf("Expected log mode to be false, got %v", mockDB.logMode)
					}
				}
			}
		})
	}
}

func mockDSN(s string, err error) func() (string, error) {
	return func() (string, error) {
		return s, err
	}
}


/*
ROOST_METHOD_HASH=NewTestDB_7feb2c4a7a
ROOST_METHOD_SIG_HASH=NewTestDB_1b71546d9d

FUNCTION_DEF=func NewTestDB() (*gorm.DB, error) 

 */
func (m *mockDB) LogMode(enable bool) *gorm.DB {
	return m.DB
}

func (m *mockDB) SetMaxIdleConns(n int) *gorm.DB {
	return m.DB
}

func TestNewTestDb(t *testing.T) {
	originalGodotenvLoad := godotenv.Load
	originalGormOpen := gorm.Open
	originalSqlOpen := sql.Open
	originalTxdbRegister := txdb.Register

	defer func() {
		godotenv.Load = originalGodotenvLoad
		gorm.Open = originalGormOpen
		sql.Open = originalSqlOpen
		txdb.Register = originalTxdbRegister
	}()

	tests := []struct {
		name           string
		envFileExists  bool
		envFileContent string
		gormOpenError  error
		sqlOpenError   error
		wantErr        bool
	}{
		{
			name:           "Successful Database Connection",
			envFileExists:  true,
			envFileContent: "DB_USER=testuser\nDB_PASSWORD=testpass\nDB_NAME=testdb\nDB_HOST=localhost\nDB_PORT=3306",
			gormOpenError:  nil,
			sqlOpenError:   nil,
			wantErr:        false,
		},
		{
			name:          "Environment File Not Found",
			envFileExists: false,
			wantErr:       true,
		},
		{
			name:           "Invalid Database Credentials",
			envFileExists:  true,
			envFileContent: "DB_USER=invalid\nDB_PASSWORD=invalid\nDB_NAME=invalid\nDB_HOST=invalid\nDB_PORT=invalid",
			gormOpenError:  errors.New("invalid credentials"),
			wantErr:        true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			if tt.envFileExists {
				err := os.WriteFile("../env/test.env", []byte(tt.envFileContent), 0644)
				if err != nil {
					t.Fatal(err)
				}
				defer os.Remove("../env/test.env")
			}

			godotenv.Load = func(filenames ...string) error {
				if !tt.envFileExists {
					return errors.New("env file not found")
				}
				return nil
			}

			gorm.Open = func(dialect string, args ...interface{}) (*gorm.DB, error) {
				if tt.gormOpenError != nil {
					return nil, tt.gormOpenError
				}
				return &gorm.DB{}, nil
			}

			sql.Open = func(driverName, dataSourceName string) (*sql.DB, error) {
				if tt.sqlOpenError != nil {
					return nil, tt.sqlOpenError
				}
				return &sql.DB{}, nil
			}

			txdb.Register = func(name, driver, dsn string) {}

			txdbInitialized = false
			autoMigrateCalled = false

			db, err := NewTestDB()

			if (err != nil) != tt.wantErr {
				t.Errorf("NewTestDB() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr {
				if db == nil {
					t.Error("NewTestDB() returned nil db")
				}
				if !autoMigrateCalled {
					t.Error("AutoMigrate was not called")
				}
			}
		})
	}

	t.Run("Concurrent Access", func(t *testing.T) {
		const numGoroutines = 10
		var wg sync.WaitGroup
		results := make(chan error, numGoroutines)

		for i := 0; i < numGoroutines; i++ {
			wg.Add(1)
			go func() {
				defer wg.Done()
				_, err := NewTestDB()
				results <- err
			}()
		}

		wg.Wait()
		close(results)

		for err := range results {
			if err != nil {
				t.Errorf("Concurrent NewTestDB() failed: %v", err)
			}
		}
	})

	t.Run("Repeated Calls", func(t *testing.T) {
		for i := 0; i < 3; i++ {
			db, err := NewTestDB()
			if err != nil {
				t.Errorf("Repeated call %d to NewTestDB() failed: %v", i+1, err)
			}
			if db == nil {
				t.Errorf("Repeated call %d to NewTestDB() returned nil db", i+1)
			}
		}
	})
}


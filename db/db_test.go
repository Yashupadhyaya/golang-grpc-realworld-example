package db

import (
	"os"
	"testing"
	"errors"
	"sync"
	"github.com/jinzhu/gorm"
	"github.com/raahii/golang-grpc-realworld-example/model"
	"io/ioutil"
	"database/sql"
	"time"
	_ "github.com/jinzhu/gorm/dialects/mysql"
	"github.com/DATA-DOG/go-txdb"
	"github.com/joho/godotenv"
)




var autoMigrateCalled bool

type mockDB struct {
	gorm.DB
	migrationError error
	migratedModels []interface{}
	mutex          sync.Mutex
}
type mockDB struct {
	gorm.DB
	closeCalled int
	closeError  error
}
type mockDB struct {
	gorm.DB
	createError  error
	createdUsers int
}
type MockDB struct {
	gorm.DB
	OpenError           error
	SetMaxIdleConnsFunc func(n int)
	LogModeFunc         func(enable bool) *gorm.DB
}
type mockDB struct {
	gorm.DB
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

			if got != tt.expected {
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
func (m *mockDB) AutoMigrate(values ...interface{}) *gorm.DB {
	m.mutex.Lock()
	defer m.mutex.Unlock()
	if m.migrationError != nil {
		return &gorm.DB{Error: m.migrationError}
	}
	m.migratedModels = append(m.migratedModels, values...)
	return &gorm.DB{}
}

func TestAutoMigrate(t *testing.T) {
	tests := []struct {
		name          string
		setupMockDB   func() *mockDB
		expectedError error
	}{
		{
			name: "Successful Auto-Migration",
			setupMockDB: func() *mockDB {
				return &mockDB{}
			},
			expectedError: nil,
		},
		{
			name: "Database Connection Error",
			setupMockDB: func() *mockDB {
				return &mockDB{migrationError: errors.New("connection error")}
			},
			expectedError: errors.New("connection error"),
		},
		{
			name: "Partial Migration Failure",
			setupMockDB: func() *mockDB {
				db := &mockDB{}
				db.migrationError = errors.New("partial migration failure")
				return db
			},
			expectedError: errors.New("partial migration failure"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockDB := tt.setupMockDB()
			err := AutoMigrate(mockDB)

			if (err != nil && tt.expectedError == nil) || (err == nil && tt.expectedError != nil) {
				t.Errorf("AutoMigrate() error = %v, expectedError %v", err, tt.expectedError)
				return
			}

			if err != nil && tt.expectedError != nil && err.Error() != tt.expectedError.Error() {
				t.Errorf("AutoMigrate() error = %v, expectedError %v", err, tt.expectedError)
			}

			if err == nil {
				expectedModels := []interface{}{
					&model.User{},
					&model.Article{},
					&model.Tag{},
					&model.Comment{},
				}
				if len(mockDB.migratedModels) != len(expectedModels) {
					t.Errorf("AutoMigrate() migrated %d models, expected %d", len(mockDB.migratedModels), len(expectedModels))
				}

			}
		})
	}
}

func TestConcurrentAutoMigration(t *testing.T) {
	mockDB := &mockDB{}
	concurrency := 10
	var wg sync.WaitGroup
	wg.Add(concurrency)

	for i := 0; i < concurrency; i++ {
		go func() {
			defer wg.Done()
			err := AutoMigrate(mockDB)
			if err != nil {
				t.Errorf("Concurrent AutoMigrate() returned an error: %v", err)
			}
		}()
	}

	wg.Wait()

	expectedMigrationCount := 4 * concurrency
	if len(mockDB.migratedModels) != expectedMigrationCount {
		t.Errorf("Concurrent AutoMigrate() migrated %d models, expected %d", len(mockDB.migratedModels), expectedMigrationCount)
	}
}


/*
ROOST_METHOD_HASH=DropTestDB_4c6b54d5e5
ROOST_METHOD_SIG_HASH=DropTestDB_69b51a825b

FUNCTION_DEF=func DropTestDB(d *gorm.DB) error 

 */
func (m *mockDB) Close() error {
	m.closeCalled++
	return m.closeError
}

func TestDropTestDb(t *testing.T) {
	tests := []struct {
		name    string
		db      *gorm.DB
		wantErr bool
	}{
		{
			name:    "Successfully Close Database Connection",
			db:      &gorm.DB{Value: &mockDB{}},
			wantErr: false,
		},
		{
			name:    "Handle Nil Database Pointer",
			db:      nil,
			wantErr: false,
		},
		{
			name:    "Handle Close Method Error",
			db:      &gorm.DB{Value: &mockDB{closeError: errors.New("close error")}},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := DropTestDB(tt.db)
			if (err != nil) != tt.wantErr {
				t.Errorf("DropTestDB() error = %v, wantErr %v", err, tt.wantErr)
			}

			if tt.db != nil {
				if mock, ok := tt.db.Value.(*mockDB); ok {
					if mock.closeCalled != 1 {
						t.Errorf("Close() called %d times, want 1", mock.closeCalled)
					}
				}
			}
		})
	}
}

func TestDropTestDbConcurrency(t *testing.T) {
	const numGoroutines = 10
	var wg sync.WaitGroup
	wg.Add(numGoroutines)

	for i := 0; i < numGoroutines; i++ {
		go func() {
			defer wg.Done()
			mock := &mockDB{}
			db := &gorm.DB{Value: mock}
			err := DropTestDB(db)
			if err != nil {
				t.Errorf("DropTestDB() error = %v, wantErr false", err)
			}
			if mock.closeCalled != 1 {
				t.Errorf("Close() called %d times, want 1", mock.closeCalled)
			}
		}()
	}

	wg.Wait()
}

func TestDropTestDbIdempotency(t *testing.T) {
	mock := &mockDB{}
	db := &gorm.DB{Value: mock}

	for i := 0; i < 2; i++ {
		err := DropTestDB(db)
		if err != nil {
			t.Errorf("DropTestDB() error = %v, wantErr false", err)
		}
	}

	if mock.closeCalled != 1 {
		t.Errorf("Close() called %d times, want 1", mock.closeCalled)
	}
}

func TestDropTestDbNoFurtherOperations(t *testing.T) {
	mock := &mockDB{}
	db := &gorm.DB{Value: mock}

	err := DropTestDB(db)
	if err != nil {
		t.Errorf("DropTestDB() error = %v, wantErr false", err)
	}

	if mock.closeCalled != 1 {
		t.Errorf("Database operation succeeded after DropTestDB, expected it to fail")
	}
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
		setupFile     func() error
		expectedError error
		expectedUsers int
	}{
		{
			name: "Successful Seeding of Users",
			setupMock: func() *mockDB {
				return &mockDB{}
			},
			setupFile: func() error {
				return ioutil.WriteFile("db/seed/users.toml", []byte(`
					[[Users]]
					username = "user1"
					email = "user1@example.com"
					password = "password1"

					[[Users]]
					username = "user2"
					email = "user2@example.com"
					password = "password2"
				`), 0644)
			},
			expectedError: nil,
			expectedUsers: 2,
		},
		{
			name: "File Not Found Error",
			setupMock: func() *mockDB {
				return &mockDB{}
			},
			setupFile: func() error {
				return os.Remove("db/seed/users.toml")
			},
			expectedError: errors.New("open db/seed/users.toml: no such file or directory"),
			expectedUsers: 0,
		},
		{
			name: "Invalid TOML Format",
			setupMock: func() *mockDB {
				return &mockDB{}
			},
			setupFile: func() error {
				return ioutil.WriteFile("db/seed/users.toml", []byte(`
					invalid toml
				`), 0644)
			},
			expectedError: errors.New("toml: line 2: expected key separator '=', but got 't' instead"),
			expectedUsers: 0,
		},
		{
			name: "Database Insertion Error",
			setupMock: func() *mockDB {
				return &mockDB{createError: errors.New("database error")}
			},
			setupFile: func() error {
				return ioutil.WriteFile("db/seed/users.toml", []byte(`
					[[Users]]
					username = "user1"
					email = "user1@example.com"
					password = "password1"
				`), 0644)
			},
			expectedError: errors.New("database error"),
			expectedUsers: 0,
		},
		{
			name: "Empty Users File",
			setupMock: func() *mockDB {
				return &mockDB{}
			},
			setupFile: func() error {
				return ioutil.WriteFile("db/seed/users.toml", []byte(``), 0644)
			},
			expectedError: nil,
			expectedUsers: 0,
		},
		{
			name: "Large Number of Users",
			setupMock: func() *mockDB {
				return &mockDB{}
			},
			setupFile: func() error {
				var content string
				for i := 0; i < 10000; i++ {
					content += fmt.Sprintf(`
						[[Users]]
						username = "user%d"
						email = "user%d@example.com"
						password = "password%d"
					`, i, i, i)
				}
				return ioutil.WriteFile("db/seed/users.toml", []byte(content), 0644)
			},
			expectedError: nil,
			expectedUsers: 10000,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			mockDB := tt.setupMock()
			err := tt.setupFile()
			if err != nil {
				t.Fatalf("Failed to setup file: %v", err)
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
func (m *MockDB) DB() *sql.DB {
	return &sql.DB{}
}

func (m *MockDB) LogMode(enable bool) *gorm.DB {
	if m.LogModeFunc != nil {
		return m.LogModeFunc(enable)
	}
	return &m.DB
}

func (m *MockDB) Open(dialect string, args ...interface{}) (db *gorm.DB, err error) {
	if m.OpenError != nil {
		return nil, m.OpenError
	}
	return &m.DB, nil
}

func (m *MockDB) SetMaxIdleConns(n int) {
	if m.SetMaxIdleConnsFunc != nil {
		m.SetMaxIdleConnsFunc(n)
	}
}

func TestNew(t *testing.T) {
	originalGormOpen := gorm.Open
	defer func() { gorm.Open = originalGormOpen }()

	tests := []struct {
		name            string
		setupMock       func() *MockDB
		setupEnv        func()
		expectedDB      bool
		expectedError   bool
		maxIdleConns    int
		logMode         bool
		concurrentCalls int
	}{
		{
			name: "Successful Database Connection",
			setupMock: func() *MockDB {
				return &MockDB{}
			},
			setupEnv: func() {
				os.Setenv("DB_HOST", "localhost")
				os.Setenv("DB_PORT", "3306")
				os.Setenv("DB_USER", "user")
				os.Setenv("DB_PASSWORD", "password")
				os.Setenv("DB_NAME", "testdb")
			},
			expectedDB:    true,
			expectedError: false,
			maxIdleConns:  3,
			logMode:       false,
		},
		{
			name: "Database Connection Retry",
			setupMock: func() *MockDB {
				mock := &MockDB{}
				callCount := 0
				gorm.Open = func(dialect string, args ...interface{}) (db *gorm.DB, err error) {
					callCount++
					if callCount < 3 {
						return nil, errors.New("connection failed")
					}
					return &mock.DB, nil
				}
				return mock
			},
			setupEnv: func() {
				os.Setenv("DB_HOST", "localhost")
				os.Setenv("DB_PORT", "3306")
				os.Setenv("DB_USER", "user")
				os.Setenv("DB_PASSWORD", "password")
				os.Setenv("DB_NAME", "testdb")
			},
			expectedDB:    true,
			expectedError: false,
		},
		{
			name: "Connection Failure After Max Retries",
			setupMock: func() *MockDB {
				mock := &MockDB{}
				gorm.Open = func(dialect string, args ...interface{}) (db *gorm.DB, err error) {
					return nil, errors.New("connection failed")
				}
				return mock
			},
			setupEnv: func() {
				os.Setenv("DB_HOST", "localhost")
				os.Setenv("DB_PORT", "3306")
				os.Setenv("DB_USER", "user")
				os.Setenv("DB_PASSWORD", "password")
				os.Setenv("DB_NAME", "testdb")
			},
			expectedDB:    false,
			expectedError: true,
		},
		{
			name: "Invalid DSN Configuration",
			setupMock: func() *MockDB {
				return &MockDB{}
			},
			setupEnv: func() {
				os.Unsetenv("DB_HOST")
				os.Unsetenv("DB_PORT")
				os.Unsetenv("DB_USER")
				os.Unsetenv("DB_PASSWORD")
				os.Unsetenv("DB_NAME")
			},
			expectedDB:    false,
			expectedError: true,
		},
		{
			name: "Correct Database Connection Settings",
			setupMock: func() *MockDB {
				mock := &MockDB{}
				mock.SetMaxIdleConnsFunc = func(n int) {
					if n != 3 {
						t.Errorf("Expected MaxIdleConns to be 3, got %d", n)
					}
				}
				mock.LogModeFunc = func(enable bool) *gorm.DB {
					if enable {
						t.Error("Expected LogMode to be false")
					}
					return &mock.DB
				}
				return mock
			},
			setupEnv: func() {
				os.Setenv("DB_HOST", "localhost")
				os.Setenv("DB_PORT", "3306")
				os.Setenv("DB_USER", "user")
				os.Setenv("DB_PASSWORD", "password")
				os.Setenv("DB_NAME", "testdb")
			},
			expectedDB:    true,
			expectedError: false,
		},
		{
			name: "Concurrent Access Safety",
			setupMock: func() *MockDB {
				return &MockDB{}
			},
			setupEnv: func() {
				os.Setenv("DB_HOST", "localhost")
				os.Setenv("DB_PORT", "3306")
				os.Setenv("DB_USER", "user")
				os.Setenv("DB_PASSWORD", "password")
				os.Setenv("DB_NAME", "testdb")
			},
			expectedDB:      true,
			expectedError:   false,
			concurrentCalls: 10,
		},
		{
			name: "Environment Variable Dependency",
			setupMock: func() *MockDB {
				return &MockDB{}
			},
			setupEnv: func() {
				os.Unsetenv("DB_HOST")
			},
			expectedDB:    false,
			expectedError: true,
		},
		{
			name: "Database Driver Compatibility",
			setupMock: func() *MockDB {
				mock := &MockDB{}
				gorm.Open = func(dialect string, args ...interface{}) (db *gorm.DB, err error) {
					if dialect != "_" {
						return nil, errors.New("invalid dialect")
					}
					return &mock.DB, nil
				}
				return mock
			},
			setupEnv: func() {
				os.Setenv("DB_HOST", "localhost")
				os.Setenv("DB_PORT", "3306")
				os.Setenv("DB_USER", "user")
				os.Setenv("DB_PASSWORD", "password")
				os.Setenv("DB_NAME", "testdb")
			},
			expectedDB:    true,
			expectedError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mock := tt.setupMock()
			gorm.Open = mock.Open
			tt.setupEnv()

			if tt.concurrentCalls > 0 {
				var wg sync.WaitGroup
				for i := 0; i < tt.concurrentCalls; i++ {
					wg.Add(1)
					go func() {
						defer wg.Done()
						db, err := New()
						if (db == nil) == tt.expectedDB {
							t.Errorf("New() returned unexpected db: got %v, want %v", db, tt.expectedDB)
						}
						if (err != nil) != tt.expectedError {
							t.Errorf("New() returned unexpected error: got %v, want %v", err, tt.expectedError)
						}
					}()
				}
				wg.Wait()
			} else {
				db, err := New()
				if (db == nil) == tt.expectedDB {
					t.Errorf("New() returned unexpected db: got %v, want %v", db, tt.expectedDB)
				}
				if (err != nil) != tt.expectedError {
					t.Errorf("New() returned unexpected error: got %v, want %v", err, tt.expectedError)
				}
			}
		})
	}
}


/*
ROOST_METHOD_HASH=NewTestDB_7feb2c4a7a
ROOST_METHOD_SIG_HASH=NewTestDB_1b71546d9d

FUNCTION_DEF=func NewTestDB() (*gorm.DB, error) 

 */
func (m *mockDB) DB() *sql.DB {
	return &sql.DB{}
}

func (m *mockDB) LogMode(enable bool) *gorm.DB {
	return &m.DB
}

func TestNewTestDb(t *testing.T) {
	originalGodotenvLoad := godotenv.Load
	originalGormOpen := gorm.Open
	originalSqlOpen := sql.Open
	originalTxdbRegister := txdb.Register
	originalAutoMigrate := AutoMigrate

	defer func() {
		godotenv.Load = originalGodotenvLoad
		gorm.Open = originalGormOpen
		sql.Open = originalSqlOpen
		txdb.Register = originalTxdbRegister
		AutoMigrate = originalAutoMigrate
	}()

	tests := []struct {
		name           string
		setupMock      func()
		expectedDB     bool
		expectedError  bool
		validateResult func(*testing.T, *gorm.DB, error)
	}{
		{
			name: "Successful Database Connection and Initialization",
			setupMock: func() {
				godotenv.Load = func(filenames ...string) error { return nil }
				gorm.Open = func(dialect string, args ...interface{}) (*gorm.DB, error) {
					return &gorm.DB{}, nil
				}
				sql.Open = func(driverName, dataSourceName string) (*sql.DB, error) {
					return &sql.DB{}, nil
				}
				txdb.Register = func(name, driver, dsn string) {}
				AutoMigrate = mockAutoMigrate
			},
			expectedDB:    true,
			expectedError: false,
			validateResult: func(t *testing.T, db *gorm.DB, err error) {
				if !autoMigrateCalled {
					t.Error("AutoMigrate was not called")
				}
			},
		},
		{
			name: "Environment File Not Found",
			setupMock: func() {
				godotenv.Load = func(filenames ...string) error {
					return errors.New("env file not found")
				}
			},
			expectedDB:    false,
			expectedError: true,
		},
		{
			name: "Invalid Database Credentials",
			setupMock: func() {
				godotenv.Load = func(filenames ...string) error { return nil }
				gorm.Open = func(dialect string, args ...interface{}) (*gorm.DB, error) {
					return nil, errors.New("invalid credentials")
				}
			},
			expectedDB:    false,
			expectedError: true,
		},
		{
			name: "Database Connection Limit",
			setupMock: func() {
				godotenv.Load = func(filenames ...string) error { return nil }
				gorm.Open = func(dialect string, args ...interface{}) (*gorm.DB, error) {
					return &mockDB{}, nil
				}
				sql.Open = func(driverName, dataSourceName string) (*sql.DB, error) {
					return &sql.DB{}, nil
				}
				txdb.Register = func(name, driver, dsn string) {}
			},
			expectedDB:    true,
			expectedError: false,
			validateResult: func(t *testing.T, db *gorm.DB, err error) {

			},
		},
		{
			name: "LogMode Setting",
			setupMock: func() {
				godotenv.Load = func(filenames ...string) error { return nil }
				gorm.Open = func(dialect string, args ...interface{}) (*gorm.DB, error) {
					return &mockDB{}, nil
				}
				sql.Open = func(driverName, dataSourceName string) (*sql.DB, error) {
					return &sql.DB{}, nil
				}
				txdb.Register = func(name, driver, dsn string) {}
			},
			expectedDB:    true,
			expectedError: false,
			validateResult: func(t *testing.T, db *gorm.DB, err error) {

			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			txdbInitialized = false
			tt.setupMock()

			db, err := NewTestDB()

			if (db != nil) != tt.expectedDB {
				t.Errorf("NewTestDB() returned unexpected db status, got: %v, want: %v", db != nil, tt.expectedDB)
			}

			if (err != nil) != tt.expectedError {
				t.Errorf("NewTestDB() returned unexpected error status, got: %v, want: %v", err != nil, tt.expectedError)
			}

			if tt.validateResult != nil {
				tt.validateResult(t, db, err)
			}
		})
	}
}

func TestNewTestDbConcurrency(t *testing.T) {
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

	godotenv.Load = func(filenames ...string) error { return nil }
	gorm.Open = func(dialect string, args ...interface{}) (*gorm.DB, error) {
		return &gorm.DB{}, nil
	}
	sql.Open = func(driverName, dataSourceName string) (*sql.DB, error) {
		return &sql.DB{}, nil
	}

	var registerCount int
	var registerMutex sync.Mutex
	txdb.Register = func(name, driver, dsn string) {
		registerMutex.Lock()
		registerCount++
		registerMutex.Unlock()
	}

	const goroutines = 10
	var wg sync.WaitGroup
	wg.Add(goroutines)

	for i := 0; i < goroutines; i++ {
		go func() {
			defer wg.Done()
			_, err := NewTestDB()
			if err != nil {
				t.Errorf("NewTestDB() returned an error: %v", err)
			}
		}()
	}

	wg.Wait()

	if registerCount != 1 {
		t.Errorf("txdb.Register was called %d times, expected 1", registerCount)
	}
}

func mockAutoMigrate(db *gorm.DB) {
	autoMigrateCalled = true
}


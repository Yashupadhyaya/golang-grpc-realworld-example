package db

import (
	"errors"
	"io/ioutil"
	"os"
	"testing"

	"github.com/BurntSushi/toml"
	"github.com/jinzhu/gorm"
	"github.com/raahii/golang-grpc-realworld-example/model"
)

type mockDB struct {
	*gorm.DB
	createError error
}

func (m *mockDB) Create(value interface{}) *gorm.DB {
	return &gorm.DB{Error: m.createError}
}

func TestSeed(t *testing.T) {
	tests := []struct {
		name           string
		setupMock      func() *mockDB
		setupFile      func() error
		expectedError  error
		expectedUsers  int
		cleanupFile    func()
	}{
		{
			name: "Successful Seeding of Users",
			setupMock: func() *mockDB {
				return &mockDB{}
			},
			setupFile: func() error {
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
			cleanupFile: func() {
				os.Remove("db/seed/users.toml")
			},
		},
		{
			name: "File Not Found Error",
			setupMock: func() *mockDB {
				return &mockDB{}
			},
			setupFile: func() error {
				return nil // Don't create the file
			},
			expectedError: errors.New("open db/seed/users.toml: no such file or directory"),
			expectedUsers: 0,
			cleanupFile:   func() {},
		},
		{
			name: "Invalid TOML Format",
			setupMock: func() *mockDB {
				return &mockDB{}
			},
			setupFile: func() error {
				content := `
				[[Users]
				username = "user1"
				email = "user1@example.com"
				password = "password1"
				`
				return ioutil.WriteFile("db/seed/users.toml", []byte(content), 0644)
			},
			expectedError: errors.New("toml: line 2: expected '=' after key name; found '[' instead"),
			expectedUsers: 0,
			cleanupFile: func() {
				os.Remove("db/seed/users.toml")
			},
		},
		{
			name: "Database Insertion Error",
			setupMock: func() *mockDB {
				return &mockDB{createError: errors.New("database insertion error")}
			},
			setupFile: func() error {
				content := `
				[[Users]]
				username = "user1"
				email = "user1@example.com"
				password = "password1"
				`
				return ioutil.WriteFile("db/seed/users.toml", []byte(content), 0644)
			},
			expectedError: errors.New("database insertion error"),
			expectedUsers: 0,
			cleanupFile: func() {
				os.Remove("db/seed/users.toml")
			},
		},
		{
			name: "Empty TOML File",
			setupMock: func() *mockDB {
				return &mockDB{}
			},
			setupFile: func() error {
				return ioutil.WriteFile("db/seed/users.toml", []byte(""), 0644)
			},
			expectedError: nil,
			expectedUsers: 0,
			cleanupFile: func() {
				os.Remove("db/seed/users.toml")
			},
		},
		{
			name: "Large Dataset Handling",
			setupMock: func() *mockDB {
				return &mockDB{}
			},
			setupFile: func() error {
				content := ""
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
			cleanupFile: func() {
				os.Remove("db/seed/users.toml")
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockDB := tt.setupMock()
			err := tt.setupFile()
			if err != nil {
				t.Fatalf("Failed to setup file: %v", err)
			}
			defer tt.cleanupFile()

			err = Seed(mockDB)

			if (err != nil && tt.expectedError == nil) || (err == nil && tt.expectedError != nil) || (err != nil && tt.expectedError != nil && err.Error() != tt.expectedError.Error()) {
				t.Errorf("Seed() error = %v, expectedError %v", err, tt.expectedError)
			}

			// TODO: Add assertions to check the number of users inserted
			// This would require modifying the mockDB to track insertions
		})
	}
}

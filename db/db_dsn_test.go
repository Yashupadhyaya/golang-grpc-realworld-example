package db

import (
	"os"
	"testing"
)

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
			name: "Empty Values in Environment Variables",
			envVars: map[string]string{
				"DB_HOST":     "",
				"DB_USER":     "user",
				"DB_PASSWORD": "password",
				"DB_NAME":     "testdb",
				"DB_PORT":     "3306",
			},
			wantErr: true,
			errMsg:  "$DB_HOST is not set",
		},
		{
			name: "Correct DSN Format",
			envVars: map[string]string{
				"DB_HOST":     "testhost",
				"DB_USER":     "testuser",
				"DB_PASSWORD": "testpass",
				"DB_NAME":     "testname",
				"DB_PORT":     "5432",
			},
			expected: "testuser:testpass@(testhost:5432)/testname?charset=utf8mb4&parseTime=True&loc=Local",
			wantErr:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Clear environment variables
			os.Clearenv()

			// Set environment variables for the test
			for key, value := range tt.envVars {
				os.Setenv(key, value)
			}

			// Call the function
			got, err := dsn()

			// Check for errors
			if (err != nil) != tt.wantErr {
				t.Errorf("dsn() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			// Check error message if expecting an error
			if tt.wantErr && err.Error() != tt.errMsg {
				t.Errorf("dsn() error message = %v, want %v", err.Error(), tt.errMsg)
				return
			}

			// Check the returned DSN string
			if got != tt.expected {
				t.Errorf("dsn() = %v, want %v", got, tt.expected)
			}
		})
	}
}

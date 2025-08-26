package auth

import (
	debug "runtime/debug"
	sync "sync"
	testing "testing"
	time "time"
)

func TestGenerateTokenWithTime(t *testing.T) {
	tests := []struct {
		name      string
		userID    uint
		timeInput time.Time
		expectErr bool
		expected  func(token string, err error) bool
	}{
		{
			name:      "Valid ID and Time Input",
			userID:    123,
			timeInput: time.Now(),
			expectErr: false,
			expected: func(token string, err error) bool {
				return err == nil && len(token) > 0
			},
		},
		{
			name:      "Invalid ID (zero value)",
			userID:    0,
			timeInput: time.Now(),
			expectErr: true,
			expected: func(token string, err error) bool {
				return err != nil || len(token) == 0
			},
		},
		{
			name:      "Invalid Time Input",
			userID:    123,
			timeInput: time.Time{},
			expectErr: true,
			expected: func(token string, err error) bool {
				return err != nil || len(token) == 0
			},
		},
		{
			name:      "Handling Large IDs",
			userID:    uint(^uint(0)),
			timeInput: time.Now(),
			expectErr: false,
			expected: func(token string, err error) bool {
				return err == nil && len(token) > 0
			},
		},
		{
			name:      "Near Midnight Time Input",
			userID:    123,
			timeInput: time.Date(2023, 10, 15, 23, 59, 59, 0, time.UTC),
			expectErr: false,
			expected: func(token string, err error) bool {
				return err == nil && len(token) > 0
			},
		},
		{
			name:      "Negative Scenario: Past Date",
			userID:    123,
			timeInput: time.Date(2010, 1, 1, 9, 0, 0, 0, time.UTC),
			expectErr: true,
			expected: func(token string, err error) bool {
				return err != nil || len(token) == 0
			},
		},
		{
			name:      "Negative Scenario: ID Overflow",
			userID:    uint(0xFFFFFFFFFFFFFFFF),
			timeInput: time.Now(),
			expectErr: true,
			expected: func(token string, err error) bool {
				return err != nil || len(token) == 0
			},
		},
	}

	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			defer func() {
				if r := recover(); r != nil {
					t.Logf("Panic encountered, failing test: %v\n%s", r, string(debug.Stack()))
					t.Fail()
				}
			}()

			token, err := GenerateTokenWithTime(testCase.userID, testCase.timeInput)

			if testCase.expectErr && err == nil {
				t.Fatalf("expected error but got none")
			}
			if !testCase.expectErr && err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if !testCase.expected(token, err) {
				t.Fatalf("test failed for %s. Unexpected result with token: %s, error: %v", testCase.name, token, err)
			}
			t.Logf("Test %s succeeded with token: %s", testCase.name, token)
		})
	}

	t.Run("Concurrent Token Generation", func(t *testing.T) {
		defer func() {
			if r := recover(); r != nil {
				t.Logf("Panic encountered, failing test: %v\n%s", r, string(debug.Stack()))
				t.Fail()
			}
		}()

		var wg sync.WaitGroup
		numThreads := 10
		successCount := 0
		failed := false
		tokenChannel := make(chan string, numThreads)

		for i := 0; i < numThreads; i++ {
			wg.Add(1)
			go func(threadID int) {
				defer wg.Done()
				token, err := GenerateTokenWithTime(uint(threadID), time.Now())
				if err != nil {
					t.Errorf("unexpected error in goroutine %d: %v", threadID, err)
					failed = true
				} else {
					tokenChannel <- token
				}
			}(i)
		}

		wg.Wait()
		close(tokenChannel)

		for token := range tokenChannel {
			if len(token) > 0 {
				successCount++
			}
		}

		if failed {
			t.Fatalf("Concurrent execution failed due to unexpected errors.")
		}

		if successCount != numThreads {
			t.Fatalf("Expected %d successful token generations, got %d", numThreads, successCount)
		}

		t.Logf("Concurrent Token Generation succeeded with %d tokens", successCount)
	})
}

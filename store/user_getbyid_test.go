package store

import (
	"errors"
	"testing"
	"time"
	"github.com/jinzhu/gorm"
	"github.com/raahii/golang-grpc-realworld-example/model"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

type Callback struct {
	logger     logger
	creates    []*func(scope *Scope)
	updates    []*func(scope *Scope)
	deletes    []*func(scope *Scope)
	queries    []*func(scope *Scope)
	rowQueries []*func(scope *Scope)
	processors []*CallbackProcessor
}


type Scope struct {
	Search          *search
	Value           interface{}
	SQL             string
	SQLVars         []interface{}
	db              *DB
	instanceID      string
	primaryKeyField *Field
	skipLeft        bool
	fields          *[]*Field
	selectAttrs     *[]string
}



type Call struct {
	Parent *Mock

	// The name of the method that was or will be called.
	Method string

	// Holds the arguments of the method.
	Arguments Arguments

	// Holds the arguments that should be returned when
	// this method is called.
	ReturnArguments Arguments

	// Holds the caller info for the On() call
	callerInfo []string

	// The number of times to return the return arguments when setting
	// expectations. 0 means to always return the value.
	Repeatability int

	// Amount of times this call has been called
	totalCalls int

	// Call to this method can be optional
	optional bool

	// Holds a channel that will be used to block the Return until it either
	// receives a message or is closed. nil means it returns immediately.
	WaitFor <-chan time.Time

	waitTime time.Duration

	// Holds a handler used to manipulate arguments content that are passed by
	// reference. It's useful when mocking methods such as unmarshalers or
	// decoders.
	RunFn func(Arguments)
}

type T struct {
	common
	isEnvSet bool
	context  *testContext // For running tests and subtests.
}




type MockDB struct {
	mock.Mock
}




type Callback struct {
	logger     logger
	creates    []*func(scope *Scope)
	updates    []*func(scope *Scope)
	deletes    []*func(scope *Scope)
	queries    []*func(scope *Scope)
	rowQueries []*func(scope *Scope)
	processors []*CallbackProcessor
}


type Logger struct {
	LogWriter
}


type Scope struct {
	Search          *search
	Value           interface{}
	SQL             string
	SQLVars         []interface{}
	db              *DB
	instanceID      string
	primaryKeyField *Field
	skipLeft        bool
	fields          *[]*Field
	selectAttrs     *[]string
}




type T struct {
	common
	isEnvSet bool
	context  *testContext // For running tests and subtests.
}
func (m *MockDB) AddError(err error) error {
	args := m.Called(err)
	return args.Error(0)
}
func (m *MockDB) Callback() *gorm.Callback {
	args := m.Called()
	return args.Get(0).(*gorm.Callback)
}
func (m *MockDB) CommonDB() gorm.SQLCommon {
	args := m.Called()
	return args.Get(0).(gorm.SQLCommon)
}
func (m *MockDB) DB() *gorm.DB {
	args := m.Called()
	return args.Get(0).(*gorm.DB)
}
func (m *MockDB) Find(out interface{}, where ...interface{}) *gorm.DB {
	args := m.Called(out, where)
	return args.Get(0).(*gorm.DB)
}
func (m *MockDB) LogMode(enable bool) *gorm.DB {
	args := m.Called(enable)
	return args.Get(0).(*gorm.DB)
}
func (m *MockDB) New() *gorm.DB {
	args := m.Called()
	return args.Get(0).(*gorm.DB)
}
func (m *MockDB) NewScope(value interface{}) *gorm.Scope {
	args := m.Called(value)
	return args.Get(0).(*gorm.Scope)
}
func (m *MockDB) SetLogger(log gorm.Logger) {
	m.Called(log)
}
func (m *MockDB) SingularTable(enable bool) {
	m.Called(enable)
}
func TestUserStoreGetByID(t *testing.T) {
	tests := []struct {
		name      string
		id        uint
		mockSetup func(*MockDB)
		want      *model.User
		wantErr   error
	}{
		{
			name: "Successfully retrieve a user by ID",
			id:   1,
			mockSetup: func(m *MockDB) {
				m.On("Find", mock.AnythingOfType("*model.User"), uint(1)).Return(&gorm.DB{Error: nil}).Run(func(args mock.Arguments) {
					arg := args.Get(0).(*model.User)
					*arg = model.User{
						Model:    gorm.Model{ID: 1, CreatedAt: time.Now(), UpdatedAt: time.Now()},
						Username: "testuser",
						Email:    "test@example.com",
						Password: "password",
						Bio:      "Test bio",
						Image:    "test.jpg",
					}
				})
			},
			want: &model.User{
				Model:    gorm.Model{ID: 1},
				Username: "testuser",
				Email:    "test@example.com",
				Password: "password",
				Bio:      "Test bio",
				Image:    "test.jpg",
			},
			wantErr: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockDB := new(MockDB)
			tt.mockSetup(mockDB)

			s := &UserStore{
				db: mockDB,
			}

			got, err := s.GetByID(tt.id)

			if tt.wantErr != nil {
				assert.Error(t, err)
				assert.Equal(t, tt.wantErr.Error(), err.Error())
			} else {
				assert.NoError(t, err)
			}

			if tt.want != nil {
				assert.NotNil(t, got)
				assert.Equal(t, tt.want.ID, got.ID)
				assert.Equal(t, tt.want.Username, got.Username)
				assert.Equal(t, tt.want.Email, got.Email)
				assert.Equal(t, tt.want.Password, got.Password)
				assert.Equal(t, tt.want.Bio, got.Bio)
				assert.Equal(t, tt.want.Image, got.Image)
				assert.Equal(t, len(tt.want.Follows), len(got.Follows))
				assert.Equal(t, len(tt.want.FavoriteArticles), len(got.FavoriteArticles))
			} else {
				assert.Nil(t, got)
			}

			mockDB.AssertExpectations(t)
		})
	}
}
func (m *MockDB) Where(query interface{}, args ...interface{}) *gorm.DB {
	mockArgs := m.Called(query, args)
	return mockArgs.Get(0).(*gorm.DB)
}

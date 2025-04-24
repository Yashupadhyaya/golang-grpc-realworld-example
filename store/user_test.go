package store

import (
	"errors"
	"math"
	"testing"
	"github.com/jinzhu/gorm"
	"github.com/raahii/golang-grpc-realworld-example/model"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)





type MockDB struct {
	mock.Mock
}


/*
ROOST_METHOD_HASH=GetByID_bbf946112e
ROOST_METHOD_SIG_HASH=GetByID_728dd55ed1


 */
func (m *MockDB) Find(out interface{}, where ...interface{}) *gorm.DB {
	args := m.Called(out, where)
	return args.Get(0).(*gorm.DB)
}

func TestUserStoreGetByID(t *testing.T) {
	tests := []struct {
		name    string
		id      uint
		mockDB  func() *MockDB
		want    *model.User
		wantErr bool
	}{
		{
			name: "Successfully retrieve a user by ID",
			id:   1,
			mockDB: func() *MockDB {
				db := new(MockDB)
				db.On("Find", mock.AnythingOfType("*model.User"), uint(1)).Return(&gorm.DB{Error: nil}).Run(func(args mock.Arguments) {
					arg := args.Get(0).(*model.User)
					*arg = model.User{
						Model:    gorm.Model{ID: 1},
						Username: "testuser",
						Email:    "test@example.com",
						Password: "password",
					}
				})
				return db
			},
			want: &model.User{
				Model:    gorm.Model{ID: 1},
				Username: "testuser",
				Email:    "test@example.com",
				Password: "password",
			},
			wantErr: false,
		},
		{
			name: "Attempt to retrieve a non-existent user",
			id:   999,
			mockDB: func() *MockDB {
				db := new(MockDB)
				db.On("Find", mock.AnythingOfType("*model.User"), uint(999)).Return(&gorm.DB{Error: gorm.ErrRecordNotFound})
				return db
			},
			want:    nil,
			wantErr: true,
		},
		{
			name: "Handle database connection error",
			id:   1,
			mockDB: func() *MockDB {
				db := new(MockDB)
				db.On("Find", mock.AnythingOfType("*model.User"), uint(1)).Return(&gorm.DB{Error: errors.New("database connection error")})
				return db
			},
			want:    nil,
			wantErr: true,
		},
		{
			name: "Retrieve a user with minimum fields populated",
			id:   2,
			mockDB: func() *MockDB {
				db := new(MockDB)
				db.On("Find", mock.AnythingOfType("*model.User"), uint(2)).Return(&gorm.DB{Error: nil}).Run(func(args mock.Arguments) {
					arg := args.Get(0).(*model.User)
					*arg = model.User{
						Model:    gorm.Model{ID: 2},
						Username: "minuser",
						Email:    "min@example.com",
						Password: "minpass",
					}
				})
				return db
			},
			want: &model.User{
				Model:    gorm.Model{ID: 2},
				Username: "minuser",
				Email:    "min@example.com",
				Password: "minpass",
			},
			wantErr: false,
		},
		{
			name: "Retrieve a user with all fields populated",
			id:   3,
			mockDB: func() *MockDB {
				db := new(MockDB)
				db.On("Find", mock.AnythingOfType("*model.User"), uint(3)).Return(&gorm.DB{Error: nil}).Run(func(args mock.Arguments) {
					arg := args.Get(0).(*model.User)
					*arg = model.User{
						Model:    gorm.Model{ID: 3},
						Username: "fulluser",
						Email:    "full@example.com",
						Password: "fullpass",
						Bio:      "Full bio",
						Image:    "full.jpg",
						Follows:  []model.User{{Model: gorm.Model{ID: 1}}},
						FavoriteArticles: []model.Article{{
							Model: gorm.Model{ID: 1},
							Title: "Test Article",
						}},
					}
				})
				return db
			},
			want: &model.User{
				Model:    gorm.Model{ID: 3},
				Username: "fulluser",
				Email:    "full@example.com",
				Password: "fullpass",
				Bio:      "Full bio",
				Image:    "full.jpg",
				Follows:  []model.User{{Model: gorm.Model{ID: 1}}},
				FavoriteArticles: []model.Article{{
					Model: gorm.Model{ID: 1},
					Title: "Test Article",
				}},
			},
			wantErr: false,
		},
		{
			name: "Handle very large user ID",
			id:   math.MaxUint32,
			mockDB: func() *MockDB {
				db := new(MockDB)
				db.On("Find", mock.AnythingOfType("*model.User"), uint(math.MaxUint32)).Return(&gorm.DB{Error: nil}).Run(func(args mock.Arguments) {
					arg := args.Get(0).(*model.User)
					*arg = model.User{
						Model:    gorm.Model{ID: math.MaxUint32},
						Username: "largeuser",
						Email:    "large@example.com",
						Password: "largepass",
					}
				})
				return db
			},
			want: &model.User{
				Model:    gorm.Model{ID: math.MaxUint32},
				Username: "largeuser",
				Email:    "large@example.com",
				Password: "largepass",
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockDB := tt.mockDB()
			s := &UserStore{
				db: mockDB,
			}

			got, err := s.GetByID(tt.id)

			if tt.wantErr {
				assert.Error(t, err)
				assert.Nil(t, got)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.want, got)
			}

			mockDB.AssertExpectations(t)
		})
	}
}


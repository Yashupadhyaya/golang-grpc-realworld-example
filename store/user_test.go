package store

import (
	"errors"
	"testing"
	"github.com/jinzhu/gorm"
	"github.com/raahii/golang-grpc-realworld-example/model"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)





type mockDB struct {
	mock.Mock
}


/*
ROOST_METHOD_HASH=GetByID_bbf946112e
ROOST_METHOD_SIG_HASH=GetByID_728dd55ed1


 */
func (m *mockDB) Find(out interface{}, where ...interface{}) *gorm.DB {
	args := m.Called(out, where)
	return args.Get(0).(*gorm.DB)
}

func TestUserStoreGetByID(t *testing.T) {
	tests := []struct {
		name      string
		id        uint
		mockSetup func(*mockDB)
		want      *model.User
		wantErr   bool
	}{
		{
			name: "Successfully retrieve a user by ID",
			id:   1,
			mockSetup: func(m *mockDB) {
				m.On("Find", mock.AnythingOfType("*model.User"), uint(1)).Return(&gorm.DB{Error: nil}).Run(func(args mock.Arguments) {
					arg := args.Get(0).(*model.User)
					*arg = model.User{
						Model:    gorm.Model{ID: 1},
						Username: "testuser",
						Email:    "test@example.com",
						Password: "password",
					}
				})
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
			mockSetup: func(m *mockDB) {
				m.On("Find", mock.AnythingOfType("*model.User"), uint(999)).Return(&gorm.DB{Error: gorm.ErrRecordNotFound})
			},
			want:    nil,
			wantErr: true,
		},
		{
			name: "Handle database connection error",
			id:   2,
			mockSetup: func(m *mockDB) {
				m.On("Find", mock.AnythingOfType("*model.User"), uint(2)).Return(&gorm.DB{Error: errors.New("database connection error")})
			},
			want:    nil,
			wantErr: true,
		},
		{
			name: "Retrieve a user with minimum data",
			id:   3,
			mockSetup: func(m *mockDB) {
				m.On("Find", mock.AnythingOfType("*model.User"), uint(3)).Return(&gorm.DB{Error: nil}).Run(func(args mock.Arguments) {
					arg := args.Get(0).(*model.User)
					*arg = model.User{
						Model:    gorm.Model{ID: 3},
						Username: "minuser",
						Email:    "min@example.com",
						Password: "minpass",
					}
				})
			},
			want: &model.User{
				Model:    gorm.Model{ID: 3},
				Username: "minuser",
				Email:    "min@example.com",
				Password: "minpass",
			},
			wantErr: false,
		},
		{
			name: "Retrieve a user with all fields populated",
			id:   4,
			mockSetup: func(m *mockDB) {
				m.On("Find", mock.AnythingOfType("*model.User"), uint(4)).Return(&gorm.DB{Error: nil}).Run(func(args mock.Arguments) {
					arg := args.Get(0).(*model.User)
					*arg = model.User{
						Model:            gorm.Model{ID: 4},
						Username:         "fulluser",
						Email:            "full@example.com",
						Password:         "fullpass",
						Bio:              "Full bio",
						Image:            "full.jpg",
						Follows:          []model.User{{Model: gorm.Model{ID: 5}}},
						FavoriteArticles: []model.Article{{Model: gorm.Model{ID: 1}}},
					}
				})
			},
			want: &model.User{
				Model:            gorm.Model{ID: 4},
				Username:         "fulluser",
				Email:            "full@example.com",
				Password:         "fullpass",
				Bio:              "Full bio",
				Image:            "full.jpg",
				Follows:          []model.User{{Model: gorm.Model{ID: 5}}},
				FavoriteArticles: []model.Article{{Model: gorm.Model{ID: 1}}},
			},
			wantErr: false,
		},
		{
			name: "Handle zero ID input",
			id:   0,
			mockSetup: func(m *mockDB) {
				m.On("Find", mock.AnythingOfType("*model.User"), uint(0)).Return(&gorm.DB{Error: gorm.ErrRecordNotFound})
			},
			want:    nil,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockDB := new(mockDB)
			tt.mockSetup(mockDB)

			s := &UserStore{
				db: mockDB,
			}

			got, err := s.GetByID(tt.id)

			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}

			assert.Equal(t, tt.want, got)
			mockDB.AssertExpectations(t)
		})
	}
}


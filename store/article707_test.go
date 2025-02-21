package store

import (
	"database/sql"
	"errors"
	"sync"
	"testing"

	"github.com/jinzhu/gorm"
	"github.com/raahii/golang-grpc-realworld-example/model"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

type MockArticleStore struct {
	db *MockDB
}
type MockDB struct {
	mock.Mock
}

/*
ROOST_METHOD_HASH=ArticleStore_CreateComment_b16d4a71d4
ROOST_METHOD_SIG_HASH=ArticleStore_CreateComment_7475736b06

FUNCTION_DEF=func (s *ArticleStore) CreateComment(m *model.Comment) error // CreateComment creates a comment of the article
*/
func (s *MockArticleStore) CreateComment(m *model.Comment) error {
	return s.db.Create(m).Error
}

func (m *MockDB) Create(value interface{}) *gorm.DB {
	args := m.Called(value)
	return args.Get(0).(*gorm.DB)
}

func TestArticleStoreCreateComment(t *testing.T) {
	tests := []struct {
		name    string
		comment *model.Comment
		dbError error
		wantErr bool
	}{
		{
			name: "Successfully Create a New Comment",
			comment: &model.Comment{
				Body:      "Test comment",
				UserID:    1,
				ArticleID: 1,
			},
			dbError: nil,
			wantErr: false,
		},
		{
			name: "Attempt to Create a Comment with Invalid Data",
			comment: &model.Comment{
				Body:      "",
				UserID:    1,
				ArticleID: 1,
			},
			dbError: errors.New("validation error"),
			wantErr: true,
		},
		{
			name: "Database Error During Comment Creation",
			comment: &model.Comment{
				Body:      "Test comment",
				UserID:    1,
				ArticleID: 1,
			},
			dbError: errors.New("database error"),
			wantErr: true,
		},
		{
			name: "Create Comment with Maximum Allowed Length",
			comment: &model.Comment{
				Body:      string(make([]byte, 1000)),
				UserID:    1,
				ArticleID: 1,
			},
			dbError: nil,
			wantErr: false,
		},
		{
			name: "Attempt to Create a Comment for a Non-existent Article",
			comment: &model.Comment{
				Body:      "Test comment",
				UserID:    1,
				ArticleID: 9999,
			},
			dbError: errors.New("foreign key constraint violation"),
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockDB := new(MockDB)
			store := &MockArticleStore{
				db: mockDB,
			}

			mockDB.On("Create", mock.AnythingOfType("*model.Comment")).Return(&gorm.DB{Error: tt.dbError})

			err := store.CreateComment(tt.comment)

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
ROOST_METHOD_HASH=ArticleStore_DeleteFavorite_29c18a04a8
ROOST_METHOD_SIG_HASH=ArticleStore_DeleteFavorite_53deb5e792

FUNCTION_DEF=func (s *ArticleStore) DeleteFavorite(a *model.Article, u *model.User) error // DeleteFavorite unfavorite an article
*/
func TestArticleStoreDeleteFavoriteConcurrent(t *testing.T) {
	article := &model.Article{FavoritesCount: 100}
	users := make([]*model.User, 10)
	for i := range users {
		users[i] = &model.User{}
	}

	mockDB := new(mockDB)
	mockAssoc := new(mockAssociation)

	tx := &gorm.DB{}
	mockDB.On("Begin").Return(tx)
	mockDB.On("Model", mock.Anything).Return(tx)
	mockDB.On("Association", "FavoritedUsers").Return(mockAssoc)
	mockAssoc.On("Delete", mock.Anything).Return(nil)
	mockDB.On("Update", "favorites_count", mock.Anything).Return(tx)
	mockDB.On("Commit").Return(nil)

	db := &gorm.DB{Value: mockDB}
	store := &ArticleStore{db: db}

	var wg sync.WaitGroup
	for _, user := range users {
		wg.Add(1)
		go func(u *model.User) {
			defer wg.Done()
			_ = store.DeleteFavorite(article, u)
		}(user)
	}
	wg.Wait()

	assert.Equal(t, int32(90), article.FavoritesCount)
	mockDB.AssertExpectations(t)
	mockAssoc.AssertExpectations(t)
}

func TestArticleStoreDeleteFavoriteScenarios(t *testing.T) {
	tests := []struct {
		name          string
		article       *model.Article
		user          *model.User
		setupMock     func(*mockDB, *mockAssociation)
		expectedError error
		expectedCount int32
	}{
		{
			name:    "Successfully Delete a Favorite Article",
			article: &model.Article{FavoritesCount: 1},
			user:    &model.User{},
			setupMock: func(db *mockDB, assoc *mockAssociation) {
				tx := &gorm.DB{}
				db.On("Begin").Return(tx)
				db.On("Model", mock.Anything).Return(tx)
				db.On("Association", "FavoritedUsers").Return(assoc)
				assoc.On("Delete", mock.Anything).Return(nil)
				db.On("Update", "favorites_count", gorm.Expr("favorites_count - ?", 1)).Return(tx)
				db.On("Commit").Return(nil)
			},
			expectedError: nil,
			expectedCount: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockDB := new(mockDB)
			mockAssoc := new(mockAssociation)
			tt.setupMock(mockDB, mockAssoc)
			db := &gorm.DB{Value: mockDB}
			store := &ArticleStore{db: db}
			err := store.DeleteFavorite(tt.article, tt.user)
			assert.Equal(t, tt.expectedError, err)
			assert.Equal(t, tt.expectedCount, tt.article.FavoritesCount)
			mockDB.AssertExpectations(t)
			mockAssoc.AssertExpectations(t)
		})
	}
}

func (m *mockAssociation) Append(values ...interface{}) error {
	ret := m.Called(values)
	return ret.Error(0)
}

func (m *mockAssociation) Clear() error {
	ret := m.Called()
	return ret.Error(0)
}

func (m *mockAssociation) Count() int {
	ret := m.Called()
	return ret.Int(0)
}

func (m *mockAssociation) Delete(values ...interface{}) error {
	ret := m.Called(values)
	return ret.Error(0)
}

func (m *mockAssociation) Find(out interface{}) error {
	ret := m.Called(out)
	return ret.Error(0)
}

func (m *mockAssociation) Replace(values ...interface{}) error {
	ret := m.Called(values)
	return ret.Error(0)
}

func (m *mockDB) Exec(query string, args ...interface{}) (sql.Result, error) {
	args = append([]interface{}{query}, args...)
	ret := m.Called(args...)
	return ret.Get(0).(sql.Result), ret.Error(1)
}

func (m *mockDB) Prepare(query string) (*sql.Stmt, error) {
	ret := m.Called(query)
	return ret.Get(0).(*sql.Stmt), ret.Error(1)
}

func (m *mockDB) Query(query string, args ...interface{}) (*sql.Rows, error) {
	args = append([]interface{}{query}, args...)
	ret := m.Called(args...)
	return ret.Get(0).(*sql.Rows), ret.Error(1)
}

func (m *mockDB) QueryRow(query string, args ...interface{}) *sql.Row {
	args = append([]interface{}{query}, args...)
	ret := m.Called(args...)
	return ret.Get(0).(*sql.Row)
}

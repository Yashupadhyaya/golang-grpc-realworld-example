package store

import (
	"reflect"
	"testing"
	"github.com/jinzhu/gorm"
	"github.com/raahii/golang-grpc-realworld-example/model"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/DATA-DOG/go-sqlmock"
	"errors"
	"github.com/stretchr/testify/mock"
	"fmt"
	"time"
	"sync"
)

type T struct {
	common
	isEnvSet bool
	context  *testContext
}
type ExpectedBegin struct {
	commonExpectation
	delay time.Duration
}
type ExpectedCommit struct {
	commonExpectation
}
type ExpectedExec struct {
	queryBasedExpectation
	result driver.Result
	delay  time.Duration
}
type ExpectedRollback struct {
	commonExpectation
}
type Association struct {
	Error  error
	scope  *Scope
	column string
	field  *Field
}
type Tag struct {
	gorm.Model
	Name string `gorm:"not null"`
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
type mockDB struct {
	mock.Mock
}
type MockDB struct {
	mock.Mock
}
type MockDB struct {
	mock.Mock
	Error error
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
/*
ROOST_METHOD_HASH=NewArticleStore_6be2824012
ROOST_METHOD_SIG_HASH=NewArticleStore_3fe6f79a92


 */
func TestNewArticleStore(t *testing.T) {
	tests := []struct {
		name string
		db   *gorm.DB
		want *ArticleStore
	}{
		{
			name: "Create ArticleStore with Valid DB Connection",
			db:   &gorm.DB{},
			want: &ArticleStore{db: &gorm.DB{}},
		},
		{
			name: "Create ArticleStore with Nil DB Connection",
			db:   nil,
			want: &ArticleStore{db: nil},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := NewArticleStore(tt.db)
			if got == nil {
				t.Errorf("NewArticleStore() returned nil")
				return
			}
			if !reflect.DeepEqual(got.db, tt.want.db) {
				t.Errorf("NewArticleStore() = %v, want %v", got.db, tt.want.db)
			}
		})
	}

	t.Run("Verify ArticleStore Immutability", func(t *testing.T) {
		db := &gorm.DB{}
		store1 := NewArticleStore(db)
		store2 := NewArticleStore(db)
		if store1 == store2 {
			t.Errorf("NewArticleStore() returned the same instance for multiple calls")
		}
		if store1.db != store2.db {
			t.Errorf("NewArticleStore() did not use the same DB reference for multiple calls")
		}
	})

	t.Run("Check DB Field Accessibility", func(t *testing.T) {
		db := &gorm.DB{}
		store := NewArticleStore(db)
		if store.db != db {
			t.Errorf("NewArticleStore() did not set the db field correctly")
		}
	})

	t.Run("Performance Test for Multiple Instantiations", func(t *testing.T) {
		db := &gorm.DB{}
		for i := 0; i < 1000; i++ {
			NewArticleStore(db)
		}

	})
}


/*
ROOST_METHOD_HASH=Create_0a911e138d
ROOST_METHOD_SIG_HASH=Create_723c594377


 */
func TestArticleStoreCreate(t *testing.T) {
	tests := []struct {
		name    string
		article *model.Article
		setupDB func(*gorm.DB)
		wantErr bool
	}{
		{
			name: "Successfully Create a New Article",
			article: &model.Article{
				Title:       "Test Article",
				Description: "This is a test article",
				Body:        "This is the body of the test article",
				UserID:      1,
			},
			setupDB: func(db *gorm.DB) {

			},
			wantErr: false,
		},
		{
			name: "Attempt to Create an Article with Missing Required Fields",
			article: &model.Article{

				UserID: 1,
			},
			setupDB: func(db *gorm.DB) {

			},
			wantErr: true,
		},
		{
			name: "Create an Article with Maximum Length Fields",
			article: &model.Article{
				Title:       string(make([]byte, 255)),
				Description: string(make([]byte, 1000)),
				Body:        string(make([]byte, 10000)),
				UserID:      1,
			},
			setupDB: func(db *gorm.DB) {

			},
			wantErr: false,
		},
		{
			name: "Attempt to Create an Article with Duplicate Title",
			article: &model.Article{
				Title:       "Duplicate Title",
				Description: "This is a test article",
				Body:        "This is the body of the test article",
				UserID:      1,
			},
			setupDB: func(db *gorm.DB) {
				db.Create(&model.Article{
					Title:       "Duplicate Title",
					Description: "Existing article",
					Body:        "Existing body",
					UserID:      2,
				})
			},
			wantErr: true,
		},
		{
			name: "Create an Article with Associated Tags and Author",
			article: &model.Article{
				Title:       "Article with Associations",
				Description: "This article has tags and an author",
				Body:        "Body of the article with associations",
				UserID:      1,
				Tags: []model.Tag{
					{Name: "Tag1"},
					{Name: "Tag2"},
				},
				Author: model.User{Model: gorm.Model{ID: 1}, Username: "testuser"},
			},
			setupDB: func(db *gorm.DB) {
				db.Create(&model.User{Model: gorm.Model{ID: 1}, Username: "testuser"})
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			db, mock, err := sqlmock.New()
			require.NoError(t, err)
			defer db.Close()

			gormDB, err := gorm.Open("sqlite3", db)
			require.NoError(t, err)
			defer gormDB.Close()

			tt.setupDB(gormDB)

			store := &ArticleStore{db: gormDB}

			if !tt.wantErr {
				mock.ExpectBegin()
				mock.ExpectExec("INSERT INTO `articles`").WillReturnResult(sqlmock.NewResult(1, 1))
				mock.ExpectCommit()
			} else {
				mock.ExpectBegin()
				mock.ExpectExec("INSERT INTO `articles`").WillReturnError(gorm.ErrRecordNotFound)
				mock.ExpectRollback()
			}

			err = store.Create(tt.article)

			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}

			assert.NoError(t, mock.ExpectationsWereMet())
		})
	}
}


/*
ROOST_METHOD_HASH=CreateComment_58d394e2c6
ROOST_METHOD_SIG_HASH=CreateComment_28b95f60a6


 */
func TestArticleStoreCreateComment(t *testing.T) {
	tests := []struct {
		name    string
		setup   func(*gorm.DB)
		comment *model.Comment
		wantErr bool
	}{
		{
			name: "Successfully Create a Comment",
			setup: func(db *gorm.DB) {
				db.Create(&model.User{Model: gorm.Model{ID: 1}})
				db.Create(&model.Article{Model: gorm.Model{ID: 1}})
			},
			comment: &model.Comment{
				Body:      "Test comment",
				UserID:    1,
				ArticleID: 1,
			},
			wantErr: false,
		},
		{
			name: "Attempt to Create a Comment with Missing Required Fields",
			setup: func(db *gorm.DB) {
				db.Create(&model.User{Model: gorm.Model{ID: 1}})
				db.Create(&model.Article{Model: gorm.Model{ID: 1}})
			},
			comment: &model.Comment{

				UserID:    1,
				ArticleID: 1,
			},
			wantErr: true,
		},
		{
			name: "Create Comment with Maximum Length Body",
			setup: func(db *gorm.DB) {
				db.Create(&model.User{Model: gorm.Model{ID: 1}})
				db.Create(&model.Article{Model: gorm.Model{ID: 1}})
			},
			comment: &model.Comment{
				Body:      string(make([]byte, 1000)),
				UserID:    1,
				ArticleID: 1,
			},
			wantErr: false,
		},
		{
			name: "Create Comment for Non-Existent Article",
			setup: func(db *gorm.DB) {
				db.Create(&model.User{Model: gorm.Model{ID: 1}})
			},
			comment: &model.Comment{
				Body:      "Test comment",
				UserID:    1,
				ArticleID: 9999,
			},
			wantErr: true,
		},
		{
			name: "Create Comment with Non-Existent User",
			setup: func(db *gorm.DB) {
				db.Create(&model.Article{Model: gorm.Model{ID: 1}})
			},
			comment: &model.Comment{
				Body:      "Test comment",
				UserID:    9999,
				ArticleID: 1,
			},
			wantErr: true,
		},
		{
			name: "Create Comment with Special Characters in Body",
			setup: func(db *gorm.DB) {
				db.Create(&model.User{Model: gorm.Model{ID: 1}})
				db.Create(&model.Article{Model: gorm.Model{ID: 1}})
			},
			comment: &model.Comment{
				Body:      "Test comment with special characters: !@#$%^&*()_+ and emoji 😊",
				UserID:    1,
				ArticleID: 1,
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			db, err := gorm.Open("sqlite3", ":memory:")
			require.NoError(t, err)
			defer db.Close()

			err = db.AutoMigrate(&model.Comment{}, &model.User{}, &model.Article{}).Error
			require.NoError(t, err)

			if tt.setup != nil {
				tt.setup(db)
			}

			s := &ArticleStore{db: db}

			err = s.CreateComment(tt.comment)

			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)

				var createdComment model.Comment
				err = db.First(&createdComment, "body = ?", tt.comment.Body).Error
				assert.NoError(t, err)
				assert.Equal(t, tt.comment.Body, createdComment.Body)
				assert.Equal(t, tt.comment.UserID, createdComment.UserID)
				assert.Equal(t, tt.comment.ArticleID, createdComment.ArticleID)
			}
		})
	}

	t.Run("Create Multiple Comments in Quick Succession", func(t *testing.T) {
		db, err := gorm.Open("sqlite3", ":memory:")
		require.NoError(t, err)
		defer db.Close()

		err = db.AutoMigrate(&model.Comment{}, &model.User{}, &model.Article{}).Error
		require.NoError(t, err)

		db.Create(&model.User{Model: gorm.Model{ID: 1}})
		db.Create(&model.Article{Model: gorm.Model{ID: 1}})

		s := &ArticleStore{db: db}

		comments := []*model.Comment{
			{Body: "Comment 1", UserID: 1, ArticleID: 1},
			{Body: "Comment 2", UserID: 1, ArticleID: 1},
			{Body: "Comment 3", UserID: 1, ArticleID: 1},
		}

		for _, comment := range comments {
			err := s.CreateComment(comment)
			assert.NoError(t, err)
		}

		var count int
		db.Model(&model.Comment{}).Count(&count)
		assert.Equal(t, len(comments), count)
	})
}


/*
ROOST_METHOD_HASH=Delete_a8dc14c210
ROOST_METHOD_SIG_HASH=Delete_a4cc8044b1


 */
func TestArticleStoreDelete(t *testing.T) {
	tests := []struct {
		name          string
		setupDB       func(*gorm.DB)
		article       *model.Article
		expectedError error
		validate      func(*testing.T, *gorm.DB)
	}{
		{
			name: "Successfully Delete an Existing Article",
			setupDB: func(db *gorm.DB) {
				article := &model.Article{Model: gorm.Model{ID: 1}, Title: "Test Article", Description: "Test Description", Body: "Test Body"}
				db.Create(article)
			},
			article:       &model.Article{Model: gorm.Model{ID: 1}},
			expectedError: nil,
			validate: func(t *testing.T, db *gorm.DB) {
				var count int64
				db.Model(&model.Article{}).Count(&count)
				assert.Equal(t, int64(0), count)
			},
		},
		{
			name:          "Attempt to Delete a Non-existent Article",
			setupDB:       func(db *gorm.DB) {},
			article:       &model.Article{Model: gorm.Model{ID: 999}},
			expectedError: gorm.ErrRecordNotFound,
			validate: func(t *testing.T, db *gorm.DB) {
				var count int64
				db.Model(&model.Article{}).Count(&count)
				assert.Equal(t, int64(0), count)
			},
		},
		{
			name: "Delete an Article with Associated Tags",
			setupDB: func(db *gorm.DB) {
				article := &model.Article{Model: gorm.Model{ID: 1}, Title: "Test Article", Description: "Test Description", Body: "Test Body"}
				tag := &model.Tag{Model: gorm.Model{ID: 1}, Name: "TestTag"}
				db.Create(article)
				db.Create(tag)
				db.Model(article).Association("Tags").Append(tag)
			},
			article:       &model.Article{Model: gorm.Model{ID: 1}},
			expectedError: nil,
			validate: func(t *testing.T, db *gorm.DB) {
				var articleCount, tagCount, associationCount int64
				db.Model(&model.Article{}).Count(&articleCount)
				db.Model(&model.Tag{}).Count(&tagCount)
				db.Table("article_tags").Count(&associationCount)
				assert.Equal(t, int64(0), articleCount)
				assert.Equal(t, int64(1), tagCount)
				assert.Equal(t, int64(0), associationCount)
			},
		},
		{
			name: "Delete an Article with Comments",
			setupDB: func(db *gorm.DB) {
				article := &model.Article{Model: gorm.Model{ID: 1}, Title: "Test Article", Description: "Test Description", Body: "Test Body"}
				comment := &model.Comment{Model: gorm.Model{ID: 1}, Body: "Test Comment", ArticleID: 1}
				db.Create(article)
				db.Create(comment)
			},
			article:       &model.Article{Model: gorm.Model{ID: 1}},
			expectedError: nil,
			validate: func(t *testing.T, db *gorm.DB) {
				var articleCount, commentCount int64
				db.Model(&model.Article{}).Count(&articleCount)
				db.Model(&model.Comment{}).Count(&commentCount)
				assert.Equal(t, int64(0), articleCount)
				assert.Equal(t, int64(0), commentCount)
			},
		},
		{
			name: "Delete an Article with Favorites",
			setupDB: func(db *gorm.DB) {
				article := &model.Article{Model: gorm.Model{ID: 1}, Title: "Test Article", Description: "Test Description", Body: "Test Body"}
				user := &model.User{Model: gorm.Model{ID: 1}, Username: "testuser"}
				db.Create(article)
				db.Create(user)
				db.Model(article).Association("FavoritedUsers").Append(user)
			},
			article:       &model.Article{Model: gorm.Model{ID: 1}},
			expectedError: nil,
			validate: func(t *testing.T, db *gorm.DB) {
				var articleCount, userCount, favoriteCount int64
				db.Model(&model.Article{}).Count(&articleCount)
				db.Model(&model.User{}).Count(&userCount)
				db.Table("favorite_articles").Count(&favoriteCount)
				assert.Equal(t, int64(0), articleCount)
				assert.Equal(t, int64(1), userCount)
				assert.Equal(t, int64(0), favoriteCount)
			},
		},
		{
			name: "Database Error During Deletion",
			setupDB: func(db *gorm.DB) {

				db.Close()
			},
			article:       &model.Article{Model: gorm.Model{ID: 1}},
			expectedError: errors.New("sql: database is closed"),
			validate:      func(t *testing.T, db *gorm.DB) {},
		},
		{
			name: "Delete an Article with NULL Fields",
			setupDB: func(db *gorm.DB) {
				article := &model.Article{Model: gorm.Model{ID: 1}, Title: "Test Article", Description: "Test Description"}
				db.Create(article)
			},
			article:       &model.Article{Model: gorm.Model{ID: 1}},
			expectedError: nil,
			validate: func(t *testing.T, db *gorm.DB) {
				var count int64
				db.Model(&model.Article{}).Count(&count)
				assert.Equal(t, int64(0), count)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			db, err := gorm.Open("sqlite3", ":memory:")
			require.NoError(t, err)
			defer db.Close()

			db.AutoMigrate(&model.Article{}, &model.Tag{}, &model.User{}, &model.Comment{})

			tt.setupDB(db)

			store := &ArticleStore{db: db}
			err = store.Delete(tt.article)

			if tt.expectedError != nil {
				assert.Error(t, err)
				assert.Equal(t, tt.expectedError.Error(), err.Error())
			} else {
				assert.NoError(t, err)
			}

			tt.validate(t, db)
		})
	}
}


/*
ROOST_METHOD_HASH=DeleteComment_b345e525a7
ROOST_METHOD_SIG_HASH=DeleteComment_732762ff12


 */
func (m *mockDB) Delete(value interface{}, where ...interface{}) *gorm.DB {
	args := m.Called(value)
	return args.Get(0).(*gorm.DB)
}

func TestArticleStoreDeleteComment(t *testing.T) {
	tests := []struct {
		name    string
		comment *model.Comment
		dbError error
		wantErr bool
	}{
		{
			name: "Successfully Delete an Existing Comment",
			comment: &model.Comment{
				Model: gorm.Model{ID: 1},
				Body:  "Test comment",
			},
			dbError: nil,
			wantErr: false,
		},
		{
			name: "Attempt to Delete a Non-existent Comment",
			comment: &model.Comment{
				Model: gorm.Model{ID: 999},
				Body:  "Non-existent comment",
			},
			dbError: gorm.ErrRecordNotFound,
			wantErr: true,
		},
		{
			name: "Delete Comment with Database Connection Error",
			comment: &model.Comment{
				Model: gorm.Model{ID: 2},
				Body:  "Another test comment",
			},
			dbError: errors.New("database connection error"),
			wantErr: true,
		},
		{
			name:    "Delete Comment with Null Comment Pointer",
			comment: nil,
			dbError: nil,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockDB := new(mockDB)
			if tt.comment != nil {
				mockDB.On("Delete", tt.comment).Return(&gorm.DB{Error: tt.dbError})
			}

			store := &ArticleStore{
				db: mockDB,
			}

			err := store.DeleteComment(tt.comment)

			if tt.wantErr {
				assert.Error(t, err)
				if tt.dbError != nil {
					assert.Equal(t, tt.dbError, err)
				}
			} else {
				assert.NoError(t, err)
			}

			mockDB.AssertExpectations(t)
		})
	}
}


/*
ROOST_METHOD_HASH=GetCommentByID_4bc82104a6
ROOST_METHOD_SIG_HASH=GetCommentByID_333cab101b


 */
func TestArticleStoreGetCommentByID(t *testing.T) {
	tests := []struct {
		name            string
		setupDB         func(*gorm.DB)
		inputID         uint
		expectedError   error
		expectedComment *model.Comment
	}{
		{
			name: "Successfully retrieve an existing comment",
			setupDB: func(db *gorm.DB) {
				comment := &model.Comment{
					Model:     gorm.Model{ID: 1},
					Body:      "Test comment",
					UserID:    1,
					ArticleID: 1,
				}
				db.Create(comment)
			},
			inputID:       1,
			expectedError: nil,
			expectedComment: &model.Comment{
				Model:     gorm.Model{ID: 1},
				Body:      "Test comment",
				UserID:    1,
				ArticleID: 1,
			},
		},
		{
			name: "Attempt to retrieve a non-existent comment",
			setupDB: func(db *gorm.DB) {

			},
			inputID:         999,
			expectedError:   gorm.ErrRecordNotFound,
			expectedComment: nil,
		},
		{
			name: "Handle database connection error",
			setupDB: func(db *gorm.DB) {

				db.AddError(errors.New("database connection error"))
			},
			inputID:         1,
			expectedError:   errors.New("database connection error"),
			expectedComment: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			db, err := gorm.Open("sqlite3", ":memory:")
			assert.NoError(t, err)
			defer db.Close()

			tt.setupDB(db)

			store := &ArticleStore{db: db}

			comment, err := store.GetCommentByID(tt.inputID)

			if tt.expectedError != nil {
				assert.Error(t, err)
				assert.Equal(t, tt.expectedError.Error(), err.Error())
			} else {
				assert.NoError(t, err)
			}

			assert.Equal(t, tt.expectedComment, comment)
		})
	}
}


/*
ROOST_METHOD_HASH=GetTags_ac049ebded
ROOST_METHOD_SIG_HASH=GetTags_25034b82b0


 */
func TestArticleStoreGetTags(t *testing.T) {
	tests := []struct {
		name    string
		dbSetup func(*gorm.DB)
		want    []model.Tag
		wantErr bool
	}{
		{
			name: "Successfully Retrieve All Tags",
			dbSetup: func(db *gorm.DB) {
				db.Create(&model.Tag{Name: "tag1"})
				db.Create(&model.Tag{Name: "tag2"})
				db.Create(&model.Tag{Name: "tag3"})
			},
			want: []model.Tag{
				{Name: "tag1"},
				{Name: "tag2"},
				{Name: "tag3"},
			},
			wantErr: false,
		},
		{
			name:    "Empty Tag List",
			dbSetup: func(db *gorm.DB) {},
			want:    []model.Tag{},
			wantErr: false,
		},
		{
			name: "Database Connection Error",
			dbSetup: func(db *gorm.DB) {
				db.AddError(errors.New("database connection error"))
			},
			want:    []model.Tag{},
			wantErr: true,
		},
		{
			name: "Large Number of Tags",
			dbSetup: func(db *gorm.DB) {
				for i := 0; i < 10000; i++ {
					db.Create(&model.Tag{Name: fmt.Sprintf("tag%d", i)})
				}
			},
			want:    nil,
			wantErr: false,
		},
		{
			name: "Duplicate Tags in Database",
			dbSetup: func(db *gorm.DB) {
				db.Create(&model.Tag{Name: "tag1"})
				db.Create(&model.Tag{Name: "tag2"})
				db.Create(&model.Tag{Name: "tag1"})
			},
			want: []model.Tag{
				{Name: "tag1"},
				{Name: "tag2"},
				{Name: "tag1"},
			},
			wantErr: false,
		},
		{
			name: "Partial Database Failure",
			dbSetup: func(db *gorm.DB) {
				db.Create(&model.Tag{Name: "tag1"})
				db.Create(&model.Tag{Name: "tag2"})
				db.AddError(errors.New("connection lost"))
			},
			want:    []model.Tag{},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			db, err := gorm.Open("sqlite3", ":memory:")
			if err != nil {
				t.Fatalf("failed to open database: %v", err)
			}
			defer db.Close()

			tt.dbSetup(db)

			s := &ArticleStore{db: db}

			got, err := s.GetTags()

			if (err != nil) != tt.wantErr {
				t.Errorf("ArticleStore.GetTags() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if tt.name == "Large Number of Tags" {
				if len(got) != 10000 {
					t.Errorf("ArticleStore.GetTags() got %d tags, want 10000", len(got))
				}
			} else if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("ArticleStore.GetTags() = %v, want %v", got, tt.want)
			}
		})
	}
}


/*
ROOST_METHOD_HASH=GetByID_36e92ad6eb
ROOST_METHOD_SIG_HASH=GetByID_9616e43e52


 */
func (m *MockDB) Error() error {
	args := m.Called()
	return args.Error(0)
}

func (m *MockDB) Find(out interface{}, where ...interface{}) *gorm.DB {
	args := m.Called(out, where)
	return args.Get(0).(*gorm.DB)
}

func (m *MockDB) Preload(column string, conditions ...interface{}) *gorm.DB {
	args := m.Called(column, conditions)
	return args.Get(0).(*gorm.DB)
}

func TestArticleStoreGetByID(t *testing.T) {
	tests := []struct {
		name            string
		id              uint
		mockSetup       func(*MockDB)
		expectedError   error
		expectedArticle *model.Article
	}{
		{
			name: "Successfully retrieve an existing article by ID",
			id:   1,
			mockSetup: func(mockDB *MockDB) {
				mockDB.On("Preload", "Tags").Return(mockDB)
				mockDB.On("Preload", "Author").Return(mockDB)
				mockDB.On("Find", mock.AnythingOfType("*model.Article"), uint(1)).Run(func(args mock.Arguments) {
					arg := args.Get(0).(*model.Article)
					*arg = model.Article{
						Model:  gorm.Model{ID: 1, CreatedAt: time.Now(), UpdatedAt: time.Now()},
						Title:  "Test Article",
						Tags:   []model.Tag{{Name: "test"}},
						Author: model.User{Model: gorm.Model{ID: 1}, Username: "testuser"},
					}
				}).Return(mockDB)
				mockDB.On("Error").Return(nil)
			},
			expectedError: nil,
			expectedArticle: &model.Article{
				Model:  gorm.Model{ID: 1},
				Title:  "Test Article",
				Tags:   []model.Tag{{Name: "test"}},
				Author: model.User{Model: gorm.Model{ID: 1}, Username: "testuser"},
			},
		},
		{
			name: "Attempt to retrieve a non-existent article",
			id:   999,
			mockSetup: func(mockDB *MockDB) {
				mockDB.On("Preload", "Tags").Return(mockDB)
				mockDB.On("Preload", "Author").Return(mockDB)
				mockDB.On("Find", mock.AnythingOfType("*model.Article"), uint(999)).Return(mockDB)
				mockDB.On("Error").Return(gorm.ErrRecordNotFound)
			},
			expectedError:   gorm.ErrRecordNotFound,
			expectedArticle: nil,
		},
		{
			name: "Handle database connection error",
			id:   1,
			mockSetup: func(mockDB *MockDB) {
				mockDB.On("Preload", "Tags").Return(mockDB)
				mockDB.On("Preload", "Author").Return(mockDB)
				mockDB.On("Find", mock.AnythingOfType("*model.Article"), uint(1)).Return(mockDB)
				mockDB.On("Error").Return(errors.New("database connection error"))
			},
			expectedError:   errors.New("database connection error"),
			expectedArticle: nil,
		},
		{
			name: "Verify preloading of Tags and Author",
			id:   2,
			mockSetup: func(mockDB *MockDB) {
				mockDB.On("Preload", "Tags").Return(mockDB)
				mockDB.On("Preload", "Author").Return(mockDB)
				mockDB.On("Find", mock.AnythingOfType("*model.Article"), uint(2)).Run(func(args mock.Arguments) {
					arg := args.Get(0).(*model.Article)
					*arg = model.Article{
						Model:  gorm.Model{ID: 2, CreatedAt: time.Now(), UpdatedAt: time.Now()},
						Title:  "Article with Tags",
						Tags:   []model.Tag{{Name: "tag1"}, {Name: "tag2"}},
						Author: model.User{Model: gorm.Model{ID: 2}, Username: "author"},
					}
				}).Return(mockDB)
				mockDB.On("Error").Return(nil)
			},
			expectedError: nil,
			expectedArticle: &model.Article{
				Model:  gorm.Model{ID: 2},
				Title:  "Article with Tags",
				Tags:   []model.Tag{{Name: "tag1"}, {Name: "tag2"}},
				Author: model.User{Model: gorm.Model{ID: 2}, Username: "author"},
			},
		},
		{
			name: "Retrieve article with no tags",
			id:   3,
			mockSetup: func(mockDB *MockDB) {
				mockDB.On("Preload", "Tags").Return(mockDB)
				mockDB.On("Preload", "Author").Return(mockDB)
				mockDB.On("Find", mock.AnythingOfType("*model.Article"), uint(3)).Run(func(args mock.Arguments) {
					arg := args.Get(0).(*model.Article)
					*arg = model.Article{
						Model:  gorm.Model{ID: 3, CreatedAt: time.Now(), UpdatedAt: time.Now()},
						Title:  "Article without Tags",
						Tags:   []model.Tag{},
						Author: model.User{Model: gorm.Model{ID: 3}, Username: "author"},
					}
				}).Return(mockDB)
				mockDB.On("Error").Return(nil)
			},
			expectedError: nil,
			expectedArticle: &model.Article{
				Model:  gorm.Model{ID: 3},
				Title:  "Article without Tags",
				Tags:   []model.Tag{},
				Author: model.User{Model: gorm.Model{ID: 3}, Username: "author"},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockDB := new(MockDB)
			tt.mockSetup(mockDB)

			store := &ArticleStore{
				db: struct {
					*MockDB
					*gorm.DB
				}{MockDB: mockDB},
			}

			article, err := store.GetByID(tt.id)

			if tt.expectedError != nil {
				assert.Error(t, err)
				assert.Equal(t, tt.expectedError.Error(), err.Error())
			} else {
				assert.NoError(t, err)
			}

			assert.Equal(t, tt.expectedArticle, article)

			mockDB.AssertExpectations(t)
		})
	}
}


/*
ROOST_METHOD_HASH=Update_51145aa965
ROOST_METHOD_SIG_HASH=Update_6c1b5471fe


 */
func (m *mockDB) Model(value interface{}) *gorm.DB {
	args := m.Called(value)
	return args.Get(0).(*gorm.DB)
}

func TestArticleStoreUpdate(t *testing.T) {
	tests := []struct {
		name    string
		article *model.Article
		dbError error
		wantErr bool
	}{
		{
			name: "Successfully Update an Existing Article",
			article: &model.Article{
				Model: gorm.Model{ID: 1},
				Title: "Updated Title",
				Body:  "Updated Body",
			},
			dbError: nil,
			wantErr: false,
		},
		{
			name: "Attempt to Update a Non-existent Article",
			article: &model.Article{
				Model: gorm.Model{ID: 999},
				Title: "Non-existent Article",
			},
			dbError: gorm.ErrRecordNotFound,
			wantErr: true,
		},
		{
			name: "Update Article with Invalid Data",
			article: &model.Article{
				Model: gorm.Model{ID: 2},
				Title: "",
			},
			dbError: errors.New("validation error"),
			wantErr: true,
		},
		{
			name: "Update Article with No Changes",
			article: &model.Article{
				Model: gorm.Model{ID: 3},
				Title: "Unchanged Title",
				Body:  "Unchanged Body",
			},
			dbError: nil,
			wantErr: false,
		},
		{
			name: "Update Article with Modified Relationships",
			article: &model.Article{
				Model: gorm.Model{ID: 4},
				Title: "Article with Modified Relationships",
				Tags:  []model.Tag{{Name: "NewTag"}},
			},
			dbError: nil,
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockDB := new(mockDB)
			store := &ArticleStore{db: mockDB}

			mockDB.On("Model", mock.AnythingOfType("*model.Article")).Return(mockDB)
			mockDB.On("Update", mock.AnythingOfType("*model.Article")).Return(&gorm.DB{Error: tt.dbError})

			err := store.Update(tt.article)

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

func (m *mockDB) Update(attrs ...interface{}) *gorm.DB {
	args := m.Called(attrs...)
	return args.Get(0).(*gorm.DB)
}


/*
ROOST_METHOD_HASH=GetComments_e24a0f1b73
ROOST_METHOD_SIG_HASH=GetComments_fa6661983e


 */
func NewMockDB() *MockDB {
	return &MockDB{}
}

func TestArticleStoreGetComments(t *testing.T) {
	tests := []struct {
		name           string
		article        *model.Article
		mockSetup      func(*MockDB)
		expectedResult []model.Comment
		expectedError  error
	}{
		{
			name: "Retrieve Comments for an Article with Existing Comments",
			article: &model.Article{
				Model: gorm.Model{ID: 1},
			},
			mockSetup: func(mockDB *MockDB) {
				mockDB.On("Preload", "Author").Return(mockDB)
				mockDB.On("Where", "article_id = ?", uint(1)).Return(mockDB)
				mockDB.On("Find", mock.AnythingOfType("*[]model.Comment"), mock.Anything).Run(func(args mock.Arguments) {
					arg := args.Get(0).(*[]model.Comment)
					*arg = []model.Comment{
						{Model: gorm.Model{ID: 1}, Body: "Comment 1", Author: model.User{Model: gorm.Model{ID: 1}, Username: "user1"}},
						{Model: gorm.Model{ID: 2}, Body: "Comment 2", Author: model.User{Model: gorm.Model{ID: 2}, Username: "user2"}},
					}
				}).Return(mockDB)
			},
			expectedResult: []model.Comment{
				{Model: gorm.Model{ID: 1}, Body: "Comment 1", Author: model.User{Model: gorm.Model{ID: 1}, Username: "user1"}},
				{Model: gorm.Model{ID: 2}, Body: "Comment 2", Author: model.User{Model: gorm.Model{ID: 2}, Username: "user2"}},
			},
			expectedError: nil,
		},
		{
			name: "Retrieve Comments for an Article with No Comments",
			article: &model.Article{
				Model: gorm.Model{ID: 2},
			},
			mockSetup: func(mockDB *MockDB) {
				mockDB.On("Preload", "Author").Return(mockDB)
				mockDB.On("Where", "article_id = ?", uint(2)).Return(mockDB)
				mockDB.On("Find", mock.AnythingOfType("*[]model.Comment"), mock.Anything).Run(func(args mock.Arguments) {
					arg := args.Get(0).(*[]model.Comment)
					*arg = []model.Comment{}
				}).Return(mockDB)
			},
			expectedResult: []model.Comment{},
			expectedError:  nil,
		},
		{
			name: "Attempt to Retrieve Comments for a Non-existent Article",
			article: &model.Article{
				Model: gorm.Model{ID: 999},
			},
			mockSetup: func(mockDB *MockDB) {
				mockDB.On("Preload", "Author").Return(mockDB)
				mockDB.On("Where", "article_id = ?", uint(999)).Return(mockDB)
				mockDB.On("Find", mock.AnythingOfType("*[]model.Comment"), mock.Anything).Return(&gorm.DB{Error: gorm.ErrRecordNotFound})
			},
			expectedResult: []model.Comment{},
			expectedError:  gorm.ErrRecordNotFound,
		},
		{
			name: "Database Connection Error Handling",
			article: &model.Article{
				Model: gorm.Model{ID: 3},
			},
			mockSetup: func(mockDB *MockDB) {
				mockDB.On("Preload", "Author").Return(mockDB)
				mockDB.On("Where", "article_id = ?", uint(3)).Return(mockDB)
				mockDB.On("Find", mock.AnythingOfType("*[]model.Comment"), mock.Anything).Return(&gorm.DB{Error: errors.New("database connection error")})
			},
			expectedResult: []model.Comment{},
			expectedError:  errors.New("database connection error"),
		},
		{
			name: "Large Number of Comments Retrieval",
			article: &model.Article{
				Model: gorm.Model{ID: 4},
			},
			mockSetup: func(mockDB *MockDB) {
				mockDB.On("Preload", "Author").Return(mockDB)
				mockDB.On("Where", "article_id = ?", uint(4)).Return(mockDB)
				mockDB.On("Find", mock.AnythingOfType("*[]model.Comment"), mock.Anything).Run(func(args mock.Arguments) {
					arg := args.Get(0).(*[]model.Comment)
					*arg = make([]model.Comment, 1000)
					for i := 0; i < 1000; i++ {
						(*arg)[i] = model.Comment{
							Model:  gorm.Model{ID: uint(i + 1)},
							Body:   "Comment body",
							Author: model.User{Model: gorm.Model{ID: uint(i%10 + 1)}, Username: "user"},
						}
					}
				}).Return(mockDB)
			},
			expectedResult: func() []model.Comment {
				comments := make([]model.Comment, 1000)
				for i := 0; i < 1000; i++ {
					comments[i] = model.Comment{
						Model:  gorm.Model{ID: uint(i + 1)},
						Body:   "Comment body",
						Author: model.User{Model: gorm.Model{ID: uint(i%10 + 1)}, Username: "user"},
					}
				}
				return comments
			}(),
			expectedError: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockDB := NewMockDB()
			tt.mockSetup(mockDB)

			store := &ArticleStore{db: mockDB}

			result, err := store.GetComments(tt.article)

			assert.Equal(t, tt.expectedError, err)
			assert.Equal(t, tt.expectedResult, result)

			mockDB.AssertExpectations(t)
		})
	}
}

func (m *MockDB) Where(query interface{}, args ...interface{}) *gorm.DB {
	arguments := m.Called(query, args)
	return arguments.Get(0).(*gorm.DB)
}


/*
ROOST_METHOD_HASH=IsFavorited_7ef7d3ed9e
ROOST_METHOD_SIG_HASH=IsFavorited_f34d52378f


 */
func (m *MockDB) Count(value interface{}) *gorm.DB {
	args := m.Called(value)
	return args.Get(0).(*gorm.DB)
}

func (m *MockDB) Table(name string) *gorm.DB {
	args := m.Called(name)
	return args.Get(0).(*gorm.DB)
}

func TestArticleStoreIsFavorited(t *testing.T) {
	tests := []struct {
		name        string
		setupMock   func(*MockDB)
		article     *model.Article
		user        *model.User
		expected    bool
		expectedErr error
	}{
		{
			name: "Article is favorited by the user",
			setupMock: func(db *MockDB) {
				db.On("Table", "favorite_articles").Return(db)
				db.On("Where", "article_id = ? AND user_id = ?", uint(1), uint(1)).Return(db)
				db.On("Count", mock.AnythingOfType("*int")).Run(func(args mock.Arguments) {
					*args.Get(0).(*int) = 1
				}).Return(db)
			},
			article:     &model.Article{Model: gorm.Model{ID: 1}},
			user:        &model.User{Model: gorm.Model{ID: 1}},
			expected:    true,
			expectedErr: nil,
		},
		{
			name: "Article is not favorited by the user",
			setupMock: func(db *MockDB) {
				db.On("Table", "favorite_articles").Return(db)
				db.On("Where", "article_id = ? AND user_id = ?", uint(1), uint(1)).Return(db)
				db.On("Count", mock.AnythingOfType("*int")).Run(func(args mock.Arguments) {
					*args.Get(0).(*int) = 0
				}).Return(db)
			},
			article:     &model.Article{Model: gorm.Model{ID: 1}},
			user:        &model.User{Model: gorm.Model{ID: 1}},
			expected:    false,
			expectedErr: nil,
		},
		{
			name:        "Nil Article parameter",
			setupMock:   func(db *MockDB) {},
			article:     nil,
			user:        &model.User{Model: gorm.Model{ID: 1}},
			expected:    false,
			expectedErr: nil,
		},
		{
			name:        "Nil User parameter",
			setupMock:   func(db *MockDB) {},
			article:     &model.Article{Model: gorm.Model{ID: 1}},
			user:        nil,
			expected:    false,
			expectedErr: nil,
		},
		{
			name: "Database error",
			setupMock: func(db *MockDB) {
				db.On("Table", "favorite_articles").Return(db)
				db.On("Where", "article_id = ? AND user_id = ?", uint(1), uint(1)).Return(db)
				db.On("Count", mock.AnythingOfType("*int")).Return(db).Error(errors.New("database error"))
			},
			article:     &model.Article{Model: gorm.Model{ID: 1}},
			user:        &model.User{Model: gorm.Model{ID: 1}},
			expected:    false,
			expectedErr: errors.New("database error"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			mockDB := new(MockDB)
			tt.setupMock(mockDB)

			store := &ArticleStore{db: mockDB}

			result, err := store.IsFavorited(tt.article, tt.user)

			assert.Equal(t, tt.expected, result)
			if tt.expectedErr != nil {
				assert.EqualError(t, err, tt.expectedErr.Error())
			} else {
				assert.NoError(t, err)
			}

			mockDB.AssertExpectations(t)
		})
	}
}

func (m *MockDB) Where(query interface{}, args ...interface{}) *gorm.DB {
	mockArgs := m.Called(query, args)
	return mockArgs.Get(0).(*gorm.DB)
}


/*
ROOST_METHOD_HASH=GetFeedArticles_9c4f57afe4
ROOST_METHOD_SIG_HASH=GetFeedArticles_cadca0e51b


 */
func TestArticleStoreGetFeedArticles(t *testing.T) {
	tests := []struct {
		name     string
		userIDs  []uint
		limit    int64
		offset   int64
		mockDB   func() *gorm.DB
		expected []model.Article
		wantErr  bool
	}{
		{
			name:    "Successful Retrieval of Feed Articles",
			userIDs: []uint{1, 2},
			limit:   10,
			offset:  0,
			mockDB: func() *gorm.DB {
				db := &gorm.DB{}
				db.AddError(nil)

				return db
			},
			expected: []model.Article{
				{Model: gorm.Model{ID: 1}, Title: "Article 1", UserID: 1, Author: model.User{Model: gorm.Model{ID: 1}, Username: "user1"}},
				{Model: gorm.Model{ID: 2}, Title: "Article 2", UserID: 2, Author: model.User{Model: gorm.Model{ID: 2}, Username: "user2"}},
			},
			wantErr: false,
		},
		{
			name:    "Empty Result Set",
			userIDs: []uint{99, 100},
			limit:   10,
			offset:  0,
			mockDB: func() *gorm.DB {
				db := &gorm.DB{}
				db.AddError(nil)

				return db
			},
			expected: []model.Article{},
			wantErr:  false,
		},
		{
			name:    "Pagination with Offset and Limit",
			userIDs: []uint{1, 2, 3},
			limit:   2,
			offset:  1,
			mockDB: func() *gorm.DB {
				db := &gorm.DB{}
				db.AddError(nil)

				return db
			},
			expected: []model.Article{
				{Model: gorm.Model{ID: 2}, Title: "Article 2", UserID: 1, Author: model.User{Model: gorm.Model{ID: 1}, Username: "user1"}},
				{Model: gorm.Model{ID: 3}, Title: "Article 3", UserID: 2, Author: model.User{Model: gorm.Model{ID: 2}, Username: "user2"}},
			},
			wantErr: false,
		},
		{
			name:    "Error Handling for Database Failures",
			userIDs: []uint{1, 2},
			limit:   10,
			offset:  0,
			mockDB: func() *gorm.DB {
				db := &gorm.DB{}
				db.AddError(errors.New("database connection failed"))
				return db
			},
			expected: nil,
			wantErr:  true,
		},
		{
			name: "Large Number of User IDs",
			userIDs: func() []uint {
				ids := make([]uint, 10000)
				for i := range ids {
					ids[i] = uint(i + 1)
				}
				return ids
			}(),
			limit:  50,
			offset: 0,
			mockDB: func() *gorm.DB {
				db := &gorm.DB{}
				db.AddError(nil)

				return db
			},
			expected: func() []model.Article {
				articles := make([]model.Article, 50)
				for i := range articles {
					articles[i] = model.Article{
						Model:  gorm.Model{ID: uint(i + 1)},
						Title:  "Article",
						UserID: uint(i + 1),
						Author: model.User{Model: gorm.Model{ID: uint(i + 1)}, Username: "user"},
					}
				}
				return articles
			}(),
			wantErr: false,
		},
		{
			name:    "Preloading of Author Information",
			userIDs: []uint{1, 2},
			limit:   10,
			offset:  0,
			mockDB: func() *gorm.DB {
				db := &gorm.DB{}
				db.AddError(nil)

				return db
			},
			expected: []model.Article{
				{Model: gorm.Model{ID: 1}, Title: "Article 1", UserID: 1, Author: model.User{Model: gorm.Model{ID: 1}, Username: "user1", Email: "user1@example.com"}},
				{Model: gorm.Model{ID: 2}, Title: "Article 2", UserID: 2, Author: model.User{Model: gorm.Model{ID: 2}, Username: "user2", Email: "user2@example.com"}},
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			store := &ArticleStore{
				db: tt.mockDB(),
			}

			got, err := store.GetFeedArticles(tt.userIDs, tt.limit, tt.offset)

			if (err != nil) != tt.wantErr {
				t.Errorf("ArticleStore.GetFeedArticles() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !reflect.DeepEqual(got, tt.expected) {
				t.Errorf("ArticleStore.GetFeedArticles() = %v, want %v", got, tt.expected)
			}

			if tt.name == "Preloading of Author Information" {
				for _, article := range got {
					if article.Author.ID == 0 || article.Author.Username == "" || article.Author.Email == "" {
						t.Errorf("Author information not properly preloaded for article: %v", article)
					}
				}
			}
		})
	}
}


/*
ROOST_METHOD_HASH=AddFavorite_2b0cb9d894
ROOST_METHOD_SIG_HASH=AddFavorite_c4dea0ee90


 */
func (m *MockAssociation) Append(values ...interface{}) error {
	args := m.Called(values...)
	return args.Error(0)
}

func (m *MockDB) Association(column string) *gorm.Association {
	args := m.Called(column)
	return args.Get(0).(*gorm.Association)
}

func (m *MockDB) Begin() *gorm.DB {
	args := m.Called()
	return args.Get(0).(*gorm.DB)
}

func (m *MockDB) Commit() *gorm.DB {
	args := m.Called()
	return args.Get(0).(*gorm.DB)
}

func (m *MockDB) Model(value interface{}) *gorm.DB {
	args := m.Called(value)
	return args.Get(0).(*gorm.DB)
}

func (m *MockDB) Rollback() *gorm.DB {
	args := m.Called()
	return args.Get(0).(*gorm.DB)
}

func TestArticleStoreAddFavorite(t *testing.T) {
	tests := []struct {
		name           string
		setupMock      func(*MockDB)
		article        *model.Article
		user           *model.User
		expectedError  error
		expectedCount  int32
		expectedUsers  []model.User
		concurrentTest bool
	}{
		{
			name: "Successfully Add Favorite",
			setupMock: func(mockDB *MockDB) {
				mockDB.On("Begin").Return(mockDB)
				mockDB.On("Model", mock.Anything).Return(mockDB)
				mockAssoc := &MockAssociation{}
				mockAssoc.On("Append", mock.Anything).Return(nil)
				mockDB.On("Association", "FavoritedUsers").Return(mockAssoc)
				mockDB.On("Update", "favorites_count", mock.Anything).Return(mockDB)
				mockDB.On("Commit").Return(mockDB)
			},
			article:       &model.Article{FavoritesCount: 0},
			user:          &model.User{Model: gorm.Model{ID: 1}},
			expectedError: nil,
			expectedCount: 1,
			expectedUsers: []model.User{{Model: gorm.Model{ID: 1}}},
		},
		{
			name: "Add Favorite for Already Favorited Article",
			setupMock: func(mockDB *MockDB) {
				mockDB.On("Begin").Return(mockDB)
				mockDB.On("Model", mock.Anything).Return(mockDB)
				mockAssoc := &MockAssociation{}
				mockAssoc.On("Append", mock.Anything).Return(nil)
				mockDB.On("Association", "FavoritedUsers").Return(mockAssoc)
				mockDB.On("Update", "favorites_count", mock.Anything).Return(mockDB)
				mockDB.On("Commit").Return(mockDB)
			},
			article:       &model.Article{FavoritesCount: 1, FavoritedUsers: []model.User{{Model: gorm.Model{ID: 1}}}},
			user:          &model.User{Model: gorm.Model{ID: 1}},
			expectedError: nil,
			expectedCount: 2,
			expectedUsers: []model.User{{Model: gorm.Model{ID: 1}}},
		},
		{
			name: "Database Error During User Association",
			setupMock: func(mockDB *MockDB) {
				mockDB.On("Begin").Return(mockDB)
				mockDB.On("Model", mock.Anything).Return(mockDB)
				mockAssoc := &MockAssociation{}
				mockAssoc.On("Append", mock.Anything).Return(errors.New("database error"))
				mockDB.On("Association", "FavoritedUsers").Return(mockAssoc)
				mockDB.On("Rollback").Return(mockDB)
			},
			article:       &model.Article{FavoritesCount: 0},
			user:          &model.User{Model: gorm.Model{ID: 1}},
			expectedError: errors.New("database error"),
			expectedCount: 0,
			expectedUsers: []model.User{},
		},
		{
			name: "Database Error During FavoritesCount Update",
			setupMock: func(mockDB *MockDB) {
				mockDB.On("Begin").Return(mockDB)
				mockDB.On("Model", mock.Anything).Return(mockDB)
				mockAssoc := &MockAssociation{}
				mockAssoc.On("Append", mock.Anything).Return(nil)
				mockDB.On("Association", "FavoritedUsers").Return(mockAssoc)
				mockDB.On("Update", "favorites_count", mock.Anything).Return(mockDB)
				mockDB.Error = errors.New("update error")
				mockDB.On("Rollback").Return(mockDB)
			},
			article:       &model.Article{FavoritesCount: 0},
			user:          &model.User{Model: gorm.Model{ID: 1}},
			expectedError: errors.New("update error"),
			expectedCount: 0,
			expectedUsers: []model.User{},
		},
		{
			name:          "Add Favorite with Nil Article",
			setupMock:     func(mockDB *MockDB) {},
			article:       nil,
			user:          &model.User{Model: gorm.Model{ID: 1}},
			expectedError: errors.New("invalid argument: article is nil"),
			expectedCount: 0,
			expectedUsers: []model.User{},
		},
		{
			name:          "Add Favorite with Nil User",
			setupMock:     func(mockDB *MockDB) {},
			article:       &model.Article{FavoritesCount: 0},
			user:          nil,
			expectedError: errors.New("invalid argument: user is nil"),
			expectedCount: 0,
			expectedUsers: []model.User{},
		},
		{
			name: "Concurrent Favorite Additions",
			setupMock: func(mockDB *MockDB) {
				mockDB.On("Begin").Return(mockDB)
				mockDB.On("Model", mock.Anything).Return(mockDB)
				mockAssoc := &MockAssociation{}
				mockAssoc.On("Append", mock.Anything).Return(nil)
				mockDB.On("Association", "FavoritedUsers").Return(mockAssoc)
				mockDB.On("Update", "favorites_count", mock.Anything).Return(mockDB)
				mockDB.On("Commit").Return(mockDB)
			},
			article:        &model.Article{FavoritesCount: 0},
			user:           &model.User{Model: gorm.Model{ID: 1}},
			expectedError:  nil,
			expectedCount:  5,
			expectedUsers:  []model.User{{Model: gorm.Model{ID: 1}}, {Model: gorm.Model{ID: 2}}, {Model: gorm.Model{ID: 3}}, {Model: gorm.Model{ID: 4}}, {Model: gorm.Model{ID: 5}}},
			concurrentTest: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockDB := new(MockDB)
			tt.setupMock(mockDB)

			store := &ArticleStore{db: mockDB}

			if tt.concurrentTest {
				var wg sync.WaitGroup
				for i := 0; i < 5; i++ {
					wg.Add(1)
					go func(id uint) {
						defer wg.Done()
						user := &model.User{Model: gorm.Model{ID: id}}
						err := store.AddFavorite(tt.article, user)
						assert.NoError(t, err)
					}(uint(i + 1))
				}
				wg.Wait()
			} else {
				err := store.AddFavorite(tt.article, tt.user)
				assert.Equal(t, tt.expectedError, err)
			}

			if tt.article != nil {
				assert.Equal(t, tt.expectedCount, tt.article.FavoritesCount)
				assert.ElementsMatch(t, tt.expectedUsers, tt.article.FavoritedUsers)
			}

			mockDB.AssertExpectations(t)
		})
	}
}

func (m *MockDB) Update(column string, value interface{}) *gorm.DB {
	args := m.Called(column, value)
	return args.Get(0).(*gorm.DB)
}


/*
ROOST_METHOD_HASH=DeleteFavorite_a856bcbb70
ROOST_METHOD_SIG_HASH=DeleteFavorite_f7e5c0626f


 */
func (m *MockAssociation) Delete(values ...interface{}) *gorm.Association {
	args := m.Called(values...)
	return args.Get(0).(*gorm.Association)
}

func TestArticleStoreDeleteFavorite(t *testing.T) {
	tests := []struct {
		name            string
		setupMock       func(*MockDB, *MockAssociation)
		article         *model.Article
		user            *model.User
		expectedError   error
		expectedCount   int32
		concurrentCalls int
	}{
		{
			name: "Successfully Delete a Favorite",
			setupMock: func(db *MockDB, assoc *MockAssociation) {
				db.On("Begin").Return(db)
				db.On("Model", mock.Anything).Return(db)
				db.On("Association", "FavoritedUsers").Return(assoc)
				assoc.On("Delete", mock.Anything).Return(assoc)
				db.On("Update", "favorites_count", mock.Anything).Return(db)
				db.On("Commit").Return(db)
				db.On("Rollback").Return(db)
			},
			article:       &model.Article{FavoritesCount: 1},
			user:          &model.User{},
			expectedError: nil,
			expectedCount: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockDB := new(MockDB)
			mockAssoc := new(MockAssociation)
			tt.setupMock(mockDB, mockAssoc)

			store := &ArticleStore{db: *mockDB}

			if tt.concurrentCalls > 0 {
				var wg sync.WaitGroup
				for i := 0; i < tt.concurrentCalls; i++ {
					wg.Add(1)
					go func() {
						defer wg.Done()
						err := store.DeleteFavorite(tt.article, tt.user)
						assert.Equal(t, tt.expectedError, err)
					}()
				}
				wg.Wait()
				assert.Equal(t, tt.expectedCount, tt.article.FavoritesCount)
			} else {
				err := store.DeleteFavorite(tt.article, tt.user)
				assert.Equal(t, tt.expectedError, err)
				assert.Equal(t, tt.expectedCount, tt.article.FavoritesCount)
			}

			mockDB.AssertExpectations(t)
			mockAssoc.AssertExpectations(t)
		})
	}
}


/*
ROOST_METHOD_HASH=GetArticles_6382a4fe7a
ROOST_METHOD_SIG_HASH=GetArticles_1a0b3b0e8b


 */
func TestArticleStoreGetArticles(t *testing.T) {

	mockDB := &gorm.DB{}

	tests := []struct {
		name        string
		tagName     string
		username    string
		favoritedBy *model.User
		limit       int64
		offset      int64
		mockSetup   func(*gorm.DB)
		expected    []model.Article
		expectedErr error
	}{
		{
			name:        "Retrieve Articles Without Any Filters",
			tagName:     "",
			username:    "",
			favoritedBy: nil,
			limit:       10,
			offset:      0,
			mockSetup: func(db *gorm.DB) {
				db.AddError(nil)
				db.Callback().Query().Register("mock_query", func(scope *gorm.Scope) {
					if scope.HasError() {
						return
					}
					scope.InstanceSet("mock_data", []model.Article{
						{Model: gorm.Model{ID: 1}, Title: "Article 1"},
						{Model: gorm.Model{ID: 2}, Title: "Article 2"},
					})
				})
			},
			expected: []model.Article{
				{Model: gorm.Model{ID: 1}, Title: "Article 1"},
				{Model: gorm.Model{ID: 2}, Title: "Article 2"},
			},
			expectedErr: nil,
		},
		{
			name:        "Filter Articles by Tag Name",
			tagName:     "golang",
			username:    "",
			favoritedBy: nil,
			limit:       10,
			offset:      0,
			mockSetup: func(db *gorm.DB) {
				db.AddError(nil)
				db.Callback().Query().Register("mock_query", func(scope *gorm.Scope) {
					if scope.HasError() {
						return
					}
					scope.InstanceSet("mock_data", []model.Article{
						{Model: gorm.Model{ID: 3}, Title: "Golang Article"},
					})
				})
			},
			expected: []model.Article{
				{Model: gorm.Model{ID: 3}, Title: "Golang Article"},
			},
			expectedErr: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			mockDB = &gorm.DB{}
			tt.mockSetup(mockDB)

			store := &ArticleStore{db: mockDB}

			articles, err := store.GetArticles(tt.tagName, tt.username, tt.favoritedBy, tt.limit, tt.offset)

			assert.Equal(t, tt.expected, articles)
			assert.Equal(t, tt.expectedErr, err)
		})
	}
}


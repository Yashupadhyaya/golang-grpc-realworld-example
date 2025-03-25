package store

import (
	fmt "fmt"
	filepath "path/filepath"
	runtime "runtime"
	testing "testing"

	"github.com/DATA-DOG/go-sqlmock"
	gorm "github.com/jinzhu/gorm"
	model "github.com/raahii/golang-grpc-realworld-example/model"
	assert "github.com/stretchr/testify/assert"
)

type MockArticleStore struct {
	db *gorm.DB
}

/*
ROOST_METHOD_HASH=Create_c9b61e3f60
ROOST_METHOD_SIG_HASH=Create_b9fba017bc

FUNCTION_DEF=func Create(m *model.Article) string
*/
func TestCreate(t *testing.T) {

	var testTable = []struct {
		description string
		article     *model.Article
		output      string
	}{
		{"Successful creation of an article", &model.Article{Title: "Test title", Description: "Test description", Body: "Test body"}, "just for testing"},
		{"Article model with missing fields", &model.Article{Title: "", Description: "", Body: ""}, "just for testing"},
		{"Article model with invalid fields", &model.Article{Title: "1", Description: "2", Body: "3"}, "just for testing"},
	}

	for _, test := range testTable {

		t.Run(test.description, func(t *testing.T) {
			defer func() {
				if r := recover(); r != nil {
					t.Logf("Panic encountered so failing test. %v", r)
					t.Fail()
				}
			}()
			as := ArticleStore{}
			actual := as.Create(test.article)
			assert.Equal(t, test.output, actual)
		})
	}
}

/*
ROOST_METHOD_HASH=ArticleStore_Create_1273475ade
ROOST_METHOD_SIG_HASH=ArticleStore_Create_a27282cad5

FUNCTION_DEF=func (s *ArticleStore) Create(m *model.Article) error // Create creates an article
*/
func (s *MockArticleStore) Create(m *model.Article) error {
	return s.db.Create(&m).Error
}

func TestArticleStoreCreate(t *testing.T) {

	tests := []struct {
		name      string
		article   model.Article
		mockFunc  func(mock sqlmock.Sqlmock)
		wantError bool
	}{
		{
			name:    "Successful creation of an article",
			article: model.Article{Title: "Test Article", Description: "Test Description", Body: "Test Body"},
			mockFunc: func(mock sqlmock.Sqlmock) {
				mock.ExpectBegin()
				mock.ExpectExec("INSERT INTO").
					WithArgs("Test Article", "Test Description", "Test Body").
					WillReturnResult(sqlmock.NewResult(1, 1))
				mock.ExpectCommit()
			},
			wantError: false,
		},
		{
			name:    "Create an article with invalid article parameters",
			article: model.Article{},
			mockFunc: func(mock sqlmock.Sqlmock) {
				mock.ExpectBegin()
				mock.ExpectExec("INSERT INTO").
					WithArgs().
					WillReturnError(fmt.Errorf("unable to insert article"))
				mock.ExpectRollback()
			},
			wantError: true,
		},
		{
			name:    "Creation of an article with a failing database",
			article: model.Article{Title: "Test Article", Description: "Test Description", Body: "Test Body"},
			mockFunc: func(mock sqlmock.Sqlmock) {
				mock.ExpectBegin().WillReturnError(fmt.Errorf("DB connection error"))
			},
			wantError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			defer func() {
				if r := recover(); r != nil {
					_, file, line, _ := runtime.Caller(6)
					t.Logf("Panic encountered while running test: %s. %v at %s:%d", tt.name, r, filepath.Base(file), line)
					t.Fail()
				}
			}()

			rawDB, mock, err := sqlmock.New()
			if err != nil {
				t.Fatalf("Error while initializing sqlmock. Error: %v", err)
			}
			defer rawDB.Close()

			gormDB, err := gorm.Open("postgres", rawDB)
			if err != nil {
				t.Fatalf("Error while opening gorm DB. Error: %v", err)
			}
			defer gormDB.Close()

			tt.mockFunc(mock)

			artcStore := &MockArticleStore{
				db: gormDB,
			}

			err = artcStore.Create(&tt.article)

			if tt.wantError {
				if err == nil {
					t.Errorf("Expected error but didn't receive any")
				}
				t.Logf("Expected error received: %v", err)
			} else {
				if err != nil {
					t.Errorf("Didn't expect error but received: %v", err)
				}
			}
		})
	}
}

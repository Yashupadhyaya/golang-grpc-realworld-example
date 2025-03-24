package store

import (
	fmt "fmt"
	os "os"
	testing "testing"
	gosqlmock "github.com/DATA-DOG/go-sqlmock"
	gorm "github.com/jinzhu/gorm"
	assert "github.com/stretchr/testify/assert"
	model "github.com/raahii/golang-grpc-realworld-example/model"
)








/*
ROOST_METHOD_HASH=ArticleStore_Create_1273475ade
ROOST_METHOD_SIG_HASH=ArticleStore_Create_a27282cad5

FUNCTION_DEF=func (s *ArticleStore) Create(m *model.Article) error // Create creates an article


*/
func TestArticleStoreCreate(t *testing.T) {
	var testCases = []struct {
		name          string
		input         *model.Article
		mockBehaviour func(mock sqlmock.Sqlmock, input *model.Article)
		expectedError bool
	}{
		{
			name:  "Successful article creation",
			input: &model.Article{},
			mockBehaviour: func(mock sqlmock.Sqlmock, input *model.Article) {

			},
			expectedError: false,
		},
		{
			name:  "Article creation with invalid data",
			input: &model.Article{},
			mockBehaviour: func(mock sqlmock.Sqlmock, input *model.Article) {

			},
			expectedError: true,
		},
		{
			name:  "Article creation with database errors",
			input: &model.Article{},
			mockBehaviour: func(mock sqlmock.Sqlmock, input *model.Article) {

			},
			expectedError: true,
		},
	}

	for _, test := range testCases {
		t.Run(test.name, func(t *testing.T) {
			defer func() {
				if r := recover(); r != nil {

					t.Logf("Panic happened in test scenario: %s with error: %v", test.name, r)
					t.Fail()
				}
			}()

			db, mock, _ := sqlmock.New()
			gormDB, _ := gorm.Open("postgres", db)
			defer db.Close()

			test.mockBehaviour(mock, test.input)

			s := ArticleStore{

				db: gormDB,
			}

			err := s.Create(test.input)

			if test.expectedError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}


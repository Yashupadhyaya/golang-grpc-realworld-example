package store

import (
	fmt "fmt"
	testing "testing"

	sqlmock "github.com/DATA-DOG/go-sqlmock"
	gorm "github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/postgres"
	model "github.com/raahii/golang-grpc-realworld-example/model"
	require "github.com/stretchr/testify/require"
	suite "github.com/stretchr/testify/suite"
)

type Suite struct {
	suite.Suite
	DB   *gorm.DB
	mock sqlmock.Sqlmock

	articleStore *ArticleStore
	article      *model.Article
	userStore    *UserStore
}

/*
ROOST_METHOD_HASH=ArticleStore_DeleteFavorite_29c18a04a8
ROOST_METHOD_SIG_HASH=ArticleStore_DeleteFavorite_53deb5e792

FUNCTION_DEF=func (s *ArticleStore) DeleteFavorite(a *model.Article, u *model.User) error // DeleteFavorite unfavorite an article
*/
func (s *Suite) SetupSuite() {
	var (
		db  *gorm.DB
		err error
	)

	rawDB, _, err := sqlmock.New()
	if err != nil {
		s.T().Fatal(err)
	}

	db, err = gorm.Open("_", rawDB)
	if err != nil {
		s.T().Fatal(err)
	}

	s.DB = db
	s.articleStore = &ArticleStore{
		db: s.DB,
	}

	s.article = &model.Article{
		FavoritesCount: 2,
	}
	s.userStore = &UserStore{
		db: s.DB,
	}
}

func (s *Suite) TearDownSuite() {
	s.DB.Close()
}

func (s *Suite) TestArticleStoreDeleteFavorite() {
	t := s.T()

	{
		t.Log("Scenario 1: Successful Deletion of User from Article's 'FavoritedUsers' and Decrease of 'favorites_count' ")
		s.mock.ExpectBegin()
		s.mock.ExpectExec("UPDATE").WillReturnResult(sqlmock.NewResult(1, 1))
		s.mock.ExpectExec("UPDATE").WillReturnResult(sqlmock.NewResult(1, 1))
		s.mock.ExpectCommit()

		user := &model.User{Username: "user1"}
		err := s.articleStore.DeleteFavorite(s.article, user)
		require.Nil(t, err)
		require.Equal(t, s.article.FavoritesCount, 1)
	}

	{
		t.Log("Scenario 2: Failure due to User not in 'FavoritedUsers' of Article")
		s.mock.ExpectBegin()
		s.mock.ExpectExec("UPDATE").WillReturnError(fmt.Errorf("user not found"))
		s.mock.ExpectRollback()

		user := &model.User{Username: "user2"}
		err := s.articleStore.DeleteFavorite(s.article, user)
		require.NotNil(t, err)
	}

	{
		t.Log("Scenario 3: Rollback due to second transaction error")
		s.mock.ExpectBegin()
		s.mock.ExpectExec("UPDATE").WillReturnResult(sqlmock.NewResult(1, 1))
		s.mock.ExpectExec("UPDATE").WillReturnError(fmt.Errorf("update count failed"))
		s.mock.ExpectRollback()

		user := &model.User{Username: "user3"}
		err := s.articleStore.DeleteFavorite(s.article, user)
		require.NotNil(t, err)
	}
}

func TestArticleStoreDeleteFavorite(t *testing.T) {
	suite.Run(t, new(Suite))
}

package account

import (
	"testing"
	"time"

	"github.com/stretchr/testify/suite"
	"gorm.io/gorm"

	"github.com/bnb-chain/zkbnb/common/test"
)

type Suite struct {
	suite.Suite
	dao AccountModel
	db  *test.Database
}

func TestSuite(t *testing.T) {
	suite.Run(t, new(Suite))
}

func (s *Suite) SetupSuite() {
	dbName := "zkbnb"
	db, err := test.RunDB(dbName)
	s.Require().NoError(err)
	s.db = db
	s.dao = NewAccountModel(db.DB)
}

func (s *Suite) TearDownSuite() {
	err := s.db.StopDB()
	s.Require().NoError(err)
}

func (s *Suite) SetupTest() {
	err := s.db.ClearDB([]string{AccountTableName})
	s.Require().NoError(err)
	err = s.db.InitDB()
	s.Require().NoError(err)
}

func (s *Suite) TearDownTest() {
	err := s.db.ClearDB([]string{AccountTableName})
	s.Require().NoError(err)
}

func (s *Suite) TestGetAccount() {
	item := Account{
		Model: gorm.Model{
			CreatedAt: time.Unix(10, 0),
			UpdatedAt: time.Unix(10, 0),
		},
		AccountIndex:    1,
		AccountName:     "name",
		AccountNameHash: "hash",
		PublicKey:       "key",
	}
	s.Require().NoError(s.dao.UpdateAccountsInTransact(s.db.DB, []*Account{&item}))

	account, err := s.dao.GetAccountByIndex(item.AccountIndex)
	s.Require().NoError(err)
	s.Equal(int64(1), account.AccountIndex)

	account, err = s.dao.GetAccountByPk(item.PublicKey)
	s.Require().NoError(err)
	s.Equal("key", account.PublicKey)

	account, err = s.dao.GetAccountByName(item.AccountName)
	s.Require().NoError(err)
	s.Equal("name", account.AccountName)

	account, err = s.dao.GetAccountByNameHash(item.AccountNameHash)
	s.Require().NoError(err)
	s.Equal("hash", account.AccountNameHash)
}

func (s *Suite) TestUpdateTxsInTransact() {
	item := Account{
		Model: gorm.Model{
			CreatedAt: time.Unix(10, 0),
			UpdatedAt: time.Unix(10, 0),
		},
		AccountIndex:    1,
		AccountName:     "name",
		AccountNameHash: "hash",
		PublicKey:       "key",
	}

	s.Require().NoError(s.dao.UpdateAccountsInTransact(s.db.DB, []*Account{&item}))

	item.Status = AccountStatusConfirmed
	err := s.dao.UpdateAccountsInTransact(s.db.DB, []*Account{&item})
	s.Require().NoError(err)

	itemRes := &Account{}
	dbtx := s.db.DB.Where("1=1").Take(itemRes)
	s.Require().NoError(dbtx.Error)
	s.Equal(AccountStatusConfirmed, itemRes.Status)
	s.Greater(itemRes.UpdatedAt.Unix(), int64(10))
}

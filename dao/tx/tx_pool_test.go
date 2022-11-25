package tx

import (
	"testing"
	"time"

	"github.com/stretchr/testify/suite"
	"gorm.io/gorm"

	"github.com/bnb-chain/zkbnb/common/test"
)

type PoolSuite struct {
	suite.Suite
	dao TxPoolModel
	db  *test.Database
}

func TestPoolSuite(t *testing.T) {
	suite.Run(t, new(PoolSuite))
}

func (s *PoolSuite) SetupSuite() {
	dbName := "zkbnb"
	db, err := test.RunDB(dbName)
	s.Require().NoError(err)
	s.db = db
	s.dao = NewTxPoolModel(db.DB)
}

func (s *PoolSuite) TearDownSuite() {
	err := s.db.StopDB()
	s.Require().NoError(err)
}

func (s *PoolSuite) SetupTest() {
	err := s.db.ClearDB([]string{PoolTxTableName})
	s.Require().NoError(err)
	err = s.db.InitDB()
	s.Require().NoError(err)
}

func (s *PoolSuite) TearDownTest() {
	err := s.db.ClearDB([]string{PoolTxTableName})
	s.Require().NoError(err)
}

func (s *PoolSuite) TestGetTxsByStatus() {
	request := Tx{
		Model: gorm.Model{
			CreatedAt: time.Unix(10, 0),
			UpdatedAt: time.Unix(10, 0),
		},
		TxStatus: StatusPending,
	}

	s.Require().NoError(s.dao.CreateTxs([]*Tx{&request}))

	txs, err := s.dao.GetTxsByStatus(StatusPending)
	s.Require().NoError(err)
	s.Require().Len(txs, 1)
}

func (s *PoolSuite) TestUpdateTxsInTransact() {
	request := Tx{
		Model: gorm.Model{
			CreatedAt: time.Unix(10, 0),
			UpdatedAt: time.Unix(10, 0),
		},
	}

	s.Require().NoError(s.dao.CreateTxs([]*Tx{&request}))

	request.AccountIndex = 10
	err := s.dao.UpdateTxsInTransact(s.db.DB, []*Tx{&request})
	s.Require().NoError(err)

	requestRes := &PoolTx{}
	dbtx := s.db.DB.Where("1=1").Take(requestRes)
	s.Require().NoError(dbtx.Error)
	s.Greater(requestRes.UpdatedAt.Unix(), int64(10))
}

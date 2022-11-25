package tx

import (
	"testing"
	"time"

	"github.com/stretchr/testify/suite"
	"gorm.io/gorm"

	"github.com/bnb-chain/zkbnb/common/test"
)

type Suite struct {
	suite.Suite
	dao TxModel
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
	s.dao = NewTxModel(db.DB)
}

func (s *Suite) TearDownSuite() {
	err := s.db.StopDB()
	s.Require().NoError(err)
}

func (s *Suite) SetupTest() {
	err := s.db.ClearDB([]string{TxTableName})
	s.Require().NoError(err)
	err = s.db.InitDB()
	s.Require().NoError(err)
}

func (s *Suite) TearDownTest() {
	err := s.db.ClearDB([]string{TxTableName})
	s.Require().NoError(err)
}

func (s *Suite) TestUpdateTxsStatusInTransact() {
	item := Tx{
		Model: gorm.Model{
			CreatedAt: time.Unix(10, 0),
			UpdatedAt: time.Unix(10, 0),
		},
		TxStatus:    StatusPacked,
		BlockHeight: 100,
	}

	dbTx := s.db.DB.Create([]*Tx{&item})
	s.Require().NoError(dbTx.Error)

	blockTxStatus := map[int64]int{100: StatusCommitted}
	err := s.dao.UpdateTxsStatusInTransact(s.db.DB, blockTxStatus)
	s.Require().NoError(err)

	itemRes := &Tx{}
	dbtx := s.db.DB.Where("1=1").Take(itemRes)
	s.Require().NoError(dbtx.Error)
	s.Greater(itemRes.UpdatedAt.Unix(), int64(10))
	s.Equal(StatusCommitted, itemRes.TxStatus)
}

package block

import (
	"testing"
	"time"

	"github.com/stretchr/testify/suite"
	"gorm.io/gorm"

	"github.com/bnb-chain/zkbnb/common/test"
	"github.com/bnb-chain/zkbnb/dao/tx"
)

type Suite struct {
	suite.Suite
	dao BlockModel
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
	s.dao = NewBlockModel(db.DB)
}

func (s *Suite) TearDownSuite() {
	err := s.db.StopDB()
	s.Require().NoError(err)
}

func (s *Suite) SetupTest() {
	err := s.db.ClearDB([]string{BlockTableName, tx.TxTableName, tx.TxDetailTableName})
	s.Require().NoError(err)
	err = s.db.InitDB()
	s.Require().NoError(err)
}

func (s *Suite) TearDownTest() {
	err := s.db.ClearDB([]string{BlockTableName, tx.TxTableName, tx.TxDetailTableName})
	s.Require().NoError(err)
}

func (s *Suite) TestUpdateBlockInTransact() {
	item := Block{
		Model: gorm.Model{
			CreatedAt: time.Unix(10, 0),
			UpdatedAt: time.Unix(10, 0),
		},
		BlockHeight: 100,
		BlockStatus: StatusCommitted,
	}

	s.Require().NoError(s.dao.CreateBlockInTransact(s.db.DB, &item))

	item.Txs = []*tx.Tx{
		{BaseTx: tx.BaseTx{
			TxHash:  "hash1",
			BlockId: item.ID,
			TxDetails: []*tx.TxDetail{
				{
					AccountIndex: 1,
				},
				{
					AccountIndex: 2,
				},
			},
		}},
		{BaseTx: tx.BaseTx{
			TxHash:  "hash2",
			BlockId: item.ID,
			TxDetails: []*tx.TxDetail{
				{
					AccountIndex: 3,
				},
				{
					AccountIndex: 4,
				},
			},
		}},
	}

	err := s.dao.UpdateBlockInTransact(s.db.DB, &item)
	s.Require().NoError(err)

	items, err := s.dao.GetPendingBlocksBetween(item.BlockHeight, item.BlockHeight)
	s.Require().NoError(err)
	itemRes := items[0]
	s.Greater(itemRes.UpdatedAt.Unix(), int64(10))
	s.Require().Len(itemRes.Txs, 2)
	hashes := []string{}
	for i := 0; i < len(itemRes.Txs); i++ {
		hashes = append(hashes, itemRes.Txs[i].TxHash)
	}
	s.ElementsMatch([]string{"hash1", "hash2"}, hashes)
	accountIndexes := []int64{}
	for _, d := range itemRes.Txs[0].TxDetails {
		accountIndexes = append(accountIndexes, d.AccountIndex)
	}
	s.ElementsMatch([]int64{1, 2}, accountIndexes)
	accountIndexes = []int64{}
	for _, d := range itemRes.Txs[1].TxDetails {
		accountIndexes = append(accountIndexes, d.AccountIndex)
	}
	s.ElementsMatch([]int64{3, 4}, accountIndexes)
}

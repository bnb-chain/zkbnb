package nft

import (
	"testing"
	"time"

	"github.com/stretchr/testify/suite"
	"gorm.io/gorm"

	"github.com/bnb-chain/zkbnb/common/test"
)

type Suite struct {
	suite.Suite
	dao L2NftModel
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
	s.dao = NewL2NftModel(db.DB)
}

func (s *Suite) TearDownSuite() {
	err := s.db.StopDB()
	s.Require().NoError(err)
}

func (s *Suite) SetupTest() {
	err := s.db.ClearDB([]string{L2NftTableName})
	s.Require().NoError(err)
	err = s.db.InitDB()
	s.Require().NoError(err)
}

func (s *Suite) TearDownTest() {
	err := s.db.ClearDB([]string{L2NftTableName})
	s.Require().NoError(err)
}

func (s *Suite) TestUpdateNftsInTransact() {
	item := L2Nft{
		Model: gorm.Model{
			CreatedAt: time.Unix(10, 0),
			UpdatedAt: time.Unix(10, 0),
		},
		NftIndex: 1,
	}

	s.Require().NoError(s.dao.UpdateNftsInTransact(s.db.DB, []*L2Nft{&item}))

	item.OwnerAccountIndex = 1
	err := s.dao.UpdateNftsInTransact(s.db.DB, []*L2Nft{&item})
	s.Require().NoError(err)

	itemRes := &L2Nft{}
	dbtx := s.db.DB.Where("1=1").Take(itemRes)
	s.Require().NoError(dbtx.Error)
	s.Equal(int64(1), itemRes.OwnerAccountIndex)
	s.Greater(itemRes.UpdatedAt.Unix(), int64(10))
}

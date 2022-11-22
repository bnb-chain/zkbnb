package nft

import (
	"fmt"
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
	dbTx := s.db.DB.Exec(`
			DROP PROCEDURE IF EXISTS test(IN creator_account_index INT, IN owner_account_index INT, IN nft_content_hash VARCHAR, IN creator_treasury_rate INT, IN collection_id INT, IN updated_at TIMESTAMP, IN p_nft_index INT, INOUT row_num INT);
		`)
	s.Require().NoError(dbTx.Error)
	dbTx = s.db.DB.Exec(`
	CREATE OR REPLACE PROCEDURE test(IN p_creator_account_index INT, IN p_owner_account_index INT, IN p_nft_content_hash VARCHAR, IN p_creator_treasury_rate INT, IN p_collection_id INT, IN p_updated_at TIMESTAMP, IN p_nft_index INT, INOUT row_num INT)
	LANGUAGE plpgsql
AS $$
BEGIN
	UPDATE l2_nft SET creator_account_index=p_creator_account_index, owner_account_index=p_owner_account_index, nft_content_hash=p_nft_content_hash,
		creator_treasury_rate=p_creator_treasury_rate, collection_id=p_collection_id,
		updated_at=p_updated_at
		WHERE nft_index=p_nft_index;
	GET DIAGNOSTICS row_num = ROW_COUNT;
END;
$$;
`)
	s.Require().NoError(dbTx.Error)
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
	start := time.Now().UnixNano()
	err := s.dao.UpdateNftsInTransact(s.db.DB, []*L2Nft{&item})
	end := time.Now().UnixNano()
	fmt.Printf("%d\n", end-start)
	s.Require().NoError(err)

	itemRes := &L2Nft{}
	dbtx := s.db.DB.Where("1=1").Take(itemRes)
	s.Require().NoError(dbtx.Error)
	s.Equal(int64(1), itemRes.OwnerAccountIndex)
	s.Greater(itemRes.UpdatedAt.Unix(), int64(10))
}

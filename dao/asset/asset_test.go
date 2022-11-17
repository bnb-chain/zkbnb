package asset

import (
	"testing"
	"time"

	"github.com/stretchr/testify/suite"
	"gorm.io/gorm"

	"github.com/bnb-chain/zkbnb/common/test"
)

type Suite struct {
	suite.Suite
	dao AssetModel
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
	s.dao = NewAssetModel(db.DB)
}

func (s *Suite) TearDownSuite() {
	err := s.db.StopDB()
	s.Require().NoError(err)
}

func (s *Suite) SetupTest() {
	err := s.db.ClearDB([]string{TableName})
	s.Require().NoError(err)
	err = s.db.InitDB()
	s.Require().NoError(err)
}

func (s *Suite) TearDownTest() {
	err := s.db.ClearDB([]string{TableName})
	s.Require().NoError(err)
}

func (s *Suite) TestDeleteAssetsInTransact() {
	a := &Asset{
		Model: gorm.Model{
			CreatedAt: time.Unix(10, 0),
			UpdatedAt: time.Unix(10, 0),
		},
		AssetId:     1,
		AssetName:   "BNB",
		AssetSymbol: "BNB",
		L1Address:   "0x1",
		Decimals:    10,
		Status:      StatusActive,
		IsGasAsset:  1,
	}

	s.Require().NoError(s.dao.CreateAssetsInTransact(s.db.DB, []*Asset{a}))

	err := s.dao.DeleteAssetsInTransact(s.db.DB, []*Asset{a})
	s.Require().NoError(err)

	res := []*Asset{}
	dbtx := s.db.DB.Unscoped().Where("1=1").Find(&res)
	s.Require().NoError(dbtx.Error)
	s.Require().Len(res, 1)
	s.Greater(res[0].DeletedAt.Time.Unix(), int64(0))
}

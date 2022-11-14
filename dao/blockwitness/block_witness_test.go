package blockwitness

import (
	"testing"
	"time"

	"github.com/stretchr/testify/suite"
	"gorm.io/gorm"

	"github.com/bnb-chain/zkbnb/common/test"
)

type blockWitnessSuite struct {
	suite.Suite
	dao BlockWitnessModel
	db  *test.Database
}

func TestBlockWitnessSuite(t *testing.T) {
	suite.Run(t, new(blockWitnessSuite))
}

func (s *blockWitnessSuite) SetupSuite() {
	dbName := "zkbnb"
	db, err := test.RunDB(dbName)
	s.Require().NoError(err)
	s.db = db
	s.dao = NewBlockWitnessModel(db.DB)
}

func (s *blockWitnessSuite) TearDownSuite() {
	err := s.db.StopDB()
	s.Require().NoError(err)
}

func (s *blockWitnessSuite) SetupTest() {
	err := s.db.ClearDB([]string{TableName})
	s.Require().NoError(err)
	err = s.db.InitDB()
	s.Require().NoError(err)
}

func (s *blockWitnessSuite) TearDownTest() {
	err := s.db.ClearDB([]string{TableName})
	s.Require().NoError(err)
}

func (s *blockWitnessSuite) TestUpdateBlockWitnessStatus() {
	witness := &BlockWitness{
		Model: gorm.Model{
			CreatedAt: time.Unix(10, 0),
			UpdatedAt: time.Unix(10, 0),
		},
		Height:      1,
		WitnessData: "mock",
		Status:      StatusPublished,
	}
	s.Require().NoError(s.dao.CreateBlockWitness(witness))

	witness, err := s.dao.GetBlockWitnessByHeight(1)
	s.Require().NoError(err)
	s.Require().Equal(int64(10), witness.UpdatedAt.Unix())

	err = s.dao.UpdateBlockWitnessStatus(witness, StatusReceived)
	s.Require().NoError(err)

	witnessRes, err := s.dao.GetBlockWitnessByHeight(1)
	s.Require().NoError(err)
	s.Equal(int64(StatusReceived), witnessRes.Status)
	// Check updated_at updated
	s.Greater(witnessRes.UpdatedAt.Unix(), int64(10))
}

package proof

import (
	"testing"
	"time"

	"github.com/stretchr/testify/suite"
	"gorm.io/gorm"

	"github.com/bnb-chain/zkbnb/common/test"
)

type Suite struct {
	suite.Suite
	dao ProofModel
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
	s.dao = NewProofModel(db.DB)
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

func (s *Suite) TestUpdateProofsInTransact() {
	p := &Proof{
		Model: gorm.Model{
			CreatedAt: time.Unix(10, 0),
			UpdatedAt: time.Unix(10, 0),
		},
		ProofInfo:   "",
		BlockNumber: 1,
		Status:      NotSent,
	}

	s.Require().NoError(s.dao.CreateProof(p))

	m := map[int64]int{1: NotConfirmed}
	err := s.dao.UpdateProofsInTransact(s.db.DB, m)
	s.Require().NoError(err)

	pRes := []*Proof{}
	dbtx := s.db.DB.Where("1=1").Find(&pRes)
	s.Require().NoError(dbtx.Error)
	s.Require().Len(pRes, 1)
	s.Equal(int64(NotConfirmed), pRes[0].Status)
	s.Greater(pRes[0].UpdatedAt.Unix(), int64(10))
}

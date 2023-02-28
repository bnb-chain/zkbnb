package priorityrequest

import (
	"testing"
	"time"

	"github.com/stretchr/testify/suite"
	"gorm.io/gorm"

	"github.com/bnb-chain/zkbnb/common/test"
)

type Suite struct {
	suite.Suite
	dao PriorityRequestModel
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
	s.dao = NewPriorityRequestModel(db.DB)
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

func (s *Suite) TestUpdateHandledPriorityRequestsInTransact() {
	request := &PriorityRequest{
		Model: gorm.Model{
			CreatedAt: time.Unix(10, 0),
			UpdatedAt: time.Unix(10, 0),
		},
		L2TxHash:      "",
		L1BlockHeight: 1,
		Status:        PendingStatus,
	}

	s.Require().NoError(s.dao.CreatePriorityRequestsInTransact(s.db.DB, []*PriorityRequest{request}))

	request.L2TxHash = "hash"
	err := s.dao.UpdateHandledPriorityRequestsInTransact(s.db.DB, []*PriorityRequest{request})
	s.Require().NoError(err)

	requestRes := &PriorityRequest{}
	dbtx := s.db.DB.Where("1=1").Take(requestRes)
	s.Require().NoError(dbtx.Error)
	s.Equal(HandledStatus, requestRes.Status)
	s.Equal("hash", requestRes.L2TxHash)
	s.Greater(requestRes.UpdatedAt.Unix(), int64(10))
}

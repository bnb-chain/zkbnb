package sysconfig

import (
	"testing"
	"time"

	"github.com/stretchr/testify/suite"
	"gorm.io/gorm"

	"github.com/bnb-chain/zkbnb/common/test"
)

type Suite struct {
	suite.Suite
	dao SysConfigModel
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
	s.dao = NewSysConfigModel(db.DB)
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
	config := &SysConfig{
		Model: gorm.Model{
			CreatedAt: time.Unix(10, 0),
			UpdatedAt: time.Unix(10, 0),
		},
		Name:  "color",
		Value: "red",
	}

	rowNum, err := s.dao.CreateSysConfigs([]*SysConfig{config})
	s.Require().NoError(err)
	s.Require().Equal(int64(1), rowNum)

	config.Value = "green"
	err = s.dao.UpdateSysConfigsInTransact(s.db.DB, []*SysConfig{config})
	s.Require().NoError(err)

	configRes, err := s.dao.GetSysConfigByName("color")
	s.Equal("green", configRes.Value)
	s.Greater(configRes.UpdatedAt.Unix(), int64(10))
}

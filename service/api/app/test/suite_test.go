package test

import (
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/suite"
	"github.com/zeromicro/go-zero/core/conf"
	"github.com/zeromicro/go-zero/core/logx"
	"github.com/zeromicro/go-zero/rest"

	"github.com/bnb-chain/zkbas/service/api/app/internal/config"
	"github.com/bnb-chain/zkbas/service/api/app/internal/handler"
	"github.com/bnb-chain/zkbas/service/api/app/internal/svc"
)

type AppSuite struct {
	suite.Suite
	server *rest.Server
	url    string
}

func TestAppSuite(t *testing.T) {
	suite.Run(t, new(AppSuite))

}
func (s *AppSuite) SetupSuite() {
	configFile := "app.yaml"
	var c config.Config
	conf.MustLoad(configFile, &c)
	logx.DisableStat()

	c.Port += 1000
	ctx := svc.NewServiceContext(c)

	//goland:noinspection HttpUrlsUsage
	s.url = fmt.Sprintf("http://%s:%d", c.Host, c.Port)
	//s.url = "http://172.22.41.67:8888" //use external service for test
	s.server = rest.MustNewServer(c.RestConf, rest.WithCors())

	handler.RegisterHandlers(s.server, ctx)
	logx.Infof("Starting server at %s", s.url)
	go s.server.Start()
	time.Sleep(1 * time.Second)
}
func (s *AppSuite) TearDownSuite() {
	logx.Infof("Shutting down server at %s", s.url)
	s.server.Stop()
}

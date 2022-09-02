package test

import (
	"fmt"
	"os/exec"
	"testing"
	"time"

	"github.com/stretchr/testify/suite"
	"github.com/zeromicro/go-zero/core/logx"
	"github.com/zeromicro/go-zero/core/service"
	"github.com/zeromicro/go-zero/core/stores/cache"
	"github.com/zeromicro/go-zero/core/stores/redis"
	"github.com/zeromicro/go-zero/rest"

	"github.com/bnb-chain/zkbas/service/apiserver/internal/config"
	"github.com/bnb-chain/zkbas/service/apiserver/internal/handler"
	"github.com/bnb-chain/zkbas/service/apiserver/internal/svc"
)

type ApiServerSuite struct {
	suite.Suite
	server *rest.Server
	url    string
}

func TestAppSuite(t *testing.T) {
	suite.Run(t, new(ApiServerSuite))
}

func testDBSetup() {
	testDBShutdown()
	time.Sleep(5 * time.Second)
	cmd := exec.Command("docker", "run", "--name", "postgres-ut-apiserver", "-p", "5433:5432",
		"-e", "POSTGRES_PASSWORD=Zkbas@123", "-e", "POSTGRES_USER=postgres", "-e", "POSTGRES_DB=zkbas",
		"-e", "PGDATA=/var/lib/postgresql/pgdata", "-d", "ghcr.io/bnb-chain/zkbas/zkbas-ut-postgres:0.0.2")
	if err := cmd.Run(); err != nil {
		panic(err)
	}
	time.Sleep(10 * time.Second)
}

func testDBShutdown() {
	cmd := exec.Command("docker", "kill", "postgres-ut-apiserver")
	//nolint:errcheck
	cmd.Run()
	time.Sleep(time.Second)
	cmd = exec.Command("docker", "rm", "postgres-ut")
	//nolint:errcheck
	cmd.Run()
}

func (s *ApiServerSuite) SetupSuite() {
	testDBSetup()
	c := config.Config{
		RestConf: rest.RestConf{
			Host: "127.0.0.1",
			Port: 9888,
			ServiceConf: service.ServiceConf{
				Name: "api-server",
			},
		},
		LogConf: logx.LogConf{},
	}
	c.Postgres = struct{ DataSource string }{DataSource: "host=127.0.0.1 user=postgres password=Zkbas@123 dbname=zkbas port=5433 sslmode=disable"}
	c.CacheRedis = cache.CacheConf{}
	c.CacheRedis = append(c.CacheRedis, cache.NodeConf{
		RedisConf: redis.RedisConf{Host: "127.0.0.1"},
	})
	logx.DisableStat()

	ctx := svc.NewServiceContext(c)

	s.url = fmt.Sprintf("http://127.0.0.1:%d", c.Port)
	s.server = rest.MustNewServer(c.RestConf, rest.WithCors())

	handler.RegisterHandlers(s.server, ctx)
	logx.Infof("Starting server at %s", s.url)
	go s.server.Start()
	time.Sleep(1 * time.Second)
}

func (s *ApiServerSuite) TearDownSuite() {
	logx.Infof("Shutting down server at %s", s.url)
	s.server.Stop()
	testDBShutdown()
}

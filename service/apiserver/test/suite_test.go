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

	"github.com/bnb-chain/zkbnb/service/apiserver/internal/config"
	"github.com/bnb-chain/zkbnb/service/apiserver/internal/handler"
	"github.com/bnb-chain/zkbnb/service/apiserver/internal/svc"
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
		"-e", "POSTGRES_PASSWORD=ZkBNB@123", "-e", "POSTGRES_USER=postgres", "-e", "POSTGRES_DB=zkbnb",
		"-e", "PGDATA=/var/lib/postgresql/pgdata", "-d", "ghcr.io/bnb-chain/zkbnb/zkbnb-ut-postgres:0.0.2")
	if err := cmd.Run(); err != nil {
		panic(err)
	}
	time.Sleep(5 * time.Second)
}

func testDBShutdown() {
	cmd := exec.Command("docker", "kill", "postgres-ut-apiserver")
	//nolint:errcheck
	cmd.Run()
	time.Sleep(time.Second)
	cmd = exec.Command("docker", "rm", "postgres-ut-apiserver")
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
		TxPool: struct {
			MaxPendingTxCount int
		}{
			MaxPendingTxCount: 10000,
		},
		LogConf: logx.LogConf{},
		CoinMarketCap: struct {
			Url   string
			Token string
		}{Url: "https://pro-api.coinmarketcap.com/v1/cryptocurrency/quotes/latest?symbol=",
			Token: "1731a7cc-6e1b-458c-8780-ce8249d4fd3b", //personal token, free plan
		},
		MemCache: struct {
			AccountExpiration int
			AssetExpiration   int
			BlockExpiration   int
			TxExpiration      int
			PriceExpiration   int
		}{AccountExpiration: 10000, AssetExpiration: 10000, BlockExpiration: 10000, TxExpiration: 10000, PriceExpiration: 10000},
	}
	c.Postgres = struct{ DataSource string }{DataSource: "host=127.0.0.1 user=postgres password=ZkBNB@123 dbname=zkbnb port=5433 sslmode=disable"}
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

package prometheus

import (
	"github.com/bnb-chain/zkbnb/common/metrics"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

var _ metrics.MetricsServer = (*PrometheusServer)(nil)

func NewPrometheusServer(srv *metrics.RunOnceHttpMux, addr string) metrics.MetricsServer {
	srv.Handle("/metrics", promhttp.Handler())
	return &PrometheusServer{
		srv:  srv,
		addr: addr,
	}
}

type PrometheusServer struct {
	srv  *metrics.RunOnceHttpMux
	addr string
}

func (server PrometheusServer) Start() error {
	return server.srv.ListenAndServe(server.addr)
}

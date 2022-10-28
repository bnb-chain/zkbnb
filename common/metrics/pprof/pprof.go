package pprof

import (
	"net/http/pprof"

	"github.com/bnb-chain/zkbnb/common/metrics"
)

var _ metrics.MetricsServer = (*PProfServer)(nil)

func NewPProfServer(srv *metrics.RunOnceHttpMux, addr string) metrics.MetricsServer {
	srv.Handle("/debug/pprof/goroutine", pprof.Handler("goroutine"))
	srv.Handle("/debug/pprof/threadcreate", pprof.Handler("threadcreate"))
	srv.Handle("/debug/pprof/heap", pprof.Handler("heap"))
	srv.Handle("/debug/pprof/allocs", pprof.Handler("allocs"))
	srv.Handle("/debug/pprof/block", pprof.Handler("block"))
	srv.Handle("/debug/pprof/mutex", pprof.Handler("mutex"))
	return &PProfServer{
		srv:  srv,
		addr: addr,
	}
}

type PProfServer struct {
	srv  *metrics.RunOnceHttpMux
	addr string
}

func (server *PProfServer) Start() error {
	return server.srv.ListenAndServe(server.addr)
}

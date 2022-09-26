package metrics

import (
	"net/http"
	"sync"
)

func NewRunOnceHttpMux(mux *http.ServeMux) *RunOnceHttpMux {
	return &RunOnceHttpMux{
		ServeMux: mux,
		once:     &sync.Once{},
	}
}

type RunOnceHttpMux struct {
	*http.ServeMux
	once *sync.Once
}

func (mux *RunOnceHttpMux) ListenAndServe(addr string) error {
	var err error
	mux.once.Do(func() {
		err = http.ListenAndServe(addr, mux.ServeMux)
	})

	return err
}

type MetricsServer interface {
	Start() error
}

package main

import (
	promsrv "github.com/lab259/go-rscsrv-prometheus"
	"github.com/prometheus/client_golang/prometheus"
)

type PromService struct {
	promsrv.Service
	HelloSent prometheus.Counter
	WorldSent prometheus.Counter
}

// Name implements the rscsrv.Service interface.
func (srv *PromService) Name() string {
	return "Prometheus Service"
}

// Start implements the rscsrv.Startable interface.
func (srv *PromService) Start() error {
	srv.HelloSent = srv.NewCounter(prometheus.CounterOpts{
		Name: "hello_sent",
		Help: "Number of hellos what were sent",
	})
	srv.WorldSent = srv.NewCounter(prometheus.CounterOpts{
		Name: "world_sent",
		Help: "Number of worlds what were sent",
	})
	return nil
}

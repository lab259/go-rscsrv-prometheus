package services

import (
	"github.com/lab259/go-rscsrv-prometheus/promsql"
	_ "github.com/lib/pq"
)

var DefaultDriverCollector DriverCollectorService

type DriverCollectorService struct {
	promsql.DriverCollector
}

// Name implements the rscsrv.Service interface.
func (srv *DriverCollectorService) Name() string {
	return "Driver Collector Service"
}

func (service *DriverCollectorService) Start() error {
	collector, err := promsql.Register(promsql.DriverCollectorOpts{
		DriverName: "postgres",
	})
	if err != nil {
		return err
	}
	service.DriverCollector = *collector

	return DefaultPromService.Register(&service.DriverCollector)
}

// Restart restarts the Prometheus service.
func (service *DriverCollectorService) Restart() error {
	if err := service.Stop(); err != nil {
		return err
	}
	return service.Start()
}

// Stop stops the Prometheus service.
func (service *DriverCollectorService) Stop() error {
	return nil
}

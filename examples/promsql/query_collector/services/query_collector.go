package services

import (
	"github.com/lab259/go-rscsrv-prometheus/promquery"
	_ "github.com/lib/pq"
)

var DefaultQueryCollectorService QueryCollectorService

type QueryCollectorService struct {
	*promquery.QueryCollector
}

// Name implements the rscsrv.Service interface.
func (srv *QueryCollectorService) Name() string {
	return "Query Collector Service"
}

func (service *QueryCollectorService) Start() error {
	service.QueryCollector = promquery.NewQueryCollector(&promquery.QueryCollectorOpts{
		Prefix: "query_collector",
	})

	return DefaultPromService.Register(service.QueryCollector)
}

// Restart restarts the Prometheus service.
func (service *QueryCollectorService) Restart() error {
	if err := service.Stop(); err != nil {
		return err
	}
	return service.Start()
}

// Stop stops the Prometheus service.
func (service *QueryCollectorService) Stop() error {
	return nil
}

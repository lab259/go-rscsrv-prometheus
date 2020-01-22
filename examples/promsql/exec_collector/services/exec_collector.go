package services

import (
	"github.com/lab259/go-rscsrv-prometheus/promexec"
	_ "github.com/lib/pq"
)

var DefaultExecCollectorService ExecCollectorService

type ExecCollectorService struct {
	*promexec.ExecCollector
}

// Name implements the rscsrv.Service interface.
func (srv *ExecCollectorService) Name() string {
	return "Exec Collector Service"
}

func (service *ExecCollectorService) Start() error {
	service.ExecCollector = promexec.NewExecCollector(&promexec.ExecCollectorOpts{
		Prefix: "exec_collector",
	})

	return DefaultPromService.Register(service.ExecCollector)
}

// Restart restarts the Prometheus service.
func (service *ExecCollectorService) Restart() error {
	if err := service.Stop(); err != nil {
		return err
	}
	return service.Start()
}

// Stop stops the Prometheus service.
func (service *ExecCollectorService) Stop() error {
	return nil
}

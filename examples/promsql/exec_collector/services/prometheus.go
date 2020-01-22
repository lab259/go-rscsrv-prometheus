package services

import (
	promsrv "github.com/lab259/go-rscsrv-prometheus"
)

var DefaultPromService PromService

type PromService struct {
	promsrv.Service
}

// Name implements the rscsrv.Service interface.
func (srv *PromService) Name() string {
	return "Prometheus Service"
}

func (service *PromService) Start() error {
	return nil
}

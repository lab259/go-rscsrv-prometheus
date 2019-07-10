package main

import (
	"fmt"

	"github.com/lab259/go-rscsrv"
	"github.com/lab259/go-rscsrv-prometheus/promhermes"
	h "github.com/lab259/hermes"
	"github.com/lab259/hermes/middlewares"
)

var DefaultPromService PromService

func main() {
	serviceStarter := rscsrv.DefaultServiceStarter(&DefaultPromService)

	if err := serviceStarter.Start(); err != nil {
		panic(err)
	}

	router := h.DefaultRouter()
	router.Use(middlewares.RecoverableMiddleware, middlewares.LoggingMiddleware)
	router.Get("/hello", hello)
	router.Get("/world", world)
	router.Get("/metrics", promhermes.Handler(&DefaultPromService))

	app := h.NewApplication(h.ApplicationConfig{
		ServiceStarter: serviceStarter,
		HTTP: h.FasthttpServiceConfiguration{
			Bind: ":3000",
		},
	}, router)

	fmt.Println("Go to http://localhost:3000/hello")
	fmt.Println("Go to http://localhost:3000/world")
	fmt.Println("Go to http://localhost:3000/metrics")

	app.Start()
}

func hello(req h.Request, res h.Response) h.Result {
	DefaultPromService.HelloSent.Inc()
	return res.Data("Hello,...")
}

func world(req h.Request, res h.Response) h.Result {
	DefaultPromService.WorldSent.Inc()
	return res.Data("... world!")
}

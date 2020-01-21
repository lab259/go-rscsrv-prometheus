package main

import (
	"database/sql"
	"fmt"
	"log"
	"time"

	"github.com/lab259/go-rscsrv"
	"github.com/lab259/go-rscsrv-prometheus/examples/promsql/driver_collector/services"
	"github.com/lab259/go-rscsrv-prometheus/promhermes"
	h "github.com/lab259/hermes"
	"github.com/lab259/hermes/middlewares"
)

func main() {
	serviceStarter := rscsrv.DefaultServiceStarter(
		&services.DefaultDriverCollector,
		&services.DefaultPromService,
	)
	if err := serviceStarter.Start(); err != nil {
		panic(err)
	}

	router := h.DefaultRouter()
	router.Use(middlewares.RecoverableMiddleware, middlewares.LoggingMiddleware)
	router.Get("/metrics", promhermes.Handler(&services.DefaultPromService))

	app := h.NewApplication(h.ApplicationConfig{
		ServiceStarter: serviceStarter,
		HTTP: h.FasthttpServiceConfiguration{
			Bind: ":3000",
		},
	}, router)

	log.Println("Go to http://localhost:3000/metrics")

	// Updating metrics
	go func() {
		psqlInfo := fmt.Sprintf("user=postgres password=postgres dbname=pg-test sslmode=disable")
		db, err := sql.Open(services.DefaultDriverCollector.DriverName, psqlInfo)
		if err != nil {
			panic(err)
		}
		defer db.Close()

		err = db.Ping()
		if err != nil {
			panic(err)
		}

		db.Exec(`DROP TABLE users`)

		_, err = db.Exec(`
		CREATE TABLE users (
			id int NOT NULL PRIMARY KEY,
			name text NOT NULL
			);
			
			INSERT INTO users (id, name)
			VALUES (1, 'john')
			`)
		if err != nil {
			panic(err)
		}

		var opt int = 1

		for {
			time.Sleep(3 * time.Second)
			switch opt {
			case 1:
				log.Println("Executing query")
				_, err := db.Query("select name from users")
				if err != nil {
					panic(err)
				}
				opt++
			case 2:
				log.Println("Executing update")
				_, err := db.Exec("update users set name = 'Charles' where id = 1")
				if err != nil {
					panic(err)
				}
				opt = 1
			default:
				panic("invalid option")
			}
		}
	}()
	app.Start()
}

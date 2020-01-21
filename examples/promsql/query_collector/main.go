package main

import (
	"database/sql"
	"fmt"
	"log"
	"math/rand"
	"time"

	"github.com/lab259/go-rscsrv"
	"github.com/lab259/go-rscsrv-prometheus/examples/promsql/query_collector/services"
	"github.com/lab259/go-rscsrv-prometheus/promhermes"
	"github.com/lab259/go-rscsrv-prometheus/promsql"
	h "github.com/lab259/hermes"
	"github.com/lab259/hermes/middlewares"
)

func main() {
	serviceStarter := rscsrv.DefaultServiceStarter(
		&services.DefaultPromService,
		&services.DefaultQueryCollectorService,
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
		db, err := sql.Open("postgres", psqlInfo)
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

		usersQuery := services.DefaultQueryCollectorService.NewNamedQuery("users")

		usersSQL := promsql.NewSQLQuery(usersQuery, db)

		for {
			time.Sleep(3 * time.Second)
			opt := rand.Intn(100)

			if opt <= 50 {
				log.Println("Executing correct query")
				_, err := usersSQL.Query("select name from users")
				if err != nil {
					panic(err)
				}

			} else {
				log.Println("Executing failing query")
				usersSQL.Query("wrong query")
			}
		}
	}()
	app.Start()
}

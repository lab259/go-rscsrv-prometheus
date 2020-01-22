package main

import (
	"database/sql"
	"fmt"
	"log"
	"math/rand"
	"time"

	"github.com/lab259/go-rscsrv"
	"github.com/lab259/go-rscsrv-prometheus/examples/promsql/exec_collector/services"
	"github.com/lab259/go-rscsrv-prometheus/promhermes"
	"github.com/lab259/go-rscsrv-prometheus/promsql"
	h "github.com/lab259/hermes"
	"github.com/lab259/hermes/middlewares"
)

func main() {
	serviceStarter := rscsrv.DefaultServiceStarter(
		&services.DefaultPromService,
		&services.DefaultExecCollectorService,
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

		insertUserExec := services.DefaultExecCollectorService.NewNamedExec("insertUser")
		deleteUserExec := services.DefaultExecCollectorService.NewNamedExec("deleteUser")

		insertUserSQL := promsql.NewSQLExec(insertUserExec, db)
		deleteUserSQL := promsql.NewSQLExec(deleteUserExec, db)

		var lastInsertedID int64 = 1
		var lastDeletedID int64 = 1
		for {
			time.Sleep(3 * time.Second)
			opt := rand.Intn(100)

			if opt <= 50 {
				log.Println("Executing insert")
				cmd := fmt.Sprintf("insert into users (id, name) values ( %d, 'John' )", lastInsertedID)
				log.Println("CMD: ", cmd)
				res, err := insertUserSQL.Exec(cmd)
				if err != nil {
					log.Println("Inserting ", lastInsertedID, " with error", err)
				}
				row, err := res.RowsAffected()
				if err != nil {
					log.Println("error recovering rows affected ", err)
				} else {
					log.Println("Row affected ", row)
				}
				lastInsertedID++
			} else {
				log.Println("Executing delete")
				cmd := fmt.Sprintf("delete from users where id = %d", lastDeletedID)
				log.Println("CMD: ", cmd)
				res, err := deleteUserSQL.Exec(cmd)
				if err != nil {
					log.Println("Deleting ", lastDeletedID, " with error", err)
				}
				row, err := res.RowsAffected()
				if err != nil {
					log.Println("error recovering rows affected ", err)
				} else {
					log.Println("Row affected ", row)
				}
				lastDeletedID++

			}
		}
	}()
	app.Start()
}

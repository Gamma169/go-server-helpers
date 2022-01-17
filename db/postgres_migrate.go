package db

import (
	"database/sql"
	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"log"
	"math/rand"
	"time"
)

func InitPostgresMigrations(dbConn *sql.DB, maxMsToWait int, isRunningLocally bool, debug bool) {
	if debug {
		log.Println("Doing Migrations")
	}

	if !isRunningLocally {
		msToWait := rand.Intn(maxMsToWait)
		if debug {
			log.Println("Waiting this many mSec before running migrations: ", msToWait)
		}
		// If we have multiple services starting up at the same time
		// we don't want the migrations to overlap
		time.Sleep(time.Duration(msToWait) * time.Millisecond)
	}

	driver, err := postgres.WithInstance(dbConn, &postgres.Config{})
	if err != nil {
		log.Println("Error:  Couldn't create migrations driver")
		panic(err)
	}

	m, err := migrate.NewWithDatabaseInstance(
		"file://./migrations/",
		"postgres", driver)
	if err != nil {
		log.Println("Error:  Couldn't run migrations")
		panic(err)
	}

	if err := m.Up(); err != nil {
		if err.Error() == "no change" {
			if debug {
				log.Println("no change")
			}
		} else {
			log.Println("Error with Migrations")
			panic(err)
		}
	}

	if debug {
		log.Println("Migrations Successful")
	}
}

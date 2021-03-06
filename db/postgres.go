package db

import (
	"database/sql"
	"fmt"
	envs "github.com/Gamma169/go-server-helpers/environments"
	_ "github.com/lib/pq"
	"log"
)

func CheckRequiredPostgresEnvs(envVarPrefix string) {
	// Not sure if I should use the getOptionalEnv function here or just os.LookupEnv
	// Because if I use getOptionalEnv and it doesn't exist, we output the logs for it twice
	// I think that's fine, but I need to think on it
	if envs.GetOptionalEnv(envVarPrefix+"DATABASE_URL", "") == "" {
		envs.GetRequiredEnv(envVarPrefix + "DATABASE_NAME")
		envs.GetRequiredEnv(envVarPrefix + "DATABASE_HOST")
		envs.GetRequiredEnv(envVarPrefix + "DATABASE_USER")
	}
}

func InitPostgres(envVarPrefix string, debug bool) (dbConn *sql.DB) {
	if debug {
		log.Println("Establishing connection with postgres database")
	}
	var err error

	if dbURL := envs.GetOptionalEnv(envVarPrefix+"DATABASE_URL", ""); dbURL != "" {
		dbConn, err = sql.Open("postgres", dbURL)
	} else {
		dbConn, err = sql.Open("postgres",
			fmt.Sprintf("user='%s' password='%s' dbname='%s' host='%s' port=%s sslmode=%s",
				envs.GetRequiredEnv(envVarPrefix+"DATABASE_USER"),
				envs.GetOptionalEnv(envVarPrefix+"DATABASE_PASSWORD", ""),
				envs.GetRequiredEnv(envVarPrefix+"DATABASE_NAME"),
				envs.GetRequiredEnv(envVarPrefix+"DATABASE_HOST"),
				envs.GetOptionalEnv(envVarPrefix+"DATABASE_PORT", "5432"),
				envs.GetOptionalEnv(envVarPrefix+"SSL_MODE", "disable")))
	}

	if err != nil {
		log.Println("Error with sql Open statement")
		panic(err)
	}

	ValidateDBConnOrPanic(dbConn, debug)
	if debug {
		log.Println("Connection sucessfully established")
	}
	return
}

func CheckDBConnection(dbConn *sql.DB, maxTries int, secondsToWait int, debug bool) error {
	return CheckAndRetry(dbConn.Ping, maxTries, secondsToWait, debug)
}

func ValidateDBConnOrPanic(dbConn *sql.DB, debug bool) {
	if err := CheckDBConnection(dbConn, 2, 3, debug); err != nil {
		log.Println("Error: Could not connect to DB")
		panic(err)
	}
}

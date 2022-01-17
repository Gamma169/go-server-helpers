package db

import (
	"database/sql"
	"errors"
	"fmt"
	envs "github.com/Gamma169/go-server-helpers/environments"
	_ "github.com/lib/pq"
	"log"
	"reflect"
	"strings"
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

// Generally this function isn't necessary because we use prepared statements, but more safety is good
func CheckStructFieldsForInjection(st interface{}) error {
	t := reflect.TypeOf(st)
	for i := 0; i < t.NumField(); i++ {
		if t.Field(i).Type.Kind() == reflect.String {
			r := reflect.ValueOf(st)
			if strings.Contains(reflect.Indirect(r).Field(i).String(), ";") {
				return errors.New("Found semicolon in string struct field")
			}
		}
	}
	return nil
}

// Function mostly from
// https://coderedirect.com/questions/432349/golang-dynamic-access-to-a-struct-property
// A function that splits a string based on a delimiter and assigns it to a slice in a struct's field
// Ex: a struct 's' with field 'MyArr', and string "val1::val2::val3"
//     This function will assign ["val1", "val2", "val3"] into s.MyArr
func AssignArrayPropertyFromString(st interface{}, field string, arrString string, delimiter string) error {
	// st must be a pointer to a struct
	refSt := reflect.ValueOf(st)
	if refSt.Kind() != reflect.Ptr || refSt.Elem().Kind() != reflect.Struct {
		return errors.New("st must be pointer to struct")
	}

	// Dereference pointer
	refSt = refSt.Elem()

	// Lookup field by name
	fieldSt := refSt.FieldByName(field)
	if !fieldSt.IsValid() {
		return fmt.Errorf("not a field name: %s", field)
	}

	// Field must be exported
	if !fieldSt.CanSet() {
		return fmt.Errorf("cannot set field %s", field)
	}

	// We expect an array field
	if fieldSt.Kind() != reflect.Slice && fieldSt.Kind() != reflect.Array {
		return fmt.Errorf("%s is not a slice or array field", field)
	}

	arr := []string{}
	if arrString != "" {
		arr = strings.Split(arrString, delimiter)
	}

	fieldSt.Set(reflect.ValueOf(arr))
	return nil
}

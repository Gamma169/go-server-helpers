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
	"time"
)

func InitDB(DB *sql.DB, debug bool) {
	if debug {
		log.Println("Establishing connection with database")
	}
	var err error

	dbURL := envs.GetOptionalEnv("DATABASE_URL", "")
	if dbURL != "" {
		DB, err = sql.Open("postgres", dbURL)
	} else {
		DB, err = sql.Open("postgres",
			fmt.Sprintf("user='%s' password='%s' dbname='%s' host='%s' port=%s sslmode=%s",
				envs.GetRequiredEnv("DATABASE_USER"),
				envs.GetOptionalEnv("DATABASE_PASSWORD", ""),
				envs.GetRequiredEnv("DATABASE_NAME"),
				envs.GetRequiredEnv("DATABASE_HOST"),
				envs.GetOptionalEnv("DATABASE_PORT", "5432"),
				envs.GetOptionalEnv("SSL_MODE", "disable")))
	}

	if err != nil {
		log.Println("Error with sql Open statement")
		panic(err)
	}

	ValidateDBConnOrPanic(DB, debug)
	if debug {
		log.Println("Connection sucessfully established")
	}
}

func CheckDBConnection(dbConn *sql.DB, maxTries int, secondsToWait int, debug bool) (err error) {
	for tries := 0; ; tries++ {
		err = dbConn.Ping()

		if err != nil {
			if tries > maxTries-1 {

				return err
			}
			if debug {
				log.Println(fmt.Sprintf("Error: Could not connect to DB -- trying again in %d seconds", secondsToWait))
			}
			time.Sleep(time.Duration(secondsToWait) * time.Second)
		}
	}
	return
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

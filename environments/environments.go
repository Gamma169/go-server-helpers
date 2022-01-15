package environments

import (
	"log"
	"os"
)

func GetRequiredEnv(envVar string) string {
	val, found := os.LookupEnv(envVar)
	if !found || val == "" {
		panic("PLEASE SET " + envVar + " ENVIRONMENT VARIABLE")
	}
	return val
}

func GetOptionalEnv(envVar string, defaultVal string) string {
	val, found := os.LookupEnv(envVar)
	if !found || val == "" {
		log.Printf("Env var: '%s' not found or empty.  Setting to default value: '%s'", envVar, defaultVal)
		return defaultVal
	}
	return val
}

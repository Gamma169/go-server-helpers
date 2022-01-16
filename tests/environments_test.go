package tests

import (
	"github.com/Gamma169/go-server-helpers/environments"
	"os"
	"testing"
)

/*********************************************
 * Helpers
 * *******************************************/

func getUnusedEnv() string {
	var randEnvKey string
	for {
		randEnvKey = randString(60)
		if _, found := os.LookupEnv(randEnvKey); !found {
			return randEnvKey
		}
	}
}

/*********************************************
 * Tests
 * *******************************************/

func TestGetOptionalEnv(t *testing.T) {

	envName := getUnusedEnv()
	defaultValue := "mock-default"
	value := environments.GetOptionalEnv(envName, defaultValue)

	assert(t, value == defaultValue, "Environment is %s ... should be default %s", value, defaultValue)

	envValue := randString(60)
	os.Setenv(envName, envValue)
	defer os.Unsetenv(envName)

	value = environments.GetOptionalEnv(envName, defaultValue)
	assert(t, value == envValue, "Environment is %s ... should be %s", value, envValue)
}

func TestGetRequiredEnvPanics(t *testing.T) {

	defer func() {
		if r := recover(); r == nil {
			t.Errorf("The GetRequiredEnvs did not panic")
		}
	}()

	envName := getUnusedEnv()
	environments.GetRequiredEnv(envName)
}

func TestRequiredEnvGetsValue(t *testing.T) {

	envName := getUnusedEnv()
	envValue := randString(60)
	os.Setenv(envName, envValue)
	defer os.Unsetenv(envName)

	value := environments.GetRequiredEnv(envName)
	assert(t, value == envValue, "Environment is %s ... should be %s", value, envValue)
}

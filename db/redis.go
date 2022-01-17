package db

import (
	"context"
	"crypto/tls"
	envs "github.com/Gamma169/go-server-helpers/environments"
	"github.com/go-redis/redis/v8"
	"log"
)

func CheckRequiredRedisEnvs(envVarPrefix string, useTLS bool) {
	envVar := "REDIS"
	if useTLS {
		envVar = "REDIS_TLS"
	}

	if redisURL := envs.GetOptionalEnv(envVarPrefix+envVar+"_URL", ""); redisURL == "" {
		envs.GetRequiredEnv(envVarPrefix + envVar + "_HOST")
	}
}

func InitRedis(envVarPrefix string, useTLS bool, debug bool) (redisClient *redis.Client) {
	if debug {
		log.Println("Establishing connection with database")
	}
	envVar := "REDIS"
	if useTLS {
		envVar = "REDIS_TLS"
	}

	var redisOptions *redis.Options
	if redisURL := envs.GetOptionalEnv(envVarPrefix+envVar+"_URL", ""); redisURL != "" {
		var err error
		if redisOptions, err = redis.ParseURL(redisURL); err != nil {
			log.Println("Error creating redis options")
			panic(err)
		}
	} else {
		redisURL := envs.GetRequiredEnv(envVarPrefix+envVar+"_HOST") + ":" + envs.GetOptionalEnv(envVarPrefix+envVar+"_PORT", "6379")
		redisPassword := envs.GetOptionalEnv(envVarPrefix+envVar+"_PASSWORD", "")
		redisUser := envs.GetOptionalEnv(envVarPrefix+envVar+"_USER", "")
		redisOptions = &redis.Options{
			Addr:     redisURL,
			Password: redisPassword,
			Username: redisUser,
		}
	}

	// TODO- possible need for heroku
	if envs.GetOptionalEnv("USE_TLS_CONFIG", "false") == "true" {
		redisOptions.TLSConfig = &tls.Config{
			InsecureSkipVerify: true,
		}
	}

	redisClient = redis.NewClient(redisOptions)

	ValidateRedisConnOrPanic(redisClient, debug)
	if debug {
		log.Println("Sucessfully established redis connection")
	}
	return
}

func CheckRedisConnection(redisClient *redis.Client, maxTries int, secondsToWait int, debug bool) error {
	wrapperFunc := func() error {
		_, err := redisClient.Ping(context.Background()).Result()
		return err
	}
	return CheckAndRetry(wrapperFunc, maxTries, secondsToWait, debug)
}

func ValidateRedisConnOrPanic(redisClient *redis.Client, debug bool) {
	if err := CheckRedisConnection(redisClient, 2, 3, debug); err != nil {
		log.Println("Error: Could not connect to Redis DB")
		panic(err)
	}
}

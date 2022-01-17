package db

import (
	"fmt"
	"log"
	"time"
)

func CheckAndRetry(checkerFunc func() error, maxTries int, secondsToWait int, debug bool) (err error) {
	for tries := 0; tries == 0 || err != nil; tries++ {
		err = checkerFunc()

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

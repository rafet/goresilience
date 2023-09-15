package main

import (
	"context"
	"log"
	"time"

	"github.com/rafet/goresilience/timeout"
)

func main() {
	// Create our runner.
	runner := timeout.New(timeout.Config{
		Timeout: 100 * time.Millisecond,
	})

	for i := 0; i < 200; i++ {
		// Execute.
		result := ""
		err := runner.Run(context.TODO(), func(_ context.Context) error {
			if time.Now().Nanosecond()%2 == 0 {
				time.Sleep(5 * time.Second)
			}
			result = "all ok"
			return nil
		})

		if err != nil {
			result = "not ok, but fallback"
		}

		log.Printf("the result is: %s", result)
	}
}

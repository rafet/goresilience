package main

import (
	"context"
	"errors"
	"fmt"

	"github.com/rafet/goresilience"
	"github.com/rafet/goresilience/bulkhead"
	"github.com/rafet/goresilience/retry"
	"github.com/rafet/goresilience/timeout"
)

func main() {
	// Create our execution chain.
	runner := goresilience.RunnerChain(
		bulkhead.NewMiddleware(bulkhead.Config{}),
		retry.NewMiddleware(retry.Config{}),
		timeout.NewMiddleware(timeout.Config{}),
	)

	// Execute.
	calledCounter := 0
	result := ""
	err := runner.Run(context.TODO(), func(_ context.Context) error {
		calledCounter++
		if calledCounter%2 == 0 {
			return errors.New("you didn't expect this error")
		}
		result = "all ok"
		return nil
	})

	if err != nil {
		result = "not ok, but fallback"
	}

	fmt.Printf("result: %s", result)
}

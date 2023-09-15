package chaos

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/rafet/goresilience"
	"github.com/rafet/goresilience/errors"
	"github.com/rafet/goresilience/metrics"
)

const (
	kindLatency = "latency"
	kindError   = "error"
)

// Injector will control how the faults will be injected in the chaos runner.
type Injector struct {
	latency      time.Duration
	errorPercent int
	mu           sync.Mutex
}

// SetLatency will set the latency on the injector.
func (i *Injector) SetLatency(t time.Duration) {
	i.mu.Lock()
	defer i.mu.Unlock()
	i.latency = t
}

// SetErrorPercent will set the error percent on the injector.
func (i *Injector) SetErrorPercent(percent int) error {
	if percent > 100 || percent < 0 {
		return fmt.Errorf("%d is not a valid percent", percent)
	}
	i.mu.Lock()
	defer i.mu.Unlock()
	i.errorPercent = percent
	return nil
}

// Config is the configuration of the chaos runner.
type Config struct {
	// Injector is the failer injector for the chaos runner.
	Injector *Injector
}

func (c *Config) defaults() {
	if c.Injector == nil {
		c.Injector = &Injector{
			latency: 100 * time.Millisecond,
		}
	}
}

type failureInjector struct {
	total  int
	errs   int
	mu     sync.Mutex
	cfg    Config
	runner goresilience.Runner
}

// New returns a new chaos runner. The chaos runner will inject failure using
// the injector. The injector controls the faults. See Injector to know what
// kind of failures are controlable.
func New(cfg Config) goresilience.Runner {
	return NewMiddleware(cfg)(nil)
}

// NewMiddleware returns a middleware that uses the Runner return
// by chaos.New.
func NewMiddleware(cfg Config) goresilience.Middleware {
	cfg.defaults()

	return func(next goresilience.Runner) goresilience.Runner {
		return &failureInjector{
			cfg:    cfg,
			runner: goresilience.SanitizeRunner(next),
		}
	}
}

func (f *failureInjector) Run(ctx context.Context, fn goresilience.Func) (err error) {
	metricsRecorder, _ := metrics.RecorderFromContext(ctx)

	// Measure the execution requests and errors.
	defer func() {
		f.mu.Lock()
		f.total++
		if err != nil {
			f.errs++
		}
		f.mu.Unlock()
	}()

	// We don't mind to lock for reading if it's stale data, eventually we will
	// get the correct values from the injector.

	// Inject latency attack.
	lat := f.cfg.Injector.latency
	if lat > 0 {
		metricsRecorder.IncChaosInjectedFailure(kindLatency)
		time.Sleep(lat)
	}

	// Inject error attack.
	var currentErrPerc int
	f.mu.Lock()
	currentErrPerc = int((float64(f.errs) / float64(f.total)) * 100)
	f.mu.Unlock()
	if currentErrPerc < f.cfg.Injector.errorPercent {
		metricsRecorder.IncChaosInjectedFailure(kindError)
		return errors.ErrFailureInjected
	}

	return f.runner.Run(ctx, fn)
}

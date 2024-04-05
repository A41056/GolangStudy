package asyncjob

import (
	"context"
	"time"
)

type Job interface {
	Execute(ctx context.Context) error
	Retry(ctx context.Context) error
	State() JobState
	SetRetryDuration(times []time.Duration)
}

const (
	defaultMaxTimeout = time.Second * 10
)

var (
	defaultRetryTimes = []time.Duration{time.Second, time.Second * 2, time.Second * 4}
)

type JobHandler func(ctx context.Context) error

type JobState int

const (
	StateInit JobState = iota
	StateRunning
	StateFailed
	StateTimeout
	StateCompleted
	StateRetryFailed
)

func (js JobState) String() string {
	return [6]string{"Init", "Running", "Failed", "Timeout", "Completed", "RetryFailed"}[js]
}

type jobConfig struct {
	Name       string
	MaxTimeout time.Duration
	Retries    []time.Duration
}

type job struct {
	config     jobConfig
	handler    JobHandler
	state      JobState
	retryIndex int
	stopChan   chan bool
}

func NewJob(handler JobHandler, options ...OptionHdl) *job {
	j := job{
		config: jobConfig{
			MaxTimeout: defaultMaxTimeout,
			Retries:    defaultRetryTimes,
		},
		handler:    handler,
		retryIndex: -1,
		state:      StateInit,
		stopChan:   make(chan bool),
	}

	for i := range options {
		options[i](&j.config)
	}

	return &j
}

func (j *job) Execute(ctx context.Context) error {
	j.state = StateRunning
	err := j.handler(ctx)
	if err != nil {
		j.state = StateFailed
		return err
	}
	j.state = StateCompleted
	return nil
}

func (j *job) Retry(ctx context.Context) error {
	if j.retryIndex == len(j.config.Retries)-1 {
		j.state = StateRetryFailed
		return nil
	}

	j.retryIndex++
	time.Sleep(j.config.Retries[j.retryIndex])

	err := j.Execute(ctx)
	if err == nil {
		j.state = StateCompleted
		return nil
	}

	if j.retryIndex == len(j.config.Retries)-1 {
		j.state = StateRetryFailed
		return err
	}

	j.state = StateFailed
	return err
}

func (j *job) State() JobState {
	return j.state
}

func (j *job) RetryIndex() int {
	return j.retryIndex
}

func (j *job) SetRetryDuration(times []time.Duration) {
	if len(times) == 0 {
		return
	}

	j.config.Retries = times
}

type OptionHdl func(*jobConfig)

func WithName(name string) OptionHdl {
	return func(c *jobConfig) {
		c.Name = name
	}
}

func WithRetriesDuration(times []time.Duration) OptionHdl {
	return func(c *jobConfig) {
		c.Retries = times
	}
}

package scheduler

import (
	"context"
	"errors"
	"math/rand"
	"time"
)

type Scheduler struct {
	jobs    map[string]*job
	running bool
}

func New() *Scheduler {
	return &Scheduler{
		jobs: map[string]*job{},
	}
}

func (this *Scheduler) Start(ctx context.Context) {
	if this.running {
		return
	}
	this.running = true

	ticker := time.NewTicker(time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			this.runJobs()
		}
	}
}

type job struct {
	lastRun time.Time
	period  time.Duration
	fn      func()
}

var ErrJobAlreadyExists = errors.New("job already exists")

func (this *Scheduler) AddJob(name string, period time.Duration, fn func()) error {
	_, alreadyExists := this.jobs[name]
	if alreadyExists {
		return ErrJobAlreadyExists
	}
	this.jobs[name] = &job{
		period: period,
		fn:     fn,

		// use a random time between now and now-period to avoid having every job run at the same time
		lastRun: time.Now().Add(time.Duration(rand.Int63n(int64(period)))),
	}
	return nil
}

func (this *Scheduler) RemoveJob(name string) {
	delete(this.jobs, name)
}

func (this *Scheduler) runJobs() {
	for _, job := range this.jobs {
		if time.Since(job.lastRun) > job.period {
			go job.fn()
			job.lastRun = time.Now()
		}
	}
}

package orchestrate

import (
	"context"
	"fmt"

	"github.com/robfig/cron/v3"
)

// Job represents a scheduled agent run.
type Job struct {
	AgentName string
	Goal      string
	CronExpr  string // e.g. "0 * * * *" for hourly
}

// Scheduler runs agents on a schedule.
type Scheduler struct {
	jobs  []Job
	runFn func(ctx context.Context, agent, goal string) error
	cron  *cron.Cron
}

// NewScheduler creates a scheduler.
func NewScheduler(runFn func(ctx context.Context, agent, goal string) error) *Scheduler {
	return &Scheduler{
		runFn: runFn,
		cron:  cron.New(cron.WithParser(cron.NewParser(cron.Minute | cron.Hour | cron.Dom | cron.Month | cron.Dow))),
	}
}

// Add adds a job.
func (s *Scheduler) Add(job Job) {
	s.jobs = append(s.jobs, job)
}

// Run starts the scheduler (blocking).
func (s *Scheduler) Run(ctx context.Context) error {
	for _, j := range s.jobs {
		job := j
		expr := job.CronExpr
		if expr == "" {
			// Backward compatible default: run every minute.
			expr = "* * * * *"
		}
		_, err := s.cron.AddFunc(expr, func() {
			if err := s.runFn(ctx, job.AgentName, job.Goal); err != nil {
				fmt.Printf("scheduler job %s: %v\n", job.AgentName, err)
			}
		})
		if err != nil {
			return fmt.Errorf("invalid cron expr for %s: %w", job.AgentName, err)
		}
	}

	s.cron.Start()
	defer s.cron.Stop()
	<-ctx.Done()
	return ctx.Err()
}

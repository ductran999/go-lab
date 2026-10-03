// Package scheduler runs named cron jobs with timeouts and graceful
// drain. A Manager holds many jobs (name → schedule + func); guards
// wrap job funcs in jobs/, not here.
package scheduler

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"github.com/robfig/cron/v3"
)

// Job is one scheduled unit: name for logs, spec cron-style, run guarded.
type Job struct {
	Name     string
	Schedule string
	Run      func(ctx context.Context) error
}

// Manager owns one cron engine per job (isolated drain per schedule).
type Manager struct {
	jobs  []Job
	crons []*cron.Cron
}

// New builds a Manager. Jobs register here, guards wrap in cmd.
func New(jobs ...Job) *Manager {
	return &Manager{jobs: jobs}
}

// StartAll registers every job with a 2s execution timeout and ticks.
func (m *Manager) StartAll(ctx context.Context) error {
	for _, j := range m.jobs {
		job := j

		c := cron.New()

		_, err := c.AddFunc(job.Schedule, func() {
			slog.Info("cron triggered", "job", job.Name)

			jobCtx, cancel := context.WithTimeout(ctx, time.Second*2)
			defer cancel()

			jobErr := job.Run(jobCtx)
			if jobErr != nil {
				slog.Error("job failed", "job", job.Name, "error_msg", jobErr.Error())

				return
			}

			slog.Info("job done", "job", job.Name)
		})
		if err != nil {
			return fmt.Errorf("schedule %s: %w", job.Name, err)
		}

		c.Start()
		m.crons = append(m.crons, c)

		slog.Info("job scheduled", "job", j.Name, "schedule", j.Schedule)
	}

	return nil
}

// StopAll halts every engine and waits for running jobs.
func (m *Manager) StopAll() {
	for _, c := range m.crons {
		ctx := c.Stop()
		<-ctx.Done()
	}

	slog.Info("all jobs stopped")
}

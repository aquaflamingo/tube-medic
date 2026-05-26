package modules

import (
	"fmt"
	"log/slog"

	"github.com/aqfl/tmleadbot/internal/config"
	"github.com/aqfl/tmleadbot/internal/core"
)

const maxRetries = 3

func ProcessJobQueue(db *core.DB, cfg *config.Config, jobTypes ...string) (int, int, error) {
	var processed, failed int
	for _, jobType := range jobTypes {
		for {
			job, err := db.DequeueJob(jobType)
			if err != nil {
				return processed, failed, fmt.Errorf("dequeue %s: %w", jobType, err)
			}
			if job == nil {
				break
			}

			slog.Info("processing job", "type", job.Type, "id", job.ID, "payload", job.Payload)
			if err := dispatchJob(db, cfg, job); err != nil {
				failed++
			} else {
				processed++
			}
		}
	}
	return processed, failed, nil
}

func dispatchJob(db *core.DB, cfg *config.Config, job *core.Job) error {
	switch job.Type {
	case "enrich_channel":
		if err := EnrichChannel(db, cfg, job.Payload); err != nil {
			if job.Retries < maxRetries {
				slog.Warn("job failed, will retry", "type", job.Type, "id", job.ID, "retries", job.Retries+1, "error", err)
				if requeueErr := db.RequeueJob(job.ID, job.Retries+1); requeueErr != nil {
					return fmt.Errorf("requeue job %d: %w (original: %v)", job.ID, requeueErr, err)
				}
				return err
			}
			slog.Error("job failed permanently", "type", job.Type, "id", job.ID, "error", err)
			if markErr := db.MarkJobFailed(job.ID, err.Error()); markErr != nil {
				return fmt.Errorf("mark job failed %d: %w (original: %v)", job.ID, markErr, err)
			}
			return err
		}
		return db.MarkJobDone(job.ID)

	default:
		err := fmt.Errorf("unknown job type: %s", job.Type)
		db.MarkJobFailed(job.ID, err.Error())
		return err
	}
}

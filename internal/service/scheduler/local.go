package scheduler

import (
	"time"

	"github.com/GianOrtiz/k8s-transparent-checkpoint-restore/internal/usecase"
)

type localScheduler struct{}

func Local() usecase.Scheduler {
	return &localScheduler{}
}

func (s *localScheduler) ScheduleCheckpoint(usecase usecase.InterceptorUseCase, scheduleIn time.Duration) error {
	ticker := time.NewTicker(scheduleIn)
	go func(ticker *time.Ticker) {
		for range ticker.C {
			usecase.Checkpoint()
		}
	}(ticker)
	return nil
}

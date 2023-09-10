package interceptor

import (
	interceptorConfig "github.com/GianOrtiz/k8s-transparent-checkpoint-restore/internal/config/interceptor"
	"github.com/GianOrtiz/k8s-transparent-checkpoint-restore/internal/delivery"
	"github.com/GianOrtiz/k8s-transparent-checkpoint-restore/internal/entity"
	"github.com/GianOrtiz/k8s-transparent-checkpoint-restore/internal/repository/interceptedrequest"
	"github.com/GianOrtiz/k8s-transparent-checkpoint-restore/internal/service/checkpoint"
	"github.com/GianOrtiz/k8s-transparent-checkpoint-restore/internal/service/scheduler"
	"github.com/GianOrtiz/k8s-transparent-checkpoint-restore/internal/service/statemanager"
	"github.com/GianOrtiz/k8s-transparent-checkpoint-restore/internal/usecase"
	"github.com/google/uuid"
)

type Interceptor struct {
	*entity.Interceptor
	useCase   usecase.InterceptorUseCase
	scheduler usecase.Scheduler
}

func New() (*Interceptor, error) {
	interceptorId := uuid.NewString()
	monitoredContainerID := uuid.NewString()

	config, err := interceptorConfig.FromEnv()
	if err != nil {
		return nil, err
	}

	interceptor := entity.Interceptor{
		ID:                    interceptorId,
		MonitoringContainerID: monitoredContainerID,
		MonitoredContainer: &entity.Container{
			ID:      monitoredContainerID,
			PID:     272389,
			HTTPUrl: config.ContainerURL.String(),
			Name:    config.ContainerName,
		},
		Config: config,
	}

	// TODO: make checkpoint service be selected via environment.
	checkpointService, err := checkpoint.CRIU(checkpoint.CRIUCheckpointServiceConfig{
		ImagesDirectory: "/home/gian/test-images",
	})
	if err != nil {
		return nil, err
	}
	// TODO: make state manager service be selected via environment.
	stateManagerService := statemanager.AlawaysAcceptingStub()
	// TODO: make intercepted request repository be selected via environment
	interceptedRequestRepository := interceptedrequest.InMemory()

	if config.Environment == interceptorConfig.KUBERNETES_ENVIRONMENT {
		interceptorUseCase, err := usecase.Interceptor(&interceptor, checkpointService, stateManagerService, interceptedRequestRepository, nil)
		if err != nil {
			return nil, err
		}
		return &Interceptor{
			useCase:     interceptorUseCase,
			Interceptor: &interceptor,
		}, nil
	} else {
		scheduler := scheduler.Local()
		interceptorUseCase, err := usecase.Interceptor(&interceptor, checkpointService, stateManagerService, interceptedRequestRepository, scheduler)
		if err != nil {
			return nil, err
		}
		return &Interceptor{
			useCase:     interceptorUseCase,
			Interceptor: &interceptor,
			scheduler:   scheduler,
		}, nil
	}
}

func (i *Interceptor) Run() error {
	if i.Config.Environment == interceptorConfig.STANDALONE_ENVIRONMENT {
		go func(interceptorUseCase usecase.InterceptorUseCase) {
			i.scheduler.ScheduleCheckpoint(interceptorUseCase, i.Config.CheckpointingInterval)
		}(i.useCase)
	}

	interceptorServer := delivery.InterceptorServer(8001, i.useCase)
	if err := interceptorServer.Run(); err != nil {
		return err
	}

	return nil
}

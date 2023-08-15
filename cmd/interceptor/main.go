package main

import (
	"net/url"
	"time"

	"github.com/GianOrtiz/k8s-transparent-checkpoint-restore/internal/config/interceptor"
	"github.com/GianOrtiz/k8s-transparent-checkpoint-restore/internal/delivery"
	"github.com/GianOrtiz/k8s-transparent-checkpoint-restore/internal/entity"
	"github.com/GianOrtiz/k8s-transparent-checkpoint-restore/internal/repository/interceptedrequest"
	"github.com/GianOrtiz/k8s-transparent-checkpoint-restore/internal/service/checkpoint"
	"github.com/GianOrtiz/k8s-transparent-checkpoint-restore/internal/service/statemanager"
	"github.com/GianOrtiz/k8s-transparent-checkpoint-restore/internal/usecase"
	"github.com/google/uuid"
)

func main() {
	monitoredContainerID := uuid.NewString()
	interceptor := entity.Interceptor{
		ID:                    uuid.NewString(),
		MonitoringContainerID: monitoredContainerID,
		MonitoredContainer: &entity.Container{
			ID:      monitoredContainerID,
			PID:     0,
			HTTPUrl: "http://localhost:8000",
			Name:    "test",
		},
		Config: &interceptor.Config{
			CheckpointingInterval: time.Duration(time.Minute * 20),
			ContainerURL:          url.URL{},
			ContainerPID:          0,
			ContainerName:         "test",
			StateManagerURL:       url.URL{},
		},
	}
	checkpointService := checkpoint.Stub()
	stateManagerService := statemanager.AlawaysAcceptingStub()
	interceptedRequestRepository := interceptedrequest.InMemory()
	interceptorUseCase, err := usecase.Interceptor(&interceptor, checkpointService, stateManagerService, interceptedRequestRepository)
	if err != nil {
		panic(err)
	}
	interceptorServer := delivery.InterceptorServer(8001, interceptorUseCase)
	interceptorServer.Run()
}

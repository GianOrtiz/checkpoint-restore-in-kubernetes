package main

import (
	"github.com/GianOrtiz/k8s-transparent-checkpoint-restore/internal/delivery"
	"github.com/GianOrtiz/k8s-transparent-checkpoint-restore/internal/entity"
	"github.com/GianOrtiz/k8s-transparent-checkpoint-restore/internal/repository/containermetadata"
	"github.com/GianOrtiz/k8s-transparent-checkpoint-restore/internal/service/restore"
	"github.com/GianOrtiz/k8s-transparent-checkpoint-restore/internal/usecase"
	"github.com/google/uuid"
)

func main() {
	containerMetadataRepository := containermetadata.InMemory()
	restoreService, err := restore.CRIU(restore.CriuRestoreServiceConfig{
		ImagesDirectory: "/home/gian/test-images",
	})
	if err != nil {
		panic(err)
	}
	stateManagerUseCase, err := usecase.StateManager(containerMetadataRepository, restoreService, &entity.Container{
		ID:      uuid.NewString(),
		PID:     1,
		HTTPUrl: "http://localhost:8000",
		Name:    "test",
	})
	if err != nil {
		panic(err)
	}
	stateManagerServer := delivery.StateManager(8002, stateManagerUseCase, entity.StateManagerConfig{DevelopmentFeaturesEnabled: true})
	stateManagerServer.Run()
}

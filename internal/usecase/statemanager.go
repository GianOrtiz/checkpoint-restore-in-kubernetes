package usecase

import "github.com/GianOrtiz/k8s-transparent-checkpoint-restore/internal/entity"

// StateManagerUseCase declares use cases for the State Manager. It declares the use cases for
// saving checkpoint images metadata and retrieving them.
type StateManagerUseCase interface {
	// SaveImageMetadata saves metadata about a checkpoint image.
	SaveImageMetadata(checkpointHash string, metadata *entity.ContainerMetadata) error
	// RetrieveImageMetadata retrieves the metadata about a checkpoint image.
	RetrieveImageMetadata(checkpointHash string) (*entity.ContainerMetadata, error)
	// Restore restores the monitored application container to a previous checkpointed
	// image.
	Restore() error
	// DevelopmentRestore development use case to restore a specific container image with
	// the given hash.
	DevelopmentRestore(containerName string, containerHash string) error
}

// ContainerMetadataRepository repository to access container metadata at a datasource.
type ContainerMetadataRepository interface {
	// Insert inserts a new metadata.
	Insert(checkpointHash string, metadata *entity.ContainerMetadata) error
	// Get retrieves a container metadata by checkpointHash.
	Get(checkpointHash string) (*entity.ContainerMetadata, error)
	// UpsertContainerLatestCheckpoint upserts the content of the latest checkpoint
	// hash the container received.
	UpsertContainerLatestCheckpoint(checkpointHash string, containerID string) error
	// LatestContainerCheckpoint retrieves the latest container checkpoint hash.
	LatestContainerCheckpoint(containerID string) (string, error)
}

type stateManagerUseCase struct {
	repository           ContainerMetadataRepository
	restoreService       entity.RestoreService
	monitoredApplication *entity.Container
}

func StateManager(repository ContainerMetadataRepository, restoreService entity.RestoreService, monitoredApplication *entity.Container) (StateManagerUseCase, error) {
	return &stateManagerUseCase{
		repository:           repository,
		restoreService:       restoreService,
		monitoredApplication: monitoredApplication,
	}, nil
}

func (uc *stateManagerUseCase) SaveImageMetadata(checkpointHash string, metadata *entity.ContainerMetadata) error {
	err := uc.repository.UpsertContainerLatestCheckpoint(checkpointHash, uc.monitoredApplication.ID)
	if err != nil {
		return err
	}

	return uc.repository.Insert(checkpointHash, metadata)
}

func (uc *stateManagerUseCase) RetrieveImageMetadata(checkpointHash string) (*entity.ContainerMetadata, error) {
	return uc.repository.Get(checkpointHash)
}

func (uc *stateManagerUseCase) Restore() error {
	checkpointHash, err := uc.repository.LatestContainerCheckpoint(uc.monitoredApplication.ID)
	if err != nil {
		return err
	}

	return uc.restoreService.Restore(&entity.RestoreConfig{
		ContainerName:  uc.monitoredApplication.Name,
		CheckpointHash: checkpointHash,
	})
}

func (uc *stateManagerUseCase) DevelopmentRestore(containerName string, containerHash string) error {
	return uc.restoreService.Restore(&entity.RestoreConfig{
		ContainerName:  containerName,
		CheckpointHash: containerHash,
	})
}

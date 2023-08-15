package usecase

import (
	"testing"
	"time"

	"github.com/GianOrtiz/k8s-transparent-checkpoint-restore/internal/entity"
	"github.com/GianOrtiz/k8s-transparent-checkpoint-restore/internal/repository/containermetadata"
	"github.com/GianOrtiz/k8s-transparent-checkpoint-restore/internal/service/restore"
	"github.com/google/uuid"
)

func TestStateManager(t *testing.T) {
	containerMetadataRepository := containermetadata.InMemory()
	restoreService := restore.AlwaysAcceptStub()
	stateManager, err := StateManager(containerMetadataRepository, restoreService, &entity.Container{
		ID:      uuid.NewString(),
		PID:     30,
		HTTPUrl: "http://localhost:8000",
		Name:    "test",
	})
	if err != nil {
		t.Fatal(err)
	}

	t.Run("should save image metadata", func(t *testing.T) {
		checkpointHash := uuid.NewString()
		containerMetadata := entity.ContainerMetadata{
			LastTimestamp:       time.Now(),
			LastRequestSolvedID: uuid.NewString(),
		}
		err = stateManager.SaveImageMetadata(checkpointHash, &containerMetadata)
		if err != nil {
			t.Errorf("should got error nil, received %v\n", err)
		}

		t.Run("should retrieve image metadata", func(t *testing.T) {
			metadata, err := containerMetadataRepository.Get(checkpointHash)
			if err != nil {
				t.Errorf("expected to get no error retrieving metadata, got %v\n", err)
			}
			if metadata.LastRequestSolvedID != containerMetadata.LastRequestSolvedID {
				t.Errorf("expected last request solved id to be %q, received %q\n", containerMetadata.LastRequestSolvedID, metadata.LastRequestSolvedID)
			}
			if metadata.LastTimestamp != containerMetadata.LastTimestamp {
				t.Errorf("expected last timestamp to be %v, received %v\n", containerMetadata.LastTimestamp, metadata.LastTimestamp)
			}
		})
	})
}

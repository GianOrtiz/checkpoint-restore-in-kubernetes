package checkpoint

import (
	"fmt"
	"os"

	"github.com/GianOrtiz/k8s-transparent-checkpoint-restore/internal/entity"
	"github.com/checkpoint-restore/go-criu/v6"
	"github.com/checkpoint-restore/go-criu/v6/rpc"
)

// CRIUCheckpointServiceConfig configuration to run checkpoint service with CRIU.
type CRIUCheckpointServiceConfig struct {
	// ImagesDirectory directory to store and retrieve checkpoint images.
	ImagesDirectory string
}

// CRIUCheckpointService uses CRIU to implement the CheckpointService interface.
type CRIUCheckpointService struct {
	*criu.Criu
	imagesDirectory string
}

// NewService creates a new service using CRIU for checkpoint/restore.
func CRIU(cfg CRIUCheckpointServiceConfig) (*CRIUCheckpointService, error) {
	criu := criu.MakeCriu()
	return &CRIUCheckpointService{
		Criu:            criu,
		imagesDirectory: cfg.ImagesDirectory,
	}, nil
}

func (service *CRIUCheckpointService) Checkpoint(config *entity.CheckpointConfig) error {
	checkpointImageDirectory := fmt.Sprintf("%s/%s", service.imagesDirectory, config.CheckpointHash)
	os.Mkdir(checkpointImageDirectory, os.ModeDir) // Creates the checkpointing directory if it is not created yet.
	imagesDir, err := os.OpenFile(checkpointImageDirectory, 0, os.ModeDir)
	if err != nil {
		return err
	}
	defer imagesDir.Close()

	imagesDirFd := int32(imagesDir.Fd())

	// Uses the dump command on CRIU to dump a new checkpoint image of the process in
	// the given directory by the configuration.
	return service.Dump(&rpc.CriuOpts{
		Pid:         &config.Container.PID,
		ImagesDirFd: &imagesDirFd,
	}, nil)
}

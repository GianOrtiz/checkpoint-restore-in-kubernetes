package criu

import (
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
	imagesDirFd int32
}

// NewService creates a new service using CRIU for checkpoint/restore.
func NewService(cfg CRIUCheckpointServiceConfig) (*CRIUCheckpointService, error) {
	imagesDir, err := os.OpenFile(cfg.ImagesDirectory, 0, os.ModeDir)
	if err != nil {
		return nil, err
	}
	criu := criu.MakeCriu()
	return &CRIUCheckpointService{
		Criu:        criu,
		imagesDirFd: int32(imagesDir.Fd()),
	}, nil
}

func (s *CRIUCheckpointService) Checkpoint(config *entity.CheckpointConfig) error {
	// Uses the dump command on CRIU to dump a new checkpoint image of the process in
	// the given directory by the configuration.
	return s.Dump(&rpc.CriuOpts{
		Pid:         &config.Container.PID,
		ImagesDirFd: &s.imagesDirFd,
	}, criu.NoNotify{})
}

package restore

import (
	"fmt"
	"os"

	"github.com/GianOrtiz/k8s-transparent-checkpoint-restore/internal/entity"
	"github.com/checkpoint-restore/go-criu/v6"
	"github.com/checkpoint-restore/go-criu/v6/rpc"
)

type CriuRestoreServiceConfig struct {
	ImagesDirectory string
}

type criuRestoreService struct {
	*criu.Criu
	imagesDirectory string
}

func CRIU(cfg CriuRestoreServiceConfig) (*criuRestoreService, error) {
	criu := criu.MakeCriu()
	return &criuRestoreService{
		Criu:            criu,
		imagesDirectory: cfg.ImagesDirectory,
	}, nil
}

func (service *criuRestoreService) Restore(cfg *entity.RestoreConfig) error {
	containerImage := fmt.Sprintf("%s-%s", cfg.ContainerName, cfg.CheckpointHash)
	checkpointImageDirectory := fmt.Sprintf("%s/%s", service.imagesDirectory, containerImage)
	imagesDir, err := os.OpenFile(checkpointImageDirectory, 0, os.ModeDir)
	if err != nil {
		return err
	}
	defer imagesDir.Close()

	imagesDirFd := int32(imagesDir.Fd())

	// Uses the CRIU restore command to restore a specific image checkpoint by its hash.
	return service.Criu.Restore(&rpc.CriuOpts{
		ImagesDirFd: &imagesDirFd,
	}, nil)
}

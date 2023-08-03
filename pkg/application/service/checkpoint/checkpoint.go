package checkpoint

import "github.com/GianOrtiz/k8s-transparent-checkpoint-restore/pkg/domain"

// CheckpointConfig is the config to use for the checkpoint.
type CheckpointConfig struct {
	// Container is the container to make the checkpoint.
	Container *domain.Container
	// WriteTo is the location to write the checkpoint image to.
	WriteTo string
	// CheckpointHash is a hash to identify this checkpoint.
	CheckpointHash string
	// Metadata is associated metada with this checkpoint image.
	Metadata map[string]interface{}
}

// CheckpointService provides facilities for checkpointing our application.
type CheckpointService interface {
	// Checkpoint makes a new checkpoint image of an application.
	Checkpoint(config *CheckpointConfig) error
}
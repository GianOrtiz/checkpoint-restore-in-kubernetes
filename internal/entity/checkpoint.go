//go:generate mockgen -source=./checkpoint.go -destination=./mock/checkpoint.go

package entity

// CheckpointConfig is the config to use for the checkpoint.
type CheckpointConfig struct {
	// Container is the container to make the checkpoint.
	Container *Container
	// CheckpointHash is a hash to identify this checkpoint.
	CheckpointHash string
}

// CheckpointService provides facilities for checkpointing our application.
type CheckpointService interface {
	// Checkpoint makes a new checkpoint image of an application.
	Checkpoint(config *CheckpointConfig) error
}

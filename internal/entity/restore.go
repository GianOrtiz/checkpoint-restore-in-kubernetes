package entity

// RestoreConfig is the configuration to use to restore the application.
type RestoreConfig struct {
	// Container is the container to make the restore of.
	Container *Container
	// CheckpointHash is the hash of the image to use.
	CheckpointHash string
}

// RestoreService restores an application from previous checkpointed images.
type RestoreService interface {
	// Restore restores the application to a previous image.
	Restore(config *RestoreConfig) error
}

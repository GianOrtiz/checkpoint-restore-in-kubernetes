//go:generate mockgen -source=./statemanager.go -destination=./mock/statemanager.go

package entity

// StateManagerConfig defines the state manager configuration.
type StateManagerConfig struct {
	// Flag to either enable or disable development features of state manager, like
	// restore endpoint o create a new restore from an image checkpoint hash.
	DevelopmentFeaturesEnabled bool
}

// StateManagerService is the service to communicate with the state manager
type StateManagerService interface {
	// SaveMedata saves metadata about the specified container.
	SaveMetadata(containerName string, metadata *ContainerMetadata) error
}

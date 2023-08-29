//go:generate mockgen -source=./statemanager.go -destination=./mock/statemanager.go

package entity

// StateManagerService is the service to communicate with the state manager
type StateManagerService interface {
	// SaveMedata saves metadata about the specified container.
	SaveMetadata(containerName string, metadata *ContainerMetadata) error
}

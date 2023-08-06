package entity

// StateManagerService is the service to communicate with the state manager
type StateManagerService interface {
	// SaveMedata saves metadata about the specified container.
	SaveMetadata(containerName string, metadata map[string]interface{}) error
}

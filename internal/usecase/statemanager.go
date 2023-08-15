package usecase

// StateManagerUseCase declares use cases for the State Manager. It declares the use cases for
// saving checkpoint images metadata and retrieving them.
type StateManagerUseCase interface {
	// SaveImageMetadata saves metadata about a checkpoint image.
	SaveImageMetadata(checkpointHash string, metadata map[string]interface{}) error
	// RetrieveImageMetadata retrieves the metadata about a checkpoint image.
	RetrieveImageMetadata(checkpointHash string) (map[string]interface{}, error)
}

// ContainerMetadataRepository repository to access container metadata at a datasource.
type ContainerMetadataRepository interface {
	// Insert inserts a new metadata.
	Insert(checkpointHash string, metadata map[string]interface{}) error
	// Get retrieves a container metadata by checkpointHash.
	Get(checkpointHash string) (map[string]interface{}, error)
}

type stateManagerUseCase struct {
	repository ContainerMetadataRepository
}

func StateManager(repository ContainerMetadataRepository) (StateManagerUseCase, error) {
	return &stateManagerUseCase{
		repository: repository,
	}, nil
}

func (uc *stateManagerUseCase) SaveImageMetadata(checkpointHash string, metadata map[string]interface{}) error {
	return uc.repository.Insert(checkpointHash, metadata)
}

func (uc *stateManagerUseCase) RetrieveImageMetadata(checkpointHash string) (map[string]interface{}, error) {
	return uc.repository.Get(checkpointHash)
}

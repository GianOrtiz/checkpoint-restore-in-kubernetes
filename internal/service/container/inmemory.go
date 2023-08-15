package container

type InMemoryContainerMetadataRepository struct {
	memory map[string](map[string]interface{})
}

func InMemory() *InMemoryContainerMetadataRepository {
	return &InMemoryContainerMetadataRepository{
		memory: make(map[string]map[string]interface{}),
	}
}

func (r *InMemoryContainerMetadataRepository) Insert(checkpointHash string, metadata map[string]interface{}) error {
	r.memory[checkpointHash] = metadata
	return nil
}

func (r *InMemoryContainerMetadataRepository) Get(checkpointHash string) (map[string]interface{}, error) {
	return r.memory[checkpointHash], nil
}

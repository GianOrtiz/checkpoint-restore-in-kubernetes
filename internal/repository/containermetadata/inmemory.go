package containermetadata

type inMemoryContainerMetadataRepository struct {
	metadataMemory                map[string]map[string]interface{}
	containerCheckpointHashMemory map[string]string
}

func InMemory() *inMemoryContainerMetadataRepository {
	return &inMemoryContainerMetadataRepository{
		metadataMemory:                make(map[string]map[string]interface{}),
		containerCheckpointHashMemory: make(map[string]string),
	}
}

func (r *inMemoryContainerMetadataRepository) Insert(checkpointHash string, metadata map[string]interface{}) error {
	r.metadataMemory[checkpointHash] = metadata
	return nil
}

func (r *inMemoryContainerMetadataRepository) Get(checkpointHash string) (map[string]interface{}, error) {
	return r.metadataMemory[checkpointHash], nil
}

func (r *inMemoryContainerMetadataRepository) UpsertContainerLatestCheckpoint(checkpointHash string, containerID string) error {
	r.containerCheckpointHashMemory[containerID] = checkpointHash
	return nil
}

func (r *inMemoryContainerMetadataRepository) LatestContainerCheckpoint(containerID string) (string, error) {
	return r.containerCheckpointHashMemory[containerID], nil
}

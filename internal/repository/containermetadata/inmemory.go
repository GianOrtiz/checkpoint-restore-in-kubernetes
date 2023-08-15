package containermetadata

import "github.com/GianOrtiz/k8s-transparent-checkpoint-restore/internal/entity"

type inMemoryContainerMetadataRepository struct {
	metadataMemory                map[string]*entity.ContainerMetadata
	containerCheckpointHashMemory map[string]string
}

func InMemory() *inMemoryContainerMetadataRepository {
	return &inMemoryContainerMetadataRepository{
		metadataMemory:                make(map[string]*entity.ContainerMetadata),
		containerCheckpointHashMemory: make(map[string]string),
	}
}

func (r *inMemoryContainerMetadataRepository) Insert(checkpointHash string, metadata *entity.ContainerMetadata) error {
	r.metadataMemory[checkpointHash] = metadata
	return nil
}

func (r *inMemoryContainerMetadataRepository) Get(checkpointHash string) (*entity.ContainerMetadata, error) {
	return r.metadataMemory[checkpointHash], nil
}

func (r *inMemoryContainerMetadataRepository) UpsertContainerLatestCheckpoint(checkpointHash string, containerID string) error {
	r.containerCheckpointHashMemory[containerID] = checkpointHash
	return nil
}

func (r *inMemoryContainerMetadataRepository) LatestContainerCheckpoint(containerID string) (string, error) {
	return r.containerCheckpointHashMemory[containerID], nil
}

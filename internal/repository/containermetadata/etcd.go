package containermetadata

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/GianOrtiz/k8s-transparent-checkpoint-restore/internal/entity"
	client "go.etcd.io/etcd/client/v3"
)

type etcdContainerMetadataRepository struct {
	etcdClient *client.Client
}

func ETCD(etcdClient *client.Client) *etcdContainerMetadataRepository {
	return &etcdContainerMetadataRepository{
		etcdClient: etcdClient,
	}
}

func (r *etcdContainerMetadataRepository) Insert(checkpointHash string, metadata *entity.ContainerMetadata) error {
	encodedContainerMetadata, err := json.Marshal(metadata)
	if err != nil {
		return err
	}

	_, err = r.etcdClient.Put(context.Background(), checkpointHash, string(encodedContainerMetadata))
	if err != nil {
		return err
	}

	return nil
}

func (r *etcdContainerMetadataRepository) Get(checkpointHash string) (*entity.ContainerMetadata, error) {
	res, err := r.etcdClient.Get(context.Background(), checkpointHash)
	if err != nil {
		return nil, err
	}

	if len(res.Kvs) > 0 {
		var metadata entity.ContainerMetadata
		err := json.Unmarshal(res.Kvs[0].Value, &metadata)
		if err != nil {
			return nil, err
		}

		return &metadata, nil
	}

	return nil, fmt.Errorf("no result for key %q", checkpointHash)
}

func (r *etcdContainerMetadataRepository) UpsertContainerLatestCheckpoint(checkpointHash string, containerID string) error {
	_, err := r.etcdClient.Put(context.Background(), containerID, checkpointHash)
	if err != nil {
		return err
	}

	return nil
}

func (r *etcdContainerMetadataRepository) LatestContainerCheckpoint(containerID string) (string, error) {
	res, err := r.etcdClient.Get(context.Background(), containerID)
	if err != nil {
		return "", err
	}

	if len(res.Kvs) > 0 {
		return string(res.Kvs[0].Value), nil
	}

	return "", fmt.Errorf("no result for key %q", containerID)
}

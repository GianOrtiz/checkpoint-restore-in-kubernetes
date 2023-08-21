package statemanager

import (
	"github.com/GianOrtiz/k8s-transparent-checkpoint-restore/internal/entity"
	"github.com/GianOrtiz/k8s-transparent-checkpoint-restore/pkg/statemanager/client"
)

type httpStateManagerService struct {
	client *client.Client
}

func HTTP(stateManagerURL string) *httpStateManagerService {
	c := client.New(stateManagerURL)
	return &httpStateManagerService{
		client: c,
	}
}

func (stateManager *httpStateManagerService) SaveMetadata(containerName string, metadata *entity.ContainerMetadata) error {
	return stateManager.client.InsertMetadata(containerName, metadata)
}

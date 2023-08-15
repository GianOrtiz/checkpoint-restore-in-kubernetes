package statemanager

import "github.com/GianOrtiz/k8s-transparent-checkpoint-restore/internal/entity"

type alwaysAcceptingStateManagerStub struct{}

func AlawaysAcceptingStub() *alwaysAcceptingStateManagerStub {
	return &alwaysAcceptingStateManagerStub{}
}

func (stateManager *alwaysAcceptingStateManagerStub) SaveMetadata(containerName string, metadata *entity.ContainerMetadata) error {
	return nil
}

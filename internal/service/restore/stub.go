package restore

import "github.com/GianOrtiz/k8s-transparent-checkpoint-restore/internal/entity"

type alwaysAcceptRestoreServiceStub struct{}

func AlwaysAcceptStub() entity.RestoreService {
	return &alwaysAcceptRestoreServiceStub{}
}

func (svc *alwaysAcceptRestoreServiceStub) Restore(cfg *entity.RestoreConfig) error {
	return nil
}

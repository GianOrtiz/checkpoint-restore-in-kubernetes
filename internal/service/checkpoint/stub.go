package checkpoint

import (
	"github.com/GianOrtiz/k8s-transparent-checkpoint-restore/internal/entity"
)

type StubCheckpointService struct {
}

func Stub() *StubCheckpointService {
	return &StubCheckpointService{}
}

func (s *StubCheckpointService) Checkpoint(config *entity.CheckpointConfig) error {
	return nil
}

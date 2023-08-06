package interceptor

import (
	"fmt"
	"testing"
)

func TestFromYAML(t *testing.T) {
	containerURL := "http://localhost:8000"
	checkpointingIntervalInMinutes := 5
	containerName := "test"
	containerPID := 1100
	stateManagerURL := "http://localhost:4000"

	yamlContent := []byte(
		fmt.Sprintf(`checkpointingInterval: %dm
containerURL: "%s"
containerPID: %d
containerName: "%s"
stateManagerURL: "%s"`,
			checkpointingIntervalInMinutes,
			containerURL,
			containerPID,
			containerName,
			stateManagerURL))
	cfg, err := FromYAML(yamlContent)
	if err != nil {
		t.Errorf("expected error nil, received %v\n", err)
	}

	if int(cfg.CheckpointingInterval.Minutes()) != checkpointingIntervalInMinutes {
		t.Errorf("expected checkpoint interval from parse to be %d, got %d\n", checkpointingIntervalInMinutes, int(cfg.CheckpointingInterval.Minutes()))
	}

	if cfg.ContainerURL.String() != containerURL {
		t.Errorf("expected parsed container URL to be %q, got %q\n", containerURL, cfg.ContainerURL.String())
	}

	if cfg.ContainerPID != int32(containerPID) {
		t.Errorf("expected parsed container PID to be %d, got %d\n", containerPID, cfg.ContainerPID)
	}

	if cfg.ContainerName != containerName {
		t.Errorf("expected parsed container name to be %q, got %q\n", containerName, cfg.ContainerName)
	}

	if cfg.StateManagerURL.String() != stateManagerURL {
		t.Errorf("expected parsed state manager url to be %q, got %q\n", stateManagerURL, cfg.StateManagerURL.String())
	}
}

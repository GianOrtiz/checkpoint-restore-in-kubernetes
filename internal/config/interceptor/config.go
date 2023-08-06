package interceptor

import (
	"net/url"
	"os"
	"time"

	"gopkg.in/yaml.v2"
)

// Config is the configuration of the Interceptor.
type Config struct {
	// CheckpointingInterval is the interval between each checkpoint the Interceptor
	// must perform in the monitored container.
	CheckpointingInterval time.Duration
	// ContainerURL the url to use for the monitored container.
	ContainerURL url.URL
	// ContainerPID the monitored container PID.
	ContainerPID int32
	// ContainerName the name of the monitored container.
	ContainerName string
	// StateManagerURL the url to use to communicate with the State Manager API.
	StateManagerURL url.URL
}

func FromYAMLFile(filename string) (*Config, error) {
	content, err := os.ReadFile(filename)
	if err != nil {
		return nil, err
	}
	return FromYAML(content)
}

func FromYAML(content []byte) (*Config, error) {
	type configYAML struct {
		CheckpointingInterval string `yaml:"checkpointingInterval"`
		ContainerURL          string `yaml:"containerURL"`
		ContainerPID          int    `yaml:"containerPID"`
		ContainerName         string `yaml:"containerName"`
		StateManagerURL       string `yaml:"stateManagerURL"`
	}

	var cfg configYAML
	err := yaml.Unmarshal(content, &cfg)
	if err != nil {
		return nil, err
	}

	checkpointingInterval, err := time.ParseDuration(cfg.CheckpointingInterval)
	if err != nil {
		return nil, err
	}

	containerURL, err := url.Parse(cfg.ContainerURL)
	if err != nil {
		return nil, err
	}

	stateManagerURL, err := url.Parse(cfg.StateManagerURL)
	if err != nil {
		return nil, err
	}

	return &Config{
		CheckpointingInterval: checkpointingInterval,
		ContainerURL:          *containerURL,
		ContainerPID:          int32(cfg.ContainerPID),
		ContainerName:         cfg.ContainerName,
		StateManagerURL:       *stateManagerURL,
	}, nil
}

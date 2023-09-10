package interceptor

import (
	"fmt"
	"net/url"
	"os"
	"time"

	"gopkg.in/yaml.v2"
)

const (
	CHECKPOINT_INTERVAL_VARIABLE_KEY = "CHECKPOINT_INTERVAL"
	CONTAINER_URL_VARIABLE_KEY       = "CONTAINER_URL"
	STATE_MANAGER_URL_VARIABLE_KEY   = "STATE_MANAGER_URL"
	CONTAINER_NAME_VARIABLE_KEY      = "CONTAINER_NAME"
	ENVIRONMENT_VARIABLE_KEY         = "ENV"
)

const (
	KUBERNETES_ENVIRONMENT = "kubernetes"
	STANDALONE_ENVIRONMENT = "standalone"
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
	// Environment environment where the Interceptor is running.
	Environment string
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

func FromEnv() (*Config, error) {
	checkpointIntervalString := os.Getenv(CHECKPOINT_INTERVAL_VARIABLE_KEY)
	if checkpointIntervalString == "" {
		return nil, fmt.Errorf("%s variable is missing", CHECKPOINT_INTERVAL_VARIABLE_KEY)
	}

	checkpointInterval, err := time.ParseDuration(checkpointIntervalString)
	if err != nil {
		return nil, fmt.Errorf("%s variable is not of valid format: %v", CHECKPOINT_INTERVAL_VARIABLE_KEY, err)
	}

	containerURL, err := url.Parse(os.Getenv(CONTAINER_URL_VARIABLE_KEY))
	if err != nil {
		return nil, fmt.Errorf("%s variable is not of valid format: %v", CONTAINER_URL_VARIABLE_KEY, err)
	}

	stateManagerURL, err := url.Parse(os.Getenv(STATE_MANAGER_URL_VARIABLE_KEY))
	if err != nil {
		return nil, fmt.Errorf("%s variable is not of valid format: %v", STATE_MANAGER_URL_VARIABLE_KEY, err)
	}

	containerName := os.Getenv(CONTAINER_NAME_VARIABLE_KEY)

	environment := os.Getenv(ENVIRONMENT_VARIABLE_KEY)
	if environment == "" {
		environment = STANDALONE_ENVIRONMENT
	}

	return &Config{
		CheckpointingInterval: checkpointInterval,
		ContainerURL:          *containerURL,
		ContainerPID:          int32(0),
		ContainerName:         containerName,
		StateManagerURL:       *stateManagerURL,
		Environment:           environment,
	}, nil
}

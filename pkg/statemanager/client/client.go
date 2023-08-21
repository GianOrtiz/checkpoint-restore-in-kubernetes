package client

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/GianOrtiz/k8s-transparent-checkpoint-restore/internal/entity"
)

const CONTAINERS_PATH = "/containers"

type Client struct {
	httpClient *http.Client
	baseURL    string
}

func New(stateManagerURL string) *Client {
	httpClient := http.Client{
		Transport: http.DefaultTransport,
	}
	return &Client{
		httpClient: &httpClient,
		baseURL:    stateManagerURL,
	}
}

func (c *Client) InsertMetadata(checkpointHash string, containerMetadata *entity.ContainerMetadata) error {
	bodyBuffer := bytes.NewBuffer([]byte{})
	err := json.NewEncoder(bodyBuffer).Encode(containerMetadata)
	if err != nil {
		return err
	}

	url := fmt.Sprintf("%s/%s/%s", c.baseURL, CONTAINERS_PATH, checkpointHash)
	req, err := http.NewRequest(http.MethodPost, url, bodyBuffer)
	if err != nil {
		return err
	}

	res, err := c.httpClient.Do(req)
	if err != nil {
		return err
	}
	defer res.Body.Close()

	if res.StatusCode != http.StatusOK {
		return fmt.Errorf("status code is %d", res.StatusCode)
	}

	return nil
}

func (c *Client) RetrieveMetadata(checkpointHash string) (*entity.ContainerMetadata, error) {
	url := fmt.Sprintf("%s/%s/%s", c.baseURL, CONTAINERS_PATH, checkpointHash)
	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}

	res, err := c.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()

	if res.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("status code is %d", res.StatusCode)
	}

	var containerMetadata entity.ContainerMetadata
	err = json.NewDecoder(res.Body).Decode(&containerMetadata)
	if err != nil {
		return nil, err
	}

	return &containerMetadata, nil
}

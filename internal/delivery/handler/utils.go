package handler

import (
	"fmt"
	"net/url"
	"strings"
)

func retrieveContainerIDFromURL(url url.URL) (string, error) {
	parts := strings.Split(url.Path, "/")
	if len(parts) >= 3 {
		containerID := parts[2]
		return containerID, nil
	}
	return "", fmt.Errorf("path does not contain container id")
}

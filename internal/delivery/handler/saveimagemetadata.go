package handler

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/GianOrtiz/k8s-transparent-checkpoint-restore/internal/entity"
	"github.com/GianOrtiz/k8s-transparent-checkpoint-restore/internal/usecase"
)

type saveImageMetadataHandler struct {
	stateManagerUseCase usecase.StateManagerUseCase
}

func SaveImageMetadata(stateManagerUseCase usecase.StateManagerUseCase) *saveImageMetadataHandler {
	return &saveImageMetadataHandler{
		stateManagerUseCase: stateManagerUseCase,
	}
}

func (handler *saveImageMetadataHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	type httpBody struct {
		ImageHash string                   `json:"image_hash"`
		Metadata  entity.ContainerMetadata `json:"metadata"`
	}

	containerID, err := retrieveContainerIDFromURL(*r.URL)
	if err != nil {
		w.WriteHeader(http.StatusNotFound)
		return
	}

	var body httpBody
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	checkpointHash := fmt.Sprintf("%s-%s", containerID, body.ImageHash)
	err = handler.stateManagerUseCase.SaveImageMetadata(checkpointHash, &body.Metadata)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
}

package handler

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/GianOrtiz/k8s-transparent-checkpoint-restore/internal/usecase"
)

type getImageMetadataHandler struct {
	stateManagerUseCase usecase.StateManagerUseCase
}

func GetImageMetadata(stateManagerUseCase usecase.StateManagerUseCase) *getImageMetadataHandler {
	return &getImageMetadataHandler{
		stateManagerUseCase: stateManagerUseCase,
	}
}

func (handler *getImageMetadataHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	containerID, err := retrieveContainerIDFromURL(*r.URL)
	if err != nil {
		w.WriteHeader(http.StatusNotFound)
		return
	}

	type httpBody struct {
		Hash string `json:"hash"`
	}

	var body httpBody
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	checkpointHash := fmt.Sprintf("%s-%s", containerID, body.Hash)
	metadata, err := handler.stateManagerUseCase.RetrieveImageMetadata(checkpointHash)
	if err != nil {
		w.WriteHeader(http.StatusNotFound)
		return
	}

	if err := json.NewEncoder(w).Encode(metadata); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
}

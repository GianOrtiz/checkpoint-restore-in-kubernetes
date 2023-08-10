package usecase

import (
	"encoding/hex"
	"fmt"
	"hash/fnv"
	"net/http"
	"time"

	"github.com/GianOrtiz/k8s-transparent-checkpoint-restore/internal/entity"
)

// InterceptorUseCase declares use cases for the Interceptor. It declares the use cases for intercepting
// requisitions and communicating with the Interceptor.
type InterceptorUseCase interface {
	// InterceptRequest intercepts an HTTP request that should have been sent to the monitored application.
	InterceptRequest(reqID string, req *http.Request) (*http.Response, error)
	// Checkpoint creates a new checkpoint of the monitored container.
	Checkpoint() error
}

type interceptorUseCase struct {
	Interceptor                  *entity.Interceptor
	CheckpointService            entity.CheckpointService
	StateManagerService          entity.StateManagerService
	InterceptedRequestRepository entity.InterceptedRequestRepository
}

func Interceptor(interceptor *entity.Interceptor, checkpointService entity.CheckpointService, stateManagerService entity.StateManagerService, interceptedRequestRepository entity.InterceptedRequestRepository) (InterceptorUseCase, error) {
	return &interceptorUseCase{
		Interceptor:                  interceptor,
		InterceptedRequestRepository: interceptedRequestRepository,
		CheckpointService:            checkpointService,
		StateManagerService:          stateManagerService,
	}, nil
}

// InterceptRequest intercepts a given request and return the response after it is
// redirected to the monitored application.
func (uc *interceptorUseCase) InterceptRequest(reqID string, req *http.Request) (*http.Response, error) {
	interceptedRequest := entity.InterceptedRequest{
		ID:      reqID,
		Request: req,
		Solved:  false,
	}
	if err := uc.InterceptedRequestRepository.Save(&interceptedRequest); err != nil {
		return nil, err
	}

	// Create the URL to access the monitored URL from the monitored application URL
	// and the content receive in the path of the intercepted request.
	url := uc.Interceptor.MonitoredContainer.HTTPUrl + req.URL.Path
	reqCopy, err := http.NewRequest(req.Method, url, req.Body)
	if err != nil {
		return nil, err
	}

	res, err := http.DefaultClient.Do(reqCopy)
	if err != nil {
		return nil, err
	}

	if err := uc.InterceptedRequestRepository.SetSolved(reqID, time.Now(), true); err != nil {
		return nil, err
	}

	return res, nil
}

// Checkpoint the monitored application into a new image.
func (uc *interceptorUseCase) Checkpoint() error {
	metadata := uc.generateMetadataForNewImage()
	if err := uc.StateManagerService.SaveMetadata(uc.Interceptor.MonitoredContainer.Name, metadata); err != nil {
		return err
	}

	checkpointHash := uc.generateHashForNewImage(uc.Interceptor.MonitoredContainer.Name)
	return uc.CheckpointService.Checkpoint(&entity.CheckpointConfig{
		Container:      uc.Interceptor.MonitoredContainer,
		CheckpointHash: checkpointHash,
	})
}

func (uc *interceptorUseCase) generateHashForNewImage(containerName string) string {
	h := fnv.New64a()

	// Hash of Timestamp rounded to previous hour
	h.Write([]byte(time.Now().UTC().String()))
	h.Write([]byte(containerName))
	hash := hex.EncodeToString(h.Sum(nil))
	return fmt.Sprintf("%s-%s", containerName, hash)
}

func (uc *interceptorUseCase) generateMetadataForNewImage() map[string]interface{} {
	lastTimestamp := time.Now()
	// TODO: add a logger to log the error?
	lastRequestSolved, _ := uc.InterceptedRequestRepository.GetLastRequestSolved()

	lastRequestSolvedID := "-1"
	if lastRequestSolved != nil {
		lastRequestSolvedID = lastRequestSolved.ID
	}

	return map[string]interface{}{
		"lastTimestamp":       lastTimestamp,
		"lastRequestSolvedID": lastRequestSolvedID,
	}
}

package usecase

import (
	"encoding/hex"
	"fmt"
	"hash/fnv"
	"net/http"
	"sync"
	"time"

	"github.com/GianOrtiz/k8s-transparent-checkpoint-restore/internal/config/interceptor"
	"github.com/GianOrtiz/k8s-transparent-checkpoint-restore/internal/entity"
)

type InterceptorState int

const (
	Caching InterceptorState = iota
	Proxying
)

// InterceptorUseCase declares use cases for the Interceptor. It declares the use cases for intercepting
// requisitions and communicating with the Interceptor.
type InterceptorUseCase interface {
	// InterceptRequest intercepts an HTTP request that should have been sent to the monitored application.
	InterceptRequest(reqID string, req *http.Request) (*http.Response, error)
	// Checkpoint creates a new checkpoint of the monitored container.
	Checkpoint() error
	// Reproject reprojects the requests to the monitored application since the given version.
	Reproject(version int) error
	// GetState gets the current state of the interceptor.
	GetState() InterceptorState
	// SetState sets the current state of the interceptor.
	SetState(state InterceptorState)
}

// Scheduler schedules tasks to be handled in the future.
type Scheduler interface {
	// ScheduleCheckpoint schedules the checkponint to be made in the future.
	ScheduleCheckpoint(usecase InterceptorUseCase, scheduleIn time.Duration) error
}

type interceptorUseCase struct {
	Interceptor                  *entity.Interceptor
	CheckpointService            entity.CheckpointService
	StateManagerService          entity.StateManagerService
	InterceptedRequestRepository entity.InterceptedRequestRepository
	Scheduler                    Scheduler
	LastVersion                  int
	Mutex                        sync.Mutex
	currentState                 InterceptorState
	waitChannel                  chan struct{}
}

func Interceptor(interceptor *entity.Interceptor, checkpointService entity.CheckpointService, stateManagerService entity.StateManagerService, interceptedRequestRepository entity.InterceptedRequestRepository, scheduler Scheduler) (InterceptorUseCase, error) {
	// Retrieve the last version of request in the database
	lastVersion, err := interceptedRequestRepository.GetLastVersion()
	if err != nil {
		return nil, err
	}

	usecase := interceptorUseCase{
		Interceptor:                  interceptor,
		InterceptedRequestRepository: interceptedRequestRepository,
		CheckpointService:            checkpointService,
		StateManagerService:          stateManagerService,
		LastVersion:                  lastVersion,
		Scheduler:                    scheduler,
		Mutex:                        sync.Mutex{},
		currentState:                 Proxying,
		waitChannel:                  make(chan struct{}),
	}
	close(usecase.waitChannel)
	return &usecase, nil
}

// InterceptRequest intercepts a given request and return the response after it is
// redirected to the monitored application.
func (uc *interceptorUseCase) InterceptRequest(reqID string, req *http.Request) (*http.Response, error) {
	// TODO: abstract this
	uc.Mutex.Lock()
	interceptedRequest := entity.InterceptedRequest{
		ID:      reqID,
		Request: req,
		Solved:  false,
		Version: uc.LastVersion + 1,
	}
	uc.LastVersion++
	uc.Mutex.Unlock()

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

	for key, values := range req.Header {
		for _, value := range values {
			reqCopy.Header.Add(key, value)
		}
	}

	if uc.currentState == Caching {
		<-uc.waitChannel
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
	err := uc.CheckpointService.Checkpoint(&entity.CheckpointConfig{
		Container:      uc.Interceptor.MonitoredContainer,
		CheckpointHash: checkpointHash,
		PodName:        uc.Interceptor.Config.KubernetesPodName,
	})
	if err != nil {
		return err
	}

	if uc.Interceptor.Config.Environment == interceptor.STANDALONE_ENVIRONMENT {
		// Reeschedule checkpoint in the future.
		return uc.Scheduler.ScheduleCheckpoint(uc, uc.Interceptor.Config.CheckpointingInterval)
	}

	return nil
}

func (uc *interceptorUseCase) GetState() InterceptorState {
	return uc.currentState
}

func (uc *interceptorUseCase) SetState(state InterceptorState) {
	if state == Caching {
		uc.waitChannel = make(chan struct{})
	} else if state == Proxying {
		if uc.waitChannel != nil {
			close(uc.waitChannel)
		}
	}
	uc.currentState = state
}

func (uc *interceptorUseCase) Reproject(version int) error {
	requests, err := uc.InterceptedRequestRepository.GetAllFromLastVersion(version)
	if err != nil {
		return err
	}

	for _, interceptedReq := range requests {
		// Create the URL to access the monitored URL from the monitored application URL
		// and the content receive in the path of the intercepted request.
		url := uc.Interceptor.MonitoredContainer.HTTPUrl + interceptedReq.Request.URL.Path
		reqCopy, err := http.NewRequest(interceptedReq.Request.Method, url, interceptedReq.Request.Body)
		if err != nil {
			return err
		}

		for key, values := range interceptedReq.Request.Header {
			for _, value := range values {
				reqCopy.Header.Add(key, value)
			}
		}

		_, err = http.DefaultClient.Do(reqCopy)
		if err != nil {
			return err
		}

		if err := uc.InterceptedRequestRepository.SetSolved(interceptedReq.ID, time.Now(), true); err != nil {
			return err
		}
	}

	return nil
}

func (uc *interceptorUseCase) generateHashForNewImage(containerName string) string {
	h := fnv.New64a()

	// Hash of Timestamp rounded to previous hour
	h.Write([]byte(time.Now().UTC().String()))
	h.Write([]byte(containerName))
	hash := hex.EncodeToString(h.Sum(nil))
	return fmt.Sprintf("%s-%s", containerName, hash)
}

func (uc *interceptorUseCase) generateMetadataForNewImage() *entity.ContainerMetadata {
	lastTimestamp := time.Now()
	// TODO: add a logger to log the error?
	lastRequestSolved, _ := uc.InterceptedRequestRepository.GetLastRequestSolved()

	lastRequestSolvedID := "-1"
	if lastRequestSolved != nil {
		lastRequestSolvedID = lastRequestSolved.ID
	}

	return &entity.ContainerMetadata{
		LastTimestamp:       lastTimestamp,
		LastRequestSolvedID: lastRequestSolvedID,
	}
}

package usecase

import (
	"net/http"

	"github.com/GianOrtiz/k8s-transparent-checkpoint-restore/internal/entity"
)

// InterceptorUseCase declares use cases for the Interceptor. It declares the use cases for intercepting
// requisitions and communicating with the Interceptor.
type InterceptorUseCase interface {
	// InterceptRequest intercepts an HTTP request that should have been sent to the monitored application.
	InterceptRequest(reqID string, req *http.Request) (*http.Response, error)
}

type interceptedRequest struct {
	id      string
	request *http.Request
	solved  bool
}

type interceptorUseCase struct {
	Interceptor       *entity.Interceptor
	CheckpointService entity.CheckpointService
	buffer            map[string]*interceptedRequest
}

func Interceptor(interceptor *entity.Interceptor) (InterceptorUseCase, error) {
	return &interceptorUseCase{
		Interceptor: interceptor,
		buffer:      make(map[string]*interceptedRequest),
	}, nil
}

// InterceptRequest intercepts a given request and return the response after it is
// redirected to the monitored application.
func (uc *interceptorUseCase) InterceptRequest(reqID string, req *http.Request) (*http.Response, error) {
	// Add the request to the buffer. TODO: use a repository to store the buffer
	uc.buffer[reqID] = &interceptedRequest{
		id:      reqID,
		request: req,
		solved:  false,
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

	uc.buffer[reqID].solved = true

	return res, nil
}

// Checkpoint the monitored application into a new image.
func (uc *interceptorUseCase) Checkpoint() error {
	return uc.CheckpointService.Checkpoint(&entity.CheckpointConfig{
		Container:      uc.Interceptor.MonitoredContainer,
		CheckpointHash: uc.generateHashForNewImage(),
		Metadata:       uc.generateMetadataForNewImage(),
	})
}

func (uc *interceptorUseCase) generateHashForNewImage() string {
	panic("not implemented")
}

func (uc *interceptorUseCase) generateMetadataForNewImage() map[string]interface{} {
	panic("not implemented")
}

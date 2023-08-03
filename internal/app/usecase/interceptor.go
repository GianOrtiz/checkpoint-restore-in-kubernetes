package usecase

import (
	"net/http"

	"github.com/GianOrtiz/k8s-transparent-checkpoint-restore/internal/app/entity"
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
	Interceptor *entity.Interceptor
	buffer      map[string]*interceptedRequest
}

func Interceptor(interceptor *entity.Interceptor) (InterceptorUseCase, error) {
	return &interceptorUseCase{
		Interceptor: interceptor,
		buffer:      make(map[string]*interceptedRequest),
	}, nil
}

func (uc *interceptorUseCase) InterceptRequest(reqID string, req *http.Request) (*http.Response, error) {
	uc.buffer[reqID] = &interceptedRequest{
		id:      reqID,
		request: req,
		solved:  false,
	}

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

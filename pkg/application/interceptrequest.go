package application

import (
	"net/http"

	"github.com/GianOrtiz/k8s-transparent-checkpoint-restore/pkg/domain"
	"github.com/google/uuid"
)

type interceptedRequest struct {
	id      string
	request *http.Request
	solved  bool
}

// InterceptRequest is the use case to intercept HTTP requests for the Interceptor.
type InterceptRequest struct {
	Interceptor *domain.Interceptor
	buffer      map[string]*interceptedRequest
}

// Execute the use case for the interception of HTTP requests. It receives the
// request, write it to a buffer, send it to the monitored container, waits for
// the response, marked as solved in the buffer and send the response back.
func (ir *InterceptRequest) Execute(req *http.Request) (*http.Response, error) {
	requestId := uuid.NewString()
	ir.buffer[requestId] = &interceptedRequest{
		id:      requestId,
		request: req,
		solved:  false,
	}

	reqCopy, err := http.NewRequest(req.Method, req.URL.String(), req.Body)
	if err != nil {
		return nil, err
	}

	res, err := http.DefaultClient.Do(reqCopy)
	if err != nil {
		return nil, err
	}

	ir.buffer[requestId].solved = true

	return res, nil
}

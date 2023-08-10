package interceptedrequest

import (
	"time"

	"github.com/GianOrtiz/k8s-transparent-checkpoint-restore/internal/entity"
)

type InMemoryInterceptedRequestRepository struct {
	requests map[string]*entity.InterceptedRequest
}

func InMemory() entity.InterceptedRequestRepository {
	return &InMemoryInterceptedRequestRepository{
		requests: make(map[string]*entity.InterceptedRequest),
	}
}

func (r *InMemoryInterceptedRequestRepository) Save(req *entity.InterceptedRequest) error {
	r.requests[req.ID] = req
	return nil
}

func (r *InMemoryInterceptedRequestRepository) SetSolved(reqID string, solvedAt time.Time, solved bool) error {
	req := r.requests[reqID]
	req.SolvedAt = &solvedAt
	req.Solved = true
	return nil
}

func (r *InMemoryInterceptedRequestRepository) GetLastRequestSolved() (*entity.InterceptedRequest, error) {
	var lastRequest *entity.InterceptedRequest
	for _, req := range r.requests {
		if lastRequest == nil {
			lastRequest = req
		} else {
			if req.SolvedAt.Compare(*lastRequest.SolvedAt) < 0 {
				lastRequest = req
			}
		}
	}
	return lastRequest, nil
}

func (r *InMemoryInterceptedRequestRepository) GetAll() ([]*entity.InterceptedRequest, error) {
	var requests []*entity.InterceptedRequest
	for _, req := range r.requests {
		requests = append(requests, req)
	}
	return requests, nil
}

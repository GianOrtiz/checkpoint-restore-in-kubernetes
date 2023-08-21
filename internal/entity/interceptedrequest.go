package entity

import (
	"net/http"
	"time"
)

// InterceptedRequest is the representation of an HTTP request intercepted by our application.
type InterceptedRequest struct {
	// ID is an unique identifier as UUID of the request, assigned to it when it comes to the application.
	ID string
	// SolvedAt is the datetime the request was solved by the monitored application.
	SolvedAt *time.Time
	// Request the representation of the HTTP request.
	Request *http.Request
	// Solved indicates whether or not the request was already solved.
	Solved bool
	// Version indicates the event version of this request in the event sourcing.
	Version int
}

type InterceptedRequestRepository interface {
	// Save saves the request to the datasource.
	Save(req *InterceptedRequest) error
	// SetSolved set the request as solved in the datasource.
	SetSolved(reqID string, solvedAt time.Time, solved bool) error
	// GetLastRequestSolved gets the last request that was solved by the application.
	GetLastRequestSolved() (*InterceptedRequest, error)
	// GetAll gets all intercepted requests in the datasource.
	GetAll() ([]*InterceptedRequest, error)
	// GetLastVersion gets the last version of request in the event source datasource.
	GetLastVersion() (int, error)
	// GetAllFromLastVersion get all intercepted request in the datasource from the given
	// last version to the newest one.
	GetAllFromLastVersion(version int) ([]*InterceptedRequest, error)
}

package entity

import "time"

type ContainerMetadata struct {
	// LastTimestamp latests timestamp of this container metadata.
	LastTimestamp time.Time `json:"last_timestamp"`
	// LastRequestSolvedID latest request id solved by the Interceptor.
	LastRequestSolvedID string `json:"last_request_solved_id"`
}

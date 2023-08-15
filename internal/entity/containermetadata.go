package entity

import "time"

type ContainerMetadata struct {
	// LastTimestamp latests timestamp of this container metadata.
	LastTimestamp time.Time
	// LastRequestSolvedID latest request id solved by the Interceptor.
	LastRequestSolvedID string
}

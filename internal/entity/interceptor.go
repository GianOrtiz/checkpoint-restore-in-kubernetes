package entity

import "github.com/GianOrtiz/k8s-transparent-checkpoint-restore/internal/config/interceptor"

// Interceptor is the structure of the Ambassador pattern to monitor containers.
// The Interceptor must intercept traffic from the container and save it in a buffer,
// every request will be marked as served once they return from the Container. The
// Interceptor must make checkpoints from the Container following its configuration.
type Interceptor struct {
	// ID is the unique identifier of interceptor as UUID.
	ID string
	// MonitoringContainerID is the id of the container the interceptor must monitor.
	MonitoringContainerID string
	// MonitoredContainer is the information about the monitored container by the interceptor.
	MonitoredContainer *Container
	// Config is the configuration of the Interceptor, containing information like
	// the interval for making checkpoints.
	Config *interceptor.Config
}

// InterceptorRepository is the definition of the data access to the Interceptor.
type InterceptorRepository interface {
	// GetById retrieves an Interceptor by its id.
	GetById(id string) (*Interceptor, error)
	// GetContainers retrieve all containers monitored by the Interceptor.
	GetMonitoredContainer(interceptor *Interceptor) (*Container, error)
	// Create creates a new Interceptor.
	Create(interceptor *Interceptor) error
	// UpdateInterceptorConfig updates the Interceptor configuration.
	UpdateInterceptorConfig(interceptorId string, config *interceptor.Config) error
}

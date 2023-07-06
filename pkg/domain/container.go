package domain

// Container is an abstraction of every application running in a container.
type Container struct {
	// ID is the unique identifier of the container in UUID.
	ID string
	// HTTPUrl is the URL to access the container to monitor.
	HTTPUrl string
}

// ContainerRepository is the definition of data access to Container.
type ContainerRepository interface {
	// GetByID gets a Container by its id.
	GetByID(id string) (*Container, error)
	// Create creates a new Container.
	Create(container *Container) error
}

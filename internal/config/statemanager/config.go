package statemanager

// StateManagerConfig defines the state manager configuration.
type StateManagerConfig struct {
	// Flag to either enable or disable development features of state manager, like
	// restore endpoint o create a new restore from an image checkpoint hash.
	DevelopmentFeaturesEnabled bool
}

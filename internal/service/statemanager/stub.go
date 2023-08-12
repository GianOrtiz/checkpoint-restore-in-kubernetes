package statemanager

type alwaysAcceptingStateManagerStub struct{}

func AlawaysAcceptingStub() *alwaysAcceptingStateManagerStub {
	return &alwaysAcceptingStateManagerStub{}
}

func (stateManager *alwaysAcceptingStateManagerStub) SaveMetadata(containerName string, metadata map[string]interface{}) error {
	return nil
}

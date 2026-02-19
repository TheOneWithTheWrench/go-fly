package internal

//go:generate moq -out mock_test.go -pkg internal_test . IndexStorage RemoteStorage RemoteFetcher Picker RefreshLauncher PruneStateStorage PruneLauncher Cloner Runner

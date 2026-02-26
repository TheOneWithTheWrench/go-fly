package internal

//go:generate moq -out mock_test.go -pkg internal_test . RemoteFetcher Picker Refresher Pruner Cloner Runner Source Refreshable Trackable Prunable LocalCleaner

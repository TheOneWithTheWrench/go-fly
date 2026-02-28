package remote

//go:generate moq -out mock_test.go -pkg remote_test . Fetcher
//go:generate moq -out cloner_mock_test.go -pkg remote . commandRunner

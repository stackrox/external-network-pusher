package commons

import (
	"fmt"
)

// FailedToCrawlSomeProviders returns an error to be raised when failed to crawl some providers
func FailedToCrawlSomeProviders(failedProviders []string) error {
	return fmt.Errorf(
		"failed to crawl some of the providers specified: %v. Please refer to logs for further debugging",
		failedProviders)
}

// FailedToCrawlAllProviders returns an error to be raised when crawler failed to crawl all providers
func FailedToCrawlAllProviders() error {
	return fmt.Errorf("failed to crawl all of the providers specified. Please look at logs for further debugging")
}

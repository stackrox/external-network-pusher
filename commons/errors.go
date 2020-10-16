package commons

import (
	"fmt"
)

func FailedToCrawlSomeProviders(failedProviders []string) error {
	return fmt.Errorf(
		"failed to crawl some of the providers specified: %v. Please refer to logs for further debugging",
		failedProviders)
}

func FailedToCrawlAllProviders() error {
	return fmt.Errorf("failed to crawl all of the providers specified. Please look at logs for further debugging")
}

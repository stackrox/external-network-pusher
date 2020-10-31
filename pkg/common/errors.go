package common

import (
	"errors"
	"fmt"
)

// NoProvidersCrawledError is returned when there is no provider successfully crawled
func NoProvidersCrawledError() error {
	return errors.New("external network sources empty. failed to crawl any provider")
}

// ProviderNameEmptyError is returned when an empty provider name is found
func ProviderNameEmptyError() error {
	return errors.New("provider name is empty")
}

// NoRegionNetworksError is returned when a provider does not have any region crawled
func NoRegionNetworksError(providerName string) error {
	return fmt.Errorf("provider %s does not have any region associated with it", providerName)
}

// EmptyRegionNameError is returned when an empty region name is found
func EmptyRegionNameError(providerName string) error {
	return fmt.Errorf("provider %s has an empty region name", providerName)
}

// NoServiceNetworksError is returned when a region does not have any service crawled
func NoServiceNetworksError(providerName, regionName string) error {
	return fmt.Errorf("provider %s has a region %s with no service names", providerName, regionName)
}

// EmptyServiceNameError is returned when an empty service name is found
func EmptyServiceNameError(providerName, regionName string) error {
	return fmt.Errorf("provider %s has a region %s with an empty service name", providerName, regionName)
}

// NoIPPrefixesError is returned when a service does not have any IP prefix crawled
func NoIPPrefixesError(providerName, regionName, serviceName string) error {
	return fmt.Errorf(
		"provider %s at region %s with service %s does not have any IP prefix",
		providerName,
		regionName,
		serviceName)
}

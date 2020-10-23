package common

import (
	"fmt"
)

// DefaultRegion is used when a vendor does not
// specify regions for the IP ranges provided
const DefaultRegion = "default"

// HeaderFileName defines the name of the header file
// uploaded to bucket. When uploading header file,
// we also append the hash of the header file to its
// name.
const HeaderFileName = "header"

// Provider is a string representing different external network providers
type Provider string

const (
	// Google is provider "enum" for Google Cloud
	Google Provider = "Google"
	// Azure is provider "enum" for Microsoft Azure Cloud
	Azure Provider = "Azure"
)

// AllProviders returns all the providers available
func AllProviders() []Provider {
	return []Provider{Google, Azure}
}

func (p Provider) String() string {
	return string(p)
}

// ToProvider converts a string representation of a provider
// to Provider type
func ToProvider(s string) (Provider, error) {
	switch s {
	case Google.String():
		return Google, nil
	case Azure.String():
		return Azure, nil
	default:
		return "", fmt.Errorf("invalid Provider: %s", s)
	}
}

// ProviderToURLs is a mapping from provider to its crawler endpoint.
// It is kept here for easier maintenance.
var ProviderToURLs = map[Provider][]string{
	Google: {"https://www.gstatic.com/ipranges/cloud.json"},
	Azure: {
		// Azure Public
		"https://www.microsoft.com/en-us/download/confirmation.aspx?id=56519",
		// Azure US Gov
		"https://www.microsoft.com/en-us/download/confirmation.aspx?id=57063",
		// Azure China
		"https://www.microsoft.com/en-us/download/confirmation.aspx?id=57062",
		// Azure Germany
		"https://www.microsoft.com/en-us/download/confirmation.aspx?id=57064",
	},
}

package common

// DefaultRegion is used when a vendor does not
// specify regions for the IP ranges provided
const DefaultRegion = "default"

// HeaderFileName defines the name of the header file
// uploaded to bucket. When uploading header file,
// we also append the hash of the header file to its
// name.
const HeaderFileName = "header"

// Provider is an enum representing different external network providers
type Provider int

const (
	// Google is provider enum for Google Cloud
	Google Provider = iota
)

// ProviderToURL is a mapping from provider to its crawler endpoint.
// It is kept here for easier maintenance.
var ProviderToURL = map[Provider]string{
	Google: "https://www.gstatic.com/ipranges/cloud.json",
}

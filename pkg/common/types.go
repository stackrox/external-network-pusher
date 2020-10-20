package common

// NetworkCrawler defines an interface for the implementation
// of Provider specific network range crawlers
type NetworkCrawler interface {
	CrawlPublicNetworkRanges() (*PublicNetworkRanges, error)
	GetHumanReadableProviderName() string
	GetObjectName() string
	GetProviderKey() Provider
}

// ServiceIPRanges contain all the IP ranges used for a specific service
// under a service region of a specific Provider
type ServiceIPRanges struct {
	// Sample IPv4 prefix: 8.8.0.0/16
	IPv4Prefixes []string `json:"ipv4Prefixes"`
	// Sample IPv6 prefix: 2600:1901::/48
	IPv6Prefixes []string `json:"ipv6Prefixes"`
}

// RegionNetworkDetail contains mapping from all the service names under a region to ServiceIPRanges
type RegionNetworkDetail struct {
	ServiceNameToIPRanges map[string]ServiceIPRanges `json:"serviceNameToIpRanges"`
}

// PublicNetworkRanges contains mapping from region names to RegionNetworkDetail
type PublicNetworkRanges struct {
	RegionToNetworkDetails map[string]RegionNetworkDetail `json:"regionToNetworkDetails"`
}

// Header defines the structure of the header file
type Header struct {
	ObjectPrefix         string            `json:"objectPrefix"`
	ObjectNameToCheckSum map[string]string `json:"objectNameToChecksum"`
}

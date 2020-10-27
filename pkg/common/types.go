package common

// NetworkCrawler defines an interface for the implementation
// of Provider specific network range crawlers
type NetworkCrawler interface {
	CrawlPublicNetworkRanges() (*ProviderNetworkRanges, error)
	GetHumanReadableProviderName() string
	GetProviderKey() Provider
}

// ServiceIPRanges contains all the IP ranges used by a specific service
type ServiceIPRanges struct {
	// ServiceName denotes the service name for the IP ranges
	ServiceName string `json:"serviceName"`
	// Sample IPv4 prefix: 8.8.0.0/16
	IPv4Prefixes []string `json:"ipv4Prefixes"`
	// Sample IPv6 prefix: 2600:1901::/48
	IPv6Prefixes []string `json:"ipv6Prefixes"`
}

// RegionNetworkDetail contains all the networks of services under a region
type RegionNetworkDetail struct {
	RegionName      string             `json:"regionName"`
	ServiceNetworks []*ServiceIPRanges `json:"serviceNetworks"`
}

// ProviderNetworkRanges contains networks for all regions of a provider
type ProviderNetworkRanges struct {
	ProviderName   string                 `json:"providerName"`
	RegionNetworks []*RegionNetworkDetail `json:"regionNetworks"`
}

// ExternalNetworkSources contains all the external networks for all providers
type ExternalNetworkSources struct {
	ProviderNetworks []*ProviderNetworkRanges `json:"providerNetworks"`
}

// AddIPv4Prefix adds the specified IPv4 prefix to the region and service name pair
func (p *ProviderNetworkRanges) AddIPv4Prefix(region, service, ipv4 string) {
	p.addIPPrefix(region, service, ipv4, true)
}

// AddIPv6Prefix adds the specified IPv6 prefix to the region and service name pair
func (p *ProviderNetworkRanges) AddIPv6Prefix(region, service, ipv6 string) {
	p.addIPPrefix(region, service, ipv6, false)
}

func (p *ProviderNetworkRanges) addIPPrefix(region, service, ip string, isIPv4 bool) {
	var regionNetwork *RegionNetworkDetail
	for _, network := range p.RegionNetworks {
		if network.RegionName == region {
			regionNetwork = network
			break
		}
	}
	if regionNetwork == nil {
		// Never seen this region before
		regionNetwork = &RegionNetworkDetail{RegionName: region}
		p.RegionNetworks = append(p.RegionNetworks, regionNetwork)
	}

	var serviceIPRanges *ServiceIPRanges
	for _, ips := range regionNetwork.ServiceNetworks {
		if ips.ServiceName == service {
			serviceIPRanges = ips
			break
		}
	}
	if serviceIPRanges == nil {
		// Never seen this service before
		serviceIPRanges = &ServiceIPRanges{ServiceName: service}
		regionNetwork.ServiceNetworks = append(regionNetwork.ServiceNetworks, serviceIPRanges)
	}

	if isIPv4 {
		serviceIPRanges.IPv4Prefixes = append(serviceIPRanges.IPv4Prefixes, ip)
	} else {
		serviceIPRanges.IPv6Prefixes = append(serviceIPRanges.IPv6Prefixes, ip)
	}
}

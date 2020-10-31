package common

import (
	"net"

	"github.com/pkg/errors"
)

// NetworkCrawler defines an interface for the implementation
// of Provider specific network range crawlers
type NetworkCrawler interface {
	CrawlPublicNetworkRanges() (*ProviderNetworkRanges, error)
	GetHumanReadableProviderName() string
	GetProviderKey() Provider
	// GetNumRequiredIPPrefixes returns number of required IP prefixes crawled by crawler
	// Used during validation of crawler outputs.
	GetNumRequiredIPPrefixes() int
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

// AddIPPrefix adds the specified IP prefix to the region and service name pair
// returns error if the IP given is not a valid IP prefix
func (p *ProviderNetworkRanges) AddIPPrefix(region, service, ipPrefix string) error {
	ip, prefix, err := net.ParseCIDR(ipPrefix)
	if err != nil || ip == nil || prefix == nil {
		return errors.Wrapf(err, "failed to parse address: %s", ip)
	}
	if ip.To4() != nil {
		p.addIPPrefix(region, service, ipPrefix, true)
	} else {
		p.addIPPrefix(region, service, ipPrefix, false)
	}
	return nil
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

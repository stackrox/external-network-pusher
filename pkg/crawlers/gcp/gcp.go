package gcp

import (
	"encoding/json"
	"log"

	"github.com/stackrox/external-network-pusher/pkg/commons"
	"github.com/stackrox/external-network-pusher/pkg/commons/utils"
)

type GcpIpSpec struct {
	Ipv4Prefix string      `json:"ipv4Prefix"`
	Ipv6Prefix string      `json:"ipv6Prefix"`
	Service string         `json:"service"`
	Scope string           `json:"scope"`
}

type GcpNetworkSpec struct {
	SyncToken string     `json:"syncToken"`
	CreationTime string  `json:"creationTime"`
	Prefixes []GcpIpSpec `json:"prefixes"`
}

type GcpNetworkCrawler struct {
	url string
}

func NewGcpNetworkCrawler () commons.NetworkCrawler {
	return &GcpNetworkCrawler{url: commons.ProviderToUrl[commons.GOOGLE]}
}

func (crawler *GcpNetworkCrawler) GetHumanReadableProviderName() string {
	return "Google Cloud"
}

func (crawler *GcpNetworkCrawler) GetObjectName() string {
	return "google-cloud-networks"
}

func (crawler *GcpNetworkCrawler) GetProviderKey() commons.Provider {
	return commons.GOOGLE
}

func (crawler *GcpNetworkCrawler) CrawlPublicNetworkRanges() (*commons.PublicNetworkRanges, error) {
	networkData, err := crawler.fetch()
	if err != nil {
		log.Printf("Failed to fetch Google network data: %+v", err)
		return nil, err
	}

	parsed, err := crawler.parseNetworks(networkData)
	if err != nil {
		log.Printf("Failed to parse Google network data: %+v", err)
		return nil, err
	}

	return parsed, nil
}

func (crawler *GcpNetworkCrawler) fetch() ([]byte, error){
	body, err := utils.HttpGet(crawler.url)
	if err != nil {
		log.Printf("Received error while trying to fetch network from Google: %+v", err)
		return nil, err
	}
	return body, nil
}

func (crawler *GcpNetworkCrawler) parseNetworks(data []byte) (*commons.PublicNetworkRanges, error) {
	var gcpNetworkSpec GcpNetworkSpec
	err := json.Unmarshal(data, &gcpNetworkSpec)
	if err != nil {
		log.Printf("Failed to parse GCP network data with error: %+v", err)
		return nil, err
	}

	regionToNetworkDetails := make(map[string]commons.RegionNetworkDetail)
	for _, gcpIpSpec := range gcpNetworkSpec.Prefixes {
		if gcpIpSpec.Ipv4Prefix == "" && gcpIpSpec.Ipv6Prefix == "" {
			continue
		}

		// Get region network spec
		var regionNetworks commons.RegionNetworkDetail
		if networks, ok := regionToNetworkDetails[gcpIpSpec.Scope]; ok {
			regionNetworks = networks
		} else {
			regionNetworks.ServiceNameToIpRanges = make(map[string]commons.ServiceIpRanges)
		}

		// Get service
		var serviceIpRanges commons.ServiceIpRanges
		if ips, ok := regionNetworks.ServiceNameToIpRanges[gcpIpSpec.Service]; ok {
			serviceIpRanges = ips
		}

		if gcpIpSpec.Ipv4Prefix != "" {
			serviceIpRanges.Ipv4Prefixes = append(serviceIpRanges.Ipv4Prefixes, gcpIpSpec.Ipv4Prefix)
		}
		if gcpIpSpec.Ipv6Prefix != "" {
			serviceIpRanges.Ipv6Prefixes = append(serviceIpRanges.Ipv6Prefixes, gcpIpSpec.Ipv6Prefix)
		}

		regionNetworks.ServiceNameToIpRanges[gcpIpSpec.Service] = serviceIpRanges
		regionToNetworkDetails[gcpIpSpec.Scope] = regionNetworks
	}

	return &commons.PublicNetworkRanges{RegionToNetworkDetails: regionToNetworkDetails}, nil
}
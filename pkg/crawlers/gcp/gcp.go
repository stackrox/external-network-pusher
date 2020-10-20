package gcp

import (
	"encoding/json"
	"log"

	"github.com/stackrox/external-network-pusher/pkg/commons"
	"github.com/stackrox/external-network-pusher/pkg/commons/utils"
)

type gcpIPSpec struct {
	Ipv4Prefix string `json:"ipv4Prefix"`
	Ipv6Prefix string `json:"ipv6Prefix"`
	Service    string `json:"service"`
	Scope      string `json:"scope"`
}

type gcpNetworkSpec struct {
	SyncToken    string      `json:"syncToken"`
	CreationTime string      `json:"creationTime"`
	Prefixes     []gcpIPSpec `json:"prefixes"`
}

type gcpNetworkCrawler struct {
	url string
}

// NewGcpNetworkCrawler returns an instance of the gcpNetworkCrawler
func NewGcpNetworkCrawler() commons.NetworkCrawler {
	return &gcpNetworkCrawler{url: commons.ProviderToURL[commons.GOOGLE]}
}

func (crawler *gcpNetworkCrawler) GetHumanReadableProviderName() string {
	return "Google Cloud"
}

func (crawler *gcpNetworkCrawler) GetObjectName() string {
	return "google-cloud-networks"
}

func (crawler *gcpNetworkCrawler) GetProviderKey() commons.Provider {
	return commons.GOOGLE
}

func (crawler *gcpNetworkCrawler) CrawlPublicNetworkRanges() (*commons.PublicNetworkRanges, error) {
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

func (crawler *gcpNetworkCrawler) fetch() ([]byte, error) {
	body, err := utils.HTTPGet(crawler.url)
	if err != nil {
		log.Printf("Received error while trying to fetch network from Google: %+v", err)
		return nil, err
	}
	return body, nil
}

func (crawler *gcpNetworkCrawler) parseNetworks(data []byte) (*commons.PublicNetworkRanges, error) {
	var gcpNetworkSpec gcpNetworkSpec
	err := json.Unmarshal(data, &gcpNetworkSpec)
	if err != nil {
		log.Printf("Failed to parse GCP network data with error: %+v", err)
		return nil, err
	}

	regionToNetworkDetails := make(map[string]commons.RegionNetworkDetail)
	for _, gcpIPSpec := range gcpNetworkSpec.Prefixes {
		if gcpIPSpec.Ipv4Prefix == "" && gcpIPSpec.Ipv6Prefix == "" {
			continue
		}

		// Get region network spec
		var regionNetworks commons.RegionNetworkDetail
		if networks, ok := regionToNetworkDetails[gcpIPSpec.Scope]; ok {
			regionNetworks = networks
		} else {
			regionNetworks.ServiceNameToIPRanges = make(map[string]commons.ServiceIPRanges)
		}

		// Get service
		var serviceIPRanges commons.ServiceIPRanges
		if ips, ok := regionNetworks.ServiceNameToIPRanges[gcpIPSpec.Service]; ok {
			serviceIPRanges = ips
		}

		if gcpIPSpec.Ipv4Prefix != "" {
			serviceIPRanges.Ipv4Prefixes = append(serviceIPRanges.Ipv4Prefixes, gcpIPSpec.Ipv4Prefix)
		}
		if gcpIPSpec.Ipv6Prefix != "" {
			serviceIPRanges.Ipv6Prefixes = append(serviceIPRanges.Ipv6Prefixes, gcpIPSpec.Ipv6Prefix)
		}

		regionNetworks.ServiceNameToIPRanges[gcpIPSpec.Service] = serviceIPRanges
		regionToNetworkDetails[gcpIPSpec.Scope] = regionNetworks
	}

	return &commons.PublicNetworkRanges{RegionToNetworkDetails: regionToNetworkDetails}, nil
}
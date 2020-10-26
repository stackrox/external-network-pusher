package gcp

import (
	"encoding/json"

	"github.com/pkg/errors"
	"github.com/stackrox/external-network-pusher/pkg/common"
	"github.com/stackrox/external-network-pusher/pkg/common/utils"
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

// NewGCPNetworkCrawler returns an instance of the gcpNetworkCrawler
func NewGCPNetworkCrawler() common.NetworkCrawler {
	return &gcpNetworkCrawler{url: common.ProviderToURLs[common.Google][0]}
}

func (c *gcpNetworkCrawler) GetHumanReadableProviderName() string {
	return "Google Cloud"
}

func (c *gcpNetworkCrawler) GetBucketObjectName() string {
	return "google-cloud-networks"
}

func (c *gcpNetworkCrawler) GetProviderKey() common.Provider {
	return common.Google
}

func (c *gcpNetworkCrawler) CrawlPublicNetworkRanges() (*common.PublicNetworkRanges, error) {
	networkData, err := c.fetch()
	if err != nil {
		return nil, errors.Wrap(err, "failed to fetch network data while crawling Google's network ranges")
	}

	parsed, err := c.parseNetworks(networkData)
	if err != nil {
		return nil, errors.Wrap(err, "failed to parse Google's network data")
	}

	return parsed, nil
}

func (c *gcpNetworkCrawler) fetch() ([]byte, error) {
	body, err := utils.HTTPGet(c.url)
	if err != nil {
		return nil, errors.Wrap(err, "failed to fetch networks from Google")
	}
	return body, nil
}

func (c *gcpNetworkCrawler) parseNetworks(data []byte) (*common.PublicNetworkRanges, error) {
	var gcpNetworkSpec gcpNetworkSpec
	err := json.Unmarshal(data, &gcpNetworkSpec)
	if err != nil {
		return nil, errors.Wrap(err, "failed to unmarshal Google's network data")
	}

	regionToNetworkDetails := make(map[string]common.RegionNetworkDetail)
	for _, gcpIPSpec := range gcpNetworkSpec.Prefixes {
		if gcpIPSpec.Ipv4Prefix == "" && gcpIPSpec.Ipv6Prefix == "" {
			continue
		}

		// Get region network spec
		var regionNetworks common.RegionNetworkDetail
		if networks, ok := regionToNetworkDetails[gcpIPSpec.Scope]; ok {
			regionNetworks = networks
		} else {
			// Never seen this region before. Create this region.
			regionNetworks =
				common.RegionNetworkDetail{
					ServiceNameToIPRanges: make(map[string]common.ServiceIPRanges),
				}
		}

		// Get service
		var serviceIPRanges common.ServiceIPRanges
		if ips, ok := regionNetworks.ServiceNameToIPRanges[gcpIPSpec.Service]; ok {
			serviceIPRanges = ips
		}

		if gcpIPSpec.Ipv4Prefix != "" {
			serviceIPRanges.IPv4Prefixes = append(serviceIPRanges.IPv4Prefixes, gcpIPSpec.Ipv4Prefix)
		}
		if gcpIPSpec.Ipv6Prefix != "" {
			serviceIPRanges.IPv6Prefixes = append(serviceIPRanges.IPv6Prefixes, gcpIPSpec.Ipv6Prefix)
		}

		regionNetworks.ServiceNameToIPRanges[gcpIPSpec.Service] = serviceIPRanges
		regionToNetworkDetails[gcpIPSpec.Scope] = regionNetworks
	}

	return &common.PublicNetworkRanges{RegionToNetworkDetails: regionToNetworkDetails}, nil
}

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

func (c *gcpNetworkCrawler) GetProviderKey() common.Provider {
	return common.Google
}

func (c *gcpNetworkCrawler) GetNumRequiredIPPrefixes() int {
	// Observed from past .json. In past we had 384 IP prefixes
	return 350
}

func (c *gcpNetworkCrawler) CrawlPublicNetworkRanges() (*common.ProviderNetworkRanges, error) {
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

func (c *gcpNetworkCrawler) parseNetworks(data []byte) (*common.ProviderNetworkRanges, error) {
	var gcpNetworkSpec gcpNetworkSpec
	err := json.Unmarshal(data, &gcpNetworkSpec)
	if err != nil {
		return nil, errors.Wrap(err, "failed to unmarshal Google's network data")
	}

	providerNetworks := common.NewProviderNetworkRanges(c.GetProviderKey().String())
	for _, gcpIPSpec := range gcpNetworkSpec.Prefixes {
		if gcpIPSpec.Ipv4Prefix == "" && gcpIPSpec.Ipv6Prefix == "" {
			continue
		}
		if gcpIPSpec.Ipv4Prefix != "" {
			err :=
				providerNetworks.AddIPPrefix(
					gcpIPSpec.Scope,
					gcpIPSpec.Service,
					gcpIPSpec.Ipv4Prefix,
					c.getComputeRedundancyFn())
			if err != nil {
				return nil, errors.Wrapf(err, "failed to add Google IPv4 Prefix: %s", gcpIPSpec.Ipv4Prefix)
			}
		}
		if gcpIPSpec.Ipv6Prefix != "" {
			err :=
				providerNetworks.AddIPPrefix(
					gcpIPSpec.Scope,
					gcpIPSpec.Service,
					gcpIPSpec.Ipv6Prefix,
					c.getComputeRedundancyFn())
			if err != nil {
				return nil, errors.Wrapf(err, "failed to add Google IPv6 Prefix: %s", gcpIPSpec.Ipv6Prefix)
			}
		}
	}

	return providerNetworks, nil
}

func (c *gcpNetworkCrawler) getComputeRedundancyFn() common.IsRedundantRegionServicePairFn {
	return common.GetDefaultRegionServicePairRedundancyCheck()
}

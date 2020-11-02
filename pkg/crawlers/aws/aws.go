package aws

import (
	"encoding/json"
	"log"

	"github.com/pkg/errors"
	"github.com/stackrox/external-network-pusher/pkg/common"
	"github.com/stackrox/external-network-pusher/pkg/common/utils"
)

type awsIPv4Spec struct {
	IPPrefix           string `json:"ip_prefix"`
	Region             string `json:"region"`
	NetworkBorderGroup string `json:"network_border_group"`
	Service            string `json:"service"`
}

type awsIPv6Spec struct {
	IPv6Prefix         string `json:"ipv6_prefix"`
	Region             string `json:"region"`
	NetworkBorderGroup string `json:"network_border_group"`
	Service            string `json:"service"`
}

type awsNetworkSpec struct {
	SyncToken    string        `json:"syncToken"`
	CreateDate   string        `json:"createDate"`
	Prefixes     []awsIPv4Spec `json:"prefixes"`
	IPv6Prefixes []awsIPv6Spec `json:"ipv6_prefixes"`
}

type awsNetworkCrawler struct {
	url string
}

// NewAWSNetworkCrawler returns an instance of the awsNetworkCrawler
func NewAWSNetworkCrawler() common.NetworkCrawler {
	return &awsNetworkCrawler{url: common.ProviderToURLs[common.Amazon][0]}
}

func (c *awsNetworkCrawler) GetHumanReadableProviderName() string {
	return "Amazon"
}

func (c *awsNetworkCrawler) GetProviderKey() common.Provider {
	return common.Amazon
}

func (c *awsNetworkCrawler) GetNumRequiredIPPrefixes() int {
	// Observed from past .json. In past we had 4217
	return 4000
}

func (c *awsNetworkCrawler) CrawlPublicNetworkRanges() (*common.ProviderNetworkRanges, error) {
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

func (c *awsNetworkCrawler) fetch() ([]byte, error) {
	body, err := utils.HTTPGet(c.url)
	if err != nil {
		return nil, errors.Wrap(err, "failed to fetch networks from Google")
	}
	return body, nil
}

func (c *awsNetworkCrawler) parseNetworks(data []byte) (*common.ProviderNetworkRanges, error) {
	var awsNetworkSpec awsNetworkSpec
	err := json.Unmarshal(data, &awsNetworkSpec)
	if err != nil {
		return nil, errors.Wrap(err, "failed to unmarshal Amazon's network data")
	}

	providerNetworks := common.ProviderNetworkRanges{ProviderName: c.GetProviderKey().String()}
	for _, ipv4Spec := range awsNetworkSpec.Prefixes {
		if ipv4Spec.IPPrefix == "" {
			// Empty IPv4. Something might be wrong here. Logging for warning
			log.Printf("Received an empty IPv4 definition: %v", ipv4Spec)
			continue
		}
		err := providerNetworks.AddIPPrefix(ipv4Spec.Region, ipv4Spec.Service, ipv4Spec.IPPrefix)
		if err != nil {
			return nil, errors.Wrapf(err, "failed to add Amazon IPv4 prefix: %s", ipv4Spec.IPPrefix)
		}
	}
	for _, ipv6Spec := range awsNetworkSpec.IPv6Prefixes {
		if ipv6Spec.IPv6Prefix == "" {
			// Empty IPv6. Something might be wrong here. Logging for warning
			log.Printf("Received an empty IPv6 definition: %v", ipv6Spec)
			continue
		}
		err := providerNetworks.AddIPPrefix(ipv6Spec.Region, ipv6Spec.Service, ipv6Spec.IPv6Prefix)
		if err != nil {
			return nil, errors.Wrapf(err, "failed to add Amazon IPv6 prefix: %s", ipv6Spec.IPv6Prefix)
		}
	}

	return &providerNetworks, nil
}

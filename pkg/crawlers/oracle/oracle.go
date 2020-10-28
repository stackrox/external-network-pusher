package oracle

import (
	"encoding/json"
	"sort"

	"github.com/pkg/errors"
	"github.com/stackrox/external-network-pusher/pkg/common"
	"github.com/stackrox/external-network-pusher/pkg/common/utils"
)

type ociNetworkCrawler struct {
	url string
}

type ociCIDRDefinition struct {
	CIDR string   `json:"cidr"`
	Tags []string `json:"tags"`
}

type ociRegionNetworkDetails struct {
	Region string              `json:"region"`
	CIDRs  []ociCIDRDefinition `json:"cidrs"`
}

type ociNetworkSpec struct {
	LastUpdatedTimestamp string                    `json:"last_updated_timestamp"`
	Regions              []ociRegionNetworkDetails `json:"regions"`
}

// NewOCINetworkCrawler returns an instance of the ociNetworkCrawler
func NewOCINetworkCrawler() common.NetworkCrawler {
	return &ociNetworkCrawler{url: common.ProviderToURLs[common.Oracle][0]}
}

func (c *ociNetworkCrawler) GetHumanReadableProviderName() string {
	return "Oracle Cloud Platform"
}

func (c *ociNetworkCrawler) GetProviderKey() common.Provider {
	return common.Oracle
}

func (c *ociNetworkCrawler) CrawlPublicNetworkRanges() (*common.ProviderNetworkRanges, error) {
	networkData, err := c.fetch()
	if err != nil {
		return nil, errors.Wrapf(
			err,
			"failed to fetch network data while crawling %s's network ranges",
			c.GetHumanReadableProviderName())
	}

	parsed, err := c.parseNetworks(networkData)
	if err != nil {
		return nil, errors.Wrapf(err, "failed to parse %s's network data", c.GetHumanReadableProviderName())
	}

	return parsed, nil
}

func (c *ociNetworkCrawler) fetch() ([]byte, error) {
	body, err := utils.HTTPGet(c.url)
	if err != nil {
		return nil, errors.Wrapf(err, "failed to fetch networks from %s", c.GetHumanReadableProviderName())
	}
	return body, nil
}

func (c *ociNetworkCrawler) parseNetworks(data []byte) (*common.ProviderNetworkRanges, error) {
	var ociNetworkSpec ociNetworkSpec
	err := json.Unmarshal(data, &ociNetworkSpec)
	if err != nil {
		return nil, errors.Wrapf(err, "failed to unmarshal %s's network data", c.GetHumanReadableProviderName())
	}

	providerNetworks := common.ProviderNetworkRanges{ProviderName: c.GetProviderKey().String()}
	for _, regionNetworks := range ociNetworkSpec.Regions {
		for _, cidrDef := range regionNetworks.CIDRs {
			// sort the tags before creating service name to make service name consistent
			service := toServiceName(cidrDef.Tags)
			err := providerNetworks.AddIPPrefix(regionNetworks.Region, service, cidrDef.CIDR)
			if err != nil {
				return nil, errors.Wrapf(
					err,
					"failed to add %s's IP prefix: %s",
					c.GetHumanReadableProviderName(),
					cidrDef.CIDR)
			}
		}
	}

	return &providerNetworks, nil
}

func toServiceName(tags []string) string {
	sort.Strings(tags)
	return utils.ToCompoundName(tags)
}

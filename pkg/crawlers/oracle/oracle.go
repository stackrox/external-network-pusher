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

func (c *ociNetworkCrawler) GetNumRequiredIPPrefixes() int {
	// Observed from past .json. In past we had 258
	return 200
}

func (c *ociNetworkCrawler) CrawlPublicNetworkRanges() (*common.ProviderNetworkRanges, error) {
	networkData, err := c.fetch()
	if err != nil {
		return nil, errors.Wrap(err, "failed to fetch network data while crawling Oracle's network ranges")
	}

	parsed, err := c.parseNetworks(networkData)
	if err != nil {
		return nil, errors.Wrap(err, "failed to parse Oracle's network data")
	}

	return parsed, nil
}

func (c *ociNetworkCrawler) fetch() ([]byte, error) {
	body, err := utils.HTTPGet(c.url)
	if err != nil {
		return nil, errors.Wrap(err, "failed to fetch networks from Oracle")
	}
	return body, nil
}

func (c *ociNetworkCrawler) parseNetworks(data []byte) (*common.ProviderNetworkRanges, error) {
	var ociNetworkSpec ociNetworkSpec
	err := json.Unmarshal(data, &ociNetworkSpec)
	if err != nil {
		return nil, errors.Wrap(err, "failed to unmarshal Oracle's network data")
	}

	providerNetworks := common.NewProviderNetworkRanges(c.GetProviderKey().String())
	for _, regionNetworks := range ociNetworkSpec.Regions {
		for _, cidrDef := range regionNetworks.CIDRs {
			// sort the tags before creating service name to make service name consistent
			service := toServiceName(cidrDef.Tags)
			err :=
				providerNetworks.AddIPPrefix(regionNetworks.Region, service, cidrDef.CIDR, c.getComputeRedundancyFn())
			if err != nil {
				return nil, errors.Wrapf(
					err,
					"failed to add Oracle's IP prefix: %s",
					cidrDef.CIDR)
			}
		}
	}

	return providerNetworks, nil
}

func toServiceName(tags []string) string {
	sort.Strings(tags)
	// Using "|" as deliminator since these are tags and tag1|tag2 as a service name
	// seems the best way to represent them, instead of "/" since this could potentially
	// be understood in a way that the tags have some sort of hierarchical relationships
	// between them.
	return utils.ToCompoundName("|", tags...)
}

func (c *ociNetworkCrawler) getComputeRedundancyFn() common.IsRedundantRegionServicePairFn {
	return common.GetDefaultRegionServicePairRedundancyCheck()
}

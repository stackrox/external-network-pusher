package cloudflare

import (
	"encoding/json"
	"strings"

	"github.com/pkg/errors"
	"github.com/stackrox/external-network-pusher/pkg/common"
	"github.com/stackrox/external-network-pusher/pkg/common/utils"
)

type cloudflareNetworkCrawler struct {
	url string
}

type cloudflareNetworkResult struct {
	// Note: the slashes within ipv4CIDRs and ipv6CIDRs
	// are escaped for some reason.
	IPv4CIDRs []string `json:"ipv4_cidrs"`
	IPv6CIDRs []string `json:"ipv6_cidrs"`
	ETag      string   `json:"etag"`
}

type cloudflareNetworkSpec struct {
	Result   cloudflareNetworkResult `json:"result"`
	Success  bool                    `json:"success"`
	Errors   []string                `json:"errors"`
	Messages []string                `json:"messages"`
}

// NewCloudflareNetworkCrawler returns an instance of the cloudflareNetworkCrawler
func NewCloudflareNetworkCrawler() common.NetworkCrawler {
	return &cloudflareNetworkCrawler{url: common.ProviderToURLs[common.Cloudflare][0]}
}

func (c *cloudflareNetworkCrawler) GetHumanReadableProviderName() string {
	return "Cloudflare"
}

func (c *cloudflareNetworkCrawler) GetProviderKey() common.Provider {
	return common.Cloudflare
}

func (c *cloudflareNetworkCrawler) GetNumRequiredIPPrefixes() int {
	// Observed from past .json. In past we had 21
	return 15
}

func (c *cloudflareNetworkCrawler) CrawlPublicNetworkRanges() (*common.ProviderNetworkRanges, error) {
	networkData, err := c.fetch()
	if err != nil {
		return nil, errors.Wrap(err, "failed to fetch network data while crawling Cloudflare's network ranges")
	}

	parsed, err := c.parseNetworks(networkData)
	if err != nil {
		return nil, errors.Wrap(err, "failed to parse Cloudflare's network data")
	}
	return parsed, nil
}

func (c *cloudflareNetworkCrawler) fetch() ([]byte, error) {
	return utils.HTTPGetWithRetry("Cloudflare", c.url)
}

func (c *cloudflareNetworkCrawler) parseNetworks(networks []byte) (*common.ProviderNetworkRanges, error) {
	var cloudflareNetworkSpec cloudflareNetworkSpec
	err := json.Unmarshal(networks, &cloudflareNetworkSpec)
	if err != nil {
		return nil, errors.Wrap(err, "failed to unmarshal Cloudflare's network data")
	}

	providerNetworks := common.NewProviderNetworkRanges(c.GetProviderKey().String())
	for _, ipv4Str := range cloudflareNetworkSpec.Result.IPv4CIDRs {
		ipv4Str = unescapeIPPrefix(ipv4Str)
		err :=
			providerNetworks.AddIPPrefix(
				common.DefaultRegion,
				common.DefaultService,
				ipv4Str,
				c.getComputeRedundancyFn())
		if err != nil {
			return nil, errors.Wrapf(err, "failed to add IPv4 prefix: %s to the Cloudflare's result", ipv4Str)
		}
	}
	for _, ipv6Str := range cloudflareNetworkSpec.Result.IPv6CIDRs {
		ipv6Str = unescapeIPPrefix(ipv6Str)
		err :=
			providerNetworks.AddIPPrefix(
				common.DefaultRegion,
				common.DefaultService,
				ipv6Str,
				c.getComputeRedundancyFn())
		if err != nil {
			return nil, errors.Wrapf(err, "failed to add IPv6 prefix: %s to the Cloudflare's result", ipv6Str)
		}
	}

	return providerNetworks, nil
}

func unescapeIPPrefix(prefix string) string {
	return strings.ReplaceAll(prefix, `\/`, "/")
}

func (c *cloudflareNetworkCrawler) getComputeRedundancyFn() common.IsRedundantRegionServicePairFn {
	return common.GetDefaultRegionServicePairRedundancyCheck()
}

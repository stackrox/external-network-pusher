package cloudflare

import (
	"bufio"
	"bytes"
	"io"
	"strings"

	"github.com/pkg/errors"
	"github.com/stackrox/external-network-pusher/pkg/common"
	"github.com/stackrox/external-network-pusher/pkg/common/utils"
)

type cloudflareNetworkCrawler struct {
	urls []string
}

// NewCloudflareNetworkCrawler returns an instance of the cloudflareNetworkCrawler
func NewCloudflareNetworkCrawler() common.NetworkCrawler {
	return &cloudflareNetworkCrawler{urls: common.ProviderToURLs[common.Cloudflare]}
}

func (c *cloudflareNetworkCrawler) GetHumanReadableProviderName() string {
	return "Cloudflare"
}

func (c *cloudflareNetworkCrawler) GetProviderKey() common.Provider {
	return common.Cloudflare
}

func (c *cloudflareNetworkCrawler) CrawlPublicNetworkRanges() (*common.ProviderNetworkRanges, error) {
	allNetworkData, err := c.fetchAll()
	if err != nil {
		return nil, errors.Wrap(err, "failed to fetch network data while crawling Cloudflare's network ranges")
	}

	parsed, err := c.parseNetworks(allNetworkData)
	if err != nil {
		return nil, errors.Wrap(err, "failed to parse Cloudflare's network data")
	}
	return parsed, nil
}

func (c *cloudflareNetworkCrawler) fetchAll() ([][]byte, error) {
	allData := make([][]byte, 0, len(c.urls))
	for _, url := range c.urls {
		body, err := utils.HTTPGet(url)
		if err != nil {
			return nil, errors.Wrapf(err, "failed to fetch networks from Cloudflare with URL: %s", url)
		}
		allData = append(allData, body)
	}
	return allData, nil
}

func (c *cloudflareNetworkCrawler) parseNetworks(allNetworks [][]byte) (*common.ProviderNetworkRanges, error) {
	providerNetworks := common.ProviderNetworkRanges{ProviderName: c.GetProviderKey().String()}
	for _, networkData := range allNetworks {
		networkPrefixes, err := c.readToLines(networkData)
		if err != nil {
			return nil, errors.Wrap(err, "failed to parse Cloudflare's network data into list of strings")
		}
		for _, prefix := range networkPrefixes {
			err := providerNetworks.AddIPPrefix(common.DefaultRegion, common.DefaultService, prefix)
			if err != nil {
				return nil, errors.Wrapf(err, "failed to add IP prefix: %s to the Cloudflare's result", prefix)
			}
		}
	}
	return &providerNetworks, nil
}

func (c *cloudflareNetworkCrawler) readToLines(data []byte) ([]string, error) {
	lines := make([]string, 0)
	reader := bufio.NewReader(bytes.NewReader(data))
	for {
		line, err := reader.ReadString('\n')
		if err != io.EOF && err != nil {
			return nil, errors.Wrapf(err, "failed to parse data to string slice: %s", string(data))
		}
		// line contains the deliminator. erase and add to the result
		line = strings.TrimSuffix(line, "\n")
		// And we only add if the line contains more than just the newline char
		if line != "" {
			lines = append(lines, line)
		}
		if err == io.EOF {
			// End of input. Return
			return lines, nil
		}
	}
}

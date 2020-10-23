package azure

import (
	"encoding/json"
	"fmt"
	"log"
	"net"
	"os/exec"

	"github.com/pkg/errors"
	"github.com/stackrox/external-network-pusher/pkg/common"
	"github.com/stackrox/external-network-pusher/pkg/common/utils"
)

// With Microsoft, it is a little different in a sense that: it has four
// different URLs to crawl from, and each represents a separate cloud.
// The four clouds (as of 10/22/2020) are: AzurePublic, AzureGovernment,
// AzureChina, AzureGermany. Thus when constructing the region name for final
// output, we construct it as the following:
//     - RegionName = "<CloudName>-<AzureRegionName>"
// EX: - RegionName = "Public-australiacentral"
// If azureCloudEntityProperties.region is empty, we just use the CloudName as
// the final region name.
//
// Similarly, for service we use the following:
//     - ServiceName = "<Platform>-<SystemService>"
// If azureCloudEntityProperties.SystemService is empty, we just use the Platform
// as the final service name.

type azureNetworkCrawler struct {
	urls []string
}

type azureCloudEntityProperties struct {
	ChangeNumber    int      `json:"changeNumber"`
	Region          string   `json:"region"`
	RegionID        int      `json:"regionId"`
	Platform        string   `json:"platform"`
	SystemService   string   `json:"systemService"`
	AddressPrefixes []string `json:"addressPrefixes"`
	NetworkFeatures []string `json:"networkFeatures"`
}

type azureCloudEntity struct {
	Name       string                     `json:"name"`
	ID         string                     `json:"id"`
	Properties azureCloudEntityProperties `json:"properties"`
}

// azureCloud represents the top level structure for
// an Azure cloud networks
type azureCloud struct {
	ChangeNumber int                `json:"changeNumber"`
	Cloud        string             `json:"cloud"`
	Values       []azureCloudEntity `json:"values"`
}

// NewAzureNetworkCrawler returns an instance of azureNetworkCrawler
func NewAzureNetworkCrawler() common.NetworkCrawler {
	return &azureNetworkCrawler{urls: common.ProviderToURLs[common.Azure]}
}

func (c *azureNetworkCrawler) GetProviderKey() common.Provider {
	return common.Azure
}

func (c *azureNetworkCrawler) GetBucketObjectName() string {
	return "azure-networks"
}

func (c *azureNetworkCrawler) GetHumanReadableProviderName() string {
	return "Microsoft Azure Cloud"
}

func (c *azureNetworkCrawler) CrawlPublicNetworkRanges() (*common.PublicNetworkRanges, error) {
	// First, fetch from all sources
	cloudInfos, err := c.fetchAll()
	if err != nil {
		return nil, err
	}

	// Parse the data into out format
	azureNetworks, err := c.parseAzureNetworks(cloudInfos)
	if err != nil {
		return nil, errors.Wrap(err, "failed to parse Azure networks")
	}
	return azureNetworks, nil
}

func (c *azureNetworkCrawler) parseAzureNetworks(cloudInfos [][]byte) (*common.PublicNetworkRanges, error) {
	regionToNetworkDetails := make(map[string]common.RegionNetworkDetail)
	for _, data := range cloudInfos {
		var cloud azureCloud
		err := json.Unmarshal(data, &cloud)
		if err != nil {
			log.Printf("Failed to unmarshal Azure networks: %v", err)
			// Keep parsing other cloud info
			continue
		}

		for _, entity := range cloud.Values {
			if len(entity.Properties.AddressPrefixes) == 0 {
				continue
			}

			regionName := c.toRegionName(cloud.Cloud, entity.Properties.Region)
			serviceName := c.toServiceName(entity.Properties.Platform, entity.Properties.SystemService)

			var regionNetworks common.RegionNetworkDetail
			if detail, ok := regionToNetworkDetails[regionName]; ok {
				regionNetworks = detail
			} else {
				regionNetworks = common.RegionNetworkDetail{
					ServiceNameToIPRanges: make(map[string]common.ServiceIPRanges),
				}
			}

			// Get service
			var serviceIPRanges common.ServiceIPRanges
			if ips, ok := regionNetworks.ServiceNameToIPRanges[serviceName]; ok {
				serviceIPRanges = ips
			}
			for _, ipStr := range entity.Properties.AddressPrefixes {
				ip, prefix, err := net.ParseCIDR(ipStr)
				if err != nil || ip == nil || prefix == nil {
					// Stop here if we have detected an invalid IP string. This
					// means we probably are doing something very wrong (using expired
					// links, Azure changed the format of the json file, etc.)
					errors.Wrapf(err, "failed to parse address: %s", ip)
					return nil, err
				}
				if ip.To4() != nil {
					// IPv4 address
					serviceIPRanges.IPv4Prefixes = append(serviceIPRanges.IPv4Prefixes, ipStr)
				} else {
					// IPv6 address
					serviceIPRanges.IPv6Prefixes = append(serviceIPRanges.IPv6Prefixes, ipStr)
				}
			}

			regionNetworks.ServiceNameToIPRanges[serviceName] = serviceIPRanges
			regionToNetworkDetails[regionName] = regionNetworks
		}
	}

	if len(regionToNetworkDetails) == 0 {
		return nil, fmt.Errorf("failed to parse any network jsons crawled")
	}

	return &common.PublicNetworkRanges{RegionToNetworkDetails: regionToNetworkDetails}, nil
}

func (c *azureNetworkCrawler) toRegionName(cloudName, regionName string) string {
	return utils.ToCompoundName(cloudName, regionName)
}

func (c *azureNetworkCrawler) toServiceName(platformName, serviceName string) string {
	return utils.ToCompoundName(platformName, serviceName)
}

func (c *azureNetworkCrawler) fetchAll() ([][]byte, error) {
	// Microsoft does not give a static URL for its IP ranges, instead, they redirect all
	// download requests to a semi-static URL with dynamic parameter (EX: <staticURL>?ID=<some ID>),
	// and the page then renders generated URLs to json files.
	jsonURLs := make([]string, 0, len(c.urls))
	for _, url := range c.urls {
		jsonURL, err := c.redirectToJSONURL(url)
		if err != nil || jsonURL == "" {
			log.Printf("Failed to crawl Azure with URL: %s, skipping... Error: %v. JSON URL: %s", url, err, jsonURL)
			// Keep crawling other URLs
			continue
		}
		log.Printf("Reiceived Azure network JSON URL: %s", jsonURL)
		jsonURLs = append(jsonURLs, jsonURL)
	}
	if len(jsonURLs) == 0 {
		return nil, fmt.Errorf("failed to fetch any JSON URLs")
	}

	contents := make([][]byte, 0, len(jsonURLs))
	for _, jsonURL := range jsonURLs {
		log.Printf("Current URL is: %s", jsonURL)
		body, err := utils.HTTPGet(jsonURL)
		if err != nil {
			log.Printf("Failed to fetch networks from Azure with URL: %s. Skipping... Error: %v", jsonURL, err)
			// Keep crawling other URLs
			continue
		}
		contents = append(contents, body)
	}
	if len(contents) == 0 {
		return nil, fmt.Errorf("failed to crawl the contents of Azure JSON URLs")
	}
	return contents, nil
}

func (c *azureNetworkCrawler) redirectToJSONURL(rawURL string) (string, error) {
	cmd := fmt.Sprintf(
		// curl the page
		"curl -Lfs \"%s\" |"+
			// Get all the <a> tags
			" grep -Eoi '<a [^>]+>' |"+
			// Get all the hrefs within <a> tags
			" grep -Eo 'href=\"[^\\\"]+\"' |"+
			// Get all the relevant download links
			" grep \"download.microsoft.com/download/\" |"+
			// Match the URL part
			" grep -m 1 -Eo '(http|https)://[^\"]+' |"+
			// Trim trailing newline char
			"tr -d '\n'",
		rawURL)
	out, err := exec.Command("bash", "-c", cmd).Output()
	if err != nil {
		errors.Wrapf(err, "failed to redirect to JSON URL while trying to crawl Azure with URL: %s", rawURL)
		return "", err
	}
	return string(out), nil
}

package azure

import (
	"encoding/json"
	"fmt"
	"log"
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

const azureCompoundNameDelim = "/"

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

func (c *azureNetworkCrawler) GetHumanReadableProviderName() string {
	return "Microsoft Azure Cloud"
}

func (c *azureNetworkCrawler) GetNumRequiredIPPrefixes() int {
	// Observed from past runs after dedupe. In the past we had 24387
	return 24000
}

func (c *azureNetworkCrawler) CrawlPublicNetworkRanges() (*common.ProviderNetworkRanges, error) {
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

func (c *azureNetworkCrawler) parseAzureNetworks(cloudInfos [][]byte) (*common.ProviderNetworkRanges, error) {
	providerNetworks := common.NewProviderNetworkRanges(c.GetProviderKey().String())
	for _, data := range cloudInfos {
		var cloud azureCloud
		err := json.Unmarshal(data, &cloud)
		if err != nil {
			return nil, errors.Wrap(err, "failed to unmarshal Azure networks")
		}

		for _, entity := range cloud.Values {
			if len(entity.Properties.AddressPrefixes) == 0 {
				continue
			}
			regionName := toRegionName(cloud.Cloud, entity.Properties.Region)
			serviceName := toServiceName(entity.Properties.Platform, entity.Properties.SystemService)

			for _, ipStr := range entity.Properties.AddressPrefixes {
				err := providerNetworks.AddIPPrefix(regionName, serviceName, ipStr, c.getComputeRedundancyFn())
				if err != nil {
					// Stop here if we have detected an invalid IP string. This
					// means we probably are doing something very wrong (using expired
					// links, Azure changed the format of the json file, etc.)
					return nil, errors.Wrapf(err, "failed to parse Azure IP address: %s", ipStr)
				}
			}
		}
	}

	return providerNetworks, nil
}

func (c *azureNetworkCrawler) getComputeRedundancyFn() common.IsRedundantRegionServicePairFn {
	return func(
		newPair *common.RegionServicePair,
		existingPair *common.RegionServicePair,
	) (*common.RegionServicePair, error) {
		// Since for an Azure IP prefix, we are doing some tricks with its region and service name
		// We determine B redundant when for two prefixes A and B:
		//         A.region.IsSubsetOfOrEqualTo(B.region) && A.service.IsSubsetOfOrEqualTo(B.service)
		// Example:
		// A: Region: Azure/useast1, Service: APIGateway;  B: Region: Azure, Service: APIGateway
		redundantPair, err := c.isSubsetOfOrEqualTo(newPair, existingPair)
		if err != nil {
			return nil, err
		}

		// If not redundant, this should be (nil, nil)
		return redundantPair, nil
	}
}

// Returns the pair that is the superset (redundant) of the other.
// If it is the same then a random one is returned
func (c *azureNetworkCrawler) isSubsetOfOrEqualTo(
	pair1, pair2 *common.RegionServicePair,
) (*common.RegionServicePair, error) {
	if pair1.Equals(pair2) {
		return pair1, nil
	}

	redundantPairBasedOnRegion, err := c.checkSubsetBasedOnRegionName(pair1, pair2)
	if err != nil {
		return nil, err
	}
	redundantPairBasedOnService, err := c.checkSubsetBasedOnServiceName(pair1, pair2)
	if err != nil {
		return nil, err
	}
	if redundantPairBasedOnRegion == nil && redundantPairBasedOnService == nil {
		return nil, nil
	}

	if (redundantPairBasedOnRegion == nil || redundantPairBasedOnRegion.Equals(pair1)) &&
		(redundantPairBasedOnService == nil || redundantPairBasedOnService.Equals(pair1)) {
		// It is impossible that both region and service results are nil (covered above).
		// Thus either based on region or based on service, or based on both, pair1 deemed redundant.
		return pair1, nil
	}
	if (redundantPairBasedOnRegion == nil || redundantPairBasedOnRegion.Equals(pair2)) &&
		(redundantPairBasedOnService == nil || redundantPairBasedOnService.Equals(pair2)) {
		// It is impossible that both region and service results are nil (covered above).
		// Thus either based on region or based on service, or based on both, pair2 deemed redundant.
		return pair2, nil
	}

	// No inclusive relationship.
	return nil, nil
}

// Returns the pair that is the superset (redundant) of the other
// Assuming the region names aren't equal.
func (c *azureNetworkCrawler) checkSubsetBasedOnRegionName(
	pair1, pair2 *common.RegionServicePair,
) (*common.RegionServicePair, error) {
	pair1CloudName, pair1RegionName, err := regionNameToCloudAndRegionNames(pair1.Region)
	if err != nil {
		return nil, err
	}
	pair2CloudName, pair2RegionName, err := regionNameToCloudAndRegionNames(pair2.Region)
	if err != nil {
		return nil, err
	}
	if pair1CloudName == pair2CloudName {
		if pair1RegionName == "" {
			// pair2 specific region name not empty. pair1 >> pair2
			return pair1, nil
		}
		if pair2RegionName == "" {
			// pair1 specific region name not empty. pair2 >> pair1
			return pair2, nil
		}
		// Reaching here means although both cloud names are the same, specific region names
		// are not empty and different.
	}

	// No inclusive relationship.
	return nil, nil
}

// Returns the pair that is the superset (redundant) of the other
// Assuming the service names aren't equal.
func (c *azureNetworkCrawler) checkSubsetBasedOnServiceName(
	pair1, pair2 *common.RegionServicePair,
) (*common.RegionServicePair, error) {
	pair1PlatformName, pair1ServiceName, err := serviceNameToPlatformAndServiceNames(pair1.Service)
	if err != nil {
		return nil, err
	}
	pair2PlatformName, pair2ServiceName, err := serviceNameToPlatformAndServiceNames(pair2.Service)
	if err != nil {
		return nil, err
	}
	if pair1PlatformName == pair2PlatformName {
		if pair1ServiceName == "" {
			// pair2 specific service name not empty. pair1 >> pair2
			return pair1, nil
		}
		if pair2ServiceName == "" {
			// pair1 specific service name not empty. pair2 >> pair1
			return pair2, nil
		}
		// Reaching here means although both platform names are the same, specific service names
		// are not empty and different.
	}

	// No inclusive relationship.
	return nil, nil
}

func regionNameToCloudAndRegionNames(regionName string) (string, string, error) {
	splitted := utils.ToTags(azureCompoundNameDelim, regionName)
	switch len(splitted) {
	case 1:
		// No specific region name provided
		return regionName, "", nil
	case 2:
		return splitted[0], splitted[1], nil
	default:
		return "", "", InvalidAzureCompoundRegionName(regionName)
	}
}

func serviceNameToPlatformAndServiceNames(serviceName string) (string, string, error) {
	splitted := utils.ToTags(azureCompoundNameDelim, serviceName)
	switch len(splitted) {
	case 1:
		// No specific service name provided
		return serviceName, "", nil
	case 2:
		return splitted[0], splitted[1], nil
	default:
		return "", "", InvalidAzureCompoundServiceName(serviceName)
	}
}

func toRegionName(cloudName, regionName string) string {
	return utils.ToCompoundName(azureCompoundNameDelim, cloudName, regionName)
}

func toServiceName(platformName, serviceName string) string {
	return utils.ToCompoundName(azureCompoundNameDelim, platformName, serviceName)
}

func (c *azureNetworkCrawler) fetchAll() ([][]byte, error) {
	// Microsoft does not give a static URL for its IP ranges, instead, they redirect all
	// download requests to a semi-static URL with dynamic parameter (EX: <staticURL>?ID=<some ID>),
	// and the page then renders generated URLs to json files.
	jsonURLs := make([]string, 0, len(c.urls))
	for _, url := range c.urls {
		jsonURL, err := c.redirectToJSONURL(url)
		if err != nil {
			return nil, errors.Wrapf(err, "failed to crawl Azure with URL: %s. Error: %v. JSON URL: %s", url, err, jsonURL)
		}
		if jsonURL == "" {
			return nil, errors.Errorf("failed to crawl Azure with URL: %s, empty JSON URL returned. This could indicate Azure's services are unavailable", url)
		}
		log.Printf("Received Azure network JSON URL: %s", jsonURL)
		jsonURLs = append(jsonURLs, jsonURL)
	}

	contents := make([][]byte, 0, len(jsonURLs))
	for _, jsonURL := range jsonURLs {
		log.Printf("Current URL is: %s", jsonURL)
		body, err := utils.HTTPGet(jsonURL)
		if err != nil {
			return nil, errors.Wrapf(err, "Failed to fetch networks from Azure with URL: %s. Error: %v", jsonURL, err)
		}
		contents = append(contents, body)
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
	out, err := exec.Command("/bin/sh", "-c", cmd).Output()
	if err != nil {
		errors.Wrapf(err, "failed to redirect to JSON URL while trying to crawl Azure with URL: %s", rawURL)
		return "", err
	}
	return string(out), nil
}

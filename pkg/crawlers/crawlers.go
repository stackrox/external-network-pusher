package crawlers

import (
	"log"

	"github.com/stackrox/external-network-pusher/pkg/common"
	"github.com/stackrox/external-network-pusher/pkg/crawlers/aws"
	"github.com/stackrox/external-network-pusher/pkg/crawlers/azure"
	"github.com/stackrox/external-network-pusher/pkg/crawlers/gcp"
)

// allCrawlers include all the crawler implementations
var allCrawlers = []common.NetworkCrawler{
	gcp.NewGCPNetworkCrawler(),
	azure.NewAzureNetworkCrawler(),
	aws.NewAWSNetworkCrawler(),
}

// Get returns list of provider specific NetworkCrawler implementations
func Get(skippedProviders []common.Provider) []common.NetworkCrawler {
	skippedProvidersSet := make(map[common.Provider]struct{})
	for _, p := range skippedProviders {
		skippedProvidersSet[p] = struct{}{}
	}
	var crawlers []common.NetworkCrawler
	for _, crawler := range allCrawlers {
		if _, ok := skippedProvidersSet[crawler.GetProviderKey()]; !ok {
			crawlers = append(crawlers, crawler)
		} else {
			log.Printf("Skipping crawling networks for %s...", crawler.GetHumanReadableProviderName())
		}
	}
	return crawlers
}

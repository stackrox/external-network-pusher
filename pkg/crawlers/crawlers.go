package crawlers

import (
	"log"

	"github.com/stackrox/external-network-pusher/pkg/common"
	"github.com/stackrox/external-network-pusher/pkg/crawlers/gcp"
)

// allCrawlers include all the crawler implementations
var allCrawlers = []common.NetworkCrawler{
	gcp.NewGCPNetworkCrawler(),
}

// Get returns list of provider specific NetworkCrawler implementations
func Get(skippedProviders map[common.Provider]struct{}) []common.NetworkCrawler {
	var crawlers []common.NetworkCrawler
	for _, crawler := range allCrawlers {
		if _, ok := skippedProviders[crawler.GetProviderKey()]; !ok {
			crawlers = append(crawlers, crawler)
		} else {
			log.Printf("Skipping crawling networks for %s...", crawler.GetHumanReadableProviderName())
		}
	}
	return crawlers
}

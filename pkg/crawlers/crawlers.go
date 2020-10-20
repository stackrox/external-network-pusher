package crawlers

import (
	"log"

	"github.com/stackrox/external-network-pusher/pkg/commons"
	"github.com/stackrox/external-network-pusher/pkg/crawlers/gcp"
)

// allCrawlers include all the crawler implementations
var allCrawlers = []commons.NetworkCrawler{
	gcp.NewGcpNetworkCrawler(),
}

// NewCrawlers returns list of provider specific NetworkCrawler implementations
func NewCrawlers(skippedProviders map[commons.Provider]bool) []commons.NetworkCrawler {
	var crawlers []commons.NetworkCrawler
	for _, crawler := range allCrawlers {
		if !skippedProviders[crawler.GetProviderKey()] {
			crawlers = append(crawlers, crawler)
		} else {
			log.Printf("Skipping crawling networks for %s...", crawler.GetHumanReadableProviderName())
		}
	}
	return crawlers
}

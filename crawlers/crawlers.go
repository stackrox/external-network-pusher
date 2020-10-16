package crawlers

import (
	"log"

	"github.com/stackrox/external-network-pusher/commons"
	"github.com/stackrox/external-network-pusher/crawlers/gcp"
)

// allCrawlers include all the crawler implementations
var allCrawlers = []commons.NetworkCrawler{
	gcp.NewGcpNetworkCrawler(),
}

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


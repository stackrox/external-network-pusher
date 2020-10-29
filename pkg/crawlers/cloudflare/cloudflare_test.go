package cloudflare

import (
	"encoding/json"
	"testing"

	"github.com/stackrox/external-network-pusher/pkg/common"
	"github.com/stackrox/external-network-pusher/pkg/common/utils"
	"github.com/stretchr/testify/require"
)

func TestCloudflareParseNetwork(t *testing.T) {
	// Cloudflare provides their IPs at URL: https://api.cloudflare.com/client/v4/ips
	// with slashes escaped. Mimic that
	ipv41, ipv42, ipv43 := "173.245.48.0\\/20", "103.21.244.0\\/22", "103.22.200.0\\/22"
	ipv61, ipv62, ipv63 := "2400:cb00::\\/32", "2606:4700::\\/32", "2803:f800::\\/32"

	testData := cloudflareNetworkSpec{
		Result: cloudflareNetworkResult{
			IPv4CIDRs: []string{ipv41, ipv42, ipv43},
			IPv6CIDRs: []string{ipv61, ipv62, ipv63},
			ETag:      utils.UnusedString,
		},
		Success:  utils.UnusedBool,
		Errors:   utils.UnusedStrSlice,
		Messages: utils.UnusedStrSlice,
	}
	networks, err := json.Marshal(testData)
	require.Nil(t, err)

	crawler := cloudflareNetworkCrawler{}
	parsedResult, err := crawler.parseNetworks(networks)
	require.Nil(t, err)
	require.Equal(t, parsedResult.ProviderName, crawler.GetProviderKey().String())

	// Just one region in total. common.DefaultRegion
	require.Equal(t, 1, len(parsedResult.RegionNetworks))

	// Check content of the region
	regionNetworks := parsedResult.RegionNetworks[0]
	// Just one service in total. common.DefaultService
	require.Equal(t, 1, len(regionNetworks.ServiceNetworks))

	// Check content of service
	serviceToIPs := utils.GetServiceNameToIPs(regionNetworks)
	escapedIPv4s := []string{unescapeIPPrefix(ipv41), unescapeIPPrefix(ipv42), unescapeIPPrefix(ipv43)}
	escapedIPv6s := []string{unescapeIPPrefix(ipv61), unescapeIPPrefix(ipv62), unescapeIPPrefix(ipv63)}
	utils.CheckServiceIPsInRegion(
		t,
		serviceToIPs,
		common.DefaultService,
		escapedIPv4s,
		escapedIPv6s)
}

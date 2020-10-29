package cloudflare

import (
	"strings"
	"testing"

	"github.com/stackrox/external-network-pusher/pkg/common"
	"github.com/stackrox/external-network-pusher/pkg/common/utils"
	"github.com/stretchr/testify/require"
)

func TestCloudflareParseNetwork(t *testing.T) {
	ipv41, ipv42, ipv43 := "173.245.48.0/20", "103.21.244.0/22", "103.22.200.0/22"
	ipv61, ipv62, ipv63 := "2400:cb00::/32", "2606:4700::/32", "2803:f800::/32"

	testIPv4Data := []byte(strings.Join([]string{ipv41, ipv42, ipv43}, "\n"))
	testIPv6Data := []byte(strings.Join([]string{ipv61, ipv62, ipv63}, "\n"))

	crawler := cloudflareNetworkCrawler{}
	parsedResult, err := crawler.parseNetworks([][]byte{testIPv4Data, testIPv6Data})
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
	utils.CheckServiceIPsInRegion(
		t,
		serviceToIPs,
		common.DefaultService,
		[]string{ipv41, ipv42, ipv43},
		[]string{ipv61, ipv62, ipv63})
}

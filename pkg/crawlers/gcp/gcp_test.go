package gcp

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestGcpParseNetwork(t *testing.T) {
	ipv41, ipv42, ipv43 := "34.80.0.0/15", "35.185.128.0/19", "34.97.0.0/16"
	ipv61, ipv62 := "2600:1901::/48", "2600:1901:1:1000::/52"
	service1, service2 := "Google Cloud", "Google Ads"
	region1, region2 := "asia-east1", "europe-west4"

	testData := GcpNetworkSpec{
		SyncToken: "1602608557449",
		CreationTime: "2020-10-13T10:02:37.449",
		Prefixes: []GcpIpSpec{
			{
				Ipv4Prefix: ipv41,
				Service:    service1,
				Scope:      region1,
			},
			{
				Ipv4Prefix: ipv42,
				Service:    service1,
				Scope:      region2,
			},
			{
				Ipv4Prefix: ipv43,
				Ipv6Prefix: ipv61,
				Service: service1,
				Scope: region1,
			},
			{
				Ipv6Prefix: ipv62,
				Service: service2,
				Scope: region1,
			},
		},
	}
	networks, err := json.Marshal(testData)
	require.Nil(t, err)

	crawler := GcpNetworkCrawler{}
	parsedResult, err := crawler.parseNetworks(networks)
	require.Nil(t, err)

	// Two regions in total
	require.Equal(t, 2, len(parsedResult.RegionToNetworkDetails))

	// Check content of the first region
	{
		firstRegionNetworks, ok := parsedResult.RegionToNetworkDetails[region1]
		require.True(t, ok)
		// Two services in total for region1
		require.Equal(t, 2, len(firstRegionNetworks.ServiceNameToIpRanges))

		// service1
		service1Networks, ok := firstRegionNetworks.ServiceNameToIpRanges[service1]
		require.True(t, ok)
		require.ElementsMatch(t, []string{ipv41, ipv43}, service1Networks.Ipv4Prefixes)
		require.ElementsMatch(t, []string{ipv61}, service1Networks.Ipv6Prefixes)

		// service2
		service2Networks, ok := firstRegionNetworks.ServiceNameToIpRanges[service2]
		require.True(t, ok)
		require.ElementsMatch(t, []string{}, service2Networks.Ipv4Prefixes)
		require.ElementsMatch(t, []string{ipv62}, service2Networks.Ipv6Prefixes)
	}

	{
		secondRegionNetworks, ok := parsedResult.RegionToNetworkDetails[region2]
		require.True(t, ok)
		// Only one service in region2
		require.Equal(t, 1, len(secondRegionNetworks.ServiceNameToIpRanges))

		// service1
		service1Networks, ok := secondRegionNetworks.ServiceNameToIpRanges[service1]
		require.True(t, ok)
		require.ElementsMatch(t, []string{ipv42}, service1Networks.Ipv4Prefixes)
		require.ElementsMatch(t, []string{}, service1Networks.Ipv6Prefixes)
	}
}

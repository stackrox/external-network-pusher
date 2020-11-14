package gcp

import (
	"encoding/json"
	"testing"

	"github.com/stackrox/external-network-pusher/pkg/common/testutils"
	"github.com/stretchr/testify/require"
)

func TestGcpParseNetwork(t *testing.T) {
	ipv41, ipv42, ipv43 := "34.80.0.0/15", "35.185.128.0/19", "34.97.0.0/16"
	ipv61, ipv62 := "2600:1901::/48", "2600:1901:1:1000::/52"
	service1, service2 := "Google Cloud", "Google Ads"
	region1, region2 := "asia-east1", "europe-west4"

	testData := gcpNetworkSpec{
		SyncToken:    testutils.UnusedString,
		CreationTime: testutils.UnusedString,
		Prefixes: []gcpIPSpec{
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
				Service:    service1,
				Scope:      region1,
			},
			{
				Ipv6Prefix: ipv62,
				Service:    service2,
				Scope:      region1,
			},
		},
	}
	networks, err := json.Marshal(testData)
	require.Nil(t, err)

	crawler := gcpNetworkCrawler{}
	parsedResult, err := crawler.parseNetworks(networks)
	require.Nil(t, err)
	require.Equal(t, parsedResult.ProviderName, crawler.GetProviderKey().String())

	// Two regions in total
	require.Equal(t, 2, len(parsedResult.RegionNetworks))
	regionNameToDetail := testutils.GetRegionNameToDetails(parsedResult)

	// Check content of the first region
	{
		firstRegionNetworks, ok := regionNameToDetail[region1]
		require.True(t, ok)
		// Two services in total for region1
		require.Equal(t, 2, len(firstRegionNetworks.ServiceNetworks))

		serviceToIPs := testutils.GetServiceNameToIPs(firstRegionNetworks)

		// service1
		testutils.CheckServiceIPsInRegion(
			t,
			serviceToIPs,
			service1,
			[]string{ipv41, ipv43},
			[]string{ipv61})

		// service2
		testutils.CheckServiceIPsInRegion(
			t,
			serviceToIPs,
			service2,
			[]string{},
			[]string{ipv62})
	}

	{
		secondRegionNetworks, ok := regionNameToDetail[region2]
		require.True(t, ok)
		// Only one service in region2
		require.Equal(t, 1, len(secondRegionNetworks.ServiceNetworks))

		serviceToIPs := testutils.GetServiceNameToIPs(secondRegionNetworks)

		// service1
		testutils.CheckServiceIPsInRegion(
			t,
			serviceToIPs,
			service1,
			[]string{ipv42},
			[]string{})
	}
}

func TestGCPRegionServiceRedundancyCheck(t *testing.T) {
	addr := "34.80.0.0/15"
	service1, service2 := "Google Cloud", "Google Ads"
	regionName := "test-region"

	testData := gcpNetworkSpec{
		SyncToken:    testutils.UnusedString,
		CreationTime: testutils.UnusedString,
		Prefixes: []gcpIPSpec{
			{
				Ipv4Prefix: addr,
				Service:    service1,
				Scope:      regionName,
			},
			{
				Ipv4Prefix: addr,
				Service:    service1,
				Scope:      regionName,
			},
			{
				Ipv4Prefix: addr,
				Service:    service2,
				Scope:      regionName,
			},
		},
	}
	networks, err := json.Marshal(testData)
	require.Nil(t, err)

	crawler := gcpNetworkCrawler{}
	parsedResult, err := crawler.parseNetworks(networks)
	require.Nil(t, err)
	require.Equal(t, parsedResult.ProviderName, crawler.GetProviderKey().String())

	// One regions in total
	require.Equal(t, 1, len(parsedResult.RegionNetworks))
	regionNameToDetail := testutils.GetRegionNameToDetails(parsedResult)

	// Check content of the region
	{
		firstRegionNetworks, ok := regionNameToDetail[regionName]
		require.True(t, ok)
		// Two services in total for region1
		require.Equal(t, 2, len(firstRegionNetworks.ServiceNetworks))

		serviceToIPs := testutils.GetServiceNameToIPs(firstRegionNetworks)

		// service1
		testutils.CheckServiceIPsInRegion(
			t,
			serviceToIPs,
			service1,
			[]string{addr},
			[]string{})

		// service2
		testutils.CheckServiceIPsInRegion(
			t,
			serviceToIPs,
			service2,
			[]string{addr},
			[]string{})
	}
}

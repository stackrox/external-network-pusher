package aws

import (
	"encoding/json"
	"testing"

	"github.com/stackrox/external-network-pusher/pkg/common/utils"
	"github.com/stretchr/testify/require"
)

var (
	UnusedString = "UNUSED"
)

/*
	IPPrefix string `json:"ip_prefix"`
	Region string `json:"region"`
	NetworkBorderGroup string `json:"network_border_group"`
	Service string `json:"service"`
*/

func TestAWSParseNetworks(t *testing.T) {
	ipv41, ipv42, ipv43 := "3.5.140.0/22", "35.180.0.0/16", "52.93.178.234/32"
	ipv61, ipv62, ipv63 := "2600:1f15::/32", "2a05:d07a:a000::/40", "240f:80ff:4000::/40"
	region1, region2, region3 := "region1", "region2", "region3"
	service1, service2, service3 := "service1", "service2", "service3"

	testData := awsNetworkSpec{
		SyncToken:  UnusedString,
		CreateDate: UnusedString,
		Prefixes: []awsIPv4Spec{
			{
				IPPrefix:           ipv41,
				Region:             region1,
				NetworkBorderGroup: UnusedString,
				Service:            service1,
			},
			{
				IPPrefix:           ipv42,
				Region:             region1,
				NetworkBorderGroup: UnusedString,
				Service:            service2,
			},
			{
				IPPrefix:           ipv43,
				Region:             region2,
				NetworkBorderGroup: UnusedString,
				Service:            service1,
			},
		},
		IPv6Prefixes: []awsIPv6Spec{
			{
				IPv6Prefix:         ipv61,
				Region:             region1,
				NetworkBorderGroup: UnusedString,
				Service:            service1,
			},
			{
				IPv6Prefix:         ipv62,
				Region:             region2,
				NetworkBorderGroup: UnusedString,
				Service:            service2,
			},
			{
				IPv6Prefix:         ipv63,
				Region:             region3,
				NetworkBorderGroup: UnusedString,
				Service:            service3,
			},
		},
	}

	networks, err := json.Marshal(testData)
	require.Nil(t, err)

	crawler := awsNetworkCrawler{}
	parsedResult, err := crawler.parseNetworks(networks)
	require.Nil(t, err)
	require.Equal(t, parsedResult.ProviderName, crawler.GetProviderKey().String())

	// Three (region1, 2, 3) regions in total
	require.Equal(t, 3, len(parsedResult.RegionNetworks))
	regionNameToDetail := utils.GetRegionNameToDetails(parsedResult)

	// region1
	{
		regionNetworks, ok := regionNameToDetail[region1]
		require.True(t, ok)
		// Two services in total for region1
		require.Equal(t, 2, len(regionNetworks.ServiceNetworks))

		serviceToIPs := utils.GetServiceNameToIPs(regionNetworks)
		// service1
		utils.CheckServiceIPsInRegion(
			t,
			serviceToIPs,
			service1,
			[]string{ipv41},
			[]string{ipv61})

		// service2
		utils.CheckServiceIPsInRegion(
			t,
			serviceToIPs,
			service2,
			[]string{ipv42},
			[]string{})
	}
	// region2
	{
		regionNetworks, ok := regionNameToDetail[region2]
		require.True(t, ok)
		// Two services in total for region2
		require.Equal(t, 2, len(regionNetworks.ServiceNetworks))

		serviceToIPs := utils.GetServiceNameToIPs(regionNetworks)
		// service1
		utils.CheckServiceIPsInRegion(
			t,
			serviceToIPs,
			service1,
			[]string{ipv43},
			[]string{})

		// service2
		utils.CheckServiceIPsInRegion(
			t,
			serviceToIPs,
			service2,
			[]string{},
			[]string{ipv62})
	}
	// region3
	{
		regionNetworks, ok := regionNameToDetail[region3]
		require.True(t, ok)
		// Only one service in region3
		require.Equal(t, 1, len(regionNetworks.ServiceNetworks))

		serviceToIPs := utils.GetServiceNameToIPs(regionNetworks)
		// service3
		utils.CheckServiceIPsInRegion(
			t,
			serviceToIPs,
			service3,
			[]string{},
			[]string{ipv63})
	}
}

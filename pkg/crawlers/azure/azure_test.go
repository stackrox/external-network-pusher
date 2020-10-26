package azure

import (
	"encoding/json"
	"testing"

	"github.com/stackrox/external-network-pusher/pkg/common/utils"
	"github.com/stretchr/testify/require"
)

var (
	UnusedInt      = -1
	UnusedStrSlice = []string{"unused1", "unused2"}
)

func TestAzureParseNetwork(t *testing.T) {
	cloud1, cloud2 := "Public", "AzureGovernment"
	service1, service2, emptyService := "ActionGroup", "AzureStorage", ""
	region1, region2, emptyRegion := "useast", "usgovcentral", ""
	platform := "Azure"

	// cloud1-region1-service1
	var (
		c1r1s1IPv41, c1r1s1IPv42 = "20.140.48.160/27", "20.140.56.160/27"
	)
	// cloud1-region1-service2
	var (
		c1r1s2IPv6 = "2001:489a:3103::140/123"
	)
	// cloud1-region1-emptyService
	var (
		c1r1NoServiceIPv4 = "20.140.64.160/27"
		c1r1NoServiceIPv6 = "2001:489a:3203::140/123"
	)
	// cloud2-region2-service2
	var (
		c2r2s2IPv4 = "20.140.72.160/27"
	)
	// cloud2-emptyRegion-emptyService
	var (
		c2NoRegionNoServiceIPv4 = "52.127.49.96/27"
		c2NoRegionNoServiceIPv6 = "2001:489a:3303::140/123"
	)

	testCloud1 := azureCloud{
		ChangeNumber: UnusedInt,
		Cloud:        cloud1,
		Values: []azureCloudEntity{
			{
				Name: service1,
				ID:   service1,
				Properties: azureCloudEntityProperties{
					ChangeNumber:    UnusedInt,
					Region:          region1,
					RegionID:        UnusedInt,
					Platform:        platform,
					SystemService:   service1,
					AddressPrefixes: []string{c1r1s1IPv41, c1r1s1IPv42},
					NetworkFeatures: UnusedStrSlice,
				},
			},
			{
				Name: service2,
				ID:   service2,
				Properties: azureCloudEntityProperties{
					ChangeNumber:    UnusedInt,
					Region:          region1,
					RegionID:        UnusedInt,
					Platform:        platform,
					SystemService:   service2,
					AddressPrefixes: []string{c1r1s2IPv6},
					NetworkFeatures: UnusedStrSlice,
				},
			},
			{
				Name: "AzureCloud." + region1,
				ID:   "AzureCloud." + region1,
				Properties: azureCloudEntityProperties{
					ChangeNumber:    UnusedInt,
					Region:          region1,
					RegionID:        UnusedInt,
					Platform:        platform,
					SystemService:   emptyService,
					AddressPrefixes: []string{c1r1NoServiceIPv4, c1r1NoServiceIPv6},
					NetworkFeatures: UnusedStrSlice,
				},
			},
		},
	}
	testCloud2 := azureCloud{
		ChangeNumber: UnusedInt,
		Cloud:        cloud2,
		Values: []azureCloudEntity{
			{
				Name: service2,
				ID:   service2,
				Properties: azureCloudEntityProperties{
					ChangeNumber:    UnusedInt,
					Region:          region2,
					RegionID:        UnusedInt,
					Platform:        platform,
					SystemService:   service2,
					AddressPrefixes: []string{c2r2s2IPv4},
					NetworkFeatures: UnusedStrSlice,
				},
			},
			{
				Name: "Azure",
				ID:   "Azure",
				Properties: azureCloudEntityProperties{
					ChangeNumber:    UnusedInt,
					Region:          emptyRegion,
					RegionID:        UnusedInt,
					Platform:        platform,
					SystemService:   emptyService,
					AddressPrefixes: []string{c2NoRegionNoServiceIPv4, c2NoRegionNoServiceIPv6},
					NetworkFeatures: UnusedStrSlice,
				},
			},
		},
	}

	cloud1Networks, err := json.Marshal(testCloud1)
	require.Nil(t, err)
	cloud2Networks, err := json.Marshal(testCloud2)
	require.Nil(t, err)

	crawler := azureNetworkCrawler{}
	parsedResult, err := crawler.parseAzureNetworks([][]byte{cloud1Networks, cloud2Networks})
	require.Nil(t, err)
	require.Equal(t, parsedResult.ProviderName, crawler.GetProviderKey().String())

	// There should be 3 regions in total (c1r1, c2r2, c2)
	require.Equal(t, 3, len(parsedResult.RegionNetworks))
	regionNameToDetail := utils.GetRegionNameToDetails(parsedResult)

	// Region1 (c1r1)
	{
		region := utils.ToCompoundName(cloud1, region1)
		regionNetworks, ok := regionNameToDetail[region]
		require.True(t, ok)
		require.Equal(t, 3, len(regionNetworks.ServiceNetworks))

		service := utils.ToCompoundName(platform, service1)
		serviceToIPs := utils.GetServiceNameToIPs(regionNetworks)
		utils.CheckServiceIPsInRegion(
			t,
			serviceToIPs,
			service,
			[]string{c1r1s1IPv41, c1r1s1IPv42},
			[]string{})

		service = utils.ToCompoundName(platform, service2)
		utils.CheckServiceIPsInRegion(
			t,
			serviceToIPs,
			service,
			[]string{},
			[]string{c1r1s2IPv6})

		service = utils.ToCompoundName(platform, emptyService)
		utils.CheckServiceIPsInRegion(
			t,
			serviceToIPs,
			service,
			[]string{c1r1NoServiceIPv4},
			[]string{c1r1NoServiceIPv6})
	}
	// Region 2 (c2r2)
	{
		region := utils.ToCompoundName(cloud2, region2)
		regionNetworks, ok := regionNameToDetail[region]
		require.True(t, ok)
		require.Equal(t, 1, len(regionNetworks.ServiceNetworks))

		service := utils.ToCompoundName(platform, service2)
		serviceToIPs := utils.GetServiceNameToIPs(regionNetworks)
		utils.CheckServiceIPsInRegion(
			t,
			serviceToIPs,
			service,
			[]string{c2r2s2IPv4},
			[]string{})
	}
	// Region 3 (c2)
	{
		region := utils.ToCompoundName(cloud2, emptyRegion)
		regionNetworks, ok := regionNameToDetail[region]
		require.True(t, ok)
		require.Equal(t, 1, len(regionNetworks.ServiceNetworks))

		service := utils.ToCompoundName(platform, emptyService)
		serviceToIPs := utils.GetServiceNameToIPs(regionNetworks)
		utils.CheckServiceIPsInRegion(
			t,
			serviceToIPs,
			service,
			[]string{c2NoRegionNoServiceIPv4},
			[]string{c2NoRegionNoServiceIPv6})
	}
}

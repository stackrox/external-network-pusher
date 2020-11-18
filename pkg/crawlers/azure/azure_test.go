package azure

import (
	"encoding/json"
	"testing"

	"github.com/stackrox/external-network-pusher/pkg/common/testutils"
	"github.com/stretchr/testify/require"
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
		ChangeNumber: testutils.UnusedInt,
		Cloud:        cloud1,
		Values: []azureCloudEntity{
			{
				Name: service1,
				ID:   service1,
				Properties: azureCloudEntityProperties{
					ChangeNumber:    testutils.UnusedInt,
					Region:          region1,
					RegionID:        testutils.UnusedInt,
					Platform:        platform,
					SystemService:   service1,
					AddressPrefixes: []string{c1r1s1IPv41, c1r1s1IPv42},
					NetworkFeatures: testutils.UnusedStrSlice,
				},
			},
			{
				Name: service2,
				ID:   service2,
				Properties: azureCloudEntityProperties{
					ChangeNumber:    testutils.UnusedInt,
					Region:          region1,
					RegionID:        testutils.UnusedInt,
					Platform:        platform,
					SystemService:   service2,
					AddressPrefixes: []string{c1r1s2IPv6},
					NetworkFeatures: testutils.UnusedStrSlice,
				},
			},
			{
				Name: "AzureCloud." + region1,
				ID:   "AzureCloud." + region1,
				Properties: azureCloudEntityProperties{
					ChangeNumber:    testutils.UnusedInt,
					Region:          region1,
					RegionID:        testutils.UnusedInt,
					Platform:        platform,
					SystemService:   emptyService,
					AddressPrefixes: []string{c1r1NoServiceIPv4, c1r1NoServiceIPv6},
					NetworkFeatures: testutils.UnusedStrSlice,
				},
			},
		},
	}
	testCloud2 := azureCloud{
		ChangeNumber: testutils.UnusedInt,
		Cloud:        cloud2,
		Values: []azureCloudEntity{
			{
				Name: service2,
				ID:   service2,
				Properties: azureCloudEntityProperties{
					ChangeNumber:    testutils.UnusedInt,
					Region:          region2,
					RegionID:        testutils.UnusedInt,
					Platform:        platform,
					SystemService:   service2,
					AddressPrefixes: []string{c2r2s2IPv4},
					NetworkFeatures: testutils.UnusedStrSlice,
				},
			},
			{
				Name: "Azure",
				ID:   "Azure",
				Properties: azureCloudEntityProperties{
					ChangeNumber:    testutils.UnusedInt,
					Region:          emptyRegion,
					RegionID:        testutils.UnusedInt,
					Platform:        platform,
					SystemService:   emptyService,
					AddressPrefixes: []string{c2NoRegionNoServiceIPv4, c2NoRegionNoServiceIPv6},
					NetworkFeatures: testutils.UnusedStrSlice,
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
	regionNameToDetail := testutils.GetRegionNameToDetails(parsedResult)

	// Region1 (c1r1)
	{
		region := toRegionName(cloud1, region1)
		regionNetworks, ok := regionNameToDetail[region]
		require.True(t, ok)
		require.Equal(t, 3, len(regionNetworks.ServiceNetworks))

		service := toServiceName(platform, service1)
		serviceToIPs := testutils.GetServiceNameToIPs(regionNetworks)
		testutils.CheckServiceIPsInRegion(
			t,
			serviceToIPs,
			service,
			[]string{c1r1s1IPv41, c1r1s1IPv42},
			[]string{})

		service = toServiceName(platform, service2)
		testutils.CheckServiceIPsInRegion(
			t,
			serviceToIPs,
			service,
			[]string{},
			[]string{c1r1s2IPv6})

		service = toServiceName(platform, emptyService)
		testutils.CheckServiceIPsInRegion(
			t,
			serviceToIPs,
			service,
			[]string{c1r1NoServiceIPv4},
			[]string{c1r1NoServiceIPv6})
	}
	// Region 2 (c2r2)
	{
		region := toRegionName(cloud2, region2)
		regionNetworks, ok := regionNameToDetail[region]
		require.True(t, ok)
		require.Equal(t, 1, len(regionNetworks.ServiceNetworks))

		service := toServiceName(platform, service2)
		serviceToIPs := testutils.GetServiceNameToIPs(regionNetworks)
		testutils.CheckServiceIPsInRegion(
			t,
			serviceToIPs,
			service,
			[]string{c2r2s2IPv4},
			[]string{})
	}
	// Region 3 (c2)
	{
		region := toRegionName(cloud2, emptyRegion)
		regionNetworks, ok := regionNameToDetail[region]
		require.True(t, ok)
		require.Equal(t, 1, len(regionNetworks.ServiceNetworks))

		service := toServiceName(platform, emptyService)
		serviceToIPs := testutils.GetServiceNameToIPs(regionNetworks)
		testutils.CheckServiceIPsInRegion(
			t,
			serviceToIPs,
			service,
			[]string{c2NoRegionNoServiceIPv4},
			[]string{c2NoRegionNoServiceIPv6})
	}
}

func TestAzureRegionServiceRedundancyCheck(t *testing.T) {
	cloudName := "test-cloud"
	regionName, emptyRegion := "test-region", ""
	platformName := "test-platform"
	serviceName, emptyService := "test-service", ""
	addr := "20.140.48.160/27"

	// Given four entities.
	// First one has:  (cloud, platform/service)
	// Second one has: (cloud, platform)
	// Third one has:  (cloud/region, platform)
	// Fourth one has: (cloud/region, platform/service)
	// After parsing, there should only be one entry left (everything but the fourth one should be cleared).
	testCloud := azureCloud{
		ChangeNumber: testutils.UnusedInt,
		Cloud:        cloudName,
		Values: []azureCloudEntity{
			{
				Name: testutils.UnusedString,
				ID:   testutils.UnusedString,
				Properties: azureCloudEntityProperties{
					ChangeNumber:    testutils.UnusedInt,
					Region:          emptyRegion,
					RegionID:        testutils.UnusedInt,
					Platform:        platformName,
					SystemService:   serviceName,
					AddressPrefixes: []string{addr},
					NetworkFeatures: testutils.UnusedStrSlice,
				},
			},
			{
				Name: testutils.UnusedString,
				ID:   testutils.UnusedString,
				Properties: azureCloudEntityProperties{
					ChangeNumber:    testutils.UnusedInt,
					Region:          emptyRegion,
					RegionID:        testutils.UnusedInt,
					Platform:        platformName,
					SystemService:   emptyService,
					AddressPrefixes: []string{addr},
					NetworkFeatures: testutils.UnusedStrSlice,
				},
			},
			{
				Name: testutils.UnusedString,
				ID:   testutils.UnusedString,
				Properties: azureCloudEntityProperties{
					ChangeNumber:    testutils.UnusedInt,
					Region:          regionName,
					RegionID:        testutils.UnusedInt,
					Platform:        platformName,
					SystemService:   emptyService,
					AddressPrefixes: []string{addr},
					NetworkFeatures: testutils.UnusedStrSlice,
				},
			},
			{
				Name: testutils.UnusedString,
				ID:   testutils.UnusedString,
				Properties: azureCloudEntityProperties{
					ChangeNumber:    testutils.UnusedInt,
					Region:          regionName,
					RegionID:        testutils.UnusedInt,
					Platform:        platformName,
					SystemService:   serviceName,
					AddressPrefixes: []string{addr},
					NetworkFeatures: testutils.UnusedStrSlice,
				},
			},
		},
	}

	cloudNetworks, err := json.Marshal(testCloud)
	require.Nil(t, err)

	crawler := azureNetworkCrawler{}
	parsedResult, err := crawler.parseAzureNetworks([][]byte{cloudNetworks})
	require.Nil(t, err)
	require.Equal(t, parsedResult.ProviderName, crawler.GetProviderKey().String())

	// One region
	require.Equal(t, 1, len(parsedResult.RegionNetworks))
	regionNameToDetail := testutils.GetRegionNameToDetails(parsedResult)

	// Check region content
	{
		region := toRegionName(cloudName, regionName)
		regionNetworks, ok := regionNameToDetail[region]
		require.True(t, ok)
		require.Equal(t, 1, len(regionNetworks.ServiceNetworks))

		service := toServiceName(platformName, serviceName)
		serviceToIPs := testutils.GetServiceNameToIPs(regionNetworks)
		testutils.CheckServiceIPsInRegion(
			t,
			serviceToIPs,
			service,
			[]string{addr},
			[]string{})
	}
}

package oracle

import (
	"encoding/json"
	"testing"

	"github.com/stackrox/external-network-pusher/pkg/common/testutils"
	"github.com/stretchr/testify/require"
)

func TestOCIParseNetwork(t *testing.T) {
	region1, region2 := "us-phoenix-1", "sa-saopaulo-1"
	tag1, tag2, tag3 := "OCI", "OSN", "OBJECT_STORAGE"
	ipv41, ipv42, ipv43, ipv44, ipv45 := "129.146.0.0/21", "129.146.64.0/18", "158.101.0.0/18", "193.123.0.0/19", "207.135.0.0/22"

	testData := ociNetworkSpec{
		LastUpdatedTimestamp: testutils.UnusedString,
		Regions: []ociRegionNetworkDetails{
			{
				Region: region1,
				CIDRs: []ociCIDRDefinition{
					{
						CIDR: ipv41,
						Tags: []string{tag1},
					},
					{
						CIDR: ipv42,
						Tags: []string{tag2, tag3},
					},
					{
						CIDR: ipv43,
						Tags: []string{tag3, tag2},
					},
				},
			},
			{
				Region: region2,
				CIDRs: []ociCIDRDefinition{
					{
						CIDR: ipv44,
						Tags: []string{tag1},
					},
					{
						CIDR: ipv45,
						Tags: []string{tag2},
					},
				},
			},
		},
	}

	networks, err := json.Marshal(testData)
	require.Nil(t, err)

	crawler := ociNetworkCrawler{}
	parsedResult, err := crawler.parseNetworks(networks)
	require.Nil(t, err)
	require.Equal(t, parsedResult.ProviderName, crawler.GetProviderKey().String())

	// Two regions in total
	require.Equal(t, 2, len(parsedResult.RegionNetworks))
	regionNameToDetail := testutils.GetRegionNameToDetails(parsedResult)

	// region1
	{
		regionNetworks, ok := regionNameToDetail[region1]
		require.True(t, ok)
		// Two services in total for region1 (tag1, tag2-tag3)
		require.Equal(t, 2, len(regionNetworks.ServiceNetworks))

		serviceToIPs := testutils.GetServiceNameToIPs(regionNetworks)
		// service1
		service := toServiceName([]string{tag1})
		testutils.CheckServiceIPsInRegion(
			t,
			serviceToIPs,
			service,
			[]string{ipv41},
			[]string{})

		// service2
		service = toServiceName([]string{tag2, tag3})
		testutils.CheckServiceIPsInRegion(
			t,
			serviceToIPs,
			service,
			[]string{ipv42, ipv43},
			[]string{})
	}
	// region2
	{
		regionNetworks, ok := regionNameToDetail[region2]
		require.True(t, ok)
		// Two services in total for region2 (tag1, tag2)
		require.Equal(t, 2, len(regionNetworks.ServiceNetworks))

		serviceToIPs := testutils.GetServiceNameToIPs(regionNetworks)
		// service1
		service := toServiceName([]string{tag1})
		testutils.CheckServiceIPsInRegion(
			t,
			serviceToIPs,
			service,
			[]string{ipv44},
			[]string{})

		// service2
		service = toServiceName([]string{tag2})
		testutils.CheckServiceIPsInRegion(
			t,
			serviceToIPs,
			service,
			[]string{ipv45},
			[]string{})
	}
}

func TestOCIRegionServiceRedundancyCheck(t *testing.T) {
	regionName := "us-phoenix-1"
	tag1, tag2, tag3 := "OCI", "OSN", "OBJECT_STORAGE"
	addr := "129.146.0.0/21"

	testData := ociNetworkSpec{
		LastUpdatedTimestamp: testutils.UnusedString,
		Regions: []ociRegionNetworkDetails{
			{
				Region: regionName,
				CIDRs: []ociCIDRDefinition{
					{
						CIDR: addr,
						Tags: []string{tag1},
					},
					{
						CIDR: addr,
						Tags: []string{tag1},
					},
					{
						CIDR: addr,
						Tags: []string{tag2, tag3},
					},
				},
			},
		},
	}

	networks, err := json.Marshal(testData)
	require.Nil(t, err)

	crawler := ociNetworkCrawler{}
	parsedResult, err := crawler.parseNetworks(networks)
	require.Nil(t, err)
	require.Equal(t, parsedResult.ProviderName, crawler.GetProviderKey().String())

	// One regions in total
	require.Equal(t, 1, len(parsedResult.RegionNetworks))
	regionNameToDetail := testutils.GetRegionNameToDetails(parsedResult)

	// Check region content
	{
		regionNetworks, ok := regionNameToDetail[regionName]
		require.True(t, ok)
		// Two services in total for region1 (tag1, tag2-tag3)
		require.Equal(t, 2, len(regionNetworks.ServiceNetworks))

		serviceToIPs := testutils.GetServiceNameToIPs(regionNetworks)
		// service1
		service := toServiceName([]string{tag1})
		testutils.CheckServiceIPsInRegion(
			t,
			serviceToIPs,
			service,
			[]string{addr},
			[]string{})

		// service2
		service = toServiceName([]string{tag2, tag3})
		testutils.CheckServiceIPsInRegion(
			t,
			serviceToIPs,
			service,
			[]string{addr},
			[]string{})
	}
}

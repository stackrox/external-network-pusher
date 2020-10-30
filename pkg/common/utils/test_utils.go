package utils

import (
	"testing"

	"github.com/stackrox/external-network-pusher/pkg/common"
	"github.com/stretchr/testify/require"
)

var (
	// UnusedInt is used by testing as a placeholder for unused ints
	UnusedInt = -1
	// UnusedStrSlice is used by testing as a placeholder for unused string slices
	UnusedStrSlice = []string{"unused1", "unused2"}
	// UnusedString is used by testing as a placeholder for unused strings
	UnusedString = "UNUSED"
)

// GetServiceNameToIPs creates a map from service name to associated networks for easier lookup
func GetServiceNameToIPs(regionNetworkDetail *common.RegionNetworkDetail) map[string]*common.ServiceIPRanges {
	result := make(map[string]*common.ServiceIPRanges)
	for _, serviceNetworks := range regionNetworkDetail.ServiceNetworks {
		result[serviceNetworks.ServiceName] = serviceNetworks
	}
	return result
}

// GetRegionNameToDetails creates a map from region name to associated network details for easier lookup
func GetRegionNameToDetails(providerNetworks *common.ProviderNetworkRanges) map[string]*common.RegionNetworkDetail {
	regionNameToDetail := make(map[string]*common.RegionNetworkDetail)
	for _, networks := range providerNetworks.RegionNetworks {
		regionNameToDetail[networks.RegionName] = networks
	}
	return regionNameToDetail
}

// CheckServiceIPsInRegion checks the existence of specified networks
func CheckServiceIPsInRegion(
	t *testing.T,
	serviceNameToIPRanges map[string]*common.ServiceIPRanges,
	service string,
	expectedIPv4s, expectedIPv6s []string) {
	ips, ok := serviceNameToIPRanges[service]
	require.True(t, ok)
	require.ElementsMatch(t, ips.IPv4Prefixes, expectedIPv4s)
	require.ElementsMatch(t, ips.IPv6Prefixes, expectedIPv6s)
}

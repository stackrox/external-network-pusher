package main

import (
	"testing"

	"github.com/stackrox/external-network-pusher/pkg/common"
	"github.com/stretchr/testify/require"
)

func TestValidateExternalNetworks(t *testing.T) {
	providerName, regionName, serviceName := "provider", "region", "service"
	providerNetwork := common.ProviderNetworkRanges{
		ProviderName:   "",
		RegionNetworks: nil,
	}
	regionNetworkDetail := common.RegionNetworkDetail{
		RegionName:      "",
		ServiceNetworks: nil,
	}
	serviceNetwork := common.ServiceIPRanges{
		ServiceName:  "",
		IPv4Prefixes: nil,
		IPv6Prefixes: nil,
	}

	var testNetworks common.ExternalNetworkSources

	err := validateExternalNetworks(&testNetworks)
	require.NotNil(t, err)
	require.Equal(t, err.Error(), common.NoProvidersCrawledError().Error())

	// No empty provider name
	testNetworks.ProviderNetworks = append(testNetworks.ProviderNetworks, &providerNetwork)
	err = validateExternalNetworks(&testNetworks)
	require.NotNil(t, err)
	require.Equal(t, err.Error(), common.ProviderNameEmptyError().Error())

	// No provider with empty regions
	providerNetwork.ProviderName = providerName
	err = validateExternalNetworks(&testNetworks)
	require.NotNil(t, err)
	require.Equal(t, err.Error(), common.NoRegionNetworksError(providerName).Error())

	// No empty region name
	providerNetwork.RegionNetworks = append(providerNetwork.RegionNetworks, &regionNetworkDetail)
	err = validateExternalNetworks(&testNetworks)
	require.NotNil(t, err)
	require.Equal(t, err.Error(), common.EmptyRegionNameError(providerName).Error())

	// No region with empty service networks
	regionNetworkDetail.RegionName = regionName
	err = validateExternalNetworks(&testNetworks)
	require.NotNil(t, err)
	require.Equal(t, err.Error(), common.NoServiceNetworksError(providerName, regionName).Error())

	// No empty service name
	regionNetworkDetail.ServiceNetworks = append(regionNetworkDetail.ServiceNetworks, &serviceNetwork)
	err = validateExternalNetworks(&testNetworks)
	require.NotNil(t, err)
	require.Equal(t, err.Error(), common.EmptyServiceNameError(providerName, regionName).Error())

	// No service should have empty IP prefixes
	serviceNetwork.ServiceName = serviceName
	err = validateExternalNetworks(&testNetworks)
	require.NotNil(t, err)
	require.Equal(t, err.Error(), common.NoIPPrefixesError(providerName, regionName, serviceName).Error())

	// Fixed all errors. Should pass the validation
	serviceNetwork.IPv4Prefixes = append(serviceNetwork.IPv4Prefixes, "0.0.0.0/24")
	err = validateExternalNetworks(&testNetworks)
	require.Nil(t, err)
}

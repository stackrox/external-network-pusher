package main

import (
	"testing"

	"github.com/stackrox/external-network-pusher/pkg/common"
	"github.com/stackrox/external-network-pusher/pkg/crawlers/gcp"
	"github.com/stretchr/testify/require"
)

func TestValidateExternalNetworks(t *testing.T) {
	crawler := gcp.NewGCPNetworkCrawler()
	providerName, regionName, serviceName := crawler.GetProviderKey().String(), "region", "service"
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
	crawlers := []common.NetworkCrawler{crawler}

	var testNetworks common.ExternalNetworkSources

	err := validateExternalNetworks(crawlers, &testNetworks)
	require.NotNil(t, err)
	require.Equal(t, err.Error(), common.NumProvidersError(0, 1).Error())

	// No empty provider name
	testNetworks.ProviderNetworks = append(testNetworks.ProviderNetworks, &providerNetwork)
	err = validateExternalNetworks(crawlers, &testNetworks)
	require.NotNil(t, err)
	require.Equal(t, err.Error(), common.ProviderNameEmptyError().Error())

	// No provider with empty regions
	providerNetwork.ProviderName = providerName
	err = validateExternalNetworks(crawlers, &testNetworks)
	require.NotNil(t, err)
	require.Equal(t, err.Error(), common.NoRegionNetworksError(providerName).Error())

	// No empty region name
	providerNetwork.RegionNetworks = append(providerNetwork.RegionNetworks, &regionNetworkDetail)
	err = validateExternalNetworks(crawlers, &testNetworks)
	require.NotNil(t, err)
	require.Equal(t, err.Error(), common.EmptyRegionNameError(providerName).Error())

	// No region with empty service networks
	regionNetworkDetail.RegionName = regionName
	err = validateExternalNetworks(crawlers, &testNetworks)
	require.NotNil(t, err)
	require.Equal(t, err.Error(), common.NoServiceNetworksError(providerName, regionName).Error())

	// No empty service name
	regionNetworkDetail.ServiceNetworks = append(regionNetworkDetail.ServiceNetworks, &serviceNetwork)
	err = validateExternalNetworks(crawlers, &testNetworks)
	require.NotNil(t, err)
	require.Equal(t, err.Error(), common.EmptyServiceNameError(providerName, regionName).Error())

	// No service should have empty IP prefixes
	serviceNetwork.ServiceName = serviceName
	err = validateExternalNetworks(crawlers, &testNetworks)
	require.NotNil(t, err)
	require.Equal(t, err.Error(), common.NoIPPrefixesError(providerName, regionName, serviceName).Error())

	// Throw error when not enough IP prefixes are crawled
	serviceNetwork.IPv4Prefixes = append(serviceNetwork.IPv4Prefixes, "0.0.0.0/24")
	err = validateExternalNetworks(crawlers, &testNetworks)
	require.NotNil(t, err)
	require.Equal(t, err.Error(), common.NotEnoughIPPrefixesError(providerName, 1, crawler.GetNumRequiredIPPrefixes()).Error())

	// Fixed all errors. Should pass the validation
	for {
		if len(serviceNetwork.IPv4Prefixes) == crawler.GetNumRequiredIPPrefixes() {
			break
		}
		serviceNetwork.IPv4Prefixes = append(serviceNetwork.IPv4Prefixes, "some prefix")
	}
	err = validateExternalNetworks(crawlers, &testNetworks)
	require.Nil(t, err)
}

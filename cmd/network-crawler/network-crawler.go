package main

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"os"
	"strings"
	"time"

	"github.com/pkg/errors"
	"github.com/stackrox/external-network-pusher/pkg/common"
	"github.com/stackrox/external-network-pusher/pkg/common/utils"
	"github.com/stackrox/external-network-pusher/pkg/crawlers"
)

// This program crawls a set of external network providers (Google, Amazon, etc.)
// and push crawled IP ranges to a specified Google Cloud bucket.
//
// For every run it creates a header file as per the structure defined
// in common/constants.go, and a folder with list of files containing
// each provider's IP ranges.

// skippedProviderFlag is a flag that takes in a list of Provider names
type skippedProviderFlag []common.Provider

func (f *skippedProviderFlag) String() string {
	strs := make([]string, 0, len(*f))
	for _, p := range *f {
		strs = append(strs, p.String())
	}
	return strings.Join(strs, ",")
}

func (f *skippedProviderFlag) Set(value string) error {
	splitted := strings.Split(value, ",")
	for _, s := range splitted {
		p, err := common.ToProvider(s)
		if err != nil {
			return err
		}
		*f = append(*f, p)
	}
	return nil
}

func main() {
	if err := run(); err != nil {
		log.Fatalf("External network pusher failed: %v", err)
		os.Exit(1)
	}
}

func run() error {
	var (
		flagBucketName       = flag.String("bucket-name", "", "GCS bucket name to upload external networks to")
		flagDryRun           = flag.Bool("dry-run", false, "Skip uploading external networks to GCS")
		flagSkippedProviders skippedProviderFlag
	)
	skippedProvidersUsage :=
		fmt.Sprintf("Comma separated list of providers. Currently acceptable providers are: %v", common.AllProviders())
	flag.Var(&flagSkippedProviders, "skipped-providers", skippedProvidersUsage)
	flag.Parse()

	if *flagDryRun {
		log.Print("Dry run specified. Instead of uploading the content to bucket will just print to stdout.")
	}

	crawlerImpls := crawlers.Get(flagSkippedProviders)
	if len(crawlerImpls) == 0 {
		log.Printf("No provider to crawl.")
		return nil
	}
	crawlingProviders := make([]string, 0, len(crawlerImpls))
	for _, crawler := range crawlerImpls {
		crawlingProviders = append(crawlingProviders, crawler.GetHumanReadableProviderName())
	}
	log.Printf("Crawling from this list of providers: %s", strings.Join(crawlingProviders, ", "))

	err := publishExternalNetworks(*flagBucketName, crawlerImpls, *flagDryRun)
	return err
}

func publishExternalNetworks(
	bucketName string,
	crawlerImpls []common.NetworkCrawler,
	isDryRun bool,
) error {
	// We use the folder name as object prefix so that all the objects
	// uploaded as part of this run appears under the same folder
	objectPrefix := getFolderName()

	var allExternalNetworks common.ExternalNetworkSources
	for _, crawler := range crawlerImpls {
		log.Print("=======")
		log.Printf("Crawing from provider %s...", crawler.GetHumanReadableProviderName())
		providerNetworkRanges, err := crawler.CrawlPublicNetworkRanges()
		if err != nil {
			log.Printf("Failed to crawl networks for %s: %v", crawler.GetHumanReadableProviderName(), err)
			// Hard stop to make the info stored in bucket absolutely correct
			return err
		}
		allExternalNetworks.ProviderNetworks = append(allExternalNetworks.ProviderNetworks, providerNetworkRanges)

		log.Printf("Successfully crawled provider %s", crawler.GetHumanReadableProviderName())
	}

	err := validateExternalNetworks(&allExternalNetworks)
	if err != nil {
		return errors.Wrap(err, "external network sources validation failed")
	}

	// Create and upload the object file
	err = uploadExternalNetworkSources(&allExternalNetworks, isDryRun, bucketName, objectPrefix)
	if err != nil {
		return errors.Wrap(err, "failed to upload data to bucket")
	}

	log.Print("Finished crawling all providers.")
	return nil
}

func validateExternalNetworks(networks *common.ExternalNetworkSources) error {
	// Validate that for each provider we at least have 1 IP prefix so that we are
	// not uploading empty data.
	// Each crawler should be responsible for its own validation of its network
	// ranges.
	if len(networks.ProviderNetworks) == 0 {
		return common.NoProvidersCrawledError()
	}
	for _, provider := range networks.ProviderNetworks {
		providerName := provider.ProviderName
		if providerName == "" {
			return common.ProviderNameEmptyError()
		}
		if len(provider.RegionNetworks) == 0 {
			return common.NoRegionNetworksError(providerName)
		}
		for _, region := range provider.RegionNetworks {
			regionName := region.RegionName
			if regionName == "" {
				return common.EmptyRegionNameError(providerName)
			}
			if len(region.ServiceNetworks) == 0 {
				return common.NoServiceNetworksError(providerName, regionName)
			}
			for _, service := range region.ServiceNetworks {
				serviceName := service.ServiceName
				if serviceName == "" {
					return common.EmptyServiceNameError(providerName, regionName)
				}
				if len(service.IPv4Prefixes) == 0 && len(service.IPv6Prefixes) == 0 {
					return common.NoIPPrefixesError(providerName, regionName, serviceName)
				}
			}
		}
	}
	return nil
}

func uploadExternalNetworkSources(
	networks *common.ExternalNetworkSources,
	isDryRun bool,
	bucketName, objectPrefix string,
) error {
	data, cksum, err := marshalAndGetCksum(networks)
	if err != nil {
		return errors.Wrap(err, "failed to marshal external networks")
	}

	if !isDryRun {
		err := uploadObjectWithPrefix(bucketName, objectPrefix, common.NetworkFileName, data)
		if err != nil {
			return errors.Wrap(err, "failed to upload network ranges")
		}
		err = uploadObjectWithPrefix(bucketName, objectPrefix, common.ChecksumFileName, []byte(cksum))
		if err != nil {
			return errors.Wrapf(err, "content upload succeeded but checksum upload has failed. Checksum: %s", cksum)
		}
		log.Print("Successfully uploaded all contents and checksum.")
		log.Print("++++++")
		log.Printf("Please check bucket: https://console.cloud.google.com/storage/browser/%s", bucketName)
		log.Print("++++++")
	} else {
		// In dry run, just print out the package name and hashes
		log.Printf(
			"Dry run specified. Folder name is: %s. Checksum computed is: %s",
			objectPrefix,
			cksum)
	}

	return nil
}

func uploadObjectWithPrefix(bucketName, objectPrefix, objectName string, data []byte) error {
	err := utils.WriteToBucket(bucketName, objectPrefix, objectName, data)
	if err != nil {
		return errors.Wrapf(
			err,
			"failed to upload content with prefix %s and name %s",
			objectPrefix,
			objectName)
	}
	return nil
}

func marshalAndGetCksum(v interface{}) ([]byte, string, error) {
	data, err := json.Marshal(v)
	if err != nil {
		return nil, "", err
	}

	// Compute checksum
	hash := sha256.Sum256(data)
	checksum := hex.EncodeToString(hash[:])
	return data, checksum, nil
}

func getFolderName() string {
	// Some Go magic here. DO NOT CHANGE THIS STRING
	return time.Now().UTC().Format("2006-01-02 15-04-05")
}

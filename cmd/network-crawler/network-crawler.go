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

	// Remember successfully crawled providers for generating a header file
	crawledProviderObjectNameToChecksum := make(map[string]string)
	for _, crawler := range crawlerImpls {
		log.Print("=======")
		log.Printf("Crawing from provider %s...", crawler.GetHumanReadableProviderName())
		networkRanges, err := crawler.CrawlPublicNetworkRanges()

		if err != nil {
			log.Printf("Failed to crawl networks for %s: %v", crawler.GetHumanReadableProviderName(), err)
			// Keep looping for other providers
			continue
		}

		data, cksum, err := marshalAndGetCksum(networkRanges)
		if err != nil {
			log.Printf("Failed to marshal data for %s: %v", crawler.GetHumanReadableProviderName(), err)
			// Keep looping for other providers
			continue
		}

		if !isDryRun {
			err :=
				uploadObjectWithPrefix(
					bucketName,
					objectPrefix,
					crawler.GetBucketObjectName(),
					crawler.GetHumanReadableProviderName(),
					data)
			if err != nil {
				log.Printf("Skipping upload for %s...", crawler.GetHumanReadableProviderName())
				// Keep looping for other providers
				continue
			}
			log.Printf(
				"Uploaded %s's network to file: %s with prefix: %s",
				crawler.GetHumanReadableProviderName(),
				crawler.GetBucketObjectName(),
				objectPrefix)
		} else {
			// In dry run, just print out the serialized json
			log.Printf("Dry run specified. Content for provider: %s", crawler.GetHumanReadableProviderName())
			log.Printf("%s", string(data))
		}

		// Remember the checksum
		crawledProviderObjectNameToChecksum[crawler.GetBucketObjectName()] = cksum
	}

	// Create header file
	header := getHeaderStruct(objectPrefix, crawledProviderObjectNameToChecksum)
	if !isDryRun && len(crawledProviderObjectNameToChecksum) != 0 {
		err := writeHeaderFile(bucketName, header)
		if err != nil {
			log.Printf("Failed to create and push header file: %s", err)
			return err
		}
		log.Print("++++++")
		log.Printf("Please check bucket: https://console.cloud.google.com/storage/browser/%s", bucketName)
		log.Print("++++++")
	} else if isDryRun {
		// In dry run, just print out the package name and hashes
		log.Printf(
			"Dry run specified. Object prefix is: %s. Object name to hashes: %v",
			objectPrefix,
			crawledProviderObjectNameToChecksum)
	}

	// Check if we were successful in crawling all specified providers
	if len(crawledProviderObjectNameToChecksum) != len(crawlerImpls) {
		var failedProviders []string
		for _, crawler := range crawlerImpls {
			if _, ok := crawledProviderObjectNameToChecksum[crawler.GetBucketObjectName()]; !ok {
				failedProviders = append(failedProviders, crawler.GetHumanReadableProviderName())
			}
		}
		return fmt.Errorf(
			"failed to crawl some of the providers specified: %s. Please refer to logs for further debugging",
			strings.Join(failedProviders, ", "))
	}

	log.Print("Successfully crawled all providers.")
	return nil
}

func uploadObjectWithPrefix(bucketName, objectPrefix, objectName, providerName string, data []byte) error {
	err := utils.WriteToBucket(bucketName, objectPrefix, objectName, data)
	if err != nil {
		log.Printf(
			"Failed to upload %s's network data to with prefix %s and name %s: %v",
			providerName,
			objectPrefix,
			objectName,
			err,
		)
	}
	return err
}

func writeHeaderFile(bucketName string, header *common.Header) error {
	data, cksum, err := marshalAndGetCksum(header)
	if err != nil {
		log.Printf("Failed to marshal header file data: %v", err)
		return err
	}

	// First check and delete any existing header file
	err = utils.DeleteObjectWithPrefix(bucketName, common.HeaderFileName)
	if err != nil {
		log.Printf("Failed to delete existing header file objects: %v", err)
		return err
	}

	headerFileName := common.HeaderFileName + "-" + cksum
	err = utils.WriteToBucket(bucketName, "", headerFileName, data)
	if err != nil {
		log.Printf("Failed to write out header file with name %s: %v", headerFileName, err)
		return err
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

func getHeaderStruct(objectPrefix string, objectNameToChecksum map[string]string) *common.Header {
	return &common.Header{
		ObjectPrefix:         objectPrefix,
		ObjectNameToCheckSum: objectNameToChecksum,
	}
}

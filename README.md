# external-network-pusher

This repo crawls list of external network providers and publishes the info
to a user specified Google bucket. The current list of Providers include:
- Google Cloud
- Amazon AWS
- Microsoft Azure
- Oracle OCI
- Cloudflare

# Code structure
cmd/network-crawler.go
- Contains the main script which crawls and publishes the external network info.

image
- Contains the Dockerfile to package this script into a docker image

pkg/common
- Contains some util functions and commons types/constants definitions. Crawler interface
  definition is also defined here under types.go

pkg/crawlers
- Contains provider specific implementations of crawler instances.

## How to build and run external-network-pusher?
To build, we can simply run
```bash
make build
```
This would create a `.gobin` directory and build a binary suited to your OS platform.

To create an image, just do
```bash
make image
```
or to push to StackRox registry under project `stackrox-hub`, do
```bash
make push
```

## How to run crawlers?
After builing the binary, just do
```bash
.gobin/network-crawler --bucket-name <GCS bucket name>
```

By default it would crawl all the providers listed above, alternatively you can specify specifc set of providers to crawl. For example, if you want to skip crawling for Google Cloud and Amazon AWS, do
```bash
.gobin/network-crawler --bucket-name <GCS bucket name> --skipped-providers Google,Amazon
```

Currently accepted values for `--skipped-providers` are: `Google`, `Amazon`, `Azure`, `Oracle`, `Cloudflare`, separated by commas.

There are some other options for the script as well (ex: `--dry-run` to run without uploading to the bucket). Please use `--help` for more details.

### Output structure
This script uploads to the user specified bucket in the following manner. Under the bucket, you should see:

external-networks/latest/networks
- Main file that contains all the provider networks.

external-networks/latest/checksum
- Contains the checksum for the above networks file

external-networks/latest/timestamp
- Contains the timestamp of when the scripts was triggered.

Later runs will not overwrite the previous run outputs. Instead, it will change the prefix from `external-networks/latest` to `external-networks/<timestamp>` where timestamp is the timestamp file value of that run. Then the scripts uploads new version of the `external-networks/latest` files.

However, as of now the scripts only keeps 10 run records in the bucket. If the script detected that there are more than 10 records, it starts deleting from the oldest one by timestamp.

### URL endpoints
Below are the URL endpoins the crawler uses per provider (defined under `pkg/common/constants.go`)

Google
- https://www.gstatic.com/ipranges/cloud.json

Amazon
- https://ip-ranges.amazonaws.com/ip-ranges.json

Azure
- AzurePublic: https://www.microsoft.com/en-us/download/confirmation.aspx?id=56519
- Azure US Gov: https://www.microsoft.com/en-us/download/confirmation.aspx?id=57063
- Azure China: https://www.microsoft.com/en-us/download/confirmation.aspx?id=57062
- Azure Germany: https://www.microsoft.com/en-us/download/confirmation.aspx?id=57064

Oracle
- https://docs.cloud.oracle.com/en-us/iaas/tools/public_ip_ranges.json

Cloudflare
- https://api.cloudflare.com/client/v4/ips

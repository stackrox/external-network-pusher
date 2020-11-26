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
After building the binary, just do
```bash
.gobin/network-crawler --bucket-name <GCS bucket name>
```
where bucket name is the name of the GCS bucket you want to upload to. It is a required field.

By default it would crawl all the providers listed above, alternatively you can crawl specific set of providers by specifying providers to skip. For example, if you want to skip crawling for Google Cloud and Amazon AWS, do
```bash
.gobin/network-crawler --bucket-name <GCS bucket name> --skipped-providers Google,Amazon
```
Please see `--help` for full list of options.


### Output structure
This script uploads to the user specified bucket in the following manner. Under the bucket, you should see:

external-networks/latest_metadata
- File which contains metadata of the latest networks. All consumers of the crawled network data should only look at this file for the filename which contains the latest networks.

external-networks/\<timestamp\>_\<dynamic_uuid\>/networks
- Main file that contains all the provider networks.

external-networks/\<timestamp\>_\<dynamic_uuid\>/checksum
- Contains the checksum for the above networks file. `latest_metadata` file also contains this info for the latest network data.

Later runs will not overwrite the previous run outputs. Instead, it will upload a new version of the network data, and change the content of the `latest_metadata` file.

As of now the script only keeps 10 run records in the bucket. If the script detected that there are more than 10 records, it starts deleting from the oldest one according to timestamp.

### URL endpoints
URL endpoints used by this crawler is defined in `pkg/common/constants.go`. Please check file for the URLs.

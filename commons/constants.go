package commons

// DEFAULT_REGION is used when a vendor does not
// specify regions for the IP ranges provided
const DEFAULT_REGION = "default"

// HEADER_FILE defines the name of the header file
// uploaded to bucket. When uploading header file,
// we also append the hash of the header file to its
// name.
const HEADER_FILE = "header"

type Provider int

const (
	GOOGLE Provider = iota
)

// The following list of URLs are kept here for easier
// maintenance
var ProviderToUrl = map[Provider]string {
	GOOGLE: "https://www.gstatic.com/ipranges/cloud.json",
}

# external-network-pusher

This repo crawls list of external network providers and publishes the info
to a user specified Google bucket. The current list of Providers include:
- Google Cloud

# Directory structure
main/
- Contains the main script which crawls and publishes the external network info.

commons/
- Contains some util functions and commons types/constants definitions. Crawler interface
  definition is also defined here.

crawlers/
- Contains the implementations of the provider specific crawler.
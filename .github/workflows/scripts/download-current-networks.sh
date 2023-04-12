#!/usr/bin/env bash

# Fetches the current latest networks before updating them through the external networks pusher

set -euo pipefail

mkdir -p /tmp/external-networks
latest_prefix="$(wget -q https://definitions.stackrox.io/external-networks/latest_prefix -O -)"
wget -O /tmp/external-networks/checksum "https://definitions.stackrox.io/${latest_prefix}/checksum"
wget -O /tmp/external-networks/networks "https://definitions.stackrox.io/${latest_prefix}/networks"
test -s /tmp/external-networks/checksum
test -s /tmp/external-networks/networks

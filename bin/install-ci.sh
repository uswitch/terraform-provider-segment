#!/usr/bin/env bash
. bin/utils.sh

PROVIDER_VERSION=$1

chmod +x terraform-provider-segment
install_provider terraform-provider-segment "$PROVIDER_VERSION" "keep"
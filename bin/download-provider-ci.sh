#!/bin/bash
set -e

. bin/utils.sh

version=${1:-$(get_provider_version)}
GITHUB_TOKEN=$TERRAFORM_PROVIDER_GITHUB_TOKEN

if [ -z $GITHUB_TOKEN ]; then
echo "TOKEN NOT SET"
fi

echo "Downloading Segment Terraform Provider $version"

assets=$(curl -s \
-H "Accept: application/vnd.github.v3+json" \
-H "Authorization: token $GITHUB_TOKEN" \
https://api.github.com/repos/uswitch/terraform-provider-segment/releases/tags/${version} | jq -r '.assets_url')

exe=$(curl -s \
-H "Accept: application/vnd.github.v3+json" \
-H "Authorization: token $GITHUB_TOKEN" \
$assets | jq -r '.[] | select(.name == "terraform-provider-segment_'${version}'_linux_amd64") | .url')

curl -L -v \
-H "Accept: application/octet-stream" \
-H "Authorization: token $GITHUB_TOKEN" \
--output terraform-provider-segment \
$exe
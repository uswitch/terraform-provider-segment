#!/usr/bin/env bash

PLATFORMS=(
  darwin.amd64
  linux.386
  linux.amd64
  windows.386
  windows.amd64
)

for platform in ${PLATFORMS[@]}; do
  platform=($(echo $platform | tr '.' ' '))
  echo "Building for ${platform[0]} ${platform[1]}"
  GOOS=${platform[0]} GOARCH=${platform[1]} go build -o ./build/terraform-provider-segment_${platform[0]}_${platform[1]}
done
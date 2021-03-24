#!/usr/bin/env bash

echo "Running shell: $0"
platform=($(echo $1 | tr '.' ' '))
echo "Building for ${platform[0]} ${platform[1]}"
export OS=${platform[0]}
export ARCH=${platform[1]}
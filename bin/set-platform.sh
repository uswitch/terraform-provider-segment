#!/bin/bash

echo "Platform: $PLATFORM"
export OS=$(echo $PLATFORM | cut -d '.' -f 1)
export ARCH=$(echo $PLATFORM | cut -d '.' -f 2)
echo "Building for $OS $ARCH"
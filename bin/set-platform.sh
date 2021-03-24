#!/bin/bash

echo "Platform: $PLATFORM"
P=$(echo $PLATFORM | tr '.' '\n')
export OS=$(echo $P | sed '1!d')
export ARCH=$(echo $P | sed '2!d')
echo "Building for $OS $ARCH"
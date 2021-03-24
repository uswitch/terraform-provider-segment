#!/bin/bash

echo "Platform: $PLATFORM"
PLATFORM=$(echo $PLATFORM | tr '.' '\n')
export OS=$(echo $PLATFORM | sed '1!d')
export ARCH=$(echo $PLATFORM | sed '2!d')
echo "Building for $OS $ARCH"
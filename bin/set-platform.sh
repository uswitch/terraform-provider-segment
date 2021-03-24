#!/bin/bash

echo "Platform: $1"
platform=$(echo $1 | tr '.' '\n')
echo $platform | sed '1!d'
export OS=$(echo $platform | sed '1!d')
export ARCH=$(echo $platform | sed '2!d')
echo "Building for $OS $ARCH"
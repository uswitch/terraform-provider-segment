#!/bin/bash

which sed
platform=$(echo $1 | tr '.' '\n')
export OS=$(echo $platform | sed '1!d')
export ARCH=$(echo $platform | sed '2!d')
echo "Building for $OS $ARCH"
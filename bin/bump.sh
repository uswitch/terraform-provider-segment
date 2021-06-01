#!/usr/bin/env bash

version=$1
type=$2
beta=${3:-}
major=$(echo $version | cut -d'.' -f1)
minor=$(echo $version | cut -d'.' -f2)
patch=$(echo $version | cut -d'.' -f3)

case $type in
  major)
  ((major=major+1))
  minor=0
  patch=0
  ;;

  minor)
  ((minor=minor+1))
  patch=0
  ;;

  patch)
  ((patch=patch+1))
  ;;

  *)
  >&2 echo "Invalid version update type. Use <major|minor|patch>"
  exit 1
esac

if [ ! -z $beta ]; then
  beta="-beta"
fi

echo $major.$minor.$patch$beta
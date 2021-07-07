#!/bin/bash
. bin/utils.sh

version=$1

# Some formatting functions to make output readable

bold() {
  tput bold
  echo -n $1
  tput sgr0
}

green() {
  tput bold
  tput setaf 2
  echo $1
  tput sgr0
}

blue() {
  tput bold
  tput setaf 4
  echo $1
  tput sgr0
}

highlight() {
  tput smso
  echo $1
  tput rmso
}

# Multiplatform utility functions

open_link() {
  local link=$1
  if [[ "${OSTYPE}" == "linux-gnu"* ]]; then
    xdg-open ${link}
  else
    open ${link}
  fi
}

get_dl_dir() {
  if [[ "${OSTYPE}" == "linux-gnu"* ]]; then
    xdg-user-dir DOWNLOAD
  else
    echo ~/Downloads
  fi
}

# Script body

bold "1. "; echo -n "Identifying your platform... "
platform_suffix=$(get_platform_suffix)
if [ -z ${platform_suffix} ]; then
  >&2 echo "Unrecognised platform"
  exit 1
fi
green "${platform_suffix}"


# Download plugin from Github release
plugin_filename="terraform-provider-segment_${version}_${platform_suffix}"
link="https://github.com/uswitch/terraform-provider-segment/releases/download/${version}/${plugin_filename}"
bold "2. "; echo -n "Downloading Segment Terraform Provider from "; blue ${link}
highlight "You can come back to this terminal once the file has been downloaded from your browser. If your download didn't work, request access to the uswitch/terraform-provider-segment repository in #segment-support"
sleep 2
open_link ${link}

# Ask user to confirm download location
dl_location="$(get_dl_dir)/${plugin_filename}"
read -p "Please confirm where the file was downloaded [${dl_location}]: " confirmed_location
if [ -z $confirmed_location ]; then
  confirmed_location=${dl_location}
fi

# Move downloaded plugin to terraform plugin dir
bold "3. "; echo -n "Installing provider"
install_provider $confirmed_location $version

bold "4. "; echo "Segment Terraform Provider was successfully installed locally."
highlight "If you're running on MacOS, you'll have to authorise its execution the first time in the Security Settings"

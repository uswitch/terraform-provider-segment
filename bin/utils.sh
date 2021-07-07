#!/bin/sh
set -e

# Retrieves the Terraform directory name based on the platform and arch
get_platform_suffix() {
  local platform_suffix=""
  if [[ $(uname -s) == "Darwin"* ]]; then
    platform_suffix="darwin_amd64"
  elif [[ $(uname -s) == "Linux"* ]]; then
    if [[ $(uname -m) == "x86_64" ]]; then
      platform_suffix="linux_amd64"
    else
      platform_suffix="linux_386"
    fi
  fi

  echo ${platform_suffix}
}

# Moves the passed provider to the Terraform plugin directory
install_provider() {
  local keep=${3:-}
  local platform_suffix=$(get_platform_suffix)
  local provider_location=$1
  local version=${2:-$(get_provider_version)}
  local plugin_dir=~/.terraform.d/plugins/uswitch.com/segment/segment/${version}/${platform_suffix}
  local plugin_path="${plugin_dir}/terraform-provider-segment"
  echo "Ensuring provider directory exists: ${plugin_dir}"
  mkdir -p "${plugin_dir}"
  
  if [ -z "$keep" ]; then
    echo "Moving provider"
    mv "${provider_location}" "${plugin_path}"
  else
    echo "Copying provider"
    cp "${provider_location}" "${plugin_path}"
  fi
  chmod +x ${plugin_path}
  echo "Provider installed in ${plugin_path}"
}

get_provider_version() {
  cat main.tf | tr '\n' ' ' | sed 's/ //g' | sed 's/.*segment={.*version="\(.\+\)"}}}.*/\1/g'
}

read -d '' JQ_COLOURS << 'EOF' || true
  def c:
  {
    "red": "\\u001b[31m",
    "mag": "\\u001b[35m",
    "green": "\\u001b[32m",
    "reset": "\\u001b[0m",
  };

  def cc($text; color):
    (c | color) + $text + c.reset;

  def r($text):
    cc($text;.red);

  def g($text):
    cc($text;.green);

  def m($text):
    cc($text;.mag);
EOF
export JQ_COLOURS

ensure_reqs() {
  if [ -z $SEGMENT_CATALOG_TOKEN ]
  then
    echo "SEGMENT_CATALOG_TOKEN needs to be set with a valid token. Ask #segment-support to get one"
    exit 1
  fi

  if ! command -v jq &> /dev/null
  then
    echo "jq is required to run this utility: https://stedolan.github.io/jq/"
    exit 1
  fi
}
echo "Triggered by $GITHUB_EVENT_NAME"
version=${GITHUB_SHA}
if [[ "$GITHUB_EVENT_NAME" == *"release"* ]]; then
  version="${GITHUB_REF/refs\/tags\//}"
fi

echo "Version: $version"
echo "::set-output name=VERSION::$version"
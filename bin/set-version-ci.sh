echo "Triggered for ref: $GITHUB_REF"
version=${GITHUB_SHA}
if [[ "$GITHUB_REF" == *"/tags/"* ]]; then
  version="${GITHUB_REF/refs\/tags\//}"
fi

echo "Version: $version"
echo "::set-output name=VERSION::$version"
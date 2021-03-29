if [[ "$GITHUB_EVENT_NAME" == *"tag"* ]]; then
  echo "Running on a tag"
  echo "::set-output name=VERSION::${GITHUB_REF/refs\/tags\//}"
else
  echo "Running on a push"
  echo ::set-output name=VERSION::${GITHUB_SHA}
fi
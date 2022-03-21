#!/usr/bin/env bash

# Set default tag name
LOCAL_DOCKER_REPO="collector"

## Get build directory ##
build_path="$(realpath "$0")"
dir_path="$(dirname "$build_path")"
project_path=$(realpath "$dir_path/../../")

# Set tag name if the docker repo is specified
if [ -n "$DOCKER_REPO" ]; then
  LOCAL_DOCKER_REPO="$DOCKER_REPO"
fi

# Get application version
VERSION=$(< "$project_path/package.json" grep version |
  head -1 |
  awk -F: '{ print $2 }' |
  sed 's/[\",]//g' |
  tr -d '[:space:]')
MINOR_VERSION="$(cut -d '.' -f 1 <<< "$VERSION")"."$(cut -d '.' -f 2 <<< "$VERSION")"
MAJOR_VERSION="$(cut -d '.' -f 1 <<< "$VERSION")"

cd "$project_path" || {
  echo "failed to change directory" 1>&2;
  exit 1;
}

docker build -t "$LOCAL_DOCKER_REPO:$VERSION" -t "$LOCAL_DOCKER_REPO:latest" -f "$dir_path/Dockerfile" .
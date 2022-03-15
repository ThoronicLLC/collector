#!/usr/bin/env bash

## Get build directory ##
build_path="$(realpath "$0")"
dir_path="$(dirname "$build_path")"
project_path=$(realpath "$dir_path/../../")

APPLICATION_NAME="collector"
GIT_COMMIT=$(git rev-list -1 HEAD)
GIT_BRANCH=$(git rev-parse --abbrev-ref HEAD)
BUILD_EPOCH=$(date +'%s')
VERSION=$(< "$project_path/package.json" grep version |
  head -1 |
  awk -F: '{ print $2 }' |
  sed 's/[\",]//g' |
  tr -d '[:space:]')

# Cleanup previous files and folders
echo "Cleaning up old bin files..."
rm "$project_path/bin/osx/collector" &>/dev/null
rm -r "$project_path/bin/osx" &>/dev/null
rm "$project_path/bin/windows/collector.exe" &>/dev/null
rm -r "$project_path/bin/windows" &>/dev/null
rm "$project_path/bin/linux/collector" &>/dev/null
rm -r "$project_path/bin/linux" &>/dev/null
rm "$project_path/bin/collector_$VERSION-osx-64bit.zip" &>/dev/null
rm "$project_path/bin/collector_$VERSION-windows-64bit.zip" &>/dev/null
rm "$project_path/bin/collector_$VERSION-linux-64bit.zip" &>/dev/null

# Create new clean directories
echo "Creating new clean bin directories..."
mkdir -p "$project_path/bin/osx/"
mkdir -p "$project_path/bin/windows/"
mkdir -p "$project_path/bin/linux/"

# Setup application build time variables
LDFLAGS="-s -w"
LDFLAGS+=" -X github.com/ThoronicLLC/collector/cmd.ApplicationName=$APPLICATION_NAME"
LDFLAGS+=" -X github.com/ThoronicLLC/collector/cmd.BuildBranch=$GIT_BRANCH"
LDFLAGS+=" -X github.com/ThoronicLLC/collector/cmd.BuildRevision=$GIT_COMMIT"
LDFLAGS+=" -X github.com/ThoronicLLC/collector/cmd.BuildVersion=$VERSION"
LDFLAGS+=" -X github.com/ThoronicLLC/collector/cmd.BuildEnv=production"
LDFLAGS+=" -X github.com/ThoronicLLC/collector/cmd.BuildDate=$BUILD_EPOCH"

# Compile for OSX
echo "Compiling OSX binary..."
CGO_ENABLED=0 GOOS=darwin GOARCH=amd64 go build \
  -o "$project_path/bin/osx/collector" \
  -ldflags "$LDFLAGS" \
  . || {
 echo "failed to build osx binary!" 1>&2;
 exit 1;
}

# Compile for Windows
echo "Compiling Windows binary..."
CGO_ENABLED=0 GOOS=windows GOARCH=amd64 go build \
  -o "$project_path/bin/windows/collector.exe" \
  -ldflags "$LDFLAGS" \
  . || {
 echo "failed to build windows binary!" 1>&2;
 exit 1;
}

# Compile for Linux
echo "Compiling Linux binary..."
CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build \
  -o "$project_path/bin/linux/collector" \
  -ldflags "$LDFLAGS" \
  . || {
 echo "failed to build linux binary!" 1>&2;
 exit 1;
}

# Package binaries into ZIP files
echo "Packaging OSX binary..."
zip "$project_path/bin/collector_$VERSION-osx-64bit.zip" "$project_path/bin/osx/collector"
echo "Packaging Windows binary..."
zip "$project_path/bin/collector_$VERSION-windows-64bit.zip" "$project_path/bin/windows/collector.exe"
echo "Packaging Linux binary..."
zip "$project_path/bin/collector_$VERSION-linux-64bit.zip" "$project_path/bin/linux/collector"

echo "Finished..."

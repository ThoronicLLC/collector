#!/usr/bin/env bash

for arg in "$@"; do
  shift
  case "$arg" in
    "--linux")   set -- "$@" "-l" ;;
    "--osx")     set -- "$@" "-o" ;;
    "--windows") set -- "$@" "-w" ;;
    "--zip")     set -- "$@" "-a" ;;
    "--verbose") set -- "$@" "-v" ;;
    "--clean")   set -- "$@" "-c" ;;
    *) ;;
  esac
done

while getopts "lowavc" opt; do
  case "$opt" in
    l) LINUX=1 ;;
    o) OSX=1 ;;
    w) WINDOWS=1 ;;
    a) ZIP=1 ;;
    v) VERBOSE=1 ;;
    c) CLEAN=1 ;;
    *) echo "unknown option $opt";;
  esac
done
shift $(( OPTIND - 1 ))

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
if [ -n "$CLEAN" ]; then
  if [ -n "$VERBOSE" ]; then
    echo "Cleaning up old bin files..."
  fi
  rm "$project_path/bin/osx/collector" &>/dev/null
  rm -r "$project_path/bin/osx" &>/dev/null
  rm "$project_path/bin/windows/collector.exe" &>/dev/null
  rm -r "$project_path/bin/windows" &>/dev/null
  rm "$project_path/bin/linux/collector" &>/dev/null
  rm -r "$project_path/bin/linux" &>/dev/null
  rm "$project_path/bin/collector_$VERSION-osx-64bit.zip" &>/dev/null
  rm "$project_path/bin/collector_$VERSION-windows-64bit.zip" &>/dev/null
  rm "$project_path/bin/collector_$VERSION-linux-64bit.zip" &>/dev/null
fi

# Create new clean directories
if [ -n "$VERBOSE" ]; then
  echo "Creating new clean bin directories..."
fi
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
if [ -n "$OSX" ]; then
  if [ -n "$VERBOSE" ]; then
    echo "Compiling OSX binary..."
  fi
  CGO_ENABLED=0 GOOS=darwin GOARCH=amd64 go build \
    -o "$project_path/bin/osx/collector" \
    -ldflags "$LDFLAGS" \
    . || {
   echo "failed to build osx binary!" 1>&2;
   exit 1;
  }
fi

# Compile for Windows
if [ -n "$WINDOWS" ]; then
  if [ -n "$VERBOSE" ]; then
    echo "Compiling Windows binary..."
  fi
  CGO_ENABLED=0 GOOS=windows GOARCH=amd64 go build \
    -o "$project_path/bin/windows/collector.exe" \
    -ldflags "$LDFLAGS" \
    . || {
   echo "failed to build windows binary!" 1>&2;
   exit 1;
  }
fi

# Compile for Linux
if [ -n "$LINUX" ]; then
  if [ -n "$VERBOSE" ]; then
    echo "Compiling Linux binary..."
  fi
  CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build \
    -o "$project_path/bin/linux/collector" \
    -ldflags "$LDFLAGS" \
    . || {
   echo "failed to build linux binary!" 1>&2;
   exit 1;
  }

  chmod +x "$project_path/bin/windows/collector.exe"
fi

# Package binaries into ZIP files
if [ -n "$ZIP" ]; then
  if [ -n "$OSX" ]; then
    if [ -n "$VERBOSE" ]; then
      echo "Packaging OSX binary..."
    fi
    zip "$project_path/bin/collector-osx-64bit.zip" "$project_path/bin/osx/collector" "$project_path/LICENSE" "$project_path/README.md"
  fi

  if [ -n "$WINDOWS" ]; then
    if [ -n "$VERBOSE" ]; then
      echo "Packaging Windows binary..."
    fi
    zip "$project_path/bin/collector-windows-64bit.zip" "$project_path/bin/windows/collector.exe" "$project_path/LICENSE" "$project_path/README.md"
  fi

  if [ -n "$LINUX" ]; then
    if [ -n "$VERBOSE" ]; then
      echo "Packaging Linux binary..."
    fi
    zip "$project_path/bin/collector-linux-64bit.zip" "$project_path/bin/linux/collector" "$project_path/LICENSE" "$project_path/README.md"
  fi
fi

echo "Finished building $APPLICATION_NAME binary!"

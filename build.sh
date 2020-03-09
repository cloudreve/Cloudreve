#!/usr/bin/env sh

REPO=$(cd $(dirname $0); pwd)
COMMIT_SHA=$(git rev-parse --short HEAD)
VERSION=$(git describe --tags)
ASSETS="false"
BINARY="false"
RELEASE="false"

debugInfo () {
  echo "Repo:           $REPO"
  echo "Build assets:   $ASSETS"
  echo "Build binary:   $BINARY"
  echo "Release:        $RELEASE"
  echo "Version:        $VERSION"
}

buildAssets () {
  cd $REPO
  rm -rf assets/build
  rm -f statik/statik.go

  cd $REPO/assets

  yarn install
  yarn run build

  if ! [ -x "$(command -v statik)" ]; then
    go get github.com/rakyll/statik
  fi

  cd $REPO
  statik -src=assets/build/  -include=*.html,*.js,*.json,*.css,*.png,*.svg,*.ico -f
}

buildBinary () {
  cd $REPO
  go build -a -o cloudreve -ldflags " -X 'github.com/HFO4/cloudreve/pkg/conf.BackendVersion=$VERSION' -X 'github.com/HFO4/cloudreve/pkg/conf.LastCommit=$COMMIT_SHA'"
}

_build() {
    local osarch=$1
    IFS=/ read -r -a arr <<<"$osarch"
    os="${arr[0]}"
    arch="${arr[1]}"

    # Go build to build the binary.
    export GOOS=$os
    export GOARCH=$arch

    go build -a -o cloudreve_$VERSION_$GOOS_$GOARCH -ldflags " -X 'github.com/HFO4/cloudreve/pkg/conf.BackendVersion=$VERSION' -X 'github.com/HFO4/cloudreve/pkg/conf.LastCommit=$COMMIT_SHA'"
}

release(){
  cd $REPO
  export CGO_ENABLED=1
  ## List of architectures and OS to test coss compilation.
  SUPPORTED_OSARCH="linux/arm64 darwin/amd64 windows/amd64 linux/arm linux/386 windows/386"

  echo "Release builds for OS/Arch: ${SUPPORTED_OSARCH}"
  for each_osarch in ${SUPPORTED_OSARCH}; do
      _build "${each_osarch}"
  done
}

usage() {
  echo "Usage: $0 [-a] [-c] [-b] [-r]" 1>&2;
  exit 1;
}

while getopts "bacr:d" o; do
  case "${o}" in
    b)
      ASSETS="true"
      BINARY="true"
      ;;
    a)
      ASSETS="true"
      ;;
    c)
      BINARY="true"
      ;;
    r)
      ASSETS="true"
      RELEASE="true"
      ;;
    d)
      DEBUG="true"
      ;;
    *)
      usage
      ;;
  esac
done
shift $((OPTIND-1))

if [ "$DEBUG" = "true" ]; then
  debugInfo
fi

if [ "$ASSETS" = "true" ]; then
  buildAssets
fi

if [ "$BINARY" = "true" ]; then
  buildBinary
fi

if [ "$RELEASE" = "true" ]; then
  release
fi
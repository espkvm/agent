#!/bin/sh
# Build the agent for every OS it runs on into dist/. No cgo, so one machine
# builds them all. Usage: sh build.sh [version]
set -e
cd "$(dirname "$0")"
v=${1:-$(git describe --tags --always --dirty 2>/dev/null || echo dev)}
rm -rf dist && mkdir dist
for t in linux/amd64 linux/arm64 linux/arm windows/amd64 windows/arm64 darwin/amd64 darwin/arm64; do
    os=${t%/*} arch=${t#*/} ext=
    [ "$os" = windows ] && ext=.exe
    CGO_ENABLED=0 GOOS=$os GOARCH=$arch go build -trimpath -ldflags "-s -w -X main.version=$v" \
        -o "dist/espkvm-agent-$os-$arch$ext" ./cmd/espkvm-agent
done
ls -l dist

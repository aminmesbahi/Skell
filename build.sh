#!/usr/bin/env bash
# build.sh — Cross-platform release build script for Skell.
#
# Usage:
#   ./build.sh [version]
#
# Examples:
#   ./build.sh v0.1.0        # build all platforms tagged v0.1.0
#   ./build.sh               # uses "dev" as version string
#
# Outputs binaries to ./dist/

set -euo pipefail

VERSION="${1:-dev}"
MODULE="github.com/aminmesbahi/skell/internal/version"
LDFLAGS="-s -w -X ${MODULE}.Version=${VERSION}"
DIST="dist"

PLATFORMS=(
  "windows/amd64"
  "windows/arm64"
  "linux/amd64"
  "linux/arm64"
  "darwin/amd64"
  "darwin/arm64"
)

echo "Building Skell ${VERSION}"
echo "Output directory: ${DIST}/"
echo ""

rm -rf "${DIST}"
mkdir -p "${DIST}"

for PLATFORM in "${PLATFORMS[@]}"; do
  GOOS="${PLATFORM%/*}"
  GOARCH="${PLATFORM#*/}"

  BINARY="skell_${GOOS}_${GOARCH}"
  if [[ "${GOOS}" == "windows" ]]; then
    BINARY="${BINARY}.exe"
  fi

  printf "  %-35s" "${BINARY}"
  GOOS="${GOOS}" GOARCH="${GOARCH}" go build \
    -trimpath \
    -ldflags "${LDFLAGS}" \
    -o "${DIST}/${BINARY}" \
    .
  echo "OK"
done

echo ""
echo "Done. Artifacts in ./${DIST}/"
ls -lh "${DIST}/"

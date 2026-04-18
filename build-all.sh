#!/usr/bin/env bash
# build-all.sh — cross-compile skell for all supported platforms.
#
# Usage:
#   ./build-all.sh [VERSION]
#
# If VERSION is not supplied the script tries to read it from the latest git
# tag (e.g. v0.1.0). Falls back to "dev".
#
# Output goes to ./dist/

set -euo pipefail

VERSION="${1:-$(git describe --tags --abbrev=0 2>/dev/null || echo "dev")}"
COMMIT="$(git rev-parse --short HEAD 2>/dev/null || echo "none")"
DATE="$(date -u +%Y-%m-%dT%H:%M:%SZ)"
MODULE="github.com/aminmesbahi/skell"
LDFLAGS="-s -w \
  -X ${MODULE}/internal/version.Version=${VERSION} \
  -X ${MODULE}/internal/version.Commit=${COMMIT} \
  -X ${MODULE}/internal/version.Date=${DATE}"

DIST="./dist"
mkdir -p "${DIST}"

declare -A TARGETS=(
  ["linux/amd64"]="skell_linux_amd64"
  ["linux/arm64"]="skell_linux_arm64"
  ["darwin/amd64"]="skell_darwin_amd64"
  ["darwin/arm64"]="skell_darwin_arm64"
  ["windows/amd64"]="skell_windows_amd64.exe"
)

echo "Building skell ${VERSION} (commit=${COMMIT})"
echo ""

for platform in "${!TARGETS[@]}"; do
  GOOS="${platform%/*}"
  GOARCH="${platform#*/}"
  OUT="${DIST}/${TARGETS[$platform]}"

  printf "  %-30s → %s\n" "${platform}" "${OUT}"
  CGO_ENABLED=0 GOOS="${GOOS}" GOARCH="${GOARCH}" \
    go build -trimpath -ldflags "${LDFLAGS}" -o "${OUT}" .
done

echo ""
echo "Generating checksums..."
(cd "${DIST}" && sha256sum skell_* > checksums.txt)

echo ""
echo "Done. Artifacts in ${DIST}/"
ls -lh "${DIST}/"

#!/usr/bin/env bash
# build.sh — Convenience wrapper. Delegates to build-all.sh.
#
# Usage:
#   ./build.sh [version]

set -euo pipefail

exec "$(dirname "$0")/build-all.sh" "${1:-}"

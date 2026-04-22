# build.ps1 — Convenience wrapper. Delegates to build-all.ps1.
#
# Usage:
#   .\build.ps1 [-Version v0.1.0]

param(
    [string]$Version = ""
)

& "$PSScriptRoot\build-all.ps1" -Version $Version

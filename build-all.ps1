# build-all.ps1 — cross-compile skell for all supported platforms.
#
# Usage:
#   .\build-all.ps1 [-Version v0.1.0]
#
# If -Version is not supplied, the script reads the latest git tag.
# Output goes to .\dist\

param(
    [string]$Version = ""
)

Set-StrictMode -Version Latest
$ErrorActionPreference = "Stop"

# Resolve version
if (-not $Version) {
    try {
        $Version = (git describe --tags --abbrev=0 2>$null).Trim()
    } catch {}
    if (-not $Version) { $Version = "dev" }
}

$Commit = try { (git rev-parse --short HEAD 2>$null).Trim() } catch { "none" }
$Date   = (Get-Date -Format "yyyy-MM-ddTHH:mm:ssZ")
$Module = "github.com/aminmesbahi/skell"
$Ldflags = "-s -w " +
    "-X ${Module}/internal/version.Version=${Version} " +
    "-X ${Module}/internal/version.Commit=${Commit} " +
    "-X ${Module}/internal/version.Date=${Date}"

$Dist = ".\dist"
New-Item -ItemType Directory -Force -Path $Dist | Out-Null

$Targets = @(
    @{ GOOS="linux";   GOARCH="amd64"; Out="skell_linux_amd64"       }
    @{ GOOS="linux";   GOARCH="arm64"; Out="skell_linux_arm64"       }
    @{ GOOS="darwin";  GOARCH="amd64"; Out="skell_darwin_amd64"      }
    @{ GOOS="darwin";  GOARCH="arm64"; Out="skell_darwin_arm64"      }
    @{ GOOS="windows"; GOARCH="amd64"; Out="skell_windows_amd64.exe" }
)

Write-Host "Building skell $Version (commit=$Commit)" -ForegroundColor Cyan
Write-Host ""

foreach ($t in $Targets) {
    $OutPath = Join-Path $Dist $t.Out
    Write-Host ("  {0,-30} -> {1}" -f "$($t.GOOS)/$($t.GOARCH)", $OutPath)

    $env:GOOS       = $t.GOOS
    $env:GOARCH     = $t.GOARCH
    $env:CGO_ENABLED = "0"

    go build -trimpath -ldflags $Ldflags -o $OutPath .

    Remove-Item Env:\GOOS
    Remove-Item Env:\GOARCH
    Remove-Item Env:\CGO_ENABLED
}

Write-Host ""
Write-Host "Generating checksums..." -ForegroundColor Cyan

Push-Location $Dist
$hashes = Get-ChildItem -Filter "skell_*" | ForEach-Object {
    $hash = (Get-FileHash $_.Name -Algorithm SHA256).Hash.ToLower()
    "$hash  $($_.Name)"
}
$hashes | Set-Content "checksums.txt" -Encoding UTF8
Pop-Location

Write-Host ""
Write-Host "Done. Artifacts in $Dist\" -ForegroundColor Green
Get-ChildItem $Dist | Select-Object Name, @{N="Size";E={"{0:N0} KB" -f ($_.Length/1KB)}}

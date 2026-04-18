# build.ps1 — Cross-platform release build script for Skell (PowerShell).
#
# Usage:
#   .\build.ps1 [-Version v0.1.0]
#
# Examples:
#   .\build.ps1 -Version v0.1.0    # build all platforms tagged v0.1.0
#   .\build.ps1                    # uses "dev" as version string
#
# Outputs binaries to .\dist\

param(
    [string]$Version = "dev"
)

$Module   = "github.com/aminmesbahi/skell/internal/version"
$LdFlags  = "-s -w -X ${Module}.Version=${Version}"
$Dist     = "dist"

$Platforms = @(
    @{ GOOS = "windows"; GOARCH = "amd64" }
    @{ GOOS = "windows"; GOARCH = "arm64" }
    @{ GOOS = "linux";   GOARCH = "amd64" }
    @{ GOOS = "linux";   GOARCH = "arm64" }
    @{ GOOS = "darwin";  GOARCH = "amd64" }
    @{ GOOS = "darwin";  GOARCH = "arm64" }
)

Write-Host "Building Skell $Version"
Write-Host "Output directory: $Dist\"
Write-Host ""

if (Test-Path $Dist) { Remove-Item -Recurse -Force $Dist }
New-Item -ItemType Directory -Path $Dist | Out-Null

$allOk = $true

foreach ($p in $Platforms) {
    $binary = "skell_$($p.GOOS)_$($p.GOARCH)"
    if ($p.GOOS -eq "windows") { $binary += ".exe" }

    $padded = $binary.PadRight(35)
    Write-Host -NoNewline "  $padded"

    $env:GOOS   = $p.GOOS
    $env:GOARCH = $p.GOARCH

    $result = & go build -trimpath -ldflags $LdFlags -o "$Dist\$binary" . 2>&1
    if ($LASTEXITCODE -ne 0) {
        Write-Host "FAILED"
        Write-Host "    $result" -ForegroundColor Red
        $allOk = $false
    } else {
        Write-Host "OK"
    }
}

# Restore env
Remove-Item Env:GOOS   -ErrorAction SilentlyContinue
Remove-Item Env:GOARCH -ErrorAction SilentlyContinue

Write-Host ""
if ($allOk) {
    Write-Host "Done. Artifacts in .\$Dist\" -ForegroundColor Green
    Get-ChildItem $Dist | Select-Object Name, @{N="Size";E={"{0:N0} KB" -f ($_.Length/1KB)}} | Format-Table -AutoSize
} else {
    Write-Host "Build completed with errors." -ForegroundColor Red
    exit 1
}

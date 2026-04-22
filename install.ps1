#Requires -Version 5.1
<#
.SYNOPSIS
    Install or uninstall skell on Windows
.DESCRIPTION
    Downloads and installs the latest skell release from GitHub.
    Pass -Uninstall to remove a previously installed skell.
.EXAMPLE
    irm https://raw.githubusercontent.com/aminmesbahi/skell/main/install.ps1 | iex
.EXAMPLE
    & ([scriptblock]::Create((irm https://raw.githubusercontent.com/aminmesbahi/skell/main/install.ps1))) -Uninstall
#>
param(
    [switch]$Uninstall
)

$ErrorActionPreference = "Stop"

# PowerShell 5.1 defaults to TLS 1.0/1.1; GitHub requires TLS 1.2+
[Net.ServicePointManager]::SecurityProtocol = [Net.SecurityProtocolType]::Tls12

$Repo       = "aminmesbahi/skell"
$BinaryName = "skell"

function Write-Info { param($Message) Write-Host $Message -ForegroundColor Green }
function Write-Warn { param($Message) Write-Host $Message -ForegroundColor Yellow }
function Write-Err  { param($Message) Write-Host $Message -ForegroundColor Red; exit 1 }

# Detect architecture
function Get-Arch {
    $arch = $env:PROCESSOR_ARCHITECTURE
    switch ($arch) {
        "AMD64" { return "amd64" }
        "ARM64" { return "arm64" }
        "x86"   { return "386"   }
        default { Write-Err "Unsupported architecture: $arch" }
    }
}

# Get latest version using redirect (avoids API rate limit)
function Get-LatestVersion {
    try {
        $response = Invoke-WebRequest -Uri "https://github.com/$Repo/releases/latest" `
            -MaximumRedirection 0 -UseBasicParsing -ErrorAction SilentlyContinue
    } catch {
        # PowerShell 5.1 throws on 3xx redirects; extract from the exception
        $response = $_.Exception.Response
    }

    $location = ""
    if ($response -and $response.Headers -and $response.Headers["Location"]) {
        $location = $response.Headers["Location"]
    }

    if ($location -match "/tag/([^/\s]+)") {
        return $Matches[1]
    }

    # Fallback to API if redirect fails
    try {
        $release = Invoke-RestMethod -Uri "https://api.github.com/repos/$Repo/releases/latest"
        return $release.tag_name
    } catch {
        Write-Err "Failed to get latest version. Please check your internet connection."
    }
}

# Get install directory
function Get-InstallDir {
    $dir = "$env:LOCALAPPDATA\Programs\skell"
    if (-not (Test-Path $dir)) {
        New-Item -ItemType Directory -Path $dir -Force | Out-Null
    }
    return $dir
}

# Add to PATH if not already present
function Add-ToPath {
    param($Dir)

    $currentPath = [Environment]::GetEnvironmentVariable("Path", "User")
    if ($currentPath -notlike "*$Dir*") {
        Write-Info "Adding $Dir to PATH..."
        [Environment]::SetEnvironmentVariable("Path", "$currentPath;$Dir", "User")
        $env:Path = "$env:Path;$Dir"
        return $true
    }
    return $false
}

function Install-Skell {
    Write-Info "Installing skell..."
    Write-Host ""

    $arch       = Get-Arch
    $version    = Get-LatestVersion
    $versionNum = $version.TrimStart("v")
    $installDir = Get-InstallDir

    $url = "https://github.com/$Repo/releases/download/$version/${BinaryName}_${versionNum}_windows_${arch}.zip"

    Write-Info "Downloading skell $version for windows/$arch..."

    # Create temp directory
    $tempDir = Join-Path $env:TEMP "skell-install-$(Get-Random)"
    New-Item -ItemType Directory -Path $tempDir -Force | Out-Null

    try {
        $zipPath = Join-Path $tempDir "skell.zip"

        # Download
        Invoke-WebRequest -Uri $url -OutFile $zipPath -UseBasicParsing

        # Extract
        Expand-Archive -Path $zipPath -DestinationPath $tempDir -Force

        # Find and move binary
        $exePath = Join-Path $tempDir "$BinaryName.exe"
        if (-not (Test-Path $exePath)) {
            Write-Err "Binary not found in archive"
        }

        $destPath = Join-Path $installDir "$BinaryName.exe"
        Move-Item -Path $exePath -Destination $destPath -Force

        # Add to PATH
        $pathAdded = Add-ToPath -Dir $installDir

        Write-Host ""
        Write-Info "Successfully installed skell to $destPath"
        Write-Host ""

        # Show version
        & $destPath --version

        Write-Host ""
        if ($pathAdded) {
            Write-Warn "PATH updated. Restart your terminal for changes to take effect."
            Write-Host ""
        }

        Write-Info "Get started:"
        Write-Info "  skell init"
        Write-Info "  skell --help"

    } finally {
        # Cleanup
        Remove-Item -Path $tempDir -Recurse -Force -ErrorAction SilentlyContinue
    }
}

function Uninstall-Skell {
    Write-Info "Uninstalling skell..."
    Write-Host ""

    $installDir = "$env:LOCALAPPDATA\Programs\skell"
    $exePath    = Join-Path $installDir "$BinaryName.exe"

    if (Test-Path $exePath) {
        Remove-Item -Path $exePath -Force
        Write-Info "Removed $exePath"
    } else {
        Write-Warn "skell binary not found at $exePath"
    }

    # Remove the install directory if it is now empty
    if ((Test-Path $installDir) -and (Get-ChildItem $installDir -Force | Measure-Object).Count -eq 0) {
        Remove-Item -Path $installDir -Force
    }

    # Remove from user PATH
    $currentPath = [Environment]::GetEnvironmentVariable("Path", "User")
    if ($currentPath -like "*$installDir*") {
        $newPath = ($currentPath -split ";" | Where-Object { $_ -ne $installDir }) -join ";"
        [Environment]::SetEnvironmentVariable("Path", $newPath, "User")
        Write-Info "Removed $installDir from PATH"
    }

    Write-Host ""
    Write-Info "skell has been uninstalled."
}

if ($Uninstall) { Uninstall-Skell } else { Install-Skell }

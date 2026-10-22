#Requires -Version 5.1
<#
.SYNOPSIS
    Install or uninstall skell on Windows
.DESCRIPTION
    Downloads and installs the latest skell release from GitHub.
    Pass -Uninstall to remove a previously installed skell. Pass -NoGui (or set
    SKELL_NO_GUI=1) to install only the command-line tool, without the desktop app.
.EXAMPLE
    irm https://raw.githubusercontent.com/aminmesbahi/skell/main/install.ps1 | iex
.EXAMPLE
    & ([scriptblock]::Create((irm https://raw.githubusercontent.com/aminmesbahi/skell/main/install.ps1))) -Uninstall
#>
param(
    [switch]$Uninstall,
    [switch]$NoGui
)

$ErrorActionPreference = "Stop"
# Windows PowerShell 5.1 renders a progress bar that slows downloads by an
# order of magnitude.
$ProgressPreference = "SilentlyContinue"
if ($env:SKELL_NO_GUI) { $NoGui = $true }

# PowerShell 5.1 defaults to TLS 1.0/1.1; GitHub requires TLS 1.2+
[Net.ServicePointManager]::SecurityProtocol = [Net.SecurityProtocolType]::Tls12

$Repo       = "aminmesbahi/skell"
$BinaryName = "skell"
$GuiBinaryName = "skell-gui"

function Write-Info { param($Message) Write-Host $Message -ForegroundColor Green }
function Write-Warn { param($Message) Write-Host $Message -ForegroundColor Yellow }
function Write-Err  { param($Message) Write-Host $Message -ForegroundColor Red; exit 1 }

# Detect architecture
function Get-Arch {
    $arch = $env:PROCESSOR_ARCHITECTURE
    switch ($arch) {
        "AMD64" { return "amd64" }
        "ARM64" { return "arm64" }
        default { Write-Err "Unsupported architecture: $arch (skell publishes windows/amd64 and windows/arm64 builds)" }
    }
}

# Get latest version using redirect (avoids API rate limit)
function Get-LatestVersion {
    # A specific release can be requested with SKELL_VERSION (e.g. v0.2.0).
    if ($env:SKELL_VERSION) {
        if ($env:SKELL_VERSION.StartsWith("v")) { return $env:SKELL_VERSION }
        return "v$($env:SKELL_VERSION)"
    }
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
    # Compare whole PATH entries (a substring match would treat C:	oolsskell2
    # as already containing C:	oolsskell).
    $entries = @($currentPath -split ";" | Where-Object { $_ })
    if (@($entries | Where-Object { $_.TrimEnd("\") -ieq $Dir.TrimEnd("\") }).Count -eq 0) {
        Write-Info "Adding $Dir to PATH..."
        [Environment]::SetEnvironmentVariable("Path", (($entries + $Dir) -join ";"), "User")
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

    # Current releases: skell_bundle_<ver>_windows_<arch>.zip (CLI + desktop app)
    # and skell_cli_<ver>_windows_<arch>.zip. Older releases used
    # skell_<ver>_windows_<arch>[_bundle].zip, which is tried when the new name
    # does not exist (for example with SKELL_VERSION=v0.14.0).
    $releaseBase     = "https://github.com/$Repo/releases/download/$version"
    $bundleUrl       = "$releaseBase/${BinaryName}_bundle_${versionNum}_windows_${arch}.zip"
    $cliUrl          = "$releaseBase/${BinaryName}_cli_${versionNum}_windows_${arch}.zip"
    $legacyBundleUrl = "$releaseBase/${BinaryName}_${versionNum}_windows_${arch}_bundle.zip"
    $legacyCliUrl    = "$releaseBase/${BinaryName}_${versionNum}_windows_${arch}.zip"

    # Downloads the first URL that exists and returns it, or $null.
    function Save-FirstAvailable {
        param([string[]]$Urls, [string]$OutFile)
        foreach ($u in $Urls) {
            try {
                Invoke-WebRequest -Uri $u -OutFile $OutFile -UseBasicParsing
                return $u
            } catch {
                continue
            }
        }
        return $null
    }

    Write-Info "Downloading skell $version for windows/$arch..."

    # Create temp directory
    $tempDir = Join-Path $env:TEMP "skell-install-$(Get-Random)"
    New-Item -ItemType Directory -Path $tempDir -Force | Out-Null

    try {
        $zipPath = Join-Path $tempDir "skell.zip"

        # Returns the expected SHA-256 for $assetName from a checksum file
        # (lines of "<sha256>  <name>"), or $null when unavailable.
        function Get-ExpectedHash {
            param($SumsUrl, $AssetName, $SumsPath)
            try {
                Invoke-WebRequest -Uri $SumsUrl -OutFile $SumsPath -UseBasicParsing
            } catch {
                return $null
            }
            foreach ($line in Get-Content $SumsPath) {
                $parts = $line -split '\s+'
                if ($parts.Count -ge 2 -and $parts[1].TrimStart('*') -eq $AssetName) { return $parts[0] }
            }
            return $null
        }

        $cliSumsUrl = "https://github.com/$Repo/releases/download/$version/checksums.txt"
        $sumsPath   = Join-Path $tempDir "checksums.txt"
        $downloadedBundle = $false
        $assetUrl   = $cliUrl
        $expected   = $null

        # The GUI bundle is assembled after GoReleaser and publishes its own
        # <bundle>.sha256; releases without it fall back to the CLI-only package.
        if ($NoGui) {
            Write-Info "Skipping the desktop app (-NoGui)."
        } else {
            $gotBundle = Save-FirstAvailable -Urls @($bundleUrl, $legacyBundleUrl) -OutFile $zipPath
            if ($gotBundle) {
                $expected = Get-ExpectedHash -SumsUrl "$gotBundle.sha256" -AssetName (Split-Path $gotBundle -Leaf) -SumsPath $sumsPath
                if ($expected) {
                    $downloadedBundle = $true
                    $assetUrl = $gotBundle
                } else {
                    Write-Warn "No checksum published for the GUI bundle of $version. Falling back to CLI-only package."
                }
            } else {
                Write-Warn "GUI bundle not available for $version yet. Falling back to CLI-only package."
            }
        }

        if (-not $downloadedBundle) {
            $gotCli = Save-FirstAvailable -Urls @($cliUrl, $legacyCliUrl) -OutFile $zipPath
            if (-not $gotCli) {
                Write-Err "Could not download the command-line package for $version from $releaseBase."
            }
            $assetUrl = $gotCli
            $expected = Get-ExpectedHash -SumsUrl $cliSumsUrl -AssetName (Split-Path $gotCli -Leaf) -SumsPath $sumsPath
        }

        # Verify the archive against the release's published SHA-256.
        $assetName = Split-Path $assetUrl -Leaf
        if (-not $expected) {
            Write-Err "No checksum published for $assetName; refusing to install an unverified binary."
        }
        $actual = (Get-FileHash -Path $zipPath -Algorithm SHA256).Hash
        if ($actual -ine $expected) {
            Write-Err "Checksum mismatch for $assetName (expected $expected, got $actual)."
        }
        Write-Info "Checksum verified."

        # Extract
        Expand-Archive -Path $zipPath -DestinationPath $tempDir -Force

        # Find and move binary
        $exePath = Join-Path $tempDir "$BinaryName.exe"
        if (-not (Test-Path $exePath)) {
            Write-Err "Binary not found in archive"
        }

        $destPath = Join-Path $installDir "$BinaryName.exe"
        Move-Item -Path $exePath -Destination $destPath -Force

        $guiSource = @(
            (Join-Path $tempDir "$GuiBinaryName.exe"),
            (Join-Path $tempDir "Skell.exe"),
            (Join-Path $tempDir "Skell-windows-amd64.exe")
        ) | Where-Object { Test-Path $_ } | Select-Object -First 1

        if ($guiSource) {
            $guiDestPath = Join-Path $installDir "$GuiBinaryName.exe"
            Move-Item -Path $guiSource -Destination $guiDestPath -Force
        } elseif ($downloadedBundle) {
            Write-Warn "GUI bundle downloaded, but GUI executable was not found in the archive."
        }

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
        if (Test-Path (Join-Path $installDir "$GuiBinaryName.exe")) {
            Write-Info "  skell gui"
        }

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
    $guiExePath = Join-Path $installDir "$GuiBinaryName.exe"
    $legacyGuiPath = Join-Path $installDir "Skell-windows-amd64.exe"

    if (Test-Path $exePath) {
        Remove-Item -Path $exePath -Force
        Write-Info "Removed $exePath"
    } else {
        Write-Warn "skell binary not found at $exePath"
    }

    if (Test-Path $guiExePath) {
        Remove-Item -Path $guiExePath -Force
        Write-Info "Removed $guiExePath"
    }

    if (Test-Path $legacyGuiPath) {
        Remove-Item -Path $legacyGuiPath -Force
        Write-Info "Removed $legacyGuiPath"
    }

    # Remove the install directory if it is now empty
    if ((Test-Path $installDir) -and (Get-ChildItem $installDir -Force | Measure-Object).Count -eq 0) {
        Remove-Item -Path $installDir -Force
    }

    # Remove from user PATH
    $currentPath = [Environment]::GetEnvironmentVariable("Path", "User")
    $entries = @($currentPath -split ";" | Where-Object { $_ })
    $kept = @($entries | Where-Object { $_.TrimEnd("\") -ine $installDir.TrimEnd("\") })
    if ($kept.Count -ne $entries.Count) {
        [Environment]::SetEnvironmentVariable("Path", ($kept -join ";"), "User")
        Write-Info "Removed $installDir from PATH"
    }

    Write-Host ""
    Write-Info "skell has been uninstalled."
}

if ($Uninstall) { Uninstall-Skell } else { Install-Skell }

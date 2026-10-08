# Install Skell

## Quick install

**macOS / Linux**

```sh
curl -fsSL https://raw.githubusercontent.com/aminmesbahi/skell/main/install.sh | sh
```

Installs to `/usr/local/bin` (asks for `sudo` if needed). Choose another folder
with `INSTALL_DIR=~/.local/bin`, or a specific release with `SKELL_VERSION=v0.2.0`.

**Windows (PowerShell)**

```powershell
irm https://raw.githubusercontent.com/aminmesbahi/skell/main/install.ps1 | iex
```

Installs `skell.exe` and the desktop app (`skell-gui.exe`) into
`%LOCALAPPDATA%\Programs\skell` and adds it to your user `PATH`.

Both installers download from GitHub Releases and **verify the SHA-256 checksum**
before installing; they refuse to install anything they can't verify.

## Package managers

| Manager | Command |
|---|---|
| Homebrew | `brew install --cask aminmesbahi/tap/skell` |
| Scoop | `scoop bucket add skell https://github.com/aminmesbahi/scoop-bucket` then `scoop install skell` |
| Go | `go install github.com/aminmesbahi/skell@latest` |

The Homebrew tap and Scoop bucket are published by the release pipeline once the
`HOMEBREW_TAP_GITHUB_TOKEN` / `SCOOP_BUCKET_GITHUB_TOKEN` secrets are configured.

## Manual download

Grab an archive from [GitHub Releases](https://github.com/aminmesbahi/skell/releases)
and put the `skell` binary on your `PATH`:

| Platform | File |
|---|---|
| Windows x64 / ARM64 (CLI + desktop app) | `skell_<ver>_windows_<arch>_bundle.zip` |
| Windows x64 / ARM64 (CLI only) | `skell_<ver>_windows_<arch>.zip` |
| macOS Apple Silicon / Intel | `skell_<ver>_darwin_<arch>.tar.gz` |
| Linux x64 / ARM64 | `skell_<ver>_linux_<arch>.tar.gz` |

Verify against `checksums.txt` (CLI archives) or `<bundle>.sha256` (desktop
bundles) from the same release.

## Shell completion

```sh
skell completion bash > /etc/bash_completion.d/skell        # bash
skell completion zsh  > "${fpath[1]}/_skell"                # zsh
skell completion fish > ~/.config/fish/completions/skell.fish
skell completion powershell | Out-String | Invoke-Expression  # PowerShell (add to $PROFILE)
```

Completion knows your installed skills, the skills your sources offer, agent
names and catalog ids.

## Updating

```sh
skell selfupdate          # verifies the release checksum before replacing itself
skell selfupdate --check  # only report whether an update exists
```

## Uninstall

```sh
# macOS / Linux (use the same INSTALL_DIR you installed with)
curl -fsSL https://raw.githubusercontent.com/aminmesbahi/skell/main/install.sh | sh -s -- uninstall
```

```powershell
# Windows
& ([scriptblock]::Create((irm https://raw.githubusercontent.com/aminmesbahi/skell/main/install.ps1))) -Uninstall
```

Skell's own data (source cache, catalog cache, audit log, global config) lives in
`~/.skell` (or `$SKELL_HOME`); delete it to remove everything.

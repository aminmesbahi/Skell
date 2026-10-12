# Install Skell

## Quick install

**macOS / Linux**

```sh
curl -fsSL https://raw.githubusercontent.com/aminmesbahi/skell/main/install.sh | sh
```

Installs to `/usr/local/bin` (asks for `sudo` if needed). Set the variables on
`sh`, not on `curl`, to choose another folder (created if it doesn't exist) or a
specific release:

```sh
curl -fsSL https://raw.githubusercontent.com/aminmesbahi/skell/main/install.sh \
  | INSTALL_DIR=~/.local/bin SKELL_VERSION=v0.13.0 sh
```

**Windows (PowerShell)**

```powershell
irm https://raw.githubusercontent.com/aminmesbahi/skell/main/install.ps1 | iex
```

Installs `skell.exe` and the desktop app (`skell-gui.exe`) into
`%LOCALAPPDATA%\Programs\skell` and adds it to your user `PATH`. For the
command-line tool only, set `$env:SKELL_NO_GUI = 1` first (the GitHub Action does
this). `SKELL_VERSION` works here too.

Both installers download from GitHub Releases and **verify the SHA-256 checksum**
before installing; they refuse to install anything they can't verify.

## Package managers

| Manager | Command |
|---|---|
| Homebrew (macOS) | `brew install --cask aminmesbahi/tap/skell` |
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

The desktop app is a separate download on macOS and Linux (`skell gui` starts
it when it sits next to the `skell` binary):

| Platform | File |
|---|---|
| macOS Apple Silicon | `Skell_<ver>_darwin_arm64.dmg` (or `.zip`) |
| Linux x64 | `Skell_<ver>_linux_amd64.tar.gz` (needs GTK 4 and WebKitGTK 6 installed) |

Verify against `checksums.txt` (CLI archives) or the matching `.sha256` file
(desktop bundles and the Linux app) from the same release.

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

Installed with `go install`, Skell reports the module version (and VCS commit) the
Go toolchain embeds, so `skell --version` is meaningful there too.
`skell selfupdate` replaces the binary in place; if you installed with Homebrew,
Scoop or `go install`, update with that tool instead.

Skell's own data (source cache, catalog cache, audit log, global config) lives in
`~/.skell` (or `$SKELL_HOME`); delete it to remove everything.

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

Every file on the [releases page](https://github.com/aminmesbahi/skell/releases)
is named `skell_<what>_<version>_<os>_<arch>`:

| Prefix | What it is |
|---|---|
| `skell_bundle_` | **Command line + desktop app together** — the easiest choice on Windows and macOS |
| `skell_gui_` | The desktop app only |
| `skell_cli_` | The command-line tool only |

| Your system | Bundle | Desktop app only | Command line only |
|---|---|---|---|
| Windows x64 / ARM64 | `skell_bundle_<ver>_windows_<arch>.zip` | `skell_gui_<ver>_windows_<arch>_installer.exe` | `skell_cli_<ver>_windows_<arch>.zip` |
| macOS Apple Silicon | `skell_bundle_<ver>_darwin_arm64.zip` | `skell_gui_<ver>_darwin_arm64.dmg` (or `.zip`) | `skell_cli_<ver>_darwin_arm64.tar.gz` |
| macOS Intel | | | `skell_cli_<ver>_darwin_amd64.tar.gz` |
| Linux x64 | | `skell_gui_<ver>_linux_amd64.tar.gz` (needs GTK 4 and WebKitGTK 6) | `skell_cli_<ver>_linux_amd64.tar.gz` |
| Linux ARM64 | | | `skell_cli_<ver>_linux_arm64.tar.gz` |

Unpack it and put the `skell` binary on your `PATH`. `skell gui` starts the
desktop app when it sits next to the `skell` binary or where its installer put
it; if you do not have it, `skell gui` offers to download the latest version from
GitHub (checked against its SHA-256, installed for your user only) and open it.
Use `skell gui --install` to do that without being asked. Releases up to v0.15 used
the older names `skell_<ver>_<os>_<arch>` and `Skell_<ver>_…`.

**Double-clicking on Windows.** `skell.exe` is a command-line tool, but opening it
from Explorer does something useful: with the desktop app (`skell-gui.exe`) next
to it, it opens the app; without it, it offers to open a command prompt in that
folder with `skell` ready to use, or to open this page. For everyday use run the
PowerShell installer above, which installs both and puts `skell` on your `PATH`.

Verify against `checksums.txt` (`skell_cli_` archives) or the matching `.sha256` file
(bundles, installers and the Linux app) from the same release.

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

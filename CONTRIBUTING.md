# Contributing to Skell

Thanks for helping! Issues and pull requests are welcome.

## Set up

- **Go** — the version in `go.mod` (the desktop app's Go module also needs a
  **64-bit** toolchain; a 32-bit Go on Windows cannot run its tests).
- **Desktop app (optional)** — [Wails v3](https://v3.wails.io) CLI and
  [bun](https://bun.sh).

```sh
git clone https://github.com/aminmesbahi/skell && cd skell
go build -o skell .         # CLI
go test ./...               # CLI tests
cd gui && go test ./...     # desktop app backend
cd gui/frontend && bun install --frozen-lockfile && bunx tsc --noEmit && bunx vitest run
```

Line endings are normalised by `.gitattributes` (LF), so Windows checkouts won't
fail `gofmt`.

## Before you open a pull request

- `gofmt -l .` and `go vet ./...` are clean; `golangci-lint run` passes (CI runs it).
- You added or updated tests for the change.
- If you changed a command or a flag, regenerate the reference — CI fails when it
  is stale:

  ```sh
  go run . gen-docs docs/reference
  ```

- Behaviour users can see is described in `README.md` / `docs/`.
- Commit messages follow [Conventional Commits](https://www.conventionalcommits.org)
  (`fix(sync): …`, `feat(gui): …`); the release notes are generated from them.

## Security

Please report vulnerabilities privately — see [SECURITY.md](SECURITY.md).

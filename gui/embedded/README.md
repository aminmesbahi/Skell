# Embedded skell CLI

The desktop app ships the `skell` command-line tool it needs, so it works even
when `skell` is not installed. The build (`wails3 build`, task `build:cli`)
puts the CLI for the target system here as `skell` / `skell.exe` before the app
is compiled; it is not committed. This file only keeps the folder in the
repository so `go vet` and the tests work without a build.

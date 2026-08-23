# Contributing

## Prerequisites

- Go 1.24+
- GNU Make (optional, for convenience targets)

## Getting started

```sh
git clone https://github.com/smeltery/awsw.git
cd awsw
make build
```

## Project layout

```
cmd/           Cobra command definitions
internal/
  config/      Profile registry, AWS config generation
  exec/        CommandExecutor interface (abstraction over os/exec)
  shell/       Shell-specific output (env export, init scripts)
```

See [architecture.md](architecture.md) for design details.

## Running tests

```sh
make test
```

This runs all tests with the race detector enabled. To see coverage:

```sh
make coverage
```

### Coverage requirements

The CI pipeline enforces **≥ 80% coverage** per package (excluding `internal/exec` and `main.go`, which are thin wrappers around `os/exec`).

## Making changes

1. Create a branch from `main`.
2. Make your changes.
3. Run `make test` and `make lint` (if available) locally.
4. Open a pull request against `main`.

### Adding a new command

1. Create `cmd/<name>.go` with a new `cobra.Command`.
2. Register it in `cmd/root.go` via `rootCmd.AddCommand()`.
3. Add tests in `cmd/cmd_test.go`.

### Adding a new profile

For end users: edit `~/.config/awsw/config.yaml` and add an entry under `profiles:`.

To change the built-in defaults: edit `DefaultConfig()` in `internal/config/loader.go` and add corresponding test cases in `internal/config/loader_test.go`.

### Changing shell output

Shell-specific logic lives in `internal/shell/shell.go`. If you add a new shell, update:
- `ParseShell()` — add the new shell type
- `ExportEnv()` — add the env var syntax
- `InitScript()` — add the wrapper function template

## Releases

Releases are automated via GitHub Actions and GoReleaser. To create a release:

1. Tag a commit on `main`: `git tag v1.x.x`
2. Push the tag: `git push origin v1.x.x`
3. The release workflow builds binaries for macOS (arm64/amd64) and Linux, and publishes a Homebrew tap.

## Code style

- Follow standard Go conventions (`gofmt`, `go vet`).
- Use table-driven tests.
- All external command execution must go through the `CommandExecutor` interface — never call `os/exec` directly from command handlers.

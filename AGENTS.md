# Repository Guidelines

## Project Structure & Module Organization

This repository is a Go module (`github.com/asjard/asjard`) targeting Go 1.25.6. Framework internals live in `core/` (bootstrap, configuration, clients, registries, servers, tracing), while reusable integrations and public adapters live in `pkg/`. General-purpose helpers are under `utils/`. Command-line tools and protobuf generators are in `cmd/`; runnable reference projects and their protobuf sources are in `_examples/`. Documentation belongs in `docs/`, and shared/generated protobuf packages are under `pkg/protobuf/`. Treat files marked `Code generated ... DO NOT EDIT` as generated artifacts; change their source or generator instead.

## Build, Test, and Development Commands

- `go test ./...` runs the full Go test suite during local iteration.
- `make test` runs the repository's CI-style checks: cleanup, cyclomatic-complexity validation, `go vet`, race-enabled tests with coverage, and benchmarks. It creates `cover.out`.
- `go tool cover -html=cover.out` opens the generated coverage report.
- `make build_gen_go_rest` builds one generator into `$GOPATH/bin`; equivalent `build_gen_*` targets cover the other generators.
- `make gen_proto` regenerates protobuf outputs through the vendored build script.
- `make github_workflows_dependices` starts the Docker services used by integration tests. Docker Compose is required.

The full CI target, `make github_workflows_test`, updates submodules and may reset their contents; use it only when that behavior is intended.

## Coding Style & Naming Conventions

Format Go changes with `gofmt` (tabs, standard import grouping) and keep `go vet -all ./...` clean. Use short, lowercase package names; exported identifiers use `CamelCase`, unexported identifiers use `camelCase`, and filenames use lowercase words separated by underscores when needed. Keep package-level APIs documented. Prefer changes within the existing `core/<area>` or `pkg/<integration>` boundary instead of creating broad utility packages.

## Testing Guidelines

Place tests beside implementation files as `*_test.go`, with functions named `TestXxx`; use `t.Run` for meaningful cases. Add regression tests for fixes and table-driven tests where inputs vary. Some store and registry tests require the Compose services. No fixed coverage threshold is configured, but changed behavior should be covered and `make test` should pass.

## Commit & Pull Request Guidelines

The visible history uses concise, imperative, lowercase subjects (for example, `add version in log`). Keep each commit focused and explain non-obvious reasoning in the body. Pull requests should summarize behavior changes, identify affected packages, link relevant issues, and list verification commands. Include screenshots only for documentation or UI output changes, and call out regenerated code, configuration changes, or new service dependencies explicitly.

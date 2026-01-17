# AGENTS

## Purpose

- This file is the quick-start guide for coding agents working in this repo.
- Keep it updated when workflows or invariants change.

## Project overview

Cascade is a CLI tool designed to apply changes across multiple git repositories efficiently. It automates the process of fetching the latest changes, creating branches, applying patches, executing scripts, or running commands, and generating pull requests. View `README.md` for details about the project.

## Architecture

- Cascade is a cobra-based CLI with a thin command layer and internal packages for git, validation, and logging.
- The `apply` command is the primary workflow; it orchestrates per-repo operations and reports success/failure.
- We favor small internal packages with clear responsibilities over shared cross-cutting abstractions.

## Repository map

- CLI semantics and flags: `cmd/`.
- Git behavior and side effects: `internal/git/`.
- Input validation rules: `internal/validation/`.
- Error log format and storage: `internal/log/`.
- End-to-end expectations: `tests/` and `README.md`.

## Platform support

- We explicitly support Linux and macOS (Darwin) runtimes.
- We do not explicitly support Windows runtimes. Cascade can be used in WSL, which keeps maintenance and development simpler.

## Testing expectations

- Cover behavior changes with unit tests; use existing tests as examples.
- If a unit test is not practical, add or extend an integration test in `tests/`.
- Integration tests require `git` in `PATH` and skip with `go test -short`.

## Coding standards

- Write concise, idiomatic Go; prefer clear standard library solutions over cleverness.
- Keep functions small and focused; avoid unnecessary abstractions.
- Run `gofmt` on Go files.

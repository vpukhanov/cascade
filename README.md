# Cascade

Cascade is a CLI tool designed to apply changes across multiple git repositories efficiently. It automates the process of fetching the latest changes, creating branches, applying patches or executing scripts, and generating pull requests.

> [!WARNING]
> Cascade is a work in progress; some features may not function as intended. To prevent data loss, only run the program on repositories without unpushed important changes.

## Installation

If you have Homebrew installed on macOS or Linux, you can install Cascade using:

```sh
brew install vpukhanov/tools/cascade
```

If you have Go installed on your system, you can install Cascade directly using the `go install` command:

```sh
go install github.com/vpukhanov/cascade@latest
```

Alternatively, you can download the binary from the [Releases page](https://github.com/vpukhanov/cascade/releases) of the repository:

1. Download the archive for your operating system and architecture.
2. Extract the archive:
   - On macOS: Double-click the .zip file or use `unzip cascade_*_Darwin_*.zip`
   - On Linux: `tar -xzf cascade_*_Linux_*.tar.gz`
   - On Windows: Extract the .zip file using File Explorer or a tool like 7-Zip
3. Move the `cascade` binary to a directory in your system's `PATH`.

## Usage

To apply changes across repositories:

```bash
cascade apply \
  --patch ./changes.patch \    # Path to patch file (or --script)
  --branch update-logging \    # New branch name
  --message "Update logging" \ # Commit message
  ./repo1 ./repo2              # Repository paths

# Alternative using a script
cascade apply \
  --script ./update.sh \
  --branch refactor-components \
  --message "Refactor components" \
  ./repo1 ./repo2

# Apply changes to a specific base branch and update it first
cascade apply \
  --patch ./changes.patch \
  --branch feature/update \
  --message "Update dependencies" \
  --base-branch main \         # Branch to apply changes to
  --pull \                     # Pull latest changes first
  ./repo1 ./repo2
```

Required parameters:

- Repository paths - One or more paths to git repositories to modify (as positional arguments)
- `--patch` or `--script` - Path to patch file or executable script
- `--branch` - Name for the new branch that will be created
- `--message` - Commit message used for the changes

Optional parameters:

- `--base-branch` - Branch to check out and apply changes to (default: current branch)
- `--pull` - Pull latest changes from remote before applying changes (default: false)

To see available commands:

```bash
cascade --help
```

To check the version:

```bash
cascade --version
```

## Development

To run the tests:

```bash
go test ./...
```

# Cascade

Cascade is a CLI tool that lets you apply changes to your codebase across multiple git repositories. It helps automate the process of making similar changes across multiple repositories by handling the fetching of latest changes, creating branches, applying patches or running scripts, and creating pull requests.

## Installation

```bash
go install cascade
```

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

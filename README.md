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
  --repos ./repo1,./repo2 \    # Comma-separated repository paths
  --branch update-logging \    # New branch name
  --message "Update logging"   # Commit message

# Alternative using a script
cascade apply \
  --script ./update.sh \
  --repos ./repo1,./repo2 \
  --branch refactor-components \
  --message "Refactor components"
```

Required flags:

- `--patch` or `--script` - Path to patch file or executable script
- `--repos` - Comma-separated list of repository paths to modify
- `--branch` - Name for the new branch that will be created
- `--message` - Commit message used for the changes

To see available commands:

```bash
cascade --help
```

To check the version:

```bash
cascade --version
```

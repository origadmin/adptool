# Issue Sync Script Documentation

This directory contains automation scripts for synchronizing local `.issues` files with remote GitHub Issues.

## Dependencies

Before running the scripts, you must install and configure the following command-line tools:

1.  **GitHub CLI (`gh`)**: The core tool for interacting with your GitHub repository.
    -   **Installation**: Please refer to the official documentation for installation instructions: <https://cli.github.com/>
    -   **Authentication**: After installation, run `gh auth login` and follow the prompts to authenticate, ensuring `gh` has permission to access your repository.

2.  **`yq`**: A lightweight command-line YAML/JSON/XML processor. We use it to safely and accurately read and write metadata (Front Matter) in the issue file headers.
    -   **Installation**: Please refer to the installation instructions on its official GitHub repository: <https://github.com/mikefarah/yq/>

## Scripts

### `sync_issues.sh`

This script is the core of the automation process. It intelligently pushes local changes to GitHub.

#### Features:

1.  **Create Remote Issues**: It iterates through the `.issues` directory. If it finds a local issue file without a `remote_id`, it creates a corresponding issue on GitHub and writes the returned issue number back to the `remote_id` field of the local file.
2.  **Close Remote Issues**: If the script finds that a local issue's status has been updated to `resolved` or `wont-fix` and it has an associated `remote_id`, the script will automatically close the issue on GitHub.

#### Usage:

```bash
# Run from the project root directory
bash scripts/sync_issues.sh
```

It is recommended to run this script after you have completed a series of local tasks and are ready to `git push` your code to the remote repository, ensuring that the remote and local states are synchronized.

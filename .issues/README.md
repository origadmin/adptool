# Local-First Issue Management Workflow

This document describes an issue management process designed for solo developers who prioritize efficiency. It centers around local files and Git, ensuring you can stay within your IDE and terminal for the vast majority of the time, while remaining fully compatible with future team collaboration and GitHub Issues integration.

## Core Concepts

- **Every Issue is a File**: Each issue (be it a bug, feature request, or documentation task) is managed as a separate Markdown file, ensuring clarity and independence.
- **Git-Driven**: The creation, status changes, and closing of all issues are tracked through Git commits, providing a clear and traceable history.
- **Local-First**: The entire process is designed to maximize local operational efficiency and avoid frequent context switching to a browser.
- **Future-Proof**: The local workflow can be synchronized with GitHub Issues at any time via a script, allowing for seamless expansion to a team-based collaboration model.

## The Workflow

### 1. Creating an Issue

When you find a new problem or have a new idea, copy the template file in the `.issues` directory and name it using the format `ID-short-description.md`.

```bash
# Example: Create a bug issue for the "user login module"
cp .issues/.template.md .issues/5-user-login-module-bug.md
```

Next, open this new file in your IDE, modify the metadata at the top (like `status`, `labels`, etc.), and fill in the detailed description.

### 2. Starting Work on an Issue

Create a dedicated Git branch for the issue. The branch name should correspond to the issue file to allow for easy tracking.

```bash
# Recommended branch name format: issue/ID
git checkout -b issue/5
```

Now you can make all related code changes on this branch.

### 3. Testing and Verification

Before declaring an issue resolved, verification is mandatory. This is a critical step for ensuring software quality.

1.  **Write Tests**: For bug fixes, one or more unit tests should be written to reproduce the bug and then prove that your fix is effective.
2.  **Perform Verification**: In the corresponding issue file, fill out the `### Verification Steps` section, detailing how to confirm the fix is effective. This serves not only as a record for yourself but also as a clear guide for future collaborators.

### 4. Completing and Closing an Issue

After the code has been verified, follow these steps to wrap up:

1.  **Update the Issue File**: Open the corresponding issue file and change the `status` field to `resolved`.
2.  **Commit All Changes**: Commit your code changes, test files, and the issue file status change together.

    ```bash
    # Add all code, test, and issue file changes
    git add .

    # Write a clear commit message
    git commit -m "fix: resolve bug in user login module\n\nResolve issue #5"
    ```

3.  **Merge the Branch**: Merge the feature branch back into the main branch.

    ```bash
    git checkout main
    git merge issue/5
    ```

4.  **(Optional) Clean Up**: Delete the merged issue branch.

    ```bash
    git branch -d issue/5
    ```

### (Optional) Syncing with GitHub Issues

When you want to push your local work records to GitHub or collaborate with others, you can use an automation script to synchronize them. For detailed instructions, please refer to the `README.md` file in the project's `scripts/` directory.

# Git-Commit Driven Issue Workflow (Final)

This document describes a robust, local-first issue management workflow. It leverages Git and a smart script to provide a highly efficient, fault-tolerant development process.

**The Guiding Principle**: GitHub is the single source of truth for an issue's *state* (open/closed). Local files are for content creation and context. We use specific keywords in our Git commit messages to link our work directly to either a remote GitHub Issue or a local issue file.

## The Workflow

### Step 1: Create a Local Issue File

As always, start by creating a local issue file from the template. This is your single point of entry.

```bash
cp .issues/.template.md .issues/10-new-critical-bug.md
```

### Step 2: Choose Your Path

You have two options, depending on whether you are online or offline.

**Path A: The Online Workflow (Recommended)**

1.  **Initial Sync**: Run the script immediately to create the issue on GitHub.
    ```bash
    bash scripts/sync_issues.sh
    ```
    The script will create the issue and populate the `remote_id` (e.g., `127`).
2.  **Branch and Fix**: Create your branch (`git checkout -b issue/127`) and do your work.
3.  **Commit with GitHub Keyword**: Commit your changes using GitHub's native closing keyword.
    ```bash
    git commit -m "fix: Resolve critical bug\n\nCloses #127"
    ```

**Path B: The Offline Workflow (The Safety Net)**

1.  **Branch and Fix**: You are offline or prefer not to sync yet. Create a branch (`git checkout -b issue/10-critical-bug`) and do your work.
2.  **Commit with Local Keyword**: When you commit, use the special `Resolves-Local` keyword, referencing the **exact filename**.
    ```bash
    git commit -m "fix: Resolve critical bug\n\nResolves-Local: 10-new-critical-bug.md"
    ```

### Step 3: Merge and Push

This step is the same for both paths.

```bash
git checkout main
git merge <your-branch-name>
git push origin main
```

### Step 4: The Final Sync

At any point after pushing, run the sync script.

```bash
bash scripts/sync_issues.sh
```

**What the script does:**
- If you followed **Path A**, it does nothing, because your issue file already has a `remote_id`.
- If you followed **Path B**, the script will:
  1. Discover your local file has no `remote_id`.
  2. Create a new issue on GitHub (e.g., #128).
  3. Update your local file with `remote_id: 128`.
  4. **Intelligently search** your `main` branch history for a commit containing `Resolves-Local: 10-new-critical-bug.md`.
  5. Find the commit and **automatically close** the newly created GitHub Issue #128, leaving a comment linking to the commit that fixed it.

This workflow provides maximum flexibility. It rewards following the standard online process while providing a powerful, automated safety net for offline work, ensuring that the remote state on GitHub always eventually matches the reality of your codebase.

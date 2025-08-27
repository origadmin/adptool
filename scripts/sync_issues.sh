#!/bin/bash

# A script to synchronize local markdown issues with a remote GitHub repository.
# It creates new issues and can retroactively close issues if their closing
# commit was merged before the issue was created on GitHub.

set -e # Exit immediately if a command exits with a non-zero status.

# --- Configuration ---
ISSUES_DIR=".issues"
# Set your main branch name here. It's used to check if a fix has been merged.
MAIN_BRANCH="main"
REPO=$(gh repo view --json nameWithOwner -q .nameWithOwner 2>/dev/null || echo "")

# --- Pre-flight Checks ---
if ! command -v gh &> /dev/null; then
    echo "Error: GitHub CLI 'gh' is not installed. Please install it from https://cli.github.com/"
    exit 1
fi
if ! command -v yq &> /dev/null; then
    echo "Error: 'yq' is not installed. Please install it from https://github.com/mikefarah/yq/"
    exit 1
fi
if [ -z "$REPO" ]; then
    echo "Error: Could not determine GitHub repository. Are you in a git repository with a remote?"
    exit 1
fi

echo "Syncing local issues with repository: $REPO"

# --- Main Loop ---
find "$ISSUES_DIR" -name "*.md" -not -name "README.md" -not -name ".template.md" | while read -r file; do
    REMOTE_ID=$(yq -N '.remote_id' "$file")

    # Process only files that have not yet been synced.
    if [ -z "$REMOTE_ID" ] || [ "$REMOTE_ID" == "null" ]; then
        echo "---"
        echo "Processing local-only file: $file"

        FILENAME=$(basename "$file" .md)
        TITLE="Issue: $FILENAME"
        LABELS=$(yq '.labels | join(",")' "$file")
        BODY=$(yq 'select(document_index == 1)' "$file")

        echo "Creating new GitHub issue..."
        NEW_ISSUE_URL=$(gh issue create --repo "$REPO" --title "$TITLE" --body "$BODY" --label "$LABELS")
        ISSUE_NUMBER=$(echo "$NEW_ISSUE_URL" | awk -F'/' '{print $NF}')

        if [ -z "$ISSUE_NUMBER" ]; then
            echo "Error: Failed to create GitHub issue for $file"
            continue
        fi

        echo "Successfully created GitHub Issue #$ISSUE_NUMBER"
        yq -i ".remote_id = $ISSUE_NUMBER" "$file"
        echo "Updated $file with remote_id: $ISSUE_NUMBER"

        # --- AUTOMATIC RETROACTIVE CLOSING LOGIC ---
        echo "Checking if a closing commit for this file already exists on the main branch..."
        
        # Define the keyword we are looking for in commit messages.
        # It must reference the exact filename.
        SEARCH_KEYWORD="Resolves-Local: $FILENAME.md"

        # Search for a commit on the main branch containing the keyword.
        # The `|| true` prevents the script from exiting if no commit is found.
        COMMIT_HASH=$(git log "$MAIN_BRANCH" --grep="$SEARCH_KEYWORD" -n 1 --format=%H || true)
        
        if [ -n "$COMMIT_HASH" ]; then
            echo "Found closing commit $COMMIT_HASH on the main branch."
            echo "Automatically closing newly created GitHub Issue #$ISSUE_NUMBER."
            gh issue close "$ISSUE_NUMBER" --repo "$REPO" --comment "Automatically closed by sync script. Resolved by commit $COMMIT_HASH."
        else
            echo "No existing closing commit found. The new issue will remain open."
        fi
    fi
done

echo "---"
echo "Issue synchronization complete."

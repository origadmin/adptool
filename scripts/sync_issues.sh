#!/bin/bash

# A script to synchronize local markdown issues with a remote GitHub repository.
#
# This script performs the following actions:
# 1. Creates a new GitHub issue for any local .md file in the .issues/ directory
#    that does not yet have a `remote_id`.
# 2. Closes a GitHub issue if the corresponding local .md file's status is
#    'resolved' or 'wont-fix'.

set -e # Exit immediately if a command exits with a non-zero status.

# --- Configuration ---
# Directory where local issues are stored.
ISSUES_DIR=".issues"
# The GitHub repository in `owner/repo` format.
# The script attempts to automatically determine this from your git remote.
# If it fails, you can manually set it here, e.g., REPO="your-user/your-repo"
REPO=$(gh repo view --json nameWithOwner -q .nameWithOwner 2>/dev/null || echo "")

# --- Pre-flight Checks ---
# Check if required tools are installed.
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
    echo "You can set the REPO variable manually in the script."
    exit 1
fi

echo "Syncing issues with repository: $REPO"
echo "Looking for issues in: $ISSUES_DIR/"

# --- Main Loop ---
# Loop through all markdown files in the issues directory.
find "$ISSUES_DIR" -name "*.md" -not -name "README.md" -not -name ".template.md" | while read -r file; do
    echo "---"
    echo "Processing file: $file"

    # Use yq to read metadata. Note the '-N' flag for numeric remote_id.
    REMOTE_ID=$(yq -N '.remote_id' "$file")
    STATUS=$(yq '.status' "$file")
    LABELS=$(yq '.labels | join(",")' "$file")

    # --- Sync Logic ---
    if [ -z "$REMOTE_ID" ] || [ "$REMOTE_ID" == "null" ]; then
        # Case 1: No remote_id. Create a new issue on GitHub.
        echo "Status: Local issue, no remote_id found. Creating new GitHub issue..."

        # Extract title from filename (e.g., '5-user-login-bug.md' -> 'Issue 5: user login bug')
        FILENAME=$(basename "$file" .md)
        TITLE="Issue #$FILENAME"

        # Extract body from the file, excluding the YAML front matter.
        BODY=$(yq 'select(document_index == 1)' "$file")

        # Create the issue and capture the URL of the new issue.
        NEW_ISSUE_URL=$(gh issue create --repo "$REPO" --title "$TITLE" --body "$BODY" --label "$LABELS")

        # Extract the issue number from the URL.
        ISSUE_NUMBER=$(echo "$NEW_ISSUE_URL" | awk -F'/' '{print $NF}')

        if [ -z "$ISSUE_NUMBER" ]; then
            echo "Error: Failed to create GitHub issue for $file"
            continue
        fi

        echo "Successfully created GitHub Issue #$ISSUE_NUMBER"

        # Write the new issue number back to the local file.
        yq -i ".remote_id = $ISSUE_NUMBER" "$file"
        echo "Updated $file with remote_id: $ISSUE_NUMBER"

    else
        # Case 2: remote_id exists. Check if the issue needs to be closed.
        echo "Status: Found remote_id #$REMOTE_ID. Checking status..."

        if [ "$STATUS" == "resolved" ] || [ "$STATUS" == "wont-fix" ]; then
            # Check the state of the remote issue before trying to close it.
            REMOTE_STATE=$(gh issue view "$REMOTE_ID" --repo "$REPO" --json state -q .state)

            if [ "$REMOTE_STATE" == "OPEN" ]; then
                echo "Local status is '$STATUS', closing GitHub Issue #$REMOTE_ID..."
                gh issue close "$REMOTE_ID" --repo "$REPO" --comment "Closing issue as per local file status: '$STATUS'."
                echo "Successfully closed GitHub Issue #$REMOTE_ID."
            else
                echo "GitHub Issue #$REMOTE_ID is already closed. No action needed."
            fi
        else
            echo "Local status is '$STATUS'. No action needed."
        fi
    fi
done

echo "---"
echo "Issue synchronization complete."

#!/bin/bash

# A script to synchronize local markdown issues with a remote GitHub repository.
# 一个用于将本地 Markdown issue 与远程 GitHub 仓库同步的脚本。
#
# This script performs the following actions:
# 该脚本执行以下操作：
# 1. Creates a new GitHub issue for any local .md file in the .issues/ directory
#    that does not yet have a `remote_id`.
#    为 .issues/ 目录中任何尚无 `remote_id` 的本地 .md 文件创建一个新的 GitHub issue。
# 2. Closes a GitHub issue if the corresponding local .md file's status is
#    'resolved' or 'wont-fix'.
#    如果本地 .md 文件对应的状态是 'resolved' 或 'wont-fix'，则关闭 GitHub 上对应的 issue。

set -e # Exit immediately if a command exits with a non-zero status. / 如果任何命令以非零状态退出，则立即退出。

# --- Configuration / 配置 ---
# Directory where local issues are stored.
# 存储本地 issue 的目录。
ISSUES_DIR=".issues"
# The GitHub repository in `owner/repo` format.
# The script attempts to automatically determine this from your git remote.
# If it fails, you can manually set it here, e.g., REPO="your-user/your-repo"
# GitHub 仓库，格式为 `owner/repo`。
# 脚本会尝试从你的 git remote 自动确定。如果失败，你可以在此处手动设置。
REPO=$(gh repo view --json nameWithOwner -q .nameWithOwner 2>/dev/null || echo "")

# --- Pre-flight Checks / 执行前检查 ---
# Check if required tools are installed.
# 检查所需工具是否已安装。
if ! command -v gh &> /dev/null; then
    echo "Error: GitHub CLI 'gh' is not installed. Please install it from https://cli.github.com/"
    echo "错误: 未安装 GitHub CLI 'gh'。请从 https://cli.github.com/ 安装。"
    exit 1
fi
if ! command -v yq &> /dev/null; then
    echo "Error: 'yq' is not installed. Please install it from https://github.com/mikefarah/yq/"
    echo "错误: 未安装 'yq'。请从 https://github.com/mikefarah/yq/ 安装。"
    exit 1
fi
if [ -z "$REPO" ]; then
    echo "Error: Could not determine GitHub repository. Are you in a git repository with a remote?"
    echo "You can set the REPO variable manually in the script."
    echo "错误: 无法确定 GitHub 仓库。你是否在一个设置了远程仓库的 git 项目中？"
    echo "你可以在脚本中手动设置 REPO 变量。"
    exit 1
fi

echo "Syncing issues with repository: $REPO"
echo "Looking for issues in: $ISSUES_DIR/"

# --- Main Loop / 主循环 ---
# Loop through all markdown files in the issues directory.
# 遍历 issues 目录下的所有 markdown 文件。
find "$ISSUES_DIR" -name "*.md" -not -name "README.md" -not -name ".template.md" | while read -r file; do
    echo "---"
    echo "Processing file: $file"

    # Use yq to read metadata. Note the '-N' flag for numeric remote_id.
    # 使用 yq 读取元数据。注意 '-N' 标志用于处理数字类型的 remote_id。
    REMOTE_ID=$(yq -N '.remote_id' "$file")
    STATUS=$(yq '.status' "$file")
    LABELS=$(yq '.labels | join(",")' "$file")

    # --- Sync Logic / 同步逻辑 ---
    if [ -z "$REMOTE_ID" ] || [ "$REMOTE_ID" == "null" ]; then
        # Case 1: No remote_id. Create a new issue on GitHub.
        # 场景 1: 没有 remote_id。在 GitHub 上创建一个新的 issue。
        echo "Status: Local issue, no remote_id found. Creating new GitHub issue..."

        # Extract title from filename (e.g., '5-user-login-bug.md' -> 'Issue 5: user login bug')
        # 从文件名中提取标题。
        FILENAME=$(basename "$file" .md)
        TITLE="Issue #$FILENAME"

        # Extract body from the file, excluding the YAML front matter.
        # 从文件中提取正文，排除 YAML front matter。
        BODY=$(yq 'select(document_index == 1)' "$file")

        # Create the issue and capture the URL of the new issue.
        # 创建 issue 并捕获新 issue 的 URL。
        NEW_ISSUE_URL=$(gh issue create --repo "$REPO" --title "$TITLE" --body "$BODY" --label "$LABELS")

        # Extract the issue number from the URL.
        # 从 URL 中提取 issue 编号。
        ISSUE_NUMBER=$(echo "$NEW_ISSUE_URL" | awk -F'/' '{print $NF}')

        if [ -z "$ISSUE_NUMBER" ]; then
            echo "Error: Failed to create GitHub issue for $file"
            continue
        fi

        echo "Successfully created GitHub Issue #$ISSUE_NUMBER"

        # Write the new issue number back to the local file.
        # 将新的 issue 编号写回本地文件。
        yq -i ".remote_id = $ISSUE_NUMBER" "$file"
        echo "Updated $file with remote_id: $ISSUE_NUMBER"

    else
        # Case 2: remote_id exists. Check if the issue needs to be closed.
        # 场景 2: remote_id 已存在。检查是否需要关闭 issue。
        echo "Status: Found remote_id #$REMOTE_ID. Checking status..."

        if [ "$STATUS" == "resolved" ] || [ "$STATUS" == "wont-fix" ]; then
            # Check the state of the remote issue before trying to close it.
            # 在尝试关闭远程 issue 之前，检查其状态。
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

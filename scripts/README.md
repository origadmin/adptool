# Issue 同步脚本说明 / Issue Sync Script Documentation

---

## 中文

本目录包含用于将本地 `.issues` 文件与远程 GitHub Issues 同步的自动化脚本。

### 依赖

在运行脚本之前，你必须安装并配置好以下命令行工具：

1.  **GitHub CLI (`gh`)**: 这是与你的 GitHub 仓库进行交互的核心工具。
    -   **安装**: 请参考官方文档进行安装：<https://cli.github.com/>
    -   **认证**: 安装后，运行 `gh auth login` 并按照提示完成认证，确保 `gh` 有权限访问你的仓库。

2.  **`yq`**: 一个轻量级的命令行 YAML/JSON/XML 处理器。我们用它来安全、准确地读取和写入 issue 文件头部的元数据 (Front Matter)。
    -   **安装**: 请参考其官方 GitHub 仓库的安装说明：<https://github.com/mikefarah/yq/>

### 脚本

#### `sync_issues.sh`

这个脚本是整个自动化流程的核心。它会智能地将本地的变更推送到 GitHub，或将 GitHub 的更新拉取到本地。

##### 功能：

1.  **创建远程 Issue**: 遍历 `.issues` 目录，如果发现一个本地 issue 文件没有 `remote_id`，它会在 GitHub 上创建一个对应的 Issue，并将返回的 Issue 编号写回本地文件的 `remote_id` 字段。
2.  **关闭远程 Issue**: 如果脚本发现一个本地 issue 的状态被更新为 `resolved` 或 `wont-fix`，并且它有关联的 `remote_id`，脚本会自动在 GitHub 上关闭这个 Issue。

##### 使用方法:

```bash
# 在项目根目录下运行
bash scripts/sync_issues.sh
```

建议在你完成了一系列本地工作，准备将代码 `git push` 到远程仓库之前，运行一次此脚本，以确保远程和本地的状态同步。

---

## English

This directory contains automation scripts for synchronizing local `.issues` files with remote GitHub Issues.

### Dependencies

Before running the scripts, you must install and configure the following command-line tools:

1.  **GitHub CLI (`gh`)**: The core tool for interacting with your GitHub repository.
    -   **Installation**: Please refer to the official documentation for installation instructions: <https://cli.github.com/>
    -   **Authentication**: After installation, run `gh auth login` and follow the prompts to authenticate, ensuring `gh` has permission to access your repository.

2.  **`yq`**: A lightweight command-line YAML/JSON/XML processor. We use it to safely and accurately read and write metadata (Front Matter) in the issue file headers.
    -   **Installation**: Please refer to the installation instructions on its official GitHub repository: <https://github.com/mikefarah/yq/>

### Scripts

#### `sync_issues.sh`

This script is the core of the automation process. It intelligently pushes local changes to GitHub or pulls updates from GitHub to your local files.

##### Features:

1.  **Create Remote Issues**: It iterates through the `.issues` directory. If it finds a local issue file without a `remote_id`, it creates a corresponding issue on GitHub and writes the returned issue number back to the `remote_id` field of the local file.
2.  **Close Remote Issues**: If the script finds that a local issue's status has been updated to `resolved` or `wont-fix` and it has an associated `remote_id`, the script will automatically close the issue on GitHub.

##### Usage:

```bash
# Run from the project root directory
bash scripts/sync_issues.sh
```

It is recommended to run this script after you have completed a series of local tasks and are ready to `git push` your code to the remote repository, ensuring that the remote and local states are synchronized.

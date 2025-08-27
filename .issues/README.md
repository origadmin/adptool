# 本地优先的 Issue 管理工作流 / Local-First Issue Management Workflow

---

## 中文

本文档描述了一套专为单人开发、追求高效率而设计的 Issue 管理流程。它以本地文件和 Git 为核心，确保您在绝大多数时间里都能停留在 IDE 和终端中，同时完全兼容未来的团队协作和 GitHub Issues 集成。

### 核心思想

- **每个 Issue 都是一个文件**: 每个问题（无论是 bug、功能需求还是文档任务）都作为一个独立的 Markdown 文件进行管理，确保问题的独立性和清晰性。
- **Git 驱动**: 所有 Issue 的创建、状态变更和关闭都通过 Git 提交来追踪，历史记录清晰、可追溯。
- **本地优先**: 整个流程的设计都旨在最大化本地操作效率，避免频繁切换到浏览器。
- **未来兼容**: 本地工作流可以随时通过脚本与 GitHub Issues 进行双向同步，无缝扩展到团队协作模式。

### 工作流程

#### 1. 创建 Issue

当你发现一个新问题或产生一个新想法时，在 `.issues` 目录下复制一份模板文件，并以 `ID-简短描述.md` 的格式命名。

```bash
# 示例：创建一个关于"用户登录模块"的 bug issue
cp .issues/.template.md .issues/5-user-login-module-bug.md
```

然后，在 IDE 中打开这个新文件，修改文件顶部的元数据（`status`, `labels` 等）并填写详细描述。

#### 2. 开始处理 Issue

为这个 Issue 创建一个专门的 Git 分支。分支名应与 Issue 文件名保持关联，以便快速溯源。

```bash
# 分支名格式推荐：issue/ID
git checkout -b issue/5
```

现在，你可以在这个分支上进行所有相关的代码修改。

#### 3. 测试与验证

在声明一个问题被修复之前，必须进行验证。这是保证软件质量的关键步骤。

1.  **编写测试**: 针对 Bug 修复，应编写一个或多个单元测试来复现该 Bug，并证明你的修复是有效的。
2.  **执行验证**: 在对应的 Issue 文件中，填写 `### 验证步骤 / Verification Steps` 部分，详细说明如何确认修复的有效性。这不仅是为你自己记录，也为未来的协作者提供了清晰的指引。

#### 4. 完成并关闭 Issue

在代码经过验证后，按以下步骤结束工作：

1.  **更新 Issue 文件**: 打开对应的 Issue 文件，将 `status` 字段修改为 `resolved`。
2.  **提交所有变更**: 将你的代码变更、测试文件和 Issue 文件的状态变更一起提交。

    ```bash
    # 添加所有代码、测试和 Issue 文件的变更
    git add .

    # 编写一个清晰的 commit message
    git commit -m "fix: 修复用户登录模块的 bug\n\nResolve issue #5"
    ```

3.  **合并分支**: 将修复分支合并回主干。

    ```bash
    git checkout main
    git merge issue/5
    ```

4.  **(可选) 清理分支**: 删除已经合并的 issue 分支。

    ```bash
    git branch -d issue/5
    ```

### (可选) 与 GitHub Issues 同步

当你希望将本地工作记录推送到 GitHub，或与他人协作时，可以使用自动化脚本来同步。详细操作请参考项目根目录下 `scripts/README.md` 中的说明。

---

## English

This document describes an issue management process designed for solo developers who prioritize efficiency. It centers around local files and Git, ensuring you can stay within your IDE and terminal for the vast majority of the time, while remaining fully compatible with future team collaboration and GitHub Issues integration.

### Core Concepts

- **Every Issue is a File**: Each issue (be it a bug, feature request, or documentation task) is managed as a separate Markdown file, ensuring clarity and independence.
- **Git-Driven**: The creation, status changes, and closing of all issues are tracked through Git commits, providing a clear and traceable history.
- **Local-First**: The entire process is designed to maximize local operational efficiency and avoid frequent context switching to a browser.
- **Future-Proof**: The local workflow can be synchronized with GitHub Issues at any time via a script, allowing for seamless expansion to a team-based collaboration model.

### The Workflow

#### 1. Creating an Issue

When you find a new problem or have a new idea, copy the template file in the `.issues` directory and name it using the format `ID-short-description.md`.

```bash
# Example: Create a bug issue for the "user login module"
cp .issues/.template.md .issues/5-user-login-module-bug.md
```

Next, open this new file in your IDE, modify the metadata at the top (like `status`, `labels`, etc.), and fill in the detailed description.

#### 2. Starting Work on an Issue

Create a dedicated Git branch for the issue. The branch name should correspond to the issue file to allow for easy tracking.

```bash
# Recommended branch name format: issue/ID
git checkout -b issue/5
```

Now you can make all related code changes on this branch.

#### 3. Testing and Verification

Before declaring an issue resolved, verification is mandatory. This is a critical step for ensuring software quality.

1.  **Write Tests**: For bug fixes, one or more unit tests should be written to reproduce the bug and then prove that your fix is effective.
2.  **Perform Verification**: In the corresponding issue file, fill out the `### 验证步骤 / Verification Steps` section, detailing how to confirm the fix is effective. This serves not only as a record for yourself but also as a clear guide for future collaborators.

#### 4. Completing and Closing an Issue

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

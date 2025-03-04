# RAG SYNC

## 注意！注意 | Important Security Notice!

### 安全提醒 | Security Reminder

⚠️ **阿里云访问密钥安全至关重要，请务必注意以下事项：**

⚠️ **Alibaba Cloud access key security is crucial, please pay attention to the following:**

1. **避免泄露密钥** - 请勿在公共代码仓库、日志或其他不安全的地方存储 AccessKey 和 SecretKey。
   
   **Avoid Key Leakage** - Do not store your AccessKey and SecretKey in public repositories, logs, or other insecure locations.

2. **使用 RAM 用户** - 强烈建议使用 RAM (Resource Access Management) 用户而非主账号的 AccessKey。
   
   **Use RAM Users** - It is strongly recommended to use RAM (Resource Access Management) user credentials instead of your primary account AccessKey.

3. **最小权限原则** - 为 RAM 用户仅分配必要的最小权限：
   
   **Principle of Least Privilege** - Assign only the minimum necessary permissions to RAM users:
   
   - Bailian 相关 API 的权限
   
   - Permissions for Bailian-related APIs

4. **百炼控制台授权** - 在阿里云百炼控制台中，确保为 RAM 用户授予相应的工作空间权限。
   
   **Bailian Console Authorization** - In the Alibaba Cloud Bailian console, ensure that the RAM user is granted appropriate workspace permissions.

### 环境建议 | Environment Recommendations

🔴 **超级管理员账户警告：**

🔴 **Super Administrator Account Warning:**

- 在开发和调试阶段可以临时使用超级管理员账户的 AccessKey
- 在生产环境中**绝对不要**使用超级管理员账户的 AccessKey
- 超级管理员账户被泄露可能导致云账户的所有资源受到威胁

- You may temporarily use the AccessKey of a super administrator account during development and debugging
- **NEVER** use super administrator AccessKey in production environments
- Leakage of a super administrator account may compromise all resources in your cloud account

### 配置 RAM 的步骤 | Steps to Configure RAM

1. 登录阿里云 RAM 控制台：https://ram.console.aliyun.com/
2. 创建 RAM 用户并生成 AccessKey
3. 为该用户授予 Bailian 和 OSS 相关权限
4. 将生成的 AccessKey 和 SecretKey 配置到 ragsync 中

1. Log in to Alibaba Cloud RAM Console: https://ram.console.aliyun.com/
2. Create a RAM user and generate AccessKey
3. Grant Bailian and OSS related permissions to this user
4. Configure the generated AccessKey and SecretKey in ragsync

## 最新更新 | Latest Updates

### v1.0.4 到 v1.0.7 重要更新 | Important Updates from v1.0.4 to v1.0.7

#### 增强的日志系统 | Enhanced Logging System
- **文件处理详细日志** - 每个文件操作都有清晰的日志前缀，如 `[File: filename]` 或 `[Dir: dirname]`
- **进度报告** - 批量处理时每 5 个文件显示一次进度
- **错误定位** - 精确定位错误发生的文件和操作步骤
- **时间戳信息** - 显示文件的修改时间和远程文件的创建时间比较结果

#### 智能文件处理 | Smart File Processing
- **自动跳过重复索引** - 检测到文件已在索引中时自动跳过
- **时间戳智能比较** - 自动比较本地和远程文件的时间戳，避免覆盖较新的文件
- **批量处理优化** - 改进了目录扫描和文件过滤的性能
- **错误恢复机制** - 批量处理时单个文件失败不影响整体进度

#### 新增命令参数 | New Command Parameters
- **--override-newest-data** - 允许覆盖较新的远程文件（需要与 --force 一起使用）
- **--ext** - 支持在批量上传时指定多个文件扩展名（如 ".txt,.pdf,.docx"）
- **--skip-index-delete** - 更新文件时保留原有索引条目

### 版本功能增强 | Version Enhancements

- **增强的日志系统**：添加了更详细的日志输出，方便快速定位问题
- **索引记录查询**：增加了查询文档是否已在索引中的功能 
- **优化的文件上传**：优化了文件上传流程，防止重复添加文档到索引
- **时间戳比较**：添加了本地文件与远程文件时间戳比较功能，避免意外覆盖较新的文件

## 项目简介 | Project Introduction

RAG SYNC 是一个用于管理阿里云百炼知识库的命令行工具，支持上传、删除、查询文件和管理知识索引等操作。

RAG SYNC is a command-line tool for managing Alibaba Cloud Bailian knowledge base, supporting operations such as uploading, deleting, querying files, and managing knowledge indices.

## 安装 | Installation

### 方式一：Go Install（推荐）| Method 1: Go Install (Recommended)

使用 Go 工具链直接安装：

Install directly using the Go toolchain:

```bash
# 安装最新版本，生成名为 ragsync 的可执行文件
# Install the latest version, generate an executable named ragsync
go install github.com/VillanCh/ragsync/cmd/ragsync@v1.0.0
```

安装完成后，可以直接在命令行中使用 `ragsync` 命令。

After installation, you can use the `ragsync` command directly in the command line.

#### 排障指南 | Troubleshooting Guide

如果 `which ragsync` 或 `where ragsync` 命令找不到已安装的程序，请检查以下几点：

If the `which ragsync` or `where ragsync` command cannot find the installed program, please check the following:

1. **检查 Go 的安装路径** | **Check Go installation path**:
   ```bash
   # 查看 Go 安装的可执行文件路径 | Check the path where Go installs executables
   go env GOPATH
   ```
   
   确保 `$GOPATH/bin` 已添加到系统的 PATH 环境变量中：
   
   Make sure `$GOPATH/bin` is added to your system's PATH environment variable:
   
   ```bash
   # 在 Linux/Mac 上 | On Linux/Mac
   echo 'export PATH=$PATH:$(go env GOPATH)/bin' >> ~/.bashrc
   source ~/.bashrc
   
   # 或对于 Zsh shell | Or for Zsh shell
   echo 'export PATH=$PATH:$(go env GOPATH)/bin' >> ~/.zshrc
   source ~/.zshrc
   
   # 在 Windows 上，请添加到系统环境变量中 | On Windows, add to system environment variables
   ```

2. **直接运行可执行文件** | **Run the executable directly**:
   ```bash
   # 找到并运行可执行文件 | Find and run the executable
   $(go env GOPATH)/bin/ragsync
   ```

### 方式二：本地编译 | Method 2: Local Build

如果您需要修改代码或从源码构建，可以按照以下步骤操作：

If you need to modify the code or build from source, you can follow these steps:

```bash
# 克隆仓库 | Clone the repository
git clone https://github.com/VillanCh/ragsync.git

# 进入项目目录 | Enter the project directory
cd ragsync

# 编译项目 | Build the project
go build -o ragsync ./ragsync.go

# 将编译后的文件移动到可执行路径（可选）| Move the compiled file to an executable path (optional)
sudo mv ragsync /usr/local/bin/
```

## 配置 | Configuration

首次使用前，需要创建配置文件：

Before using for the first time, you need to create a configuration file:

```bash
# 创建配置文件 | Create a configuration file
ragsync create-config
```

配置文件默认保存在 `~/.ragsync/ragsync.yaml`，您也可以使用 `-o` 参数指定其他路径：

The configuration file is saved in `~/.ragsync/ragsync.yaml` by default. You can also specify another path using the `-o` parameter:

```bash
ragsync create-config -o /path/to/config.yaml
```

配置文件包含以下字段：

The configuration file contains the following fields:

| 字段 | Field | 描述 | Description |
|------|-------|------|-------------|
| aliyun_access_key | aliyun_access_key | 阿里云访问密钥 ID | Alibaba Cloud Access Key ID |
| aliyun_secret_key | aliyun_secret_key | 阿里云访问密钥密码 | Alibaba Cloud Access Key Secret |
| bailian_workspace_id | bailian_workspace_id | 百炼工作空间 ID | Bailian Workspace ID |
| aliyun_bailian_endpoint | aliyun_bailian_endpoint | 百炼 API 端点 | Bailian API Endpoint |
| bailian_category_type | bailian_category_type | 分类类型 | Category Type |
| bailian_add_file_parser | bailian_add_file_parser | 文件解析器 | File Parser |
| bailian_files_default_category_id | bailian_files_default_category_id | 默认分类 ID | Default Category ID |
| bailian_knowledge_index_id | bailian_knowledge_index_id | 知识库索引 ID | Knowledge Base Index ID |

## 使用方法 | Usage

### 基本用法 | Basic Usage

```bash
# 使用指定配置文件 | Use a specific configuration file
ragsync -c /path/to/config.yaml [command]

# 查看帮助 | View help
ragsync help
```

### 验证配置 | Validate Configuration

```bash
# 验证配置文件是否有效 | Validate if the configuration file is valid
ragsync validate
```

### 上传文件 | Upload Files

```bash
# 上传文件到知识库（默认自动添加到知识索引）| Upload a file to knowledge base (automatically added to knowledge index by default)
ragsync sync --file /path/to/file.txt

# 上传但不添加到知识索引 | Upload without adding to knowledge index
ragsync sync --file /path/to/file.txt --no-index

# 强制上传（如果文件已存在则覆盖）| Force upload (overwrite if file exists)
ragsync sync --file /path/to/file.txt --force

# 强制上传但保留原有索引条目（不删除旧索引）| Force upload but preserve existing index entries (don't delete old index)
ragsync sync --file /path/to/file.txt --force --skip-index-delete

# 强制上传并覆盖较新的远程文件 | Force upload and override newer remote files
ragsync sync --file /path/to/file.txt --force --override-newest-data

# 批量上传目录中的所有支持文件 | Batch upload all supported files in a directory
ragsync sync --dir /path/to/directory

# 批量上传特定类型的文件 | Batch upload specific types of files
ragsync sync --dir /path/to/directory --ext ".txt,.pdf,.docx"

# 强制批量上传目录中的所有文件 | Force batch upload all files in a directory
ragsync sync --dir /path/to/directory --force
```

### 示例 4：批量上传目录中的文件 | Example 4: Batch upload files in a directory

```bash
# 使用默认扩展名批量上传目录中的文件
# Batch upload files in a directory with default extensions
ragsync sync --dir /path/to/documents

# 上传所有PDF和DOCX文件到知识库并添加到索引
# Upload all PDF and DOCX files to the knowledge base and add to index
ragsync sync --dir /path/to/documents --ext ".pdf,.docx"

# 强制替换所有已存在的文件，保留原有索引条目
# Force replace all existing files, preserving original index entries
ragsync sync --dir /path/to/documents --force --skip-index-delete
```

### 示例 5：处理版本冲突 | Example 5: Handling version conflicts

```bash
# 尝试上传文件（如果远程文件较新，则会自动跳过）
# Try to upload file (will skip automatically if remote file is newer)
ragsync sync --file /path/to/document.pdf

# 强制上传并覆盖远程文件，即使远程文件可能较新
# Force upload and override remote files, even if they might be newer
ragsync sync --file /path/to/document.pdf --force --override-newest-data

# 批量上传所有文件，忽略远程文件时间戳
# Batch upload all files, ignoring remote file timestamps
ragsync sync --dir /path/to/documents --force --override-newest-data
```

### 文件时间比较逻辑 | File Time Comparison Logic

当您使用 `sync` 命令上传文件时，ragsync 会自动比较本地文件的修改时间与远程文件的创建时间：

When you use the `sync` command to upload files, ragsync automatically compares the local file's modification time with the remote file's creation time:

1. **自动检测冲突** - 如果远程文件比本地文件更新，上传会被自动跳过以防止覆盖更新的内容。
   
   **Automatic Conflict Detection** - If the remote file is newer than the local file, upload will be automatically skipped to prevent overwriting newer content.

2. **选择性覆盖** - 使用 `--force` 可以覆盖已存在的文件，但默认情况下仍会检查时间戳。
   
   **Selective Overriding** - Use `--force` to override existing files, but timestamps will still be checked by default.

3. **完全覆盖** - 当您需要忽略时间戳检查时，同时使用 `--force` 和 `--override-newest-data` 选项。
   
   **Complete Overriding** - When you need to ignore timestamp checks, use both `--force` and `--override-newest-data` options.

4. **协作安全** - 此功能有助于防止在协作环境中意外覆盖他人的更新。
   
   **Collaboration Safety** - This feature helps prevent accidentally overwriting others' updates in collaborative environments.

### 列出文件 | List Files

```bash
# 列出所有文件 | List all files
ragsync list

# 按名称筛选文件 | Filter files by name
ragsync list --name "文档名称"
```

### 查询文件状态 | Check File Status

```bash
# 通过文件名查询状态 | Check status by file name
ragsync status --name "文档名称"
```

### 删除文件 | Delete Files

```bash
# 通过文件 ID 删除文件（同时从知识索引中删除）| Delete file by ID (also removes from knowledge index)
ragsync delete --id "file-id"

# 通过文件名删除文件 | Delete file by name
ragsync delete --name "文档名称"

# 强制删除（不询问确认）| Force delete (without confirmation)
ragsync delete --name "文档名称" --force

# 删除文件但保留索引条目 | Delete file but preserve index entries
ragsync delete --name "文档名称" --skip-index-delete
```

### 添加文件到知识索引 | Add File to Knowledge Index

```bash
# 通过文件 ID 添加已存在的文件到知识索引 | Add existing file to knowledge index by ID
ragsync add-job --id "file-id"

# 通过文件名添加已存在的文件到知识索引 | Add existing file to knowledge index by name
ragsync add-job --name "文档名称"

# 强制添加（不询问确认）| Force add without confirmation
ragsync add-job --name "文档名称" --force
```

### 管理索引任务 | Manage Index Jobs

```bash
# 列出所有索引任务 | List all index jobs
ragsync jobs

# 查询索引任务状态 | Check index job status
ragsync index-status --job-id "job-id"
```

## 命令参数详解 | Command Parameters

### sync（上传文件 | Upload File）

| 参数 | Parameter | 描述 | Description |
|------|-----------|------|-------------|
| --file | --file | 要上传的文件路径 | File path to upload |
| --dir | --dir | 要递归扫描并上传文件的目录路径 | Directory path to recursively scan and upload files |
| --ext | --ext | 与--dir一起使用时要上传的文件扩展名（逗号分隔，如 '.txt,.pdf,.md'）| File extensions to upload when using --dir (comma separated, e.g. '.txt,.pdf,.md') |
| --force, -f | --force, -f | 强制上传（即使文件已存在）| Force upload even if file exists |
| --override-newest-data, -o | --override-newest-data, -o | 覆盖比本地文件更新的远程文件（需要与--force一起使用）| Override remote files even if they are newer than local files (requires --force) |
| --no-index, -n | --no-index, -n | 跳过将文件添加到知识索引 | Skip adding the file to knowledge index |
| --skip-index-delete, -s | --skip-index-delete, -s | 替换文件时，跳过从知识索引中先删除文件（保留索引条目）| When replacing files, skip removing them from the knowledge index first (preserves index entries) |

### list（列出文件 | List Files）

| 参数 | Parameter | 描述 | Description |
|------|-----------|------|-------------|
| --name | --name | 按名称筛选文件（可选）| Filter files by name (optional) |

### status（查询状态 | Check Status）

| 参数 | Parameter | 描述 | Description |
|------|-----------|------|-------------|
| --name | --name | 要查询状态的文件名 | File name to check status |

### delete（删除文件 | Delete File）

| 参数 | Parameter | 描述 | Description |
|------|-----------|------|-------------|
| --id | --id | 要删除的文件 ID | File ID to delete |
| --name | --name | 要删除的文件名 | File name to search and delete |
| --force, -f | --force, -f | 强制删除（不询问确认）| Force delete without confirmation |
| --skip-index-delete, -s | --skip-index-delete, -s | 跳过从知识索引中删除文件（保留索引条目）| Skip removing the file from knowledge index before deletion (preserves index entries) |

### add-job（添加文件到知识索引 | Add File to Knowledge Index）

| 参数 | Parameter | 描述 | Description |
|------|-----------|------|-------------|
| --id | --id | 要添加到索引的文件 ID | File ID to add to index |
| --name | --name | 要添加到索引的文件名 | File name to search and add to index |
| --force, -f | --force, -f | 强制添加（不询问确认）| Force add without confirmation |

### jobs（列出索引任务 | List Index Jobs）

无参数。列出所有已保存的索引任务 ID。

No parameters. Lists all saved index job IDs.

### index-status（查询索引任务状态 | Check Index Job Status）

| 参数 | Parameter | 描述 | Description |
|------|-----------|------|-------------|
| --job-id | --job-id | 要查询的索引任务 ID | Index job ID to check |
| --auto | --auto | 自动检查状态直到任务完成或失败 | Automatically check status until the job completes or fails |
| --cleanup | --cleanup | 任务完成或失败后自动清理任务记录 | Automatically clean up job records after the job completes or fails |

## 工作流示例 | Workflow Examples

### 示例 1：添加新文件并索引 | Example 1: Add a new file and index it

```bash
# 上传文件并自动添加到知识索引（默认行为）
# Upload a file and automatically add it to the knowledge index (default behavior)
ragsync sync --file /path/to/document.pdf

# 查询索引任务状态（使用返回的 job-id）
# Check the index job status (using the returned job-id)
ragsync index-status --job-id "job-id-returned-from-sync" --auto
```

### 示例 2：分步骤上传和索引 | Example 2: Upload and index in separate steps

```bash
# 上传文件但不添加到知识索引
# Upload a file without adding it to the knowledge index
ragsync sync --file /path/to/document.pdf --no-index

# 列出文件以获取文件 ID
# List files to get the file ID
ragsync list --name "document.pdf"

# 将文件添加到知识索引
# Add the file to the knowledge index
ragsync add-job --id "file-id-from-list"
```

### 示例 3：更新现有文件并保留索引 | Example 3: Update an existing file and preserve index

```bash
# 上传新版本的文件，保留旧的索引条目
# Upload a new version of the file, preserving old index entries
ragsync sync --file /path/to/updated-document.pdf --force --skip-index-delete

# 将新文件添加到知识索引（现在会有两个索引条目指向不同版本）
# Add the new file to the knowledge index (now there will be two index entries pointing to different versions)
ragsync add-job --name "updated-document.pdf"
```

### 示例 4：批量上传目录中的文件 | Example 4: Batch upload files in a directory

```bash
# 使用默认扩展名批量上传目录中的文件
# Batch upload files in a directory with default extensions
ragsync sync --dir /path/to/documents

# 上传所有PDF和DOCX文件到知识库并添加到索引
# Upload all PDF and DOCX files to the knowledge base and add to index
ragsync sync --dir /path/to/documents --ext ".pdf,.docx"

# 强制替换所有已存在的文件，保留原有索引条目
# Force replace all existing files, preserving original index entries
ragsync sync --dir /path/to/documents --force --skip-index-delete
```

### 示例 5：处理版本冲突 | Example 5: Handling version conflicts

```bash
# 尝试上传文件（如果远程文件较新，则会自动跳过）
# Try to upload file (will skip automatically if remote file is newer)
ragsync sync --file /path/to/document.pdf

# 强制上传并覆盖远程文件，即使远程文件可能较新
# Force upload and override remote files, even if they might be newer
ragsync sync --file /path/to/document.pdf --force --override-newest-data

# 批量上传所有文件，忽略远程文件时间戳
# Batch upload all files, ignoring remote file timestamps
ragsync sync --dir /path/to/documents --force --override-newest-data
```

## 注意事项 | Notes

1. 文件上传后需要时间处理，可以使用 `status` 命令监控处理进度。
   
   Files need time to process after upload, you can use the `status` command to monitor the processing progress.

2. 文件名搜索会自动去除文件扩展名，以提高匹配的准确性。
   
   File name search automatically removes file extensions to improve matching accuracy.

3. 删除文件操作不可恢复，请谨慎操作。
   
   File deletion operations are irreversible, please operate with caution.

## 许可证 | License

MIT

## 联系方式 | Contact

如有任何问题或建议，请提交 Issue 或 Pull Request。

For any questions or suggestions, please submit an Issue or Pull Request.

## 排查问题 | Troubleshooting

### 增强的日志输出 | Enhanced Logging

最新版本的 ragsync 增加了详细的日志输出，帮助快速定位问题：

The latest version of ragsync has added detailed log output to help quickly locate issues:

1. **详细的命令参数日志** | **Detailed command parameter logs**
   - 每个命令执行时会记录所有使用的参数
   - Each command execution records all parameters used

2. **文件处理日志** | **File processing logs**
   - 每个文件处理的详细步骤都有明确的日志
   - 包含文件大小、修改时间等详细信息
   - Detailed steps for each file processing have clear logs
   - Including file size, modification time, and other detailed information

3. **错误定位日志** | **Error locating logs**
   - 文件名与操作一起记录，方便在批量处理时定位问题
   - File names are recorded along with operations, making it easy to locate problems during batch processing

### 常见错误 | Common Errors

1. **认证失败** | **Authentication Failed**
   ```
   Failed to create Bailian client: Authentication failed
   ```
   - 检查 AccessKey 和 SecretKey 配置
   - 确认用户有权限访问指定的工作空间
   - Check the AccessKey and SecretKey configuration
   - Confirm the user has permission to access the specified workspace

2. **找不到工作空间** | **Workspace Not Found**
   ```
   Failed to create Bailian client: Workspace not found
   ```
   - 确认工作空间 ID 正确
   - 检查用户是否有该工作空间的权限
   - Confirm that the workspace ID is correct
   - Check if the user has permissions for the workspace

3. **索引 ID 未配置** | **Index ID Not Configured**
   ```
   Cannot add to knowledge index: BailianKnowledgeIndexId is not configured
   ```
   - 在配置文件中设置 `bailian_knowledge_index_id` 字段
   - Set the `bailian_knowledge_index_id` field in the configuration file

4. **文件已在索引中** | **File Already in Index**
   ```
   Document is already being indexed or has been indexed. Skipping index addition.
   ```
   - 文件已经存在于索引中，不会重复添加
   - The file already exists in the index and will not be added again

### 检查文件索引状态 | Check File Indexing Status

通过查看详细日志，可以了解文件在索引中的状态：

By looking at the detailed logs, you can understand the status of files in the index:

```
[File: document.pdf] Found exact match for document: document
[File: document.pdf] Document 'document' is already being indexed or has been indexed
```

## 高级特性 | Advanced Features

### 文件时间戳比较 | File Timestamp Comparison

ragsync 会自动比较本地文件的修改时间与远程文件的创建时间：

ragsync automatically compares the modification time of local files with the creation time of remote files:

- 如果本地文件较新，默认会上传文件
- 如果远程文件较新，默认会跳过上传
- 使用 `--force --override-newest-data` 可以强制覆盖较新的远程文件

- If the local file is newer, it will upload the file by default
- If the remote file is newer, it will skip the upload by default
- Use `--force --override-newest-data` to force overwrite newer remote files

### 索引检查与去重 | Index Checking and Deduplication

在添加文件到索引前，ragsync 会检查文件是否已经在索引中：

Before adding a file to the index, ragsync checks if the file is already in the index:

- 如果文件已在索引中，将自动跳过索引添加步骤
- 避免文件重复索引，节省资源和时间
- 可以查看详细日志了解文件的索引状态

- If the file is already in the index, it will automatically skip the index addition step
- Avoids duplicate indexing of files, saving resources and time
- You can view detailed logs to understand the indexing status of files

### 新功能使用示例 | New Features Usage Examples

#### 智能时间戳处理 | Smart Timestamp Handling
```bash
# 上传文件时自动检查时间戳（如果远程文件较新会自动跳过）
ragsync sync --file document.pdf

# 查看详细的时间戳比较日志
[File: document.pdf] Local file modified: 2024-03-20 10:30:00
[File: document.pdf] Remote file created: 2024-03-20 11:00:00
[File: document.pdf] Skipping upload: remote file is newer

# 强制覆盖较新的远程文件
ragsync sync --file document.pdf --force --override-newest-data
```

#### 批量处理与进度报告 | Batch Processing and Progress Reporting
```bash
# 上传目录中的所有 PDF 和 DOCX 文件，显示进度
ragsync sync --dir /path/to/docs --ext ".pdf,.docx"

# 查看详细的进度日志
[Dir: /path/to/docs] Scanning directory...
[File: doc1.pdf] Processing... Done
[File: doc2.docx] Processing... Done
[Progress] Processed 5/20 files (25%)
```

#### 索引优化 | Index Optimization
```bash
# 更新文件但保留原有索引（适用于小幅更新）
ragsync sync --file document.pdf --force --skip-index-delete

# 自动检测并跳过已索引文件
[File: document.pdf] Checking index status...
[File: document.pdf] Document already indexed, skipping index addition
```

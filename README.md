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
go install github.com/VillanCh/ragsync/cmd/ragsync@0.2.0
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

# 批量上传目录中的所有支持文件 | Batch upload all supported files in a directory
ragsync sync --dir /path/to/directory

# 批量上传特定类型的文件 | Batch upload specific types of files
ragsync sync --dir /path/to/directory --ext ".txt,.pdf,.docx"

# 强制批量上传目录中的所有文件 | Force batch upload all files in a directory
ragsync sync --dir /path/to/directory --force
```

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

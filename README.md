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
   - OSS 存储相关权限（用于文件上传）
   
   - Permissions for Bailian-related APIs
   - OSS storage-related permissions (for file uploading)

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

RAG SYNC 是一个用于管理阿里云百炼知识库的命令行工具，支持上传、删除、查询文件等操作。

RAG SYNC is a command-line tool for managing Alibaba Cloud Bailian knowledge base, supporting operations such as uploading, deleting, and querying files.

## 安装 | Installation

```bash
# 克隆仓库 | Clone the repository
git clone https://github.com/VillanCh/ragsync.git

# 进入项目目录 | Enter the project directory
cd ragsync

# 编译项目 | Build the project
go build -o ragsync cmd/ragsync.go

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
# 上传文件到知识库 | Upload a file to the knowledge base
ragsync sync --file /path/to/file.txt

# 强制上传（如果文件已存在则覆盖）| Force upload (overwrite if file exists)
ragsync sync --file /path/to/file.txt --force
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
# 通过文件 ID 删除文件 | Delete file by ID
ragsync delete --id "file-id"

# 通过文件名删除文件 | Delete file by name
ragsync delete --name "文档名称"

# 强制删除（不询问确认）| Force delete (without confirmation)
ragsync delete --name "文档名称" --force
```

## 命令参数详解 | Command Parameters

### sync（上传文件 | Upload File）

| 参数 | Parameter | 描述 | Description |
|------|-----------|------|-------------|
| --file | --file | 要上传的文件路径 | File path to upload |
| --force, -f | --force, -f | 强制上传（即使文件已存在） | Force upload even if file exists |

### list（列出文件 | List Files）

| 参数 | Parameter | 描述 | Description |
|------|-----------|------|-------------|
| --name | --name | 按名称筛选文件（可选） | Filter files by name (optional) |

### status（查询状态 | Check Status）

| 参数 | Parameter | 描述 | Description |
|------|-----------|------|-------------|
| --name | --name | 要查询状态的文件名 | File name to check status |

### delete（删除文件 | Delete File）

| 参数 | Parameter | 描述 | Description |
|------|-----------|------|-------------|
| --id | --id | 要删除的文件 ID | File ID to delete |
| --name | --name | 要删除的文件名 | File name to delete |
| --force, -f | --force, -f | 强制删除（不询问确认） | Force delete without confirmation |

### create-config（创建配置 | Create Configuration）

| 参数 | Parameter | 描述 | Description |
|------|-----------|------|-------------|
| --output, -o | --output, -o | 配置文件输出路径 | Output path for configuration file |

## 示例 | Examples

### 完整工作流 | Complete Workflow

```bash
# 创建配置 | Create configuration
ragsync create-config

# 上传文件 | Upload file
ragsync sync --file ~/Documents/knowledge.pdf

# 查看上传的文件 | View uploaded files
ragsync list

# 检查文件处理状态 | Check file processing status
ragsync status --name "knowledge.pdf"

# 删除文件 | Delete file
ragsync delete --name "knowledge.pdf"
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

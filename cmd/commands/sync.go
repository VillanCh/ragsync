package commands

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/urfave/cli"

	"github.com/VillanCh/ragsync/common/aliyun"
	"github.com/VillanCh/ragsync/common/spec"

	"github.com/yaklang/yaklang/common/log"
	"github.com/yaklang/yaklang/common/utils"
)

// SyncCommand 上传文件命令
func SyncCommand() cli.Command {
	return cli.Command{
		Name:    "sync",
		Aliases: []string{"upload"},
		Usage:   "Apply for file upload lease",
		Flags: []cli.Flag{
			cli.StringFlag{
				Name:  "file",
				Usage: "File path to upload",
			},
			cli.StringFlag{
				Name:  "dir",
				Usage: "Directory path to recursively scan and upload files",
			},
			cli.StringFlag{
				Name:  "ext",
				Usage: "File extensions to upload when using --dir (comma separated, e.g. '.txt,.pdf,.md')",
				Value: ".txt,.md,.markdown,.json,.pdf,.doc,.docx",
			},
			cli.BoolFlag{
				Name:  "force,f",
				Usage: "Force upload even if file exists",
			},
			cli.BoolFlag{
				Name:  "override-newest-data,o",
				Usage: "Override remote files even if they are newer than local files (requires --force)",
			},
			cli.BoolFlag{
				Name:  "no-index,n",
				Usage: "Skip adding the file to knowledge index",
			},
			cli.BoolFlag{
				Name:  "skip-index-delete,s",
				Usage: "When replacing files, skip removing them from the knowledge index first (preserves index entries)",
			},
		},
		Action: executeSync,
	}
}

// executeSync 上传文件的执行逻辑
func executeSync(c *cli.Context) error {
	log.Infof("Starting sync operation...")

	// 从配置文件加载配置
	config, err := LoadConfig(c)
	if err != nil {
		log.Errorf("Failed to load configuration: %v", err)
		return err
	}
	log.Infof("Configuration loaded successfully from: %s", c.GlobalString("config"))

	filePath := c.String("file")
	dirPath := c.String("dir")

	// 打印命令参数
	log.Infof("Command parameters: file=%s, dir=%s, force=%v, override-newest-data=%v, no-index=%v, skip-index-delete=%v",
		filePath,
		dirPath,
		c.Bool("force"),
		c.Bool("override-newest-data"),
		c.Bool("no-index"),
		c.Bool("skip-index-delete"))

	// 文件和目录参数必须至少提供一个
	if filePath == "" && dirPath == "" {
		log.Errorf("No file or directory path specified")
		return utils.Errorf("Please specify either file path (--file) or directory path (--dir) to upload")
	}

	// 文件和目录参数不能同时提供
	if filePath != "" && dirPath != "" {
		log.Errorf("Both file and directory paths specified, only one is allowed")
		return utils.Errorf("Cannot specify both --file and --dir at the same time")
	}

	forceUpload := c.Bool("force")
	overrideNewestData := c.Bool("override-newest-data")
	skipIndex := c.Bool("no-index")
	skipIndexDelete := c.Bool("skip-index-delete")

	// 如果设置了 override-newest-data 但没有设置 force，给出警告
	if overrideNewestData && !forceUpload {
		log.Warnf("--override-newest-data requires --force to be effective, but --force is not set")
	}

	// 默认会添加到索引，除非指定了--no-index
	addToIndex := !skipIndex

	log.Infof("Add to index: %v", addToIndex)

	// 如果需要添加到索引，但索引ID未配置，则返回错误
	if addToIndex && config.BailianKnowledgeIndexId == "" {
		log.Errorf("BailianKnowledgeIndexId is not configured in config file")
		return utils.Errorf("Cannot add to knowledge index: BailianKnowledgeIndexId is not configured in your config file")
	}

	log.Infof("Creating Bailian client with workspace ID: %s", config.BailianWorkspaceId)
	client, err := aliyun.NewBailianClientFromConfig(config)
	if err != nil {
		log.Errorf("Failed to create Bailian client: %v", err)
		return err
	}
	log.Infof("Bailian client created successfully")

	// 如果指定了目录，则遍历目录并上传符合条件的文件
	if dirPath != "" {
		extensions := strings.Split(c.String("ext"), ",")
		// 去除可能存在的空格
		for i := range extensions {
			extensions[i] = strings.TrimSpace(extensions[i])
		}

		log.Infof("Processing directory upload with extensions: %v", extensions)
		return processDirUpload(dirPath, extensions, client, config, forceUpload, addToIndex, skipIndexDelete, overrideNewestData)
	}

	// 处理单个文件上传
	log.Infof("Processing single file upload: %s", filePath)
	return processFileUpload(filePath, client, config, forceUpload, addToIndex, skipIndexDelete, overrideNewestData)
}

// processDirUpload 处理目录递归上传
func processDirUpload(dirPath string, extensions []string, client *aliyun.BailianClient, config *spec.Config, forceUpload, addToIndex, skipIndexDelete bool, overrideNewestData bool) error {
	log.Infof("[Dir: %s] Starting directory processing", dirPath)

	// 检查目录是否存在
	dirInfo, err := os.Stat(dirPath)
	if err != nil {
		log.Errorf("[Dir: %s] Failed to access directory: %v", dirPath, err)
		return utils.Errorf("Failed to access directory: %v", err)
	}

	if !dirInfo.IsDir() {
		log.Errorf("[Dir: %s] The specified path is not a directory", dirPath)
		return utils.Errorf("The specified path is not a directory: %s", dirPath)
	}

	// 存储上传成功和失败的文件计数
	successCount := 0
	failedCount := 0
	skippedCount := 0

	log.Infof("[Dir: %s] Scanning directory", dirPath)
	log.Infof("[Dir: %s] File extensions to process: %s", dirPath, strings.Join(extensions, ", "))

	// 遍历目录中的所有文件
	err = filepath.Walk(dirPath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			log.Warnf("[Dir: %s] Error accessing path %s: %v", dirPath, path, err)
			return nil
		}

		// 跳过目录
		if info.IsDir() {
			log.Debugf("[Dir: %s] Skipping directory: %s", dirPath, path)
			return nil
		}

		// 检查文件扩展名是否符合要求
		ext := strings.ToLower(filepath.Ext(path))
		if !isExtensionAllowed(ext, extensions) {
			log.Infof("[Dir: %s] Skipping file with unsupported extension: %s (ext: %s)", dirPath, path, ext)
			skippedCount++
			return nil
		}

		log.Infof("[Dir: %s] Processing file (%d processed so far): %s", dirPath, successCount+failedCount, path)

		// 使用与单文件上传相同的逻辑处理
		err = processFileUpload(path, client, config, forceUpload, addToIndex, skipIndexDelete, overrideNewestData)
		if err != nil {
			log.Errorf("[Dir: %s] Failed to upload file %s: %v", dirPath, path, err)
			failedCount++
		} else {
			log.Infof("[Dir: %s] Successfully processed file: %s", dirPath, path)
			successCount++
		}

		// 显示进度报告
		if (successCount+failedCount)%5 == 0 {
			log.Infof("[Dir: %s] Progress: %d files processed (%d success, %d failed, %d skipped)",
				dirPath, successCount+failedCount+skippedCount, successCount, failedCount, skippedCount)
		}

		return nil
	})

	if err != nil {
		log.Errorf("[Dir: %s] Failed to walk directory: %v", dirPath, err)
		return utils.Errorf("Failed to walk directory: %v", err)
	}

	// 打印处理结果摘要
	log.Infof("[Dir: %s] Directory processing completed", dirPath)
	log.Infof("[Dir: %s] Results: %d files processed, %d uploaded successfully, %d failed, %d skipped (wrong extension)",
		dirPath, successCount+failedCount+skippedCount, successCount, failedCount, skippedCount)

	return nil
}

// isExtensionAllowed 检查文件扩展名是否在允许的列表中
func isExtensionAllowed(ext string, allowedExtensions []string) bool {
	for _, allowed := range allowedExtensions {
		if strings.EqualFold(ext, allowed) {
			return true
		}
	}
	return false
}

// processFileUpload 处理单个文件上传
func processFileUpload(filePath string, client *aliyun.BailianClient, config *spec.Config, forceUpload, addToIndex, skipIndexDelete bool, overrideNewestData bool) error {
	log.Infof("[File: %s] Starting file upload process", filePath)

	// 获取本地文件信息
	fileInfo, err := os.Stat(filePath)
	if err != nil {
		log.Errorf("[File: %s] Failed to get file information: %v", filePath, err)
		return utils.Errorf("Failed to get file information: %v", err)
	}

	log.Infof("[File: %s] File size: %d bytes, Last modified: %s",
		filePath, fileInfo.Size(), fileInfo.ModTime().Format(time.RFC3339))

	// 获取文件修改时间
	fileModTime := fileInfo.ModTime()

	log.Infof("[File: %s] Reading file content", filePath)
	fileContent, err := os.ReadFile(filePath)
	if err != nil {
		log.Errorf("[File: %s] Failed to read file content: %v", filePath, err)
		return err
	}
	log.Infof("[File: %s] File content read successfully, size: %d bytes", filePath, len(fileContent))

	fileName := filePath

	// 是否需要上传新文件（默认为true）
	needUpload := true
	// 需要添加到索引的文件ID
	var fileId string

	// 检查文件是否已存在（无论是否为强制模式）
	log.Infof("[File: %s] Checking if file already exists on server", fileName)

	// 列出所有匹配该文件名的文件
	existingFiles, err := client.ListAllFiles(fileName)
	if err != nil {
		log.Warnf("[File: %s] Failed to check existing files: %v", fileName, err)
		log.Info("[File: %s] Proceeding with upload anyway...", fileName)
	} else if len(existingFiles) > 0 {
		log.Infof("[File: %s] Found %d existing files with similar name", fileName, len(existingFiles))

		// 显示找到的文件
		fmt.Println("Found existing files with similar name:")
		fmt.Printf("\n%-40s %-50s %-15s\n", "File ID", "File Name", "Status")
		fmt.Println(strings.Repeat("-", 105))

		for _, file := range existingFiles {
			fmt.Printf("%-40s %-50s %-15s\n", file.FileId, file.FileName, file.Status)
		}

		// 显示本地文件的修改时间
		fmt.Printf("\nLocal file last modified: %s\n", fileModTime.Format(time.RFC3339))

		// 远程文件时间信息
		var remoteCreateTime time.Time
		var remoteFileTimeStr string
		var parseTimeErr error

		if len(existingFiles) > 0 {
			// 使用第一个文件的创建时间
			remoteFileTimeStr = existingFiles[0].CreateTime
			// 解析远程文件创建时间，格式通常为: "2023-08-18 11:03:35"
			remoteCreateTime, parseTimeErr = time.Parse("2006-01-02 15:04:05", remoteFileTimeStr)
			if parseTimeErr != nil {
				log.Warnf("Failed to parse remote file time: %v (value: %s)", parseTimeErr, remoteFileTimeStr)
				// 继续使用默认逻辑，假设本地文件更新
				remoteCreateTime = time.Time{}
			}
		}

		// 输出远程文件时间信息
		if remoteFileTimeStr != "" {
			fmt.Printf("Remote file created: %s\n", remoteFileTimeStr)
			if parseTimeErr == nil {
				fmt.Printf("Remote file time (parsed): %s\n", remoteCreateTime.Format(time.RFC3339))
			}
		} else {
			fmt.Println("Remote file timestamp not available.")
		}

		// 比较本地文件修改时间和远程文件创建时间
		// 只有当解析远程时间成功时才进行实际比较
		isLocalNewer := true
		if !remoteCreateTime.IsZero() {
			// 时区处理：remoteCreateTime 是北京时间 (UTC+8)
			// 如果需要调整时区，可以在这里进行
			isLocalNewer = fileModTime.After(remoteCreateTime)

			if isLocalNewer {
				log.Infof("Local file (%s) is newer than remote file (%s)",
					fileModTime.Format(time.RFC3339),
					remoteCreateTime.Format(time.RFC3339))
			} else {
				log.Infof("Remote file (%s) is newer than local file (%s)",
					remoteCreateTime.Format(time.RFC3339),
					fileModTime.Format(time.RFC3339))
			}
		} else {
			log.Infof("Cannot compare file times. Assuming local file is newer.")
		}

		// 如果设置了 --override-newest-data，则忽略时间比较结果
		if overrideNewestData && forceUpload {
			log.Infof("Overriding time comparison due to --override-newest-data flag")
			isLocalNewer = true
		}

		// 如果本地文件不比远程文件新，除非设置了强制覆盖+忽略时间，否则不上传
		if !isLocalNewer && !(forceUpload && overrideNewestData) {
			log.Infof("Local file appears to not be newer than remote file. Skipping upload.")
			needUpload = false

			// 使用最新的文件ID（如果有多个同名文件，使用第一个）
			if len(existingFiles) > 0 {
				fileId = existingFiles[0].FileId
				log.Infof("Using existing file ID: %s", fileId)
			} else {
				return utils.Errorf("No valid file ID found among existing files")
			}
		} else if forceUpload {
			// 本地文件较新或已设置强制覆盖+忽略时间，并且设置了强制上传
			if skipIndexDelete {
				log.Info("Force mode enabled. Deleting existing files (skipping index deletion)...")
			} else {
				log.Info("Force mode enabled. Deleting existing files and their index entries...")
			}

			for _, file := range existingFiles {
				log.Infof("Deleting file: %s (ID: %s)", file.FileName, file.FileId)
				// 使用DeleteFileEx方法，可以控制是否跳过索引删除
				err := client.DeleteFileEx(file.FileId, skipIndexDelete)
				if err != nil {
					log.Warnf("Failed to delete file %s: %v", file.FileId, err)
				} else {
					log.Infof("File deleted successfully: %s", file.FileId)
				}
			}
		} else {
			// 本地文件较新，但未设置强制模式，询问用户是否要覆盖
			deleteMsg := "File with the same name already exists. Local file appears to be newer."
			if skipIndexDelete {
				deleteMsg += " Do you want to delete it and upload a new version? (Index entries will be preserved)"
			} else {
				deleteMsg += " Do you want to delete it and upload a new version? (This will also update index entries)"
			}

			if !askForConfirmation(deleteMsg) {
				log.Info("Upload cancelled. Using existing file.")
				needUpload = false

				// 使用最新的文件ID（如果有多个同名文件，使用第一个）
				if len(existingFiles) > 0 {
					fileId = existingFiles[0].FileId
					log.Infof("Using existing file ID: %s", fileId)
				} else {
					return utils.Errorf("No valid file ID found among existing files")
				}
			} else {
				// 用户确认删除并上传新版本
				for _, file := range existingFiles {
					log.Infof("Deleting file: %s (ID: %s)", file.FileName, file.FileId)
					// 使用DeleteFileEx方法，可以控制是否跳过索引删除
					err := client.DeleteFileEx(file.FileId, skipIndexDelete)
					if err != nil {
						log.Warnf("Failed to delete file %s: %v", file.FileId, err)
						return err
					}
					log.Infof("File deleted successfully: %s", file.FileId)
				}
			}
		}
	}

	// 如果需要上传新文件
	if needUpload {
		log.Infof("[File: %s] Initiating file upload process", filePath)
		log.Infof("[File: %s] Applying for file upload lease", filePath)
		lis, err := client.ApplyFileUploadLease(filePath, fileContent)
		if err != nil {
			log.Errorf("[File: %s] Failed to apply for upload lease: %v", filePath, err)
			return err
		}
		log.Infof("[File: %s] Upload lease acquired successfully", filePath)

		headers := utils.InterfaceToGeneralMap(lis.Headers)
		bailianExtra, ok := headers["X-bailian-extra"]
		if !ok {
			log.Errorf("[File: %s] X-bailian-extra header not found in lease response", filePath)
			return utils.Errorf("X-bailian-extra does not exist")
		}
		contentType, ok := headers["Content-Type"]
		if !ok {
			log.Errorf("[File: %s] Content-Type header not found in lease response", filePath)
			return utils.Errorf("Content-Type does not exist")
		}

		log.Infof("[File: %s] Uploading file to URL: %s", filePath, lis.UploadURL)
		log.Infof("[File: %s] Upload method: %s, Content-Type: %s", filePath, lis.Method, contentType)

		// Upload file
		err = aliyun.UploadFile(lis.Method, lis.UploadURL, filePath, fmt.Sprint(contentType), fileContent, fmt.Sprintf("%s", bailianExtra))
		if err != nil {
			log.Errorf("[File: %s] File upload failed: %v", filePath, err)
			return err
		}
		log.Infof("[File: %s] File content uploaded successfully", filePath)

		log.Infof("[File: %s] Adding file to Bailian RAG with lease ID: %s", filePath, lis.LeaseId)
		fileId, err = client.AddFile(lis.LeaseId)
		if err != nil {
			log.Errorf("[File: %s] Failed to add file to Bailian RAG: %v", filePath, err)
			return err
		}

		log.Infof("[File: %s] File added successfully with ID: %s", filePath, fileId)
	} else {
		log.Infof("[File: %s] Using existing file, skipping upload", filePath)
	}

	// 无论是新上传的文件还是使用已有文件，如果需要添加到索引，就执行索引步骤
	if addToIndex && fileId != "" {
		log.Infof("[File: %s] Adding file (ID: %s) to knowledge index: %s",
			filePath, fileId, config.BailianKnowledgeIndexId)

		jobId, err := client.AppendDocumentToIndex(fileId)
		if err != nil {
			log.Errorf("[File: %s] Failed to add file to knowledge index: %v", filePath, err)
			return err
		}

		if jobId != "" {
			log.Infof("[File: %s] File added to knowledge index successfully. Job ID: %s", filePath, jobId)
			log.Infof("[File: %s] You can check the job status with: ragsync job --job-id %s", filePath, jobId)
		} else {
			log.Warnf("[File: %s] File was processed, but no job ID was returned. The file may still be added to the index.", filePath)
		}
	} else if !addToIndex {
		log.Infof("[File: %s] Skipping knowledge index step (--no-index was specified)", filePath)
	} else {
		log.Warnf("[File: %s] Cannot add to index: file ID is empty", filePath)
	}

	log.Infof("[File: %s] File processing completed successfully", filePath)
	return nil
}

// askForConfirmation 请求用户确认
func askForConfirmation(s string) bool {
	fmt.Printf("%s [y/N]: ", s)

	var response string
	_, err := fmt.Scanln(&response)
	if err != nil {
		return false
	}

	response = strings.ToLower(strings.TrimSpace(response))
	return response == "y" || response == "yes"
}

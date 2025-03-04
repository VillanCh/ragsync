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
			cli.StringFlag{
				Name:  "exclude",
				Usage: "Keywords to exclude files (comma separated, e.g. 'draft,temp,private')",
				Value: "temp,private,unverified,unverified_,ignored",
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

	// 解析排除关键字
	excludeKeywords := []string{}
	if c.String("exclude") != "" {
		excludeKeywords = strings.Split(c.String("exclude"), ",")
		// 去除可能存在的空格
		for i := range excludeKeywords {
			excludeKeywords[i] = strings.TrimSpace(excludeKeywords[i])
		}
		log.Infof("Exclusion keywords: %v", excludeKeywords)
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

	// 如果既没有指定文件也没有指定目录，使用配置文件中的 include_paths
	if filePath == "" && dirPath == "" {
		if len(config.IncludePaths) == 0 {
			log.Errorf("No file or directory path specified, and no include_paths in config")
			return utils.Errorf("Please specify either file path (--file) or directory path (--dir) to upload, or configure include_paths in your config file")
		}
		log.Infof("No file or directory specified, using include_paths from config: %v", config.IncludePaths)

		// 验证所有路径是否存在
		var invalidPaths []string
		for _, path := range config.IncludePaths {
			if _, err := os.Stat(path); os.IsNotExist(err) {
				invalidPaths = append(invalidPaths, path)
			}
		}

		if len(invalidPaths) > 0 {
			log.Errorf("Some include paths do not exist: %v", invalidPaths)
			return utils.Errorf("The following paths specified in include_paths do not exist: %v", invalidPaths)
		}

		// 处理所有有效的路径
		for _, path := range config.IncludePaths {
			log.Infof("Processing include path: %s", path)
			// 检查路径是文件还是目录
			pathInfo, err := os.Stat(path)
			if err != nil {
				log.Errorf("Failed to stat path %s: %v", path, err)
				continue
			}

			if pathInfo.IsDir() {
				// 如果是目录，使用目录处理逻辑
				extensions := strings.Split(c.String("ext"), ",")
				for i := range extensions {
					extensions[i] = strings.TrimSpace(extensions[i])
				}
				if err := processDirUpload(path, extensions, excludeKeywords, client, config, forceUpload, addToIndex, skipIndexDelete, overrideNewestData); err != nil {
					log.Errorf("Failed to process directory %s: %v", path, err)
					continue
				}
			} else {
				// 如果是文件，使用文件处理逻辑
				if containsExcludedKeywords(path, excludeKeywords) {
					log.Infof("[File: %s] Skipped due to exclusion keywords", path)
					continue
				}
				if err := processFileUpload(path, client, config, forceUpload, addToIndex, skipIndexDelete, overrideNewestData); err != nil {
					log.Errorf("Failed to process file %s: %v", path, err)
					continue
				}
			}
		}
		return nil
	}

	// 文件和目录参数不能同时提供
	if filePath != "" && dirPath != "" {
		log.Errorf("Both file and directory paths specified, only one is allowed")
		return utils.Errorf("Cannot specify both --file and --dir at the same time")
	}

	// 如果指定了目录，则遍历目录并上传符合条件的文件
	if dirPath != "" {
		extensions := strings.Split(c.String("ext"), ",")
		// 去除可能存在的空格
		for i := range extensions {
			extensions[i] = strings.TrimSpace(extensions[i])
		}

		log.Infof("Processing directory upload with extensions: %v", extensions)
		return processDirUpload(dirPath, extensions, excludeKeywords, client, config, forceUpload, addToIndex, skipIndexDelete, overrideNewestData)
	}

	// 处理单个文件上传
	log.Infof("Processing single file upload: %s", filePath)
	if containsExcludedKeywords(filePath, excludeKeywords) {
		log.Infof("[File: %s] Skipped due to exclusion keywords", filePath)
		return nil
	}
	return processFileUpload(filePath, client, config, forceUpload, addToIndex, skipIndexDelete, overrideNewestData)
}

// processDirUpload 处理目录递归上传
func processDirUpload(dirPath string, extensions []string, excludeKeywords []string, client *aliyun.BailianClient, config *spec.Config, forceUpload, addToIndex, skipIndexDelete bool, overrideNewestData bool) error {
	if strings.Trim(dirPath, "./") == "" {
		return utils.Errorf("Directory path cannot be empty")
	}

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

	// 获取目录的绝对路径
	absDirPath, err := filepath.Abs(dirPath)
	if err != nil {
		log.Errorf("[Dir: %s] Failed to get absolute path: %v", dirPath, err)
		return err
	}

	// 获取本地文件列表
	localFiles := make(map[string]bool)
	err = filepath.Walk(absDirPath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if !info.IsDir() {
			// 检查文件扩展名是否符合要求
			ext := strings.ToLower(filepath.Ext(path))
			if isExtensionAllowed(ext, extensions) {
				// 检查是否包含排除关键字
				if !containsExcludedKeywords(path, excludeKeywords) {
					// 使用相对路径作为键
					relPath, err := filepath.Rel(absDirPath, path)
					if err != nil {
						log.Warnf("[Dir: %s] Failed to get relative path for %s: %v", dirPath, path, err)
						return nil
					}
					key := filepath.Join(dirPath, relPath)
					log.Infof("[Dir: %s] Found file: %s", dirPath, key)
					localFiles[key] = true
				}
			}
		}
		return nil
	})
	if err != nil {
		log.Errorf("[Dir: %s] Failed to scan local directory: %v", dirPath, err)
		return err
	}

	// 获取远程文件列表
	remoteFileRaw, err := client.ListAllFiles("")
	if err != nil {
		log.Errorf("[Dir: %s] Failed to list remote files: %v", dirPath, err)
		return err
	}

	remoteFiles := make([]*aliyun.FileInfo, 0, len(remoteFileRaw))
	for _, fileDesc := range remoteFileRaw {
		dirPathWithoutDot := dirPath
		for strings.HasPrefix(dirPathWithoutDot, "./") {
			dirPathWithoutDot = strings.TrimPrefix(dirPathWithoutDot, "./")
		}
		fileWithoutDot := fileDesc.FileName
		for strings.HasPrefix(fileWithoutDot, "./") {
			fileWithoutDot = strings.TrimPrefix(fileWithoutDot, "./")
		}
		hasPrefix := strings.HasPrefix(fileWithoutDot, dirPathWithoutDot)
		if hasPrefix {
			log.Infof("[Dir: %s] Found remote file: %s", dirPath, fileDesc.FileName)
			remoteFiles = append(remoteFiles, fileDesc)
		}
	}

	// 存储上传成功和失败的文件计数
	successCount := 0
	failedCount := 0
	skippedCount := 0
	deletedCount := 0

	// 找出需要删除的远程文件
	var filesToDelete = make(map[string]string)
	var skippedFiles = make(map[string]bool)
	var uploadFiles []string
	for _, remoteFile := range remoteFiles {
		// 检查远程文件是否在本地文件列表中
		if have, ok := localFiles[remoteFile.FileName]; !ok || !have {
			log.Infof("[Dir: %s] Remote file %s not found locally, will be deleted", dirPath, remoteFile.FileName)
			filesToDelete[remoteFile.FileName] = remoteFile.FileId
			deletedCount++
			continue
		} else if remoteFile.CreateTime != "" {
			// 解析远程文件创建时间
			remoteCreateTime, parseErr := time.Parse("2006-01-02 15:04:05", remoteFile.CreateTime)
			if parseErr != nil {
				log.Warnf("[Dir: %s] Failed to parse remote file time for %s: %v", dirPath, remoteFile.FileName, parseErr)
			} else {
				// 获取本地文件信息
				localFilePath := remoteFile.FileName
				localFileInfo, statErr := os.Stat(localFilePath)
				if statErr != nil {
					log.Warnf("[Dir: %s] Failed to get local file info for %s: %v", dirPath, localFilePath, statErr)
				} else {
					// 比较本地文件修改时间和远程文件创建时间
					localModTime := localFileInfo.ModTime()
					if !localModTime.After(remoteCreateTime) && !overrideNewestData {
						log.Infof("[Dir: %s] Remote file %s is newer than local file, skipping", dirPath, remoteFile.FileName)
						skippedFiles[remoteFile.FileName] = true
						skippedCount++
						continue
					}
				}
			}
		}
	}

	log.Infof("[Dir: %s] Found %d local files", dirPath, len(localFiles))
	for localFilename := range localFiles {
		if _, ok := skippedFiles[localFilename]; ok {
			log.Infof("[Dir: %s] Skipping file because it was skipped earlier: %s", dirPath, localFilename)
			continue
		}
		if _, ok := filesToDelete[localFilename]; ok {
			log.Infof("[Dir: %s] Skipping file because it was marked for deletion: %s", dirPath, localFilename)
			continue
		}
		log.Infof("[Dir: %s] Uploading file: %s", dirPath, localFilename)
		uploadFiles = append(uploadFiles, localFilename)
	}

	// 删除不在本地的远程文件
	if len(filesToDelete) > 0 {
		log.Infof("[Dir: %s] Deleting %d remote files that don't exist locally", dirPath, len(filesToDelete))
		for _, fileId := range filesToDelete {
			if err := client.DeleteFileEx(fileId, false); err != nil {
				log.Errorf("[Dir: %s] Failed to delete remote file %s: %v", dirPath, fileId, err)
				continue
			}
			log.Infof("[Dir: %s] Successfully deleted remote file %s", dirPath, fileId)
		}
	}

	log.Infof("[Dir: %s] Scanning directory", dirPath)
	log.Infof("[Dir: %s] File extensions to process: %s", dirPath, strings.Join(extensions, ", "))

	for _, path := range uploadFiles {
		info, err := os.Stat(path)
		if err != nil {
			log.Warnf("[Dir: %s] Error accessing path %s: %v", dirPath, path, err)
			failedCount++
			continue
		}

		// 跳过目录
		if info.IsDir() {
			log.Infof("[Dir: %s] Skipping directory: %s", dirPath, path)
			skippedCount++
			continue
		}

		// 检查文件扩展名是否符合要求
		ext := strings.ToLower(filepath.Ext(path))
		if !isExtensionAllowed(ext, extensions) {
			log.Infof("[Dir: %s] Skipping file with unsupported extension: %s (ext: %s)", dirPath, path, ext)
			skippedCount++
			continue
		}

		// 检查是否包含排除关键字
		if containsExcludedKeywords(path, excludeKeywords) {
			log.Infof("[Dir: %s] Skipping file containing excluded keywords: %s", dirPath, path)
			skippedCount++
			continue
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
	}

	// 打印处理结果摘要
	log.Infof("[Dir: %s] Directory processing completed", dirPath)
	log.Infof("[Dir: %s] Results: %d files processed, %d uploaded successfully, %d failed, %d skipped (wrong extension), %d remote files deleted",
		dirPath, successCount+failedCount+skippedCount, successCount, failedCount, skippedCount, len(filesToDelete))

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

// containsExcludedKeywords 检查文件名是否包含排除关键字
func containsExcludedKeywords(filePath string, excludeKeywords []string) bool {
	if len(excludeKeywords) == 0 {
		return false
	}

	fileName := filepath.Base(filePath)
	fileNameLower := strings.ToLower(fileName)

	for _, keyword := range excludeKeywords {
		if keyword == "" {
			continue
		}
		if strings.Contains(fileNameLower, strings.ToLower(keyword)) {
			return true
		}
	}
	return false
}

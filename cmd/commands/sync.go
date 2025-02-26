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
	// 从配置文件加载配置
	config, err := LoadConfig(c)
	if err != nil {
		return err
	}

	filePath := c.String("file")
	dirPath := c.String("dir")

	// 文件和目录参数必须至少提供一个
	if filePath == "" && dirPath == "" {
		return utils.Errorf("Please specify either file path (--file) or directory path (--dir) to upload")
	}

	// 文件和目录参数不能同时提供
	if filePath != "" && dirPath != "" {
		return utils.Errorf("Cannot specify both --file and --dir at the same time")
	}

	forceUpload := c.Bool("force")
	skipIndex := c.Bool("no-index")
	skipIndexDelete := c.Bool("skip-index-delete")

	// 默认会添加到索引，除非指定了--no-index
	addToIndex := !skipIndex

	// 如果需要添加到索引，但索引ID未配置，则返回错误
	if addToIndex && config.BailianKnowledgeIndexId == "" {
		return utils.Errorf("Cannot add to knowledge index: BailianKnowledgeIndexId is not configured in your config file")
	}

	client, err := aliyun.NewBailianClientFromConfig(config)
	if err != nil {
		return err
	}

	// 如果指定了目录，则遍历目录并上传符合条件的文件
	if dirPath != "" {
		extensions := strings.Split(c.String("ext"), ",")
		// 去除可能存在的空格
		for i := range extensions {
			extensions[i] = strings.TrimSpace(extensions[i])
		}

		return processDirUpload(dirPath, extensions, client, config, forceUpload, addToIndex, skipIndexDelete)
	}

	// 处理单个文件上传
	return processFileUpload(filePath, client, config, forceUpload, addToIndex, skipIndexDelete)
}

// processDirUpload 处理目录递归上传
func processDirUpload(dirPath string, extensions []string, client *aliyun.BailianClient, config *spec.Config, forceUpload, addToIndex, skipIndexDelete bool) error {
	// 检查目录是否存在
	dirInfo, err := os.Stat(dirPath)
	if err != nil {
		return utils.Errorf("Failed to access directory: %v", err)
	}

	if !dirInfo.IsDir() {
		return utils.Errorf("The specified path is not a directory: %s", dirPath)
	}

	// 存储上传成功和失败的文件计数
	successCount := 0
	failedCount := 0
	skippedCount := 0

	log.Infof("Scanning directory: %s", dirPath)
	log.Infof("File extensions to process: %s", strings.Join(extensions, ", "))

	// 遍历目录中的所有文件
	err = filepath.Walk(dirPath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			log.Warnf("Error accessing path %s: %v", path, err)
			return nil
		}

		// 跳过目录
		if info.IsDir() {
			return nil
		}

		// 检查文件扩展名是否符合要求
		ext := strings.ToLower(filepath.Ext(path))
		if !isExtensionAllowed(ext, extensions) {
			skippedCount++
			return nil
		}

		log.Infof("Processing file: %s", path)

		// 使用与单文件上传相同的逻辑处理
		err = processFileUpload(path, client, config, forceUpload, addToIndex, skipIndexDelete)
		if err != nil {
			log.Errorf("Failed to upload file %s: %v", path, err)
			failedCount++
		} else {
			successCount++
		}

		return nil
	})

	if err != nil {
		return utils.Errorf("Failed to walk directory: %v", err)
	}

	// 打印处理结果摘要
	log.Infof("Directory processing completed: %s", dirPath)
	log.Infof("Results: %d files processed, %d uploaded successfully, %d failed, %d skipped (wrong extension)",
		successCount+failedCount+skippedCount, successCount, failedCount, skippedCount)

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
func processFileUpload(filePath string, client *aliyun.BailianClient, config *spec.Config, forceUpload, addToIndex, skipIndexDelete bool) error {
	// 获取本地文件信息
	fileInfo, err := os.Stat(filePath)
	if err != nil {
		return utils.Errorf("Failed to get file information: %v", err)
	}

	// 获取文件修改时间
	fileModTime := fileInfo.ModTime()

	fileContent, err := os.ReadFile(filePath)
	if err != nil {
		return err
	}

	fileName := filePath

	// 是否需要上传新文件（默认为true）
	needUpload := true
	// 需要添加到索引的文件ID
	var fileId string

	// 检查文件是否已存在（无论是否为强制模式）
	log.Infof("Checking if file '%s' already exists...", fileName)

	// 列出所有匹配该文件名的文件
	existingFiles, err := client.ListAllFiles(fileName)
	if err != nil {
		log.Warnf("Failed to check existing files: %v", err)
		log.Info("Proceeding with upload anyway...")
	} else if len(existingFiles) > 0 {
		// 显示找到的文件
		fmt.Println("Found existing files with similar name:")
		fmt.Printf("\n%-40s %-50s %-15s\n", "File ID", "File Name", "Status")
		fmt.Println(strings.Repeat("-", 105))

		for _, file := range existingFiles {
			fmt.Printf("%-40s %-50s %-15s\n", file.FileId, file.FileName, file.Status)
		}

		// 显示本地文件的修改时间
		fmt.Printf("\nLocal file last modified: %s\n", fileModTime.Format(time.RFC3339))
		fmt.Println("Note: Remote file timestamps are not available through the API.")

		// 如果是强制模式，直接删除文件
		if forceUpload {
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
			// 在非强制模式下询问用户是否删除文件并上传新版本
			deleteMsg := "File with the same name already exists."
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
		log.Infof("Uploading file: %s", filePath)
		lis, err := client.ApplyFileUploadLease(filePath, fileContent)
		if err != nil {
			return err
		}
		headers := utils.InterfaceToGeneralMap(lis.Headers)
		bailianExtra, ok := headers["X-bailian-extra"]
		if !ok {
			return utils.Errorf("X-bailian-extra does not exist")
		}
		contentType, ok := headers["Content-Type"]
		if !ok {
			return utils.Errorf("Content-Type does not exist")
		}

		// Upload file
		err = aliyun.UploadFile(lis.Method, lis.UploadURL, filePath, fmt.Sprint(contentType), fileContent, fmt.Sprintf("%s", bailianExtra))
		if err != nil {
			return err
		}

		log.Info("Adding file to Bailian RAG")
		fileId, err = client.AddFile(lis.LeaseId)
		if err != nil {
			return err
		}

		log.Infof("File added successfully with ID: %s", fileId)
	}

	// 无论是新上传的文件还是使用已有文件，如果需要添加到索引，就执行索引步骤
	if addToIndex && fileId != "" {
		log.Infof("Adding file to knowledge index: %s", config.BailianKnowledgeIndexId)

		jobId, err := client.AppendDocumentToIndex(fileId)
		if err != nil {
			log.Errorf("Failed to add file to knowledge index: %v", err)
			return err
		}

		if jobId != "" {
			log.Infof("File added to knowledge index successfully. Job ID: %s", jobId)
			log.Info("You can check the job status with: ragsync job --job-id " + jobId)
		} else {
			log.Warnf("File was processed, but no job ID was returned. The file may still be added to the index.")
		}
	} else if !addToIndex {
		log.Info("Skipping knowledge index step (--no-index was specified)")
	}

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

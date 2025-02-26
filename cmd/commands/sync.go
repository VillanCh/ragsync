package commands

import (
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/urfave/cli"

	"github.com/VillanCh/ragsync/common/aliyun"

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
				Usage: "Skip deleting from index when replacing existing files",
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
	if filePath == "" {
		return utils.Errorf("Please specify the file path to upload")
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

	client, err := aliyun.NewBailianClientFromConfig(config)
	if err != nil {
		return err
	}

	fileName := filePath

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
			// 在非强制模式下询问用户是否删除文件
			deleteMsg := "File with the same name already exists."
			if skipIndexDelete {
				deleteMsg += " Do you want to delete it before uploading? (Index entries will be preserved)"
			} else {
				deleteMsg += " Do you want to delete it and its index entries before uploading?"
			}

			if !askForConfirmation(deleteMsg) {
				log.Info("Upload cancelled")
				return nil
			}

			// 用户确认删除
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
	fileId, err := client.AddFile(lis.LeaseId)
	if err != nil {
		return err
	}

	log.Infof("File added successfully with ID: %s", fileId)

	// 如果用户没有选择跳过索引，则将文件添加到知识库索引
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
	} else if skipIndex {
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

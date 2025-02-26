package commands

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

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
				Name:  "force, f",
				Usage: "Force upload even if file exists",
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

	fileContent, err := os.ReadFile(filePath)
	if err != nil {
		return err
	}

	client, err := aliyun.NewBailianClientFromConfig(config)
	if err != nil {
		return err
	}

	// 获取文件名用于检查是否已存在
	fileName := filepath.Base(filePath)

	// 只有在非强制模式下才检查文件是否存在
	if !forceUpload {
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

			// 询问用户是否继续
			if !askForConfirmation("File with the same name already exists. Do you want to continue uploading?") {
				log.Info("Upload cancelled")
				return nil
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
	err = client.AddFile(lis.LeaseId)
	if err != nil {
		return err
	}

	log.Info("File added successfully")
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

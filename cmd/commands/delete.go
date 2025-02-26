package commands

import (
	"fmt"
	"strings"

	"github.com/urfave/cli"

	"github.com/VillanCh/ragsync/common/aliyun"

	"github.com/yaklang/yaklang/common/log"
	"github.com/yaklang/yaklang/common/utils"
)

// DeleteCommand 删除文件命令
func DeleteCommand() cli.Command {
	return cli.Command{
		Name:    "delete",
		Aliases: []string{"del"},
		Usage:   "Delete file from Bailian workspace",
		Flags: []cli.Flag{
			cli.StringFlag{
				Name:  "id",
				Usage: "File ID to delete",
			},
			cli.StringFlag{
				Name:  "name",
				Usage: "File name to search and delete",
			},
			cli.BoolFlag{
				Name:  "force, f",
				Usage: "Force delete without confirmation",
			},
		},
		Action: executeDelete,
	}
}

// executeDelete 删除文件的执行逻辑
func executeDelete(c *cli.Context) error {
	// 从配置文件加载配置
	config, err := LoadConfig(c)
	if err != nil {
		return err
	}

	fileId := c.String("id")
	fileName := c.String("name")
	forceDelete := c.Bool("force")

	if fileId == "" && fileName == "" {
		return utils.Errorf("Please specify either file ID (--id) or file name (--name) to delete")
	}

	client, err := aliyun.NewBailianClientFromConfig(config)
	if err != nil {
		return err
	}

	// 如果提供了文件名，需要先查找对应的文件
	if fileId == "" && fileName != "" {
		// 首先列出所有匹配的文件
		log.Infof("Searching for files with name: %s", fileName)
		files, err := client.ListAllFiles(fileName)
		if err != nil {
			return utils.Errorf("Failed to list files: %v", err)
		}

		// 检查是否找到匹配的文件
		if len(files) == 0 {
			return utils.Errorf("No files found with name: %s", fileName)
		}

		// 如果找到多个文件，让用户选择
		if len(files) > 1 {
			fmt.Println("Multiple files found with this name:")
			fmt.Printf("\n%-40s %-50s %-15s\n", "File ID", "File Name", "Status")
			fmt.Println(strings.Repeat("-", 105))

			for i, file := range files {
				fmt.Printf("[%d] %-36s %-50s %-15s\n", i+1, file.FileId, file.FileName, file.Status)
			}

			fmt.Print("\nPlease enter the number of the file to delete (or 0 to cancel): ")
			var selection int
			fmt.Scanln(&selection)

			if selection <= 0 || selection > len(files) {
				return utils.Errorf("Operation cancelled or invalid selection")
			}

			fileId = files[selection-1].FileId
			fileName = files[selection-1].FileName
		} else {
			fileId = files[0].FileId
			fileName = files[0].FileName
		}
	} else if fileId != "" && fileName == "" {
		// 如果只提供了ID，获取文件名用于确认
		fileInfo, err := client.DescribeFile(fileId)
		if err != nil {
			log.Warnf("Failed to get file information: %v", err)
			// 继续执行，但没有文件名用于确认
		} else {
			fileName = fileInfo.FileName
		}
	}

	// 确认删除操作
	if !forceDelete {
		confirmMessage := fmt.Sprintf("Are you sure you want to delete the file: %s (ID: %s)", fileName, fileId)
		if !askForConfirmation(confirmMessage) {
			log.Info("Delete operation cancelled")
			return nil
		}
	}

	// 执行删除操作
	log.Infof("Deleting file: %s (ID: %s)", fileName, fileId)
	err = client.DeleteFile(fileId)
	if err != nil {
		return err
	}

	log.Infof("File deleted successfully")
	return nil
}

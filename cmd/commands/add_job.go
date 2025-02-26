package commands

import (
	"fmt"
	"strings"

	"github.com/urfave/cli"

	"github.com/VillanCh/ragsync/common/aliyun"

	"github.com/yaklang/yaklang/common/log"
	"github.com/yaklang/yaklang/common/utils"
)

// AddJobCommand 将现有文件添加到知识索引的命令
func AddJobCommand() cli.Command {
	return cli.Command{
		Name:    "add-job",
		Aliases: []string{"aj"},
		Usage:   "Add existing file to knowledge index",
		Flags: []cli.Flag{
			cli.StringFlag{
				Name:  "id",
				Usage: "File ID to add to index",
			},
			cli.StringFlag{
				Name:  "name",
				Usage: "File name to search and add to index",
			},
			cli.BoolFlag{
				Name:  "force, f",
				Usage: "Force add without confirmation",
			},
		},
		Action: executeAddJob,
	}
}

// executeAddJob 将文件添加到知识索引的执行逻辑
func executeAddJob(c *cli.Context) error {
	// 从配置文件加载配置
	config, err := LoadConfig(c)
	if err != nil {
		return err
	}

	// 检查知识库索引ID是否存在
	if config.BailianKnowledgeIndexId == "" {
		return utils.Errorf("Knowledge Index ID not configured. Please update your configuration file.")
	}

	fileId := c.String("id")
	fileName := c.String("name")
	forceAdd := c.Bool("force")

	if fileId == "" && fileName == "" {
		return utils.Errorf("Please specify either file ID (--id) or file name (--name) to add to index")
	}

	client, err := aliyun.NewBailianClientFromConfig(config)
	if err != nil {
		return err
	}

	// 如果只提供了文件名，需要先查找对应的文件ID
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

			fmt.Print("\nPlease enter the number of the file to add to index (or 0 to cancel): ")
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
		// 如果只提供了ID，尝试获取文件名用于确认
		fileInfo, err := client.DescribeFile(fileId)
		if err != nil {
			log.Warnf("Failed to get file information: %v", err)
			return utils.Errorf("Cannot find file with ID: %s. Please make sure the file exists.", fileId)
		} else {
			fileName = fileInfo.FileName
		}
	}

	// 确认操作
	if !forceAdd {
		confirmMessage := fmt.Sprintf("Are you sure you want to add file '%s' (ID: %s) to knowledge index?", fileName, fileId)
		if !askForConfirmation(confirmMessage) {
			log.Info("Operation cancelled")
			return nil
		}
	}

	// 执行添加到索引的操作
	log.Infof("Adding file to knowledge index: %s", config.BailianKnowledgeIndexId)
	jobId, err := client.AppendDocumentToIndex(fileId)
	if err != nil {
		return utils.Errorf("Failed to add file to knowledge index: %v", err)
	}

	if jobId != "" {
		log.Infof("File added to knowledge index successfully. Job ID: %s", jobId)
		log.Info("You can check the job status with: ragsync job --job-id " + jobId)
	} else {
		log.Warnf("File was processed, but no job ID was returned. The file may still be added to the index.")
	}

	return nil
}

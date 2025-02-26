package commands

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/urfave/cli"

	"github.com/yaklang/yaklang/common/log"
	"github.com/yaklang/yaklang/common/utils"
)

// IndexJobsListCommand 列出索引任务命令
func IndexJobsListCommand() cli.Command {
	return cli.Command{
		Name:    "jobs",
		Aliases: []string{"lsj", "list-jobs"},
		Usage:   "List all saved index job IDs",
		Action:  executeIndexJobsList,
	}
}

// executeIndexJobsList 列出索引任务的执行逻辑
func executeIndexJobsList(c *cli.Context) error {
	// 获取用户主目录
	homeDir := utils.GetHomeDirDefault(".")

	// 构建任务目录路径
	jobsDir := filepath.Join(homeDir, ".ragsync", "index-jobs")

	// 检查目录是否存在
	if _, err := os.Stat(jobsDir); os.IsNotExist(err) {
		return utils.Errorf("No index jobs directory found at %s. No jobs have been saved yet.", jobsDir)
	}

	// 读取目录中的所有文件
	files, err := os.ReadDir(jobsDir)
	if err != nil {
		return utils.Errorf("Failed to read index jobs directory: %v", err)
	}

	// 检查是否有任务文件
	if len(files) == 0 {
		fmt.Println("No index jobs found.")
		return nil
	}

	// 输出任务列表
	fmt.Printf("\n%-40s %-25s\n", "Job ID", "Creation Time")
	fmt.Println(strings.Repeat("-", 70))

	for _, file := range files {
		if file.IsDir() {
			continue // 跳过目录
		}

		// 获取文件信息
		fileInfo, err := file.Info()
		if err != nil {
			log.Warnf("Failed to get info for file %s: %v", file.Name(), err)
			continue
		}

		// 获取创建时间
		creationTime := fileInfo.ModTime().Format(time.RFC3339)

		// 输出任务ID和创建时间
		fmt.Printf("%-40s %-25s\n", file.Name(), creationTime)
	}

	fmt.Printf("\nTotal jobs: %d\n", len(files))
	fmt.Println("To check job status, use: ragsync index-status --job-id <JOB_ID>")

	return nil
}

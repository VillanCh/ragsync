package commands

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/alibabacloud-go/tea/tea"
	"github.com/urfave/cli"

	"github.com/VillanCh/ragsync/common/aliyun"

	"github.com/yaklang/yaklang/common/log"
	"github.com/yaklang/yaklang/common/utils"
)

// IndexStatusCommand 任务状态命令
func IndexStatusCommand() cli.Command {
	return cli.Command{
		Name:    "job",
		Aliases: []string{"job-status"},
		Usage:   "Check the status of a document indexing job",
		Flags: []cli.Flag{
			cli.StringFlag{
				Name:     "job-id",
				Usage:    "Job ID to query (if not provided, will check all local jobs)",
				Required: false,
			},
			cli.BoolFlag{
				Name:  "cleanup",
				Usage: "Automatically delete local files for FINISH or DELETED jobs",
			},
		},
		Action: executeIndexStatus,
	}
}

// executeIndexStatus 查询索引任务状态的执行逻辑
func executeIndexStatus(c *cli.Context) error {
	// 从配置文件加载配置
	config, err := LoadConfig(c)
	if err != nil {
		return err
	}

	// 检查知识库索引ID是否存在
	if config.BailianKnowledgeIndexId == "" {
		return utils.Errorf("Knowledge Index ID not configured. Please update your configuration file.")
	}

	// 创建客户端
	client, err := aliyun.NewBailianClientFromConfig(config)
	if err != nil {
		return err
	}

	// 获取任务ID
	jobId := c.String("job-id")
	autoCleanup := c.Bool("cleanup")

	// 如果提供了特定的任务ID，则只检查该任务
	if jobId != "" {
		return checkSingleJobStatus(client, jobId, autoCleanup)
	}

	// 否则，检查所有本地保存的任务
	return checkAllLocalJobs(client, autoCleanup)
}

// checkSingleJobStatus 检查单个任务的状态
func checkSingleJobStatus(client *aliyun.BailianClient, jobId string, autoCleanup bool) error {
	// 查询任务状态
	log.Infof("Querying status for job: %s", jobId)
	response, err := client.GetIndexJobStatus(jobId)
	if err != nil {
		return utils.Errorf("Failed to query job status: %v", err)
	}

	// 显示任务状态
	if response != nil && response.Data != nil {
		fmt.Printf("\n--- Index Job Status ---\n")
		fmt.Printf("Job ID: %s\n", jobId)

		// 获取状态
		status := "Unknown"
		if response.Data.Status != nil {
			status = tea.StringValue(response.Data.Status)
			fmt.Printf("Status: %s\n", status)
		} else {
			fmt.Printf("Status: Unknown\n")
		}

		// 将完整数据转为 JSON 显示
		jsonData, _ := json.MarshalIndent(response.Data, "", "  ")
		fmt.Printf("\nDetailed Status Data:\n%s\n\n", string(jsonData))

		// 如果启用了自动清理并且任务状态是 FINISH 或 DELETED，删除本地文件
		if autoCleanup && (status == "FINISH" || status == "DELETED") {
			if err := removeLocalJobFile(jobId); err != nil {
				log.Warnf("Failed to remove local job file: %v", err)
			} else {
				log.Infof("Local job file for %s has been removed (status: %s)", jobId, status)
			}
		}

		return nil
	}

	return utils.Errorf("Empty or invalid response from service")
}

// checkAllLocalJobs 检查所有本地保存的任务状态
func checkAllLocalJobs(client *aliyun.BailianClient, autoCleanup bool) error {
	// 获取用户主目录
	homeDir := utils.GetHomeDirDefault(".")

	// 任务目录路径
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
		fmt.Println("No index jobs found locally.")
		return nil
	}

	fmt.Printf("\n%-40s %-15s %-25s\n", "Job ID", "Status", "Creation Time")
	fmt.Println(strings.Repeat("-", 85))

	finishedCount := 0
	errorCount := 0
	pendingCount := 0

	// 检查每个任务
	for _, file := range files {
		if file.IsDir() {
			continue // 跳过目录
		}

		jobId := file.Name()
		response, err := client.GetIndexJobStatus(jobId)

		// 获取文件信息
		fileInfo, _ := file.Info()
		creationTime := fileInfo.ModTime().Format(time.RFC3339)

		if err != nil {
			fmt.Printf("%-40s %-15s %-25s\n", jobId, "ERROR", creationTime)
			log.Warnf("Failed to query status for job %s: %v", jobId, err)
			errorCount++
			continue
		}

		// 获取并显示状态
		status := "Unknown"
		if response != nil && response.Data != nil && response.Data.Status != nil {
			status = tea.StringValue(response.Data.Status)
		}

		fmt.Printf("%-40s %-15s %-25s\n", jobId, status, creationTime)

		// 根据状态分类计数
		if status == "FINISH" || status == "DELETED" {
			finishedCount++
			// 如果启用了自动清理，删除已完成的任务文件
			if autoCleanup {
				if err := removeLocalJobFile(jobId); err != nil {
					log.Warnf("Failed to remove local job file for %s: %v", jobId, err)
				} else {
					log.Infof("Local job file for %s has been removed (status: %s)", jobId, status)
				}
			}
		} else if status == "ERROR" {
			errorCount++
		} else {
			pendingCount++
		}
	}

	// 显示统计信息
	fmt.Printf("\nTotal jobs: %d (Finished: %d, Error: %d, Pending: %d)\n", len(files), finishedCount, errorCount, pendingCount)

	if autoCleanup {
		fmt.Println("Auto cleanup enabled: Local files for FINISH and DELETED jobs have been removed.")
	} else {
		fmt.Println("To remove local files for completed jobs, run with --cleanup flag.")
	}

	return nil
}

// removeLocalJobFile 删除本地任务文件
func removeLocalJobFile(jobId string) error {
	homeDir := utils.GetHomeDirDefault(".")
	jobFilePath := filepath.Join(homeDir, ".ragsync", "index-jobs", jobId)

	// 检查文件是否存在
	if _, err := os.Stat(jobFilePath); os.IsNotExist(err) {
		return nil // 文件不存在，无需删除
	}

	// 删除文件
	return os.Remove(jobFilePath)
}

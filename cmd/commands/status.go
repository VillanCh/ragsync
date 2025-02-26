package commands

import (
	"fmt"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/urfave/cli"

	"github.com/VillanCh/ragsync/common/aliyun"

	"github.com/yaklang/yaklang/common/log"
	"github.com/yaklang/yaklang/common/utils"
)

// StatusCommand 检查文件状态命令
func StatusCommand() cli.Command {
	return cli.Command{
		Name:  "status",
		Usage: "Check file status (updates every 2 seconds)",
		Flags: []cli.Flag{
			cli.StringFlag{
				Name:  "name",
				Usage: "File name to check status",
			},
		},
		Action: executeStatus,
	}
}

// executeStatus 检查状态的执行逻辑
func executeStatus(c *cli.Context) error {
	// 从配置文件加载配置
	config, err := LoadConfig(c)
	if err != nil {
		return err
	}

	fileName := c.String("name")
	if fileName == "" {
		return utils.Errorf("Please specify the file name using --name parameter")
	}

	client, err := aliyun.NewBailianClientFromConfig(config)
	if err != nil {
		return err
	}

	// 首先列出所有文件
	log.Infof("Searching for file: %s", fileName)
	files, err := client.ListAllFiles(fileName)
	if err != nil {
		return utils.Errorf("Failed to list files: %v", err)
	}

	// 检查是否找到匹配的文件
	if len(files) == 0 {
		return utils.Errorf("No files found with name: %s", fileName)
	}

	// 如果找到多个文件，让用户选择
	var targetFile *aliyun.FileInfo
	if len(files) > 1 {
		fmt.Println("Multiple files found with this name:")
		fmt.Printf("\n%-40s %-50s %-15s\n", "File ID", "File Name", "Status")
		fmt.Println(strings.Repeat("-", 105))

		for i, file := range files {
			fmt.Printf("[%d] %-36s %-50s %-15s\n", i+1, file.FileId, file.FileName, file.Status)
		}

		fmt.Print("\nPlease enter the number of the file to monitor (or 0 to cancel): ")
		var selection int
		fmt.Scanln(&selection)

		if selection <= 0 || selection > len(files) {
			return utils.Errorf("Operation cancelled or invalid selection")
		}

		targetFile = files[selection-1]
	} else {
		targetFile = files[0]
	}

	log.Infof("Monitoring file: %s (ID: %s)", targetFile.FileName, targetFile.FileId)
	log.Info("Press Ctrl+C to stop monitoring")

	// 创建计时器，每2秒触发一次
	ticker := time.NewTicker(2 * time.Second)
	defer ticker.Stop()

	// 设置信号处理，监听Ctrl+C
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	// 显示初始状态
	fmt.Printf("\nFile ID: %s\n", targetFile.FileId)
	fmt.Printf("Name: %s\n", targetFile.FileName)
	fmt.Printf("Status: %s\n", targetFile.Status)
	fmt.Printf("Category ID: %s\n", targetFile.CategoryId)

	// 使用变量跟踪是否应该继续循环
	keepRunning := true

	// 设置goroutine监听信号
	go func() {
		<-sigChan
		keepRunning = false
		fmt.Println("\nMonitoring stopped")
	}()

	// 循环并每2秒更新一次状态
	for keepRunning {
		select {
		case <-ticker.C:
			fileInfo, err := client.DescribeFile(targetFile.FileId)
			if err != nil {
				log.Errorf("Failed to update status: %v", err)
				continue
			}

			// 清除之前的几行
			fmt.Print("\033[1A\033[2K") // 上移一行并清除该行
			fmt.Print("\033[1A\033[2K") // 再上移一行并清除该行
			fmt.Print("\033[1A\033[2K") // 再上移一行并清除该行
			fmt.Print("\033[1A\033[2K") // 再上移一行并清除该行

			// 显示更新的状态
			fmt.Printf("File ID: %s\n", fileInfo.FileId)
			fmt.Printf("Name: %s\n", fileInfo.FileName)
			fmt.Printf("Status: %s\n", fileInfo.Status)
			fmt.Printf("Category ID: %s\n", fileInfo.CategoryId)

			// 如果文件处理完成或失败，退出循环
			if fileInfo.Status == "COMPLETED" || fileInfo.Status == "FAILED" {
				log.Infof("File processing %s", fileInfo.Status)
				keepRunning = false
			}
		}
	}

	return nil
}

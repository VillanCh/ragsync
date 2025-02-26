package commands

import (
	"fmt"
	"strings"

	"github.com/urfave/cli"

	"github.com/VillanCh/ragsync/common/aliyun"

	"github.com/yaklang/yaklang/common/log"
	"github.com/yaklang/yaklang/common/utils"
)

// ListCommand 列出文件命令
func ListCommand() cli.Command {
	return cli.Command{
		Name:  "list",
		Usage: "List files in workspace",
		Flags: []cli.Flag{
			cli.StringFlag{
				Name:  "name",
				Usage: "Filter files by name (optional)",
				Value: "",
			},
		},
		Action: executeList,
	}
}

// executeList 列出文件的执行逻辑
func executeList(c *cli.Context) error {
	// 从配置文件加载配置
	config, err := LoadConfig(c)
	if err != nil {
		return err
	}

	client, err := aliyun.NewBailianClientFromConfig(config)
	if err != nil {
		return err
	}

	fileName := c.String("name")

	log.Infof("Listing files in workspace (filter: %s)...", fileName)
	files, err := client.ListAllFiles(fileName)
	if err != nil {
		return utils.Errorf("Failed to list files: %v", err)
	}

	// 输出文件列表
	fmt.Printf("\n%-40s %-50s %-15s\n", "File ID", "File Name", "Status")
	fmt.Println(strings.Repeat("-", 105))

	for _, file := range files {
		fmt.Printf("%-40s %-50s %-15s\n", file.FileId, file.FileName, file.Status)
	}

	fmt.Printf("\nTotal files: %d\n", len(files))
	return nil
}

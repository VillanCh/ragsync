package main

import (
	"os"
	"path/filepath"

	"github.com/urfave/cli"

	"github.com/VillanCh/ragsync/cmd/commands"

	"github.com/yaklang/yaklang/common/log"
	"github.com/yaklang/yaklang/common/utils"
)

func main() {
	app := cli.NewApp()
	app.Name = "ragsync"
	app.Usage = "Aliyun Bailian RAG Sync Tool"
	app.Version = "0.1.0"

	// 设置配置文件路径
	baseConfigDir := filepath.Join(utils.GetHomeDirDefault("."), ".ragsync")
	if utils.GetFirstExistedPath(baseConfigDir) == "" {
		os.MkdirAll(baseConfigDir, 0755)
	}

	defaultConfigPath := filepath.Join(baseConfigDir, "ragsync.yaml")

	// 只接受一个配置文件路径的参数
	app.Flags = []cli.Flag{
		cli.StringFlag{
			Name:  "config, c",
			Usage: "Configuration file path",
			Value: defaultConfigPath,
		},
	}

	// 设置命令
	app.Commands = commands.GetCommands()

	// 默认动作
	app.Action = func(c *cli.Context) error {
		// 尝试加载配置文件
		config, err := commands.LoadConfig(c)
		if err != nil {
			log.Warnf("Failed to load configuration: %v", err)
			log.Info("Please run 'ragsync create-config' to create a configuration file")
		} else {
			log.Infof("Configuration loaded successfully")
			log.Infof("Aliyun Access Key: %v", utils.ShrinkString(config.AliyunAccessKey, 5))
			log.Infof("Bailian Workspace ID: %v", config.BailianWorkspaceId)
		}

		log.Infof("ragsync version %s", app.Version)
		log.Info("Use 'ragsync help' to view available commands")
		log.Info("Use 'ragsync sync --file <file path>' to upload files to Bailian RAG")
		log.Info("Use 'ragsync list' to list files in workspace")
		log.Info("Use 'ragsync status --id <file id>' to monitor file processing status")

		return nil
	}

	err := app.Run(os.Args)
	if err != nil {
		log.Error(utils.Errorf("Error running ragsync: %v", err))
		os.Exit(1)
	}
}

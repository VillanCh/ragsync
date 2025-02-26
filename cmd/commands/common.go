package commands

import (
	"github.com/urfave/cli"

	"github.com/VillanCh/ragsync/common/spec"

	"github.com/yaklang/yaklang/common/log"
	"github.com/yaklang/yaklang/common/utils"
)

// LoadConfig 从命令行参数加载配置文件
func LoadConfig(c *cli.Context) (*spec.Config, error) {
	configPath := c.GlobalString("config")
	if configPath == "" {
		return nil, utils.Errorf("Configuration file path not specified")
	}

	log.Infof("Using configuration file: %s", configPath)
	config, err := spec.LoadConfig(configPath)
	if err != nil {
		return nil, utils.Errorf("Failed to load configuration file: %v", err)
	}

	if err := config.Validate(); err != nil {
		return nil, utils.Errorf("Invalid configuration: %v", err)
	}

	return config, nil
}

// GetCommands 获取所有命令
func GetCommands() []cli.Command {
	return []cli.Command{
		CreateConfigCommand(),
		SyncCommand(),
		ListCommand(),
		StatusCommand(),
		DeleteCommand(),
		ValidateConfigCommand(),
		IndexStatusCommand(),
		IndexJobsListCommand(),
		AddJobCommand(),
	}
}

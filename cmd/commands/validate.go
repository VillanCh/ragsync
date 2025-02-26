package commands

import (
	"fmt"
	"os"
	"strings"

	"github.com/urfave/cli"

	"github.com/VillanCh/ragsync/common/spec"

	"github.com/yaklang/yaklang/common/log"
)

// ValidateConfigCommand 验证配置文件命令
func ValidateConfigCommand() cli.Command {
	return cli.Command{
		Name:    "validate",
		Aliases: []string{"check"},
		Usage:   "Validate configuration file and exit with non-zero code if invalid",
		Flags:   []cli.Flag{},
		Action:  executeValidateConfig,
	}
}

// executeValidateConfig 验证配置文件的执行逻辑
func executeValidateConfig(c *cli.Context) error {
	configPath := c.GlobalString("config")
	if configPath == "" {
		log.Error("Configuration file path not specified")
		os.Exit(1)
		return nil
	}

	log.Infof("Validating configuration file: %s", configPath)
	config, err := spec.LoadConfig(configPath)
	if err != nil {
		log.Errorf("Failed to load configuration file: %v", err)
		os.Exit(1)
		return nil
	}

	if err := config.Validate(); err != nil {
		log.Errorf("Invalid configuration: %v", err)
		os.Exit(1)
		return nil
	}

	// 输出具体的配置信息
	fmt.Println("配置验证成功！配置详情：")
	fmt.Printf("Aliyun Access Key: %s\n", maskSensitiveString(config.AliyunAccessKey))
	fmt.Printf("Bailian Endpoint: %s\n", config.BailianEndpoint)
	fmt.Printf("Bailian Workspace ID: %s\n", config.BailianWorkspaceId)
	fmt.Printf("Bailian Category Type: %s\n", config.BailianCategoryType)
	fmt.Printf("Bailian File Parser: %s\n", config.BailianAddFileParser)
	fmt.Printf("Bailian Default Category ID: %s\n", config.BailianFilesDefaultCategoryId)

	log.Info("Configuration is valid")
	return nil
}

// maskSensitiveString 脱敏敏感字符串
func maskSensitiveString(s string) string {
	if len(s) < 8 {
		return "******"
	}
	visible := len(s) / 4
	return s[:visible] + strings.Repeat("*", len(s)-visible*2) + s[len(s)-visible:]
}

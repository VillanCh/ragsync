package commands

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	"github.com/urfave/cli"

	"github.com/VillanCh/ragsync/common/spec"

	"github.com/yaklang/yaklang/common/log"
	"github.com/yaklang/yaklang/common/utils"
)

// CreateConfigCommand 创建配置文件命令
func CreateConfigCommand() cli.Command {
	return cli.Command{
		Name:  "create-config",
		Usage: "Create configuration file",
		Flags: []cli.Flag{
			cli.StringFlag{
				Name:  "output, o",
				Usage: "Output path for the configuration file",
				Value: "",
			},
		},
		Action: executeCreateConfig,
	}
}

// executeCreateConfig 创建配置文件的执行逻辑
func executeCreateConfig(c *cli.Context) error {
	configPath := c.GlobalString("config")
	if configPath == "" {
		return utils.Errorf("Configuration file path not specified")
	}

	// 获取默认配置
	defaultCfg := spec.GetDefaultConfig()

	// 创建一个新的配置结构
	config := spec.Config{}

	// 创建一个带缓冲的读取器，用于读取控制台输入
	reader := bufio.NewReader(os.Stdin)

	// 用于读取单行输入的函数
	readLine := func() string {
		input, _ := reader.ReadString('\n')
		return strings.TrimSpace(input)
	}

	// 引导用户逐个输入配置项
	fmt.Println("Please enter configuration items sequentially (press Enter to use default values):")

	// 阿里云访问密钥
	fmt.Print("Aliyun Access Key (AliyunAccessKey): ")
	config.AliyunAccessKey = readLine()

	// 阿里云秘密密钥
	fmt.Print("Aliyun Secret Key (AliyunSecretKey): ")
	config.AliyunSecretKey = readLine()

	// 百炼服务端点
	fmt.Printf("Bailian Endpoint (BailianEndpoint) [%s]: ", defaultCfg.BailianEndpoint)
	endpoint := readLine()
	if endpoint != "" {
		config.BailianEndpoint = endpoint
	} else {
		config.BailianEndpoint = defaultCfg.BailianEndpoint
	}

	// 百炼分类类型
	fmt.Printf("Bailian Category Type (BailianCategoryType) [%s]: ", defaultCfg.BailianCategoryType)
	categoryType := readLine()
	if categoryType != "" {
		config.BailianCategoryType = categoryType
	} else {
		config.BailianCategoryType = defaultCfg.BailianCategoryType
	}

	// 百炼工作空间ID
	fmt.Print("Bailian Workspace ID (BailianWorkspaceId): ")
	config.BailianWorkspaceId = readLine()

	// 百炼文件解析器
	fmt.Printf("Bailian File Parser (BailianAddFileParser) [%s]: ", defaultCfg.BailianAddFileParser)
	fileParser := readLine()
	if fileParser != "" {
		config.BailianAddFileParser = fileParser
	} else {
		config.BailianAddFileParser = defaultCfg.BailianAddFileParser
	}

	// 百炼默认分类ID
	fmt.Printf("Bailian Default Category ID (BailianFilesDefaultCategoryId) [%s]: ", defaultCfg.BailianFilesDefaultCategoryId)
	categoryId := readLine()
	if categoryId != "" {
		config.BailianFilesDefaultCategoryId = categoryId
	} else {
		config.BailianFilesDefaultCategoryId = defaultCfg.BailianFilesDefaultCategoryId
	}

	// 验证配置
	if err := config.Validate(); err != nil {
		return utils.Errorf("Configuration validation failed: %v", err)
	}
	// 检查配置文件是否存在，如果存在则备份
	if utils.GetFirstExistedPath(configPath) != "" {
		backupPath := configPath + ".bak"
		if err := utils.CopyFile(configPath, backupPath); err != nil {
			log.Errorf("Failed to backup existing config file: %v", err)
		} else {
			log.Infof("Existing configuration backed up to: %s", backupPath)
		}
	}

	// 将配置保存到文件
	if err := spec.SaveConfig(&config, configPath); err != nil {
		return err
	}

	log.Infof("Configuration file has been successfully saved to: %s", configPath)
	return nil
}

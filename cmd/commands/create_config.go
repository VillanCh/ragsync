package commands

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	"github.com/urfave/cli"

	"github.com/VillanCh/ragsync/common/aliyun"
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
	if config.AliyunAccessKey == "" {
		return utils.Errorf("Aliyun Access Key is required")
	}

	// 阿里云密钥
	fmt.Print("Aliyun Secret Key (AliyunSecretKey): ")
	config.AliyunSecretKey = readLine()
	if config.AliyunSecretKey == "" {
		return utils.Errorf("Aliyun Secret Key is required")
	}

	// 百炼工作空间ID
	fmt.Print("Bailian Workspace ID (BailianWorkspaceId): ")
	config.BailianWorkspaceId = readLine()
	if config.BailianWorkspaceId == "" {
		return utils.Errorf("Bailian Workspace ID is required, plz check your workspace id in https://bailian.console.aliyun.com")
	}

	// 百炼知识库索引ID
	fmt.Println("\nAvailable indices:")
	indices, err := aliyun.ListIndices(config.AliyunAccessKey, config.AliyunSecretKey, config.BailianWorkspaceId)
	if err != nil {
		log.Warnf("Failed to list indices: %v", err)
		fmt.Println("Warning: Could not fetch existing indices. You can still proceed with manual input.")
	} else {
		fmt.Printf("%-40s %-50s\n", "Index ID", "Index Name")
		fmt.Println(strings.Repeat("-", 90))
		for _, idx := range indices {
			fmt.Printf("%-40s %-50s\n", idx.IndexId, idx.IndexName)
		}
		fmt.Println()
	}

	fmt.Print("Bailian Knowledge Index ID (BailianKnowledgeIndexId): ")
	indexId := readLine()

	if indexId == "" {
		fmt.Print("\nNo index ID provided. Would you like to create a new index? (y/n): ")
		createNew := readLine()
		if strings.ToLower(createNew) == "y" {
			fmt.Print("Enter new index name: ")
			indexName := readLine()
			if indexName == "" {
				return utils.Errorf("Index name cannot be empty")
			}
			err := aliyun.CreateIndex(config.AliyunAccessKey, config.AliyunSecretKey, config.BailianWorkspaceId, indexName)
			if err != nil {
				return utils.Errorf("Failed to create index: %v", err)
			}
			fmt.Println("Index created successfully. Please run the command again to select the new index.")
			return nil
		} else {
			return utils.Errorf("Bailian Knowledge Index ID is required, plz check your knowledge index id in https://bailian.console.aliyun.com")
		}
	} else {
		// 验证索引ID是否存在
		if err == nil {
			found := false
			for _, idx := range indices {
				if idx.IndexId == indexId {
					found = true
					break
				}
			}
			if !found {
				fmt.Printf("\nWarning: Index ID '%s' not found in the list of available indices.\n", indexId)
				fmt.Print("Do you want to proceed anyway? (y/n): ")
				confirm := readLine()
				if strings.ToLower(confirm) != "y" {
					return utils.Errorf("Please provide a valid index ID")
				}
			}
		}
		config.BailianKnowledgeIndexId = indexId
	}

	// 百炼默认分类ID
	fmt.Println("\nAvailable categories:")
	categories, err := aliyun.ListCategories(config.AliyunAccessKey, config.AliyunSecretKey, config.BailianWorkspaceId)
	if err != nil {
		log.Warnf("Failed to list categories: %v", err)
		fmt.Println("Warning: Could not fetch existing categories. You can still proceed with default category.")
	} else {
		fmt.Printf("%-40s %-50s\n", "Category ID", "Category Name")
		fmt.Println(strings.Repeat("-", 90))
		for _, cat := range categories {
			fmt.Printf("%-40s %-50s\n", cat.CategoryId, cat.CategoryName)
		}
		fmt.Println()
	}

	fmt.Printf("Bailian Default Category ID (BailianFilesDefaultCategoryId) [%s]: ", defaultCfg.BailianFilesDefaultCategoryId)
	categoryId := readLine()

	if categoryId == "" {
		fmt.Print("\nNo category ID provided. Would you like to create a new category? (y/n): ")
		createNew := readLine()
		if strings.ToLower(createNew) == "y" {
			fmt.Print("Enter new category name: ")
			categoryName := readLine()
			if categoryName == "" {
				return utils.Errorf("Category name cannot be empty")
			}
			err := aliyun.CreateCategory(config.AliyunAccessKey, config.AliyunSecretKey, config.BailianWorkspaceId, categoryName)
			if err != nil {
				return utils.Errorf("Failed to create category: %v", err)
			}
			fmt.Println("Category created successfully. Please run the command again to select the new category.")
			return nil
		} else {
			fmt.Print("\nWarning: Using default category ID. This may not be suitable for all use cases.\nAre you sure you want to use the default category? (y/n): ")
			confirm := readLine()
			if strings.ToLower(confirm) != "y" {
				return utils.Errorf("Please provide a valid category ID or create a new one")
			}
			config.BailianFilesDefaultCategoryId = defaultCfg.BailianFilesDefaultCategoryId
		}
	} else {
		// 验证分类ID是否存在
		if err == nil {
			found := false
			for _, cat := range categories {
				if cat.CategoryId == categoryId {
					found = true
					break
				}
			}
			if !found {
				fmt.Printf("\nWarning: Category ID '%s' not found in the list of available categories.\n", categoryId)
				fmt.Print("Do you want to proceed anyway? (y/n): ")
				confirm := readLine()
				if strings.ToLower(confirm) != "y" {
					return utils.Errorf("Please provide a valid category ID")
				}
			}
		}
		config.BailianFilesDefaultCategoryId = categoryId
	}

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

	// 百炼文件解析器
	fmt.Printf("Bailian File Parser (BailianAddFileParser) [%s]: ", defaultCfg.BailianAddFileParser)
	fileParser := readLine()
	if fileParser != "" {
		config.BailianAddFileParser = fileParser
	} else {
		config.BailianAddFileParser = defaultCfg.BailianAddFileParser
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

	// 打印配置并确认
	fmt.Println("\nPlease review your configuration:")
	fmt.Println("----------------------------------------")
	fmt.Printf("Aliyun Access Key: %s\n", config.AliyunAccessKey)
	fmt.Printf("Aliyun Secret Key: %s\n", strings.Repeat("*", len(config.AliyunSecretKey)))
	fmt.Printf("Bailian Workspace ID: %s\n", config.BailianWorkspaceId)
	fmt.Printf("Bailian Knowledge Index ID: %s\n", config.BailianKnowledgeIndexId)
	fmt.Printf("Bailian Endpoint: %s\n", config.BailianEndpoint)
	fmt.Printf("Bailian Category Type: %s\n", config.BailianCategoryType)
	fmt.Printf("Bailian File Parser: %s\n", config.BailianAddFileParser)
	fmt.Printf("Bailian Default Category ID: %s\n", config.BailianFilesDefaultCategoryId)
	fmt.Println("----------------------------------------")
	fmt.Print("\nDo you want to save this configuration? (y/n): ")
	confirm := readLine()
	if strings.ToLower(confirm) != "y" {
		return utils.Errorf("Configuration not saved")
	}

	// 将配置保存到文件
	if err := spec.SaveConfig(&config, configPath); err != nil {
		return err
	}

	log.Infof("Configuration file has been successfully saved to: %s", configPath)
	return nil
}

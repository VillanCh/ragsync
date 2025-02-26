package main

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/urfave/cli"

	"github.com/VillanCh/ragsync/common/aliyun"
	"github.com/VillanCh/ragsync/common/spec"
	"github.com/davecgh/go-spew/spew"

	"github.com/yaklang/yaklang/common/log"
	"github.com/yaklang/yaklang/common/utils"
)

func main() {
	app := cli.NewApp()
	app.Name = "ragsync"
	app.Usage = "Aliyun Bailian RAG Sync Tool"
	app.Version = "0.1.0"

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

	// 加载配置的函数
	loadConfig := func(c *cli.Context) (*spec.Config, error) {
		configPath := c.GlobalString("config")
		if configPath == "" {
			configPath = defaultConfigPath
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

	app.Commands = []cli.Command{
		{
			Name:  "create-config",
			Usage: "Create configuration file",
			Flags: []cli.Flag{
				cli.StringFlag{
					Name:  "output, o",
					Usage: "Output path for the configuration file",
					Value: "",
				},
			},
			Action: func(c *cli.Context) error {
				configPath := c.GlobalString("config")
				if configPath == "" {
					configPath = defaultConfigPath
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
			},
		},
		{
			Name:  "apply-lease",
			Usage: "Apply for file upload lease",
			Flags: []cli.Flag{
				cli.StringFlag{
					Name:  "file",
					Usage: "File path to upload",
				},
			},
			Action: func(c *cli.Context) error {
				// 从配置文件加载配置
				config, err := loadConfig(c)
				if err != nil {
					return err
				}

				client, err := aliyun.NewBailianClient(
					config.AliyunAccessKey,
					config.AliyunSecretKey,
					config.BailianEndpoint,
				)
				if err != nil {
					return err
				}

				client.SetWorkspaceId(config.BailianWorkspaceId)

				lis, err := client.ApplyFileUploadLease("test.txt", []byte("test"))
				if err != nil {
					return err
				}
				spew.Dump(lis)

				headers := utils.InterfaceToGeneralMap(lis.Headers)
				bailianExtra, ok := headers["X-bailian-extra"]
				if !ok {
					return utils.Errorf("X-bailian-extra does not exist")
				}
				contentType, ok := headers["Content-Type"]
				if !ok {
					return utils.Errorf("Content-Type does not exist")
				}

				// Upload file
				content := []byte("test")
				err = aliyun.UploadFile(lis.Method, lis.UploadURL, "test.txt", fmt.Sprint(contentType), content, fmt.Sprintf("%s", bailianExtra))
				if err != nil {
					return err
				}

				log.Info("Adding file to Bailian RAG")
				err = client.AddFile(lis.LeaseId)
				if err != nil {
					return err
				}

				log.Info("File added successfully")
				return nil
			},
		},
		{
			Name:  "upload",
			Usage: "Upload file to Bailian RAG",
			Flags: []cli.Flag{
				cli.StringFlag{
					Name:  "file",
					Usage: "File path to upload",
					Value: "",
				},
				cli.StringFlag{
					Name:  "agent-name",
					Usage: "Agent name for the file",
					Value: "",
				},
			},
			Action: func(c *cli.Context) error {
				// 从配置文件加载配置
				config, err := loadConfig(c)
				if err != nil {
					return err
				}

				agentName := c.String("agent-name")
				filePath := c.String("file")
				if filePath == "" {
					return utils.Errorf("Please specify the file path to upload")
				}

				// Create Bailian client
				client, err := aliyun.NewBailianClient(
					config.AliyunAccessKey,
					config.AliyunSecretKey,
					config.BailianEndpoint,
				)
				if err != nil {
					return err
				}

				client.SetWorkspaceId(config.BailianWorkspaceId)

				// Apply for file upload lease
				lease, err := client.ApplyFileUploadLease(agentName, []byte("test"))
				if err != nil {
					return err
				}
				log.Infof("Successfully obtained upload lease, lease ID: %s", lease.LeaseId)

				// headers := utils.InterfaceToGeneralMap(lease.Headers)
				// bailianExtra, ok := headers["X-bailian-extra"]
				// if !ok {
				// 	return utils.Errorf("X-bailian-extra does not exist")
				// }

				// // Upload file
				// content := []byte("test")
				// err = aliyun.UploadFile(lease.Method, lease.UploadURL, filePath, content, fmt.Sprintf("%s", bailianExtra))
				// if err != nil {
				// 	return err
				// }
				return nil
			},
		},
	}

	app.Action = func(c *cli.Context) error {
		// 尝试加载配置文件
		config, err := loadConfig(c)
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
		log.Info("Use 'ragsync upload --file <file path>' to upload files to Bailian RAG")

		return nil
	}

	err := app.Run(os.Args)
	if err != nil {
		log.Error(utils.Errorf("Error running ragsync: %v", err))
		os.Exit(1)
	}
}

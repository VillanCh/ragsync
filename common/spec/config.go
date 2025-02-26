package spec

import (
	"io/ioutil"
	"os"

	"github.com/yaklang/yaklang/common/log"
	"github.com/yaklang/yaklang/common/utils"
	"gopkg.in/yaml.v3"
)

type Config struct {
	AliyunAccessKey string `yaml:"aliyun_access_key"`
	AliyunSecretKey string `yaml:"aliyun_secret_key"`

	BailianEndpoint               string `yaml:"aliyun_bailian_endpoint"`           // bailian.cn-beijing.aliyuncs.com
	BailianCategoryType           string `yaml:"bailian_category_type"`             // UNSTRUCTURED
	BailianWorkspaceId            string `yaml:"bailian_workspace_id"`              // fetch from bailian.console.aliyun.com
	BailianAddFileParser          string `yaml:"bailian_add_file_parser"`           // DASHSCOPE_DOCMIND
	BailianFilesDefaultCategoryId string `yaml:"bailian_files_default_category_id"` // default
	BailianKnowledgeIndexId       string `yaml:"bailian_knowledge_index_id"`        // knowledge index id for RAG
}

// 默认配置值
var defaultConfig = Config{
	BailianEndpoint:               "bailian.cn-beijing.aliyuncs.com",
	BailianCategoryType:           "UNSTRUCTURED",
	BailianAddFileParser:          "DASHSCOPE_DOCMIND",
	BailianFilesDefaultCategoryId: "default",
}

// Validate 验证配置是否有效
func (c *Config) Validate() error {
	if c.BailianEndpoint == "" {
		return utils.Errorf("Bailian endpoint (BailianEndpoint) cannot be empty")
	}
	if c.BailianCategoryType == "" {
		return utils.Errorf("Bailian category type (BailianCategoryType) cannot be empty")
	}
	if c.BailianWorkspaceId == "" {
		return utils.Errorf("Bailian workspace ID (BailianWorkspaceId) cannot be empty")
	}
	if c.BailianAddFileParser == "" {
		return utils.Errorf("Bailian file parser (BailianAddFileParser) cannot be empty")
	}
	if c.BailianFilesDefaultCategoryId == "" {
		return utils.Errorf("Bailian default category ID (BailianFilesDefaultCategoryId) cannot be empty")
	}
	return nil
}

// GetDefaultConfig 返回默认配置
func GetDefaultConfig() Config {
	return defaultConfig
}

// LoadConfig 从YAML文件加载配置
func LoadConfig(configPath string) (*Config, error) {
	// 使用默认配置作为基础
	config := defaultConfig

	// 检查文件是否存在
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		log.Warnf("Configuration file %s does not exist, using default configuration", configPath)
		return &config, nil
	}

	// 读取配置文件
	yamlFile, err := os.ReadFile(configPath)
	if err != nil {
		return nil, utils.Errorf("Failed to read configuration file: %v", err)
	}

	// 解析YAML
	err = yaml.Unmarshal(yamlFile, &config)
	if err != nil {
		return nil, utils.Errorf("Failed to parse YAML configuration: %v", err)
	}

	// 检查必要的配置项
	if config.AliyunAccessKey == "" || config.AliyunSecretKey == "" {
		log.Warnf("Aliyun access key not set in configuration file, will try to get from environment variables")
		return nil, utils.Errorf("Aliyun access key not set in configuration file, will try to get from environment variables")
	}

	if config.BailianWorkspaceId == "" {
		log.Warnf("Bailian workspace ID not set, please specify in configuration file or through environment variables")
	}

	if config.BailianKnowledgeIndexId == "" {
		log.Warnf("Bailian knowledge index ID not set, please specify in configuration file or through environment variables")
	}

	// 验证配置
	if err := config.Validate(); err != nil {
		return nil, err
	}

	return &config, nil
}

// SaveConfig 将配置保存到YAML文件
func SaveConfig(config *Config, configPath string) error {
	yamlData, err := yaml.Marshal(config)
	if err != nil {
		return utils.Errorf("Failed to serialize configuration: %v", err)
	}

	err = ioutil.WriteFile(configPath, yamlData, 0644)
	if err != nil {
		return utils.Errorf("Failed to write configuration file: %v", err)
	}

	return nil
}

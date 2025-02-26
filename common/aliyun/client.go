package aliyun

import (
	"os"

	"github.com/VillanCh/ragsync/common/spec"
	bailian20231229 "github.com/alibabacloud-go/bailian-20231229/v2/client"
	openapi "github.com/alibabacloud-go/darabonba-openapi/v2/client"
	"github.com/alibabacloud-go/tea/tea"
	"github.com/yaklang/yaklang/common/utils"
)

// BailianClient 百炼 RAG 客户端
type BailianClient struct {
	config *spec.Config
	Client *bailian20231229.Client
}

// NewBailianClientFromConfig 从配置创建新的百炼客户端
func NewBailianClientFromConfig(config *spec.Config) (*BailianClient, error) {
	if config == nil {
		return nil, utils.Error("Configuration cannot be nil")
	}

	// 获取访问密钥
	accessKeyID := config.AliyunAccessKey
	accessKeySecret := config.AliyunSecretKey
	endpoint := config.BailianEndpoint

	// 如果未提供 accessKeyID 或 accessKeySecret，则从环境变量获取
	if accessKeyID == "" {
		accessKeyID = os.Getenv("ALIBABA_CLOUD_ACCESS_KEY_ID")
	}
	if accessKeySecret == "" {
		accessKeySecret = os.Getenv("ALIBABA_CLOUD_ACCESS_KEY_SECRET")
	}
	// 如果未提供 endpoint，则使用默认值
	if endpoint == "" {
		endpoint = "bailian.cn-beijing.aliyuncs.com"
	}

	// 确保其他配置字段有值
	defaultCfg := spec.GetDefaultConfig()
	if config.BailianCategoryType == "" {
		config.BailianCategoryType = defaultCfg.BailianCategoryType
	}
	if config.BailianAddFileParser == "" {
		config.BailianAddFileParser = defaultCfg.BailianAddFileParser
	}
	if config.BailianFilesDefaultCategoryId == "" {
		config.BailianFilesDefaultCategoryId = defaultCfg.BailianFilesDefaultCategoryId
	}

	// 创建 OpenAPI 配置
	openapiConfig := &openapi.Config{
		AccessKeyId:     tea.String(accessKeyID),
		AccessKeySecret: tea.String(accessKeySecret),
		Endpoint:        tea.String(endpoint),
	}

	client, err := bailian20231229.NewClient(openapiConfig)
	if err != nil {
		return nil, utils.Errorf("Failed to create Bailian client: %v", err)
	}

	// 更新配置中的访问密钥（如果是从环境变量中获取的）
	if config.AliyunAccessKey == "" {
		config.AliyunAccessKey = accessKeyID
	}
	if config.AliyunSecretKey == "" {
		config.AliyunSecretKey = accessKeySecret
	}
	if config.BailianEndpoint == "" {
		config.BailianEndpoint = endpoint
	}

	return &BailianClient{
		Client: client,
		config: config,
	}, nil
}

// UpdateConfig 更新客户端配置
func (client *BailianClient) UpdateConfig(config *spec.Config) {
	if config != nil {
		client.config = config
	}
}

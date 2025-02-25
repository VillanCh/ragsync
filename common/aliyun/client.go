package aliyun

import (
	"os"

	bailian20231229 "github.com/alibabacloud-go/bailian-20231229/v2/client"
	openapi "github.com/alibabacloud-go/darabonba-openapi/v2/client"
	"github.com/alibabacloud-go/tea/tea"
	"github.com/yaklang/yaklang/common/utils"
)

// BailianClient 百炼 RAG 客户端
type BailianClient struct {
	Client      *bailian20231229.Client
	workspaceId string
}

func (c *BailianClient) SetWorkspaceId(workspaceId string) {
	c.workspaceId = workspaceId
}

// NewBailianClient 创建新的百炼客户端
func NewBailianClient(accessKeyID, accessKeySecret, endpoint string) (*BailianClient, error) {
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

	config := &openapi.Config{
		AccessKeyId:     tea.String(accessKeyID),
		AccessKeySecret: tea.String(accessKeySecret),
		Endpoint:        tea.String(endpoint),
	}

	client, err := bailian20231229.NewClient(config)
	if err != nil {
		return nil, utils.Errorf("创建百炼客户端失败: %v", err)
	}

	return &BailianClient{
		Client: client,
	}, nil
}

// This file is auto-generated, don't edit it. Thanks.
package aliyun

import (
	"encoding/json"
	"strings"

	util "github.com/alibabacloud-go/tea-utils/v2/service"
	"github.com/alibabacloud-go/tea/tea"
	"github.com/yaklang/yaklang/common/log"
	"github.com/yaklang/yaklang/common/utils"
)

// FileInfo 文件信息结构体
type FileInfo struct {
	FileId     string `json:"fileId"`
	FileName   string `json:"fileName"`
	Status     string `json:"status"`
	CategoryId string `json:"categoryId"`
	CreateTime string `json:"createTime"` // 文件创建时间
	Raw        any    `json:"raw"`
}

// DescribeFile 查询文件信息
func (client *BailianClient) DescribeFile(fileId string) (*FileInfo, error) {
	if client.config == nil {
		return nil, utils.Error("Client configuration is not set")
	}

	if client.config.BailianWorkspaceId == "" {
		return nil, utils.Error("Workspace ID is not set")
	}

	if fileId == "" {
		return nil, utils.Error("File ID cannot be empty")
	}

	runtime := &util.RuntimeOptions{}
	headers := make(map[string]*string)

	// 调用 API
	response, err := client.Client.DescribeFileWithOptions(
		tea.String(client.config.BailianWorkspaceId),
		tea.String(fileId),
		headers,
		runtime,
	)

	if err != nil {
		var sdkErr *tea.SDKError
		if teaErr, ok := err.(*tea.SDKError); ok {
			sdkErr = teaErr
			// 尝试解析更详细的错误信息
			if sdkErr.Data != nil {
				var data interface{}
				decoder := json.NewDecoder(strings.NewReader(tea.StringValue(sdkErr.Data)))
				if decodeErr := decoder.Decode(&data); decodeErr == nil {
					if m, ok := data.(map[string]interface{}); ok {
						recommend, ok := m["Recommend"]
						if ok {
							log.Errorf("Detailed error information: %v", recommend)
						}
					}
				}
			}
		}
		return nil, utils.Errorf("Failed to describe file: %v", err)
	}

	// 解析响应
	if response == nil || response.Body == nil {
		return nil, utils.Errorf("Describe file response is empty")
	}

	// 检查响应是否成功
	if response.Body.Success == nil || !*response.Body.Success {
		errorMsg := "Unknown error"
		if response.Body.Message != nil {
			errorMsg = *response.Body.Message
		}
		return nil, utils.Errorf("Failed to describe file: %v", errorMsg)
	}

	// 构造文件信息结构体
	fileInfo := &FileInfo{
		Raw:        response.Body,
		FileId:     tea.StringValue(response.Body.Data.FileId),
		FileName:   tea.StringValue(response.Body.Data.FileName),
		Status:     tea.StringValue(response.Body.Data.Status),
		CategoryId: tea.StringValue(response.Body.Data.CategoryId),
		CreateTime: tea.StringValue(response.Body.Data.CreateTime),
	}

	log.Infof("File information retrieved successfully, file ID: %s, name: %s", fileInfo.FileId, fileInfo.FileName)
	return fileInfo, nil
}

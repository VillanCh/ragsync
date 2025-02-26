package aliyun

import (
	"encoding/json"
	"strings"

	bailian20231229 "github.com/alibabacloud-go/bailian-20231229/v2/client"
	util "github.com/alibabacloud-go/tea-utils/v2/service"
	"github.com/alibabacloud-go/tea/tea"
	"github.com/yaklang/yaklang/common/log"
	"github.com/yaklang/yaklang/common/utils"
)

// AddFile 将已上传的文件添加到百炼服务
func (client *BailianClient) AddFile(leaseId string) (string, error) {
	if client.config == nil {
		return "", utils.Error("Client configuration is not set")
	}

	if client.config.BailianWorkspaceId == "" {
		return "", utils.Error("Workspace ID is not set")
	}

	if leaseId == "" {
		return "", utils.Error("Lease ID cannot be empty")
	}

	// 创建添加文件请求
	addFileRequest := &bailian20231229.AddFileRequest{
		LeaseId:      tea.String(leaseId),
		Parser:       tea.String(client.config.BailianAddFileParser),
		CategoryId:   tea.String(client.config.BailianFilesDefaultCategoryId),
		CategoryType: tea.String(client.config.BailianCategoryType),
	}

	runtime := &util.RuntimeOptions{}
	headers := make(map[string]*string)

	// 调用API
	response, err := client.Client.AddFileWithOptions(
		tea.String(client.config.BailianWorkspaceId),
		addFileRequest,
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
		return "", utils.Errorf("Failed to add file: %v", err)
	}

	// 解析响应
	if response == nil || response.Body == nil {
		return "", utils.Errorf("Add file response is empty")
	}

	if tea.StringValue(response.Body.Success) != "true" {
		return "", utils.Errorf("Failed to add file: %v", response.Body.Message)
	}

	fileId := tea.StringValue(response.Body.Data.FileId)
	log.Infof("File added successfully, file ID: %s", fileId)
	return fileId, nil
}

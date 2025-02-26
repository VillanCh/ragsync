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

// DeleteFile 删除指定ID的文件
func (client *BailianClient) DeleteFile(fileId string) error {
	if client.config == nil {
		return utils.Error("Client configuration is not set")
	}

	if client.config.BailianWorkspaceId == "" {
		return utils.Error("Workspace ID is not set")
	}

	if fileId == "" {
		return utils.Error("File ID cannot be empty")
	}

	runtime := &util.RuntimeOptions{}
	headers := make(map[string]*string)

	log.Infof("Deleting file with ID: %#v in workspace: %#v", fileId, client.config.BailianWorkspaceId)

	// 调用API删除文件
	response, err := client.Client.DeleteFileWithOptions(
		tea.String(fileId),
		tea.String(client.config.BailianWorkspaceId),
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
		return utils.Errorf("Failed to delete file: %v", err)
	}

	// 验证响应
	if response == nil || response.Body == nil {
		return utils.Errorf("Delete file response is empty")
	}

	// 检查响应是否成功
	if response.Body.Success == nil || !*response.Body.Success {
		errorMsg := "Unknown error"
		if response.Body.Message != nil {
			errorMsg = *response.Body.Message
		}
		return utils.Errorf("Failed to delete file: %v", errorMsg)
	}

	log.Infof("File deleted successfully: %s", fileId)
	return nil
}

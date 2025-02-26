// This file is auto-generated, don't edit it. Thanks.
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

// DeleteIndexDocument 从知识库索引中删除指定文档
func (client *BailianClient) DeleteIndexDocument(documentId string) error {
	if client.config == nil {
		return utils.Error("Client configuration is not set")
	}

	if client.config.BailianWorkspaceId == "" {
		return utils.Error("Workspace ID is not set")
	}

	if client.config.BailianKnowledgeIndexId == "" {
		return utils.Error("Knowledge Index ID is not set")
	}

	if documentId == "" {
		return utils.Error("Document ID cannot be empty")
	}

	// 创建请求
	deleteIndexDocumentRequest := &bailian20231229.DeleteIndexDocumentRequest{
		IndexId:     tea.String(client.config.BailianKnowledgeIndexId),
		DocumentIds: []*string{tea.String(documentId)},
	}

	runtime := &util.RuntimeOptions{}
	headers := make(map[string]*string)

	log.Infof("Deleting document with ID: %s from index: %s", documentId, client.config.BailianKnowledgeIndexId)

	// 使用try-catch结构来处理可能的异常
	var response *bailian20231229.DeleteIndexDocumentResponse
	var err error

	tryErr := func() error {
		defer func() {
			if r := tea.Recover(recover()); r != nil {
				err = utils.Errorf("Panic occurred: %v", r)
			}
		}()

		// 发送请求
		response, err = client.Client.DeleteIndexDocumentWithOptions(
			tea.String(client.config.BailianWorkspaceId),
			deleteIndexDocumentRequest,
			headers,
			runtime,
		)

		if err != nil {
			return err
		}

		return nil
	}()

	if tryErr != nil {
		var sdkErr *tea.SDKError
		if teaErr, ok := tryErr.(*tea.SDKError); ok {
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
		return utils.Errorf("Failed to delete document from index: %v", tryErr)
	}

	// 验证响应
	if response == nil || response.Body == nil {
		return utils.Errorf("Delete index document response is empty")
	}

	// 检查响应是否成功
	if response.Body.Success == nil || !*response.Body.Success {
		errorMsg := "Unknown error"
		if response.Body.Message != nil {
			errorMsg = *response.Body.Message
		}
		return utils.Errorf("Failed to delete document from index: %v", errorMsg)
	}

	log.Infof("Document deleted successfully from index: %s", documentId)
	return nil
}

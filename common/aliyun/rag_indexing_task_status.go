// This file is auto-generated, don't edit it. Thanks.
package aliyun

import (
	"encoding/json"
	"fmt"
	"strings"

	bailian20231229 "github.com/alibabacloud-go/bailian-20231229/v2/client"
	util "github.com/alibabacloud-go/tea-utils/v2/service"
	"github.com/alibabacloud-go/tea/tea"
	"github.com/yaklang/yaklang/common/log"
	"github.com/yaklang/yaklang/common/utils"
)

// GetIndexJobStatus 获取索引任务状态
func (client *BailianClient) GetIndexJobStatus(jobId string) (*bailian20231229.GetIndexJobStatusResponseBody, error) {
	if client.config.BailianKnowledgeIndexId == "" {
		return nil, utils.Errorf("Bailian knowledge index ID (BailianKnowledgeIndexId) is not configured")
	}

	if jobId == "" {
		return nil, utils.Errorf("Job ID cannot be empty")
	}

	// 创建请求
	getIndexJobStatusRequest := &bailian20231229.GetIndexJobStatusRequest{
		JobId:   tea.String(jobId),
		IndexId: tea.String(client.config.BailianKnowledgeIndexId),
	}

	// 运行时选项和请求头
	runtime := &util.RuntimeOptions{}
	headers := make(map[string]*string)

	// 记录请求日志
	log.Infof("Querying job status for job ID: %s, index ID: %s", jobId, client.config.BailianKnowledgeIndexId)

	var response *bailian20231229.GetIndexJobStatusResponse
	var err error

	// 使用 try-catch 结构处理可能的错误
	tryErr := func() error {
		defer func() {
			if r := tea.Recover(recover()); r != nil {
				err = fmt.Errorf("panic recovered: %v", r)
			}
		}()

		// 发送请求
		response, err = client.Client.GetIndexJobStatusWithOptions(
			tea.String(client.config.BailianWorkspaceId),
			getIndexJobStatusRequest,
			headers,
			runtime,
		)

		if err != nil {
			return err
		}

		return nil
	}()

	// 处理错误
	if tryErr != nil {
		var sdkErr = &tea.SDKError{}
		if t, ok := tryErr.(*tea.SDKError); ok {
			sdkErr = t
		} else {
			sdkErr.Message = tea.String(tryErr.Error())
		}

		// 记录错误信息
		log.Errorf("Failed to get job status: %s", tea.StringValue(sdkErr.Message))

		// 处理诊断信息
		if sdkErr.Data != nil {
			var data interface{}
			d := json.NewDecoder(strings.NewReader(tea.StringValue(sdkErr.Data)))
			if d.Decode(&data) == nil {
				if m, ok := data.(map[string]interface{}); ok {
					if recommend, ok := m["Recommend"]; ok {
						log.Errorf("Error recommendation: %v", recommend)
					}
				}
			}
		}

		return nil, utils.Errorf("Failed to get job status: %v", tryErr)
	}

	// 处理响应
	if response != nil && response.Body != nil {
		log.Infof("Job status query successful, request ID: %s", tea.StringValue(response.Body.RequestId))
		return response.Body, nil
	}

	return nil, utils.Errorf("Empty response from job status query")
}

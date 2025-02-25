package aliyun

import (
	"encoding/json"
	"path/filepath"
	"strconv"
	"strings"

	bailian20231229 "github.com/alibabacloud-go/bailian-20231229/v2/client"
	util "github.com/alibabacloud-go/tea-utils/v2/service"
	"github.com/alibabacloud-go/tea/tea"
	"github.com/yaklang/yaklang/common/log"
	"github.com/yaklang/yaklang/common/utils"
)

// FileUploadLease 文件上传租约信息
type FileUploadLease struct {
	LeaseId   string `json:"leaseId"`
	UploadURL string `json:"uploadUrl"`
	Method    string `json:"method"`
	Headers   any    `json:"headers"`
	Raw       any    `json:"raw"`
}

// ApplyFileUploadLease 申请文件上传租约
func (client *BailianClient) ApplyFileUploadLease(fileName string, content []byte) (*FileUploadLease, error) {
	if client.workspaceId == "" {
		return nil, utils.Error("workspaceId 未设置")
	}

	ext := filepath.Ext(fileName)
	if ext == "" {
		return nil, utils.Error("文件扩展名不能为空")
	}

	md5Str := utils.CalcMd5(string(content))
	request := &bailian20231229.ApplyFileUploadLeaseRequest{
		CategoryType: tea.String("UNSTRUCTURED"),
		FileName:     tea.String(fileName),
		Md5:          tea.String(md5Str),
		SizeInBytes:  tea.String(strconv.Itoa(len(content))),
	}
	runtime := &util.RuntimeOptions{}
	headers := make(map[string]*string)

	// 调用 API
	response, err := client.Client.ApplyFileUploadLeaseWithOptions(
		tea.String("default"),
		tea.String(client.workspaceId),
		request,
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
							log.Errorf("详细错误信息: %v", recommend)
						}
					}
				}
			}
		}
		return nil, utils.Errorf("申请文件上传租约失败: %v", err)
	}

	// 解析响应
	if response == nil || response.Body == nil {
		return nil, utils.Errorf("申请文件上传租约响应为空")
	}

	if !tea.BoolValue(response.Body.Success) {
		return nil, utils.Errorf("申请文件上传租约失败: %v", response.Body.Message)
	}

	lease := &FileUploadLease{
		Raw:       response.Body,
		LeaseId:   tea.StringValue(response.Body.Data.FileUploadLeaseId),
		UploadURL: tea.StringValue(response.Body.Data.Param.Url),
		Method:    tea.StringValue(response.Body.Data.Param.Method),
		Headers:   response.Body.Data.Param.Headers,
	}

	return lease, nil
}

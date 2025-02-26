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

// ListFilesResult 列出文件的结果
type ListFilesResult struct {
	Files     []*FileInfo `json:"files"`
	NextToken string      `json:"nextToken"`
	Raw       any         `json:"raw"`
}

// ListFile 列出工作空间下的文件
func (client *BailianClient) ListFile(maxResults int32, nextToken string, fileName string) (*ListFilesResult, error) {
	if client.config == nil {
		return nil, utils.Error("Client configuration is not set")
	}

	if client.config.BailianWorkspaceId == "" {
		return nil, utils.Error("Workspace ID is not set")
	}

	// 使用配置中的默认分类ID
	categoryId := client.config.BailianFilesDefaultCategoryId

	// 创建请求
	listFileRequest := &bailian20231229.ListFileRequest{
		CategoryId: tea.String(categoryId),
		FileName:   tea.String(fileName),
		NextToken:  tea.String(nextToken),
		MaxResults: tea.Int32(maxResults),
	}

	runtime := &util.RuntimeOptions{}
	headers := make(map[string]*string)

	// 调用 API
	response, err := client.Client.ListFileWithOptions(
		tea.String(client.config.BailianWorkspaceId),
		listFileRequest,
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
		return nil, utils.Errorf("Failed to list files: %v", err)
	}

	// 解析响应
	if response == nil || response.Body == nil {
		return nil, utils.Errorf("List files response is empty")
	}

	// 检查响应是否成功
	if response.Body.Success == nil || !*response.Body.Success {
		errorMsg := "Unknown error"
		if response.Body.Message != nil {
			errorMsg = *response.Body.Message
		}
		return nil, utils.Errorf("Failed to list files: %v", errorMsg)
	}

	// 构造结果
	result := &ListFilesResult{
		Files:     make([]*FileInfo, 0),
		NextToken: tea.StringValue(response.Body.Data.NextToken),
		Raw:       response.Body,
	}

	// 处理文件列表
	if response.Body.Data.FileList != nil {
		for _, file := range response.Body.Data.FileList {
			fileInfo := &FileInfo{
				FileId:     tea.StringValue(file.FileId),
				FileName:   tea.StringValue(file.FileName),
				Status:     tea.StringValue(file.Status),
				CategoryId: tea.StringValue(file.CategoryId),
			}
			result.Files = append(result.Files, fileInfo)
		}
	}

	log.Infof("Retrieved %d files", len(result.Files))
	return result, nil
}

// ListAllFiles 列出所有文件（自动处理分页）
func (client *BailianClient) ListAllFiles(fileName string) ([]*FileInfo, error) {
	allFiles := make([]*FileInfo, 0)
	nextToken := ""

	for {
		result, err := client.ListFile(100, nextToken, fileName)
		if err != nil {
			return nil, err
		}

		allFiles = append(allFiles, result.Files...)

		// 检查是否有更多页
		if result.NextToken == "" {
			break
		}
		nextToken = result.NextToken
	}

	log.Infof("Retrieved %d files in total", len(allFiles))
	return allFiles, nil
}

// ListAllFilesAsync 异步列出所有文件（自动处理分页），返回一个 channel
func (client *BailianClient) ListAllFilesAsync(fileName string) (<-chan *FileInfo, <-chan error) {
	fileChan := make(chan *FileInfo)
	errChan := make(chan error, 1)

	go func() {
		defer close(fileChan)
		defer close(errChan)

		nextToken := ""
		totalFiles := 0

		for {
			result, err := client.ListFile(100, nextToken, fileName)
			if err != nil {
				errChan <- err
				return
			}

			// 发送文件到 channel
			for _, file := range result.Files {
				fileChan <- file
				totalFiles++
			}

			// 检查是否有更多页
			if result.NextToken == "" {
				break
			}
			nextToken = result.NextToken
		}

		log.Infof("Streamed %d files in total", totalFiles)
	}()

	return fileChan, errChan
}

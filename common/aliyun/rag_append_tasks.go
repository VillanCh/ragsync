// This file is auto-generated, don't edit it. Thanks.
package aliyun

import (
	"os"
	"path/filepath"

	bailian20231229 "github.com/alibabacloud-go/bailian-20231229/v2/client"
	util "github.com/alibabacloud-go/tea-utils/v2/service"
	"github.com/alibabacloud-go/tea/tea"
	"github.com/yaklang/yaklang/common/log"
	"github.com/yaklang/yaklang/common/utils"
)

// AppendDocumentsToIndex 将文档添加到知识库索引，并返回任务ID
func (client *BailianClient) AppendDocumentsToIndex(documentIds []string) (string, error) {
	if client.config.BailianKnowledgeIndexId == "" {
		return "", utils.Errorf("Bailian knowledge index ID (BailianKnowledgeIndexId) is not configured")
	}

	// 转换文档ID为tea.String数组
	teaDocumentIds := make([]*string, 0, len(documentIds))
	for _, id := range documentIds {
		teaDocumentIds = append(teaDocumentIds, tea.String(id))
	}

	// 创建请求
	submitIndexAddDocumentsJobRequest := &bailian20231229.SubmitIndexAddDocumentsJobRequest{
		IndexId:     tea.String(client.config.BailianKnowledgeIndexId),
		SourceType:  tea.String("DATA_CENTER_FILE"),
		DocumentIds: teaDocumentIds,
	}

	// 运行时选项和请求头
	runtime := &util.RuntimeOptions{}
	headers := make(map[string]*string)

	// 发送请求
	log.Infof("Adding %d documents to knowledge index: %s", len(documentIds), client.config.BailianKnowledgeIndexId)
	response, err := client.Client.SubmitIndexAddDocumentsJobWithOptions(
		tea.String(client.config.BailianWorkspaceId),
		submitIndexAddDocumentsJobRequest,
		headers,
		runtime,
	)

	if err != nil {
		return "", utils.Errorf("Failed to add documents to index: %v", err)
	}

	// 打印响应信息
	var jobId string
	if response != nil && response.Body != nil {
		log.Infof("Documents added successfully, request ID: %s", tea.StringValue(response.Body.RequestId))

		// 将响应数据转为JSON，然后从JSON中获取JobId
		if response.Body.Data != nil {
			jobId = tea.StringValue(response.Body.Data.Id)

			// 如果获取到了JobId，保存它
			if jobId != "" {
				log.Infof("Job ID: %s", jobId)

				// 保存任务ID到本地文件
				if err := saveJobIdToFile(jobId); err != nil {
					log.Warnf("Failed to save job ID to file: %v", err)
				}
			}
		}

		log.Infof("Response data: %+v", response.Body.Data)
	} else {
		log.Infof("Documents added successfully, but response is empty")
	}

	return jobId, nil
}

// AppendDocumentToIndex 将单个文档添加到知识库索引 (便捷方法)
func (client *BailianClient) AppendDocumentToIndex(documentId string) (string, error) {
	log.Infof("Start to appending document to knowledge index: %s, checking if it's already indexed...", documentId)

	log.Infof("Getting file[%v] info...", documentId)
	// 获取文档信息
	fileInfo, err := client.DescribeFile(documentId)
	if err != nil {
		return "", utils.Errorf("Failed to get file info for document ID %s: %v", documentId, err)
	}

	log.Infof("Checking if file[%v] is already indexed...", fileInfo.FileName)
	// 检查文档是否已经在索引中
	isExisting, err := client.CheckAndWaitForExistingIndexJob(fileInfo.FileName)
	if err != nil {
		log.Warnf("Failed to check if document is already indexed: %v", err)
		// 即使检查失败，仍然继续添加文档到索引
	} else if isExisting {
		log.Infof("Document '%s' (ID: %s) is already being indexed or has been indexed. Skipping index addition.",
			fileInfo.FileName, documentId)
		return "", nil
	}

	// 添加文档到索引
	return client.AppendDocumentsToIndex([]string{documentId})
}

// saveJobIdToFile 将任务ID保存到本地文件
func saveJobIdToFile(jobId string) error {
	// 获取用户主目录
	homeDir := utils.GetHomeDirDefault(".")

	// 创建 ~/.ragsync/index-jobs/ 目录
	jobsDir := filepath.Join(homeDir, ".ragsync", "index-jobs")
	if err := os.MkdirAll(jobsDir, 0755); err != nil {
		return utils.Errorf("Failed to create directory %s: %v", jobsDir, err)
	}

	// 创建文件，文件名为任务ID
	jobFilePath := filepath.Join(jobsDir, jobId)

	// 创建空文件
	file, err := os.Create(jobFilePath)
	if err != nil {
		return utils.Errorf("Failed to create job file %s: %v", jobFilePath, err)
	}
	defer file.Close()

	log.Infof("Job ID saved to file: %s", jobFilePath)
	return nil
}

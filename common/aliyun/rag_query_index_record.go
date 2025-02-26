// This file is auto-generated, don't edit it. Thanks.
package aliyun

import (
	"time"

	bailian20231229 "github.com/alibabacloud-go/bailian-20231229/v2/client"
	util "github.com/alibabacloud-go/tea-utils/v2/service"
	"github.com/alibabacloud-go/tea/tea"
	"github.com/yaklang/yaklang/common/log"
	"github.com/yaklang/yaklang/common/utils"
)

// IndexDocumentRecord 索引文档记录
type IndexDocumentRecord struct {
	DocumentName string `json:"documentName"`
	DocumentId   string `json:"documentId"`
	Status       string `json:"status"`
	IndexId      string `json:"indexId"`
	DocumentType string `json:"documentType"`
	Code         string `json:"code"`
	Message      string `json:"message"`
	Size         int32  `json:"size"`
	SourceId     string `json:"sourceId"`
	Raw          any    `json:"raw"`
}

// QueryIndexRecordFromDocumentName 根据文档名查询索引记录
func (client *BailianClient) QueryIndexRecordFromDocumentName(documentName string) ([]*IndexDocumentRecord, error) {
	if client.config == nil {
		return nil, utils.Error("Client configuration is not set")
	}

	if client.config.BailianWorkspaceId == "" {
		return nil, utils.Error("Workspace ID is not set")
	}

	if client.config.BailianKnowledgeIndexId == "" {
		return nil, utils.Error("Knowledge Index ID is not set")
	}

	if documentName == "" {
		return nil, utils.Error("Document name cannot be empty")
	}

	// 创建请求
	listIndexDocumentsRequest := &bailian20231229.ListIndexDocumentsRequest{
		DocumentName:   tea.String(documentName),
		IndexId:        tea.String(client.config.BailianKnowledgeIndexId),
		DocumentStatus: tea.String(""),
	}

	runtime := &util.RuntimeOptions{}
	headers := make(map[string]*string)

	// 调用 API
	response, err := client.Client.ListIndexDocumentsWithOptions(
		tea.String(client.config.BailianWorkspaceId),
		listIndexDocumentsRequest,
		headers,
		runtime,
	)

	if err != nil {
		return nil, utils.Errorf("Failed to query index records: %v", err)
	}

	// 解析响应
	if response == nil || response.Body == nil {
		return nil, utils.Errorf("Query index records response is empty")
	}

	// 检查响应是否成功
	if response.Body.Success == nil || !*response.Body.Success {
		errorMsg := "Unknown error"
		if response.Body.Message != nil {
			errorMsg = *response.Body.Message
		}
		return nil, utils.Errorf("Failed to query index records: %v", errorMsg)
	}

	// 解析结果
	var records []*IndexDocumentRecord

	if response.Body.Data != nil && response.Body.Data.Documents != nil {
		for _, doc := range response.Body.Data.Documents {
			// 检查文档名称是否完全匹配
			if tea.StringValue(doc.Name) != documentName {
				log.Infof("Skipping document with name: %s (not exact match with: %s)", tea.StringValue(doc.Name), documentName)
				continue
			}
			log.Infof("Found exact match for document: %s", documentName)
			record := &IndexDocumentRecord{
				Raw:          doc,
				DocumentName: tea.StringValue(doc.Name),
				DocumentId:   tea.StringValue(doc.Id),
				Status:       tea.StringValue(doc.Status),
				DocumentType: tea.StringValue(doc.DocumentType),
				Code:         tea.StringValue(doc.Code),
				Message:      tea.StringValue(doc.Message),
				SourceId:     tea.StringValue(doc.SourceId),
				IndexId:      client.config.BailianKnowledgeIndexId,
			}
			if doc.Size != nil {
				record.Size = *doc.Size
			}
			records = append(records, record)
		}
	}

	log.Infof("Found %d index records for document: %s", len(records), documentName)
	return records, nil
}

// CheckAndWaitForExistingIndexJob 检查文档是否已在索引中，如果是，则等待并返回true
func (client *BailianClient) CheckAndWaitForExistingIndexJob(documentName string) (bool, error) {
	log.Infof("Checking if document[%v] is already indexed... QueryIndexRecordFromDocumentName", documentName)

	// 如果提供了文件名，移除扩展名
	if documentName != "" {
		documentName = removeFileExtension(documentName)
		log.Infof("Checking index records for document with name: %s (extension removed)", documentName)
	}
	records, err := client.QueryIndexRecordFromDocumentName(documentName)
	if err != nil {
		return false, err
	}

	if len(records) > 0 {
		log.Warnf("Document '%s' is already being indexed or has been indexed", documentName)
		log.Infof("Found %d existing index records, waiting for 1 seconds...", len(records))

		// 打印出每个记录的状态
		for i, record := range records {
			log.Infof("Record %d: DocumentId=%s, Status=%s",
				i+1, record.DocumentId, record.Status)
		}

		time.Sleep(1 * time.Second)
		return true, nil
	}

	return false, nil
}

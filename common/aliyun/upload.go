package aliyun

import (
	"fmt"
	"path/filepath"

	"github.com/yaklang/yaklang/common/utils"
	"github.com/yaklang/yaklang/common/utils/lowhttp/poc"
)

// UploadFile 上传文件到指定URL
func UploadFile(method string, uploadURL string, fileName string, contentType string, content []byte, bailianExtra string) error {
	// 获取文件扩展名
	ext := filepath.Ext(fileName)
	if ext == "" {
		return utils.Errorf("File extension cannot be empty")
	}

	rsp, req, err := poc.Do(
		method, uploadURL,
		poc.WithReplaceHttpPacketBody(content, false),
		poc.WithReplaceHttpPacketHeader("Content-Type", contentType),
		poc.WithReplaceHttpPacketHeader("X-bailian-extra", bailianExtra),
	)
	if err != nil {
		return utils.Errorf("Failed to upload file: %v", err)
	}
	_ = req
	fmt.Println(string(rsp.RawRequest))
	fmt.Println(string(rsp.RawPacket))
	return nil
}

// getContentType 根据文件扩展名获取内容类型
func getContentType(filePath string) string {
	ext := filepath.Ext(filePath)
	switch ext {
	case ".pdf":
		return "application/pdf"
	case ".txt":
		return "text/plain"
	case ".doc":
		return "application/msword"
	case ".docx":
		return "application/vnd.openxmlformats-officedocument.wordprocessingml.document"
	case ".xls":
		return "application/vnd.ms-excel"
	case ".xlsx":
		return "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet"
	case ".ppt":
		return "application/vnd.ms-powerpoint"
	case ".pptx":
		return "application/vnd.openxmlformats-officedocument.presentationml.presentation"
	default:
		return "application/octet-stream"
	}
}

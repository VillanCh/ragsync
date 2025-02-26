package aliyun

import (
	"path/filepath"

	"github.com/yaklang/yaklang/common/log"
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
	_ = rsp
	log.Infof("Upload file success: %s", fileName)
	return nil
}

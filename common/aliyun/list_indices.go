package aliyun

import (
	"encoding/json"
	"strings"

	"github.com/VillanCh/ragsync/common/spec"
	bailian20231229 "github.com/alibabacloud-go/bailian-20231229/v2/client"
	util "github.com/alibabacloud-go/tea-utils/v2/service"
	"github.com/alibabacloud-go/tea/tea"
	"github.com/yaklang/yaklang/common/log"
	"github.com/yaklang/yaklang/common/utils"
)

type Index struct {
	IndexId   string
	IndexName string
	Raw       any
}

func ListIndices(accessKey, secretKey, workspaceId string) ([]Index, error) {
	config := &spec.Config{
		AliyunAccessKey:    accessKey,
		AliyunSecretKey:    secretKey,
		BailianWorkspaceId: workspaceId,
	}

	client, err := NewBailianClientFromConfig(config)
	if err != nil {
		return nil, utils.Errorf("Failed to create client: %v", err)
	}

	request := &bailian20231229.ListIndicesRequest{}
	runtime := &util.RuntimeOptions{}
	headers := make(map[string]*string)

	response, err := client.Client.ListIndicesWithOptions(
		tea.String(workspaceId),
		request,
		headers,
		runtime,
	)
	if err != nil {
		var sdkErr *tea.SDKError
		if teaErr, ok := err.(*tea.SDKError); ok {
			sdkErr = teaErr
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
		return nil, utils.Errorf("Failed to list indices: %v", err)
	}

	if response == nil || response.Body == nil {
		return nil, utils.Errorf("List indices response is empty")
	}

	if !tea.BoolValue(response.Body.Success) {
		return nil, utils.Errorf("Failed to list indices: %v", response.Body.Message)
	}

	var indices []Index
	if response.Body.Data != nil && response.Body.Data.Indices != nil {
		for _, idx := range response.Body.Data.Indices {
			indices = append(indices, Index{
				IndexId:   tea.StringValue(idx.Id),
				IndexName: tea.StringValue(idx.Name),
				Raw:       idx,
			})
		}
	}

	return indices, nil
}

func CreateIndex(accessKey, secretKey, workspaceId, name, sourceType string, categoryIds []string) error {
	config := &spec.Config{
		AliyunAccessKey:    accessKey,
		AliyunSecretKey:    secretKey,
		BailianWorkspaceId: workspaceId,
	}

	client, err := NewBailianClientFromConfig(config)
	if err != nil {
		return utils.Errorf("Failed to create client: %v", err)
	}

	request := &bailian20231229.CreateIndexRequest{
		Name:          tea.String(name),
		StructureType: tea.String("unstructured"),
		SourceType:    tea.String(sourceType),
		CategoryIds:   tea.StringSlice(categoryIds),
		SinkType:      tea.String("BUILT_IN"),
	}
	runtime := &util.RuntimeOptions{}
	headers := make(map[string]*string)

	response, err := client.Client.CreateIndexWithOptions(
		tea.String(workspaceId),
		request,
		headers,
		runtime,
	)
	if err != nil {
		var sdkErr *tea.SDKError
		if teaErr, ok := err.(*tea.SDKError); ok {
			sdkErr = teaErr
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
		return utils.Errorf("Failed to create index: %v", err)
	}

	if response == nil || response.Body == nil {
		return utils.Errorf("Create index response is empty")
	}

	if !tea.BoolValue(response.Body.Success) {
		return utils.Errorf("Failed to create index: %v", response.Body.Message)
	}

	log.Infof("Index created successfully: %s", name)
	return nil
}

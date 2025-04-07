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

func CreateCategory(accessKey, secretKey, workspaceId, name string) error {
	config := &spec.Config{
		AliyunAccessKey:    accessKey,
		AliyunSecretKey:    secretKey,
		BailianWorkspaceId: workspaceId,
	}

	client, err := NewBailianClientFromConfig(config)
	if err != nil {
		return utils.Errorf("Failed to create client: %v", err)
	}

	request := &bailian20231229.AddCategoryRequest{
		CategoryName: tea.String(name),
		CategoryType: tea.String("UNSTRUCTURED"),
	}
	runtime := &util.RuntimeOptions{}
	headers := make(map[string]*string)

	response, err := client.Client.AddCategoryWithOptions(
		tea.String(workspaceId),
		request,
		headers,
		runtime,
	)
	if err != nil {
		var sdkErr *tea.SDKError
		if teaErr, ok := err.(*tea.SDKError); ok {
			sdkErr = teaErr
			log.Errorf("SDK Error Code: %s", tea.StringValue(sdkErr.Code))
			log.Errorf("SDK Error Message: %s", tea.StringValue(sdkErr.Message))

			if sdkErr.Data != nil {
				log.Errorf("Raw Error Data: %s", tea.StringValue(sdkErr.Data))
				var data interface{}
				decoder := json.NewDecoder(strings.NewReader(tea.StringValue(sdkErr.Data)))
				if decodeErr := decoder.Decode(&data); decodeErr == nil {
					log.Errorf("Parsed Error Data: %v", data)
					if m, ok := data.(map[string]interface{}); ok {
						// Print all possible error fields
						for k, v := range m {
							log.Errorf("Error Field %s: %v", k, v)
						}

						if recommend, ok := m["Recommend"]; ok {
							log.Errorf("Recommended Solution: %v", recommend)
						}
						if code, ok := m["Code"]; ok {
							log.Errorf("Error Code: %v", code)
						}
						if message, ok := m["Message"]; ok {
							log.Errorf("Error Message: %v", message)
						}
						if requestId, ok := m["RequestId"]; ok {
							log.Errorf("Request ID: %v", requestId)
						}
					}
				} else {
					log.Errorf("Failed to parse error data: %v", decodeErr)
				}
			}
		} else {
			log.Errorf("Non-SDK Error: %v", err)
		}
		return utils.Errorf("Failed to create category: %v", err)
	}

	if response == nil || response.Body == nil {
		return utils.Errorf("Create category response is empty")
	}

	if !tea.BoolValue(response.Body.Success) {
		log.Errorf("Request ID: %s", tea.StringValue(response.Body.RequestId))
		log.Errorf("Error Code: %s", tea.StringValue(response.Body.Code))
		log.Errorf("Error Message: %s", tea.StringValue(response.Body.Message))
		return utils.Errorf("Failed to create category: %s (Request ID: %s)", tea.StringValue(response.Body.Message), tea.StringValue(response.Body.RequestId))
	}

	log.Infof("Category created successfully: %s", name)
	return nil
}

type Category struct {
	CategoryId   string
	CategoryName string
	Raw          any
}

func ListCategories(accessKey, secretKey, workspaceId string) ([]Category, error) {
	config := &spec.Config{
		AliyunAccessKey:    accessKey,
		AliyunSecretKey:    secretKey,
		BailianWorkspaceId: workspaceId,
	}

	client, err := NewBailianClientFromConfig(config)
	if err != nil {
		return nil, utils.Errorf("Failed to create client: %v", err)
	}

	request := &bailian20231229.ListCategoryRequest{
		CategoryType: tea.String("UNSTRUCTURED"),
	}
	runtime := &util.RuntimeOptions{}
	headers := make(map[string]*string)

	response, err := client.Client.ListCategoryWithOptions(
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
		return nil, utils.Errorf("Failed to list categories: %v", err)
	}

	if response == nil || response.Body == nil {
		return nil, utils.Errorf("List categories response is empty")
	}

	if !tea.BoolValue(response.Body.Success) {
		return nil, utils.Errorf("Failed to list categories: %v", response.Body.Message)
	}

	var categories []Category
	if response.Body.Data != nil && response.Body.Data.CategoryList != nil {
		for _, cat := range response.Body.Data.CategoryList {
			categories = append(categories, Category{
				CategoryId:   tea.StringValue(cat.CategoryId),
				CategoryName: tea.StringValue(cat.CategoryName),
				Raw:          cat,
			})
		}
	}

	return categories, nil
}

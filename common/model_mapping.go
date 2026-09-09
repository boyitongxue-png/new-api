package common

import (
	"errors"
	"fmt"
	"strings"
)

// ResolveModelMapping follows a channel model mapping to its final upstream
// model while rejecting cycles.
func ResolveModelMapping(modelMapping, modelName string) (string, bool, error) {
	modelMapping = strings.TrimSpace(modelMapping)
	if modelMapping == "" || modelMapping == "{}" {
		return modelName, false, nil
	}

	modelMap := make(map[string]string)
	if err := UnmarshalJsonStr(modelMapping, &modelMap); err != nil {
		return "", false, fmt.Errorf("unmarshal_model_mapping_failed")
	}

	currentModel := modelName
	visitedModels := map[string]bool{currentModel: true}
	for {
		mappedModel, exists := modelMap[currentModel]
		if !exists || mappedModel == "" {
			return currentModel, currentModel != modelName, nil
		}
		if visitedModels[mappedModel] {
			if mappedModel == currentModel {
				return currentModel, currentModel != modelName, nil
			}
			return "", false, errors.New("model_mapping_contains_cycle")
		}
		visitedModels[mappedModel] = true
		currentModel = mappedModel
	}
}

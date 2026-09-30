package ratio_setting

import (
	"encoding/json"
	"fmt"
	"math"
	"strings"

	"github.com/QuantumNous/new-api/common"
)

// MaxPricingValue keeps configurable prices and multipliers within a range
// that can be converted to quota without overflowing the 32-bit quota fields.
const MaxPricingValue = 1_000_000_000.0

func validatePricingValue(label string, value float64) error {
	if math.IsNaN(value) || math.IsInf(value, 0) || value < 0 {
		return fmt.Errorf("%s must be a finite non-negative number", label)
	}
	if value > MaxPricingValue {
		return fmt.Errorf("%s must not exceed %g", label, MaxPricingValue)
	}
	return nil
}

func validateRatioMapJSON(jsonStr, label string) error {
	var values map[string]json.RawMessage
	if err := common.UnmarshalJsonStr(jsonStr, &values); err != nil {
		return fmt.Errorf("%s must be a JSON object: %w", label, err)
	}
	if values == nil {
		return fmt.Errorf("%s must be a JSON object", label)
	}
	for name, rawValue := range values {
		if strings.TrimSpace(name) == "" {
			return fmt.Errorf("%s model name cannot be empty", label)
		}
		if common.GetJsonType(rawValue) != "number" {
			return fmt.Errorf("%s for %s must be a JSON number", label, name)
		}
		var value float64
		if err := common.Unmarshal(rawValue, &value); err != nil {
			return fmt.Errorf("%s for %s must be a JSON number: %w", label, name, err)
		}
		if err := validatePricingValue(fmt.Sprintf("%s for %s", label, name), value); err != nil {
			return err
		}
	}
	return nil
}

// ValidateModelPriceJSONString validates per-model token prices.
func ValidateModelPriceJSONString(jsonStr string) error {
	return validateRatioMapJSON(jsonStr, "model price")
}

// ValidateModelRatioJSONString validates per-model input ratios.
func ValidateModelRatioJSONString(jsonStr string) error {
	return validateRatioMapJSON(jsonStr, "model ratio")
}

// ValidateCompletionRatioJSONString validates output completion ratios.
func ValidateCompletionRatioJSONString(jsonStr string) error {
	return validateRatioMapJSON(jsonStr, "completion ratio")
}

// ValidateCacheRatioJSONString validates cached input ratios.
func ValidateCacheRatioJSONString(jsonStr string) error {
	return validateRatioMapJSON(jsonStr, "cache ratio")
}

// ValidateCreateCacheRatioJSONString validates cache creation ratios.
func ValidateCreateCacheRatioJSONString(jsonStr string) error {
	return validateRatioMapJSON(jsonStr, "create cache ratio")
}

// ValidateImageRatioJSONString validates image generation ratios.
func ValidateImageRatioJSONString(jsonStr string) error {
	return validateRatioMapJSON(jsonStr, "image ratio")
}

// ValidateAudioRatioJSONString validates audio input ratios.
func ValidateAudioRatioJSONString(jsonStr string) error {
	return validateRatioMapJSON(jsonStr, "audio ratio")
}

// ValidateAudioCompletionRatioJSONString validates audio output ratios.
func ValidateAudioCompletionRatioJSONString(jsonStr string) error {
	return validateRatioMapJSON(jsonStr, "audio completion ratio")
}

func validateNestedRatioMapJSON(jsonStr, label string) error {
	var values map[string]map[string]json.RawMessage
	if err := common.UnmarshalJsonStr(jsonStr, &values); err != nil {
		return fmt.Errorf("%s must be a JSON object: %w", label, err)
	}
	if values == nil {
		return fmt.Errorf("%s must be a JSON object", label)
	}
	for outerName, nested := range values {
		if strings.TrimSpace(outerName) == "" {
			return fmt.Errorf("%s group name cannot be empty", label)
		}
		if nested == nil {
			return fmt.Errorf("%s group %s must be a JSON object", label, outerName)
		}
		for innerName, rawValue := range nested {
			if strings.TrimSpace(innerName) == "" {
				return fmt.Errorf("%s target group name cannot be empty", label)
			}
			if common.GetJsonType(rawValue) != "number" {
				return fmt.Errorf("%s for %s/%s must be a JSON number", label, outerName, innerName)
			}
			var value float64
			if err := common.Unmarshal(rawValue, &value); err != nil {
				return fmt.Errorf("%s for %s/%s must be a JSON number: %w", label, outerName, innerName, err)
			}
			if err := validatePricingValue(fmt.Sprintf("%s for %s/%s", label, outerName, innerName), value); err != nil {
				return err
			}
		}
	}
	return nil
}

// ValidateGroupRatioJSONString validates user group price multipliers.
func ValidateGroupRatioJSONString(jsonStr string) error {
	return validateRatioMapJSON(jsonStr, "group ratio")
}

// ValidateGroupGroupRatioJSONString validates per-user-group target multipliers.
func ValidateGroupGroupRatioJSONString(jsonStr string) error {
	return validateNestedRatioMapJSON(jsonStr, "group group ratio")
}

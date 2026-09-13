package model

import (
	"fmt"
	"math"
	"sort"
	"strconv"
	"strings"
)

// ImageChannelCost returns the configured per-image cost for a resolution tier.
// Channel metadata is intentionally kept in OtherInfo so this remains backwards
// compatible with existing channel records and API clients.
func ImageChannelCost(channel *Channel, resolution string) (float64, bool) {
	if channel == nil {
		return 0, false
	}
	tier := NormalizeImageResolution(resolution)
	if tier == "" {
		return 0, false
	}
	info := channel.GetOtherInfo()
	if !supportsImageResolution(info, tier) {
		return 0, false
	}
	keys := []string{"cost" + tier, "cost_" + tier, "costPerImage" + tier, "cost_per_image_" + tier}
	for _, key := range keys {
		if value, ok := numberFromMetadata(info[key]); ok && value >= 0 && !math.IsNaN(value) && !math.IsInf(value, 0) {
			return value, true
		}
	}
	if costs, ok := info["costs"].(map[string]interface{}); ok {
		if value, ok := numberFromMetadata(costs[tier]); ok && value >= 0 && !math.IsNaN(value) && !math.IsInf(value, 0) {
			return value, true
		}
	}
	if value, ok := numberFromMetadata(info["unitCostPerImage"]); ok && value >= 0 && !math.IsNaN(value) && !math.IsInf(value, 0) {
		return value, true
	}
	return 0, false
}

// SelectLowestCostImageChannels returns candidates ordered by cost. It returns
// ok=false when no candidate has a valid cost for the requested tier, allowing
// callers to preserve the legacy priority/weight scheduler.
func SelectLowestCostImageChannels(channels []*Channel, resolution string) (selected []*Channel, ok bool) {
	if NormalizeImageResolution(resolution) == "" {
		return nil, false
	}
	type priced struct {
		channel *Channel
		cost    float64
	}
	pricedChannels := make([]priced, 0, len(channels))
	for _, channel := range channels {
		if cost, exists := ImageChannelCost(channel, resolution); exists {
			pricedChannels = append(pricedChannels, priced{channel: channel, cost: cost})
		}
	}
	if len(pricedChannels) == 0 {
		return nil, false
	}
	sort.SliceStable(pricedChannels, func(i, j int) bool {
		if pricedChannels[i].cost == pricedChannels[j].cost {
			return pricedChannels[i].channel.Id < pricedChannels[j].channel.Id
		}
		return pricedChannels[i].cost < pricedChannels[j].cost
	})
	selected = make([]*Channel, len(pricedChannels))
	for i := range pricedChannels {
		selected[i] = pricedChannels[i].channel
	}
	return selected, true
}

// NormalizeImageResolution accepts the workshop/API MIX tier names and common
// OpenAI size values. Unknown values are left empty so routing safely falls back.
func NormalizeImageResolution(raw string) string {
	s := strings.ToLower(strings.TrimSpace(raw))
	s = strings.ReplaceAll(s, " ", "")
	s = strings.ReplaceAll(s, "×", "x")
	switch s {
	case "1k", "2k", "4k", "8k":
		return s
	case "256x256", "512x512", "1024x1024", "1024x1792", "1792x1024":
		return "1k"
	}
	if strings.Contains(s, "x") {
		parts := strings.Split(s, "x")
		if len(parts) == 2 {
			w, e1 := strconv.Atoi(parts[0])
			h, e2 := strconv.Atoi(parts[1])
			if e1 == nil && e2 == nil && w > 0 && h > 0 {
				pixels := w * h
				switch {
				case pixels >= 8_000_000:
					return "4k"
				case pixels >= 2_000_000:
					return "2k"
				default:
					return "1k"
				}
			}
		}
	}
	return ""
}

func supportsImageResolution(info map[string]interface{}, tier string) bool {
	raw, ok := info["supportedResolutions"]
	if !ok {
		raw, ok = info["supported_resolutions"]
	}
	if ok {
		switch values := raw.(type) {
		case []interface{}:
			for _, value := range values {
				if NormalizeImageResolution(fmt.Sprint(value)) == tier {
					return true
				}
			}
			return false
		case []string:
			for _, value := range values {
				if NormalizeImageResolution(value) == tier {
					return true
				}
			}
			return false
		case string:
			for _, value := range strings.Split(values, ",") {
				if NormalizeImageResolution(value) == tier {
					return true
				}
			}
			return false
		}
	}
	return true
}

func numberFromMetadata(value interface{}) (float64, bool) {
	switch v := value.(type) {
	case float64:
		return v, true
	case float32:
		return float64(v), true
	case int:
		return float64(v), true
	case int64:
		return float64(v), true
	case jsonNumber:
		f, err := strconv.ParseFloat(string(v), 64)
		return f, err == nil
	case string:
		f, err := strconv.ParseFloat(strings.TrimSpace(v), 64)
		return f, err == nil
	default:
		return 0, false
	}
}

type jsonNumber string
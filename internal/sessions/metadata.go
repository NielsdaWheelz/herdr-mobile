package sessions

import (
	"context"
	"encoding/base64"
	"errors"
	"fmt"
	"strconv"
	"strings"
)

func encodeObjective(objective string) map[string]string {
	if objective == "" {
		return nil
	}
	encoded := base64.RawURLEncoding.EncodeToString([]byte(objective))
	count := (len(encoded) + 79) / 80
	parts := make(map[string]string, count+1)
	parts["skid_objective_count"] = strconv.Itoa(count)
	for index := 0; index < count; index++ {
		start := index * 80
		end := min(start+80, len(encoded))
		parts[fmt.Sprintf("skid_objective_%02d", index)] = encoded[start:end]
	}
	return parts
}

func decodeObjective(tokens map[string]string) string {
	countText, found := tokens["skid_objective_count"]
	if !found {
		return ""
	}
	count, err := strconv.Atoi(countText)
	if err != nil || count < 1 || count > 16 || strconv.Itoa(count) != countText {
		return ""
	}
	var encoded strings.Builder
	for index := 0; index < count; index++ {
		part, present := tokens[fmt.Sprintf("skid_objective_%02d", index)]
		if !present || len(part) == 0 || len(part) > 80 {
			return ""
		}
		encoded.WriteString(part)
	}
	for key := range tokens {
		if strings.HasPrefix(key, "skid_objective_") && key != "skid_objective_count" {
			indexText := strings.TrimPrefix(key, "skid_objective_")
			index, err := strconv.Atoi(indexText)
			if err != nil || index < 0 || index >= count || fmt.Sprintf("%02d", index) != indexText {
				return ""
			}
		}
	}
	decoded, err := base64.RawURLEncoding.DecodeString(encoded.String())
	if err != nil || base64.RawURLEncoding.EncodeToString(decoded) != encoded.String() || validateObjective(string(decoded)) != nil {
		return ""
	}
	return string(decoded)
}

func (manager *Manager) writeObjective(ctx context.Context, paneID, objective string) error {
	parts := encodeObjective(objective)
	if len(parts) == 0 {
		return nil
	}
	count, _ := strconv.Atoi(parts["skid_objective_count"])
	if err := manager.reportPane(ctx, paneID, map[string]string{"skid_objective_count": parts["skid_objective_count"]}); err != nil {
		return err
	}
	for start := 0; start < count; start += 16 {
		batch := make(map[string]string)
		for index := start; index < min(start+16, count); index++ {
			key := fmt.Sprintf("skid_objective_%02d", index)
			batch[key] = parts[key]
		}
		if err := manager.reportPane(ctx, paneID, batch); err != nil {
			return err
		}
	}
	current, err := manager.pane(ctx, paneID)
	if err != nil || decodeObjective(current.Tokens) != objective {
		return errors.New("objective metadata was not retained")
	}
	return nil
}

package service

import (
	"encoding/json"
	"strings"

	"github.com/basketikun/infinite-canvas/model"
	"github.com/basketikun/infinite-canvas/repository"
)

func ListGenerationLogs(q model.Query) (model.GenerationLogList, error) {
	items, total, err := repository.ListGenerationLogs(q)
	if err != nil {
		return model.GenerationLogList{}, err
	}
	for i := range items {
		items[i].Request = ""
		items[i].Response = ""
		items[i].Images = compactLogImages(items[i].Images)
	}
	return model.GenerationLogList{Items: items, Total: int(total)}, nil
}

func ListUserGenerationLogs(userID string, q model.Query) (model.GenerationLogList, error) {
	items, total, err := repository.ListUserGenerationLogs(userID, q)
	if err != nil {
		return model.GenerationLogList{}, err
	}
	for i := range items {
		items[i].Request = ""
		items[i].Response = ""
		items[i].Images = compactLogImages(items[i].Images)
	}
	return model.GenerationLogList{Items: items, Total: int(total)}, nil
}

func compactLogImages(images []string) []string {
	if len(images) == 0 {
		return images
	}
	compact := make([]string, 0, len(images))
	for _, image := range images {
		if strings.HasPrefix(image, "data:image/") {
			continue
		}
		compact = append(compact, image)
	}
	return compact
}

func SaveGenerationLog(item model.GenerationLog) (model.GenerationLog, error) {
	if item.ID == "" {
		item.ID = newID("gen")
	}
	if item.CreatedAt == "" {
		item.CreatedAt = now()
	}
	return repository.SaveGenerationLog(item)
}

func DeleteGenerationLog(id string) error { return repository.DeleteGenerationLog(id) }

func DeleteGenerationLogs(ids []string) error { return repository.DeleteGenerationLogs(ids) }

func BuildGenerationLog(userID string, path string, modelName string, requestBody []byte, responseBody []byte, status string, errMessage string) model.GenerationLog {
	return BuildGenerationLogForTask("", userID, path, modelName, requestBody, responseBody, status, errMessage)
}

func BuildGenerationLogForTask(taskID string, userID string, path string, modelName string, requestBody []byte, responseBody []byte, status string, errMessage string) model.GenerationLog {
	kind := model.GenerationLogKindChat
	if strings.Contains(path, "/images/") {
		kind = model.GenerationLogKindImage
	}
	return model.GenerationLog{ID: newID("gen"), TaskID: taskID, UserID: userID, Kind: kind, Model: modelName, Path: path, Prompt: limitRunes(extractPrompt(requestBody), 4000), Images: extractImages(responseBody), Request: limitRunes(string(requestBody), 12000), Response: limitRunes(string(responseBody), 12000), Status: status, Error: limitRunes(errMessage, 4000), CreatedAt: now()}
}

func extractPrompt(body []byte) string {
	var payload map[string]any
	if json.Unmarshal(body, &payload) != nil {
		return ""
	}
	if value, ok := payload["prompt"].(string); ok {
		return value
	}
	if messages, ok := payload["messages"].([]any); ok {
		parts := []string{}
		for _, item := range messages {
			message, ok := item.(map[string]any)
			if !ok {
				continue
			}
			role, _ := message["role"].(string)
			content := stringifyMessageContent(message["content"])
			if strings.TrimSpace(content) != "" {
				if role != "" {
					parts = append(parts, role+": "+content)
				} else {
					parts = append(parts, content)
				}
			}
		}
		return strings.Join(parts, "\\n")
	}
	return ""
}

func stringifyMessageContent(value any) string {
	switch v := value.(type) {
	case string:
		return v
	case []any:
		parts := []string{}
		for _, item := range v {
			if m, ok := item.(map[string]any); ok {
				if text, ok := m["text"].(string); ok && text != "" {
					parts = append(parts, text)
				}
			}
		}
		return strings.Join(parts, "\\n")
	default:
		return ""
	}
}

func ExtractImagesForAccounting(body []byte) []string {
	return extractImages(body)
}

func extractImages(body []byte) []string {
	var payload any
	if json.Unmarshal(body, &payload) != nil {
		return nil
	}
	seen := map[string]bool{}
	images := []string{}
	collectImages(payload, &images, seen)
	return images
}

func collectImages(value any, images *[]string, seen map[string]bool) {
	switch typed := value.(type) {
	case map[string]any:
		for _, key := range []string{"url", "image_url", "image", "b64_json", "base64"} {
			if image := normalizeExtractedImage(typed[key], key); image != "" && !seen[image] {
				seen[image] = true
				*images = append(*images, image)
			}
		}
		for _, child := range typed {
			collectImages(child, images, seen)
		}
	case []any:
		for _, child := range typed {
			collectImages(child, images, seen)
		}
	}
}

func normalizeExtractedImage(value any, key string) string {
	text, ok := value.(string)
	if !ok {
		return ""
	}
	text = strings.TrimSpace(text)
	if text == "" {
		return ""
	}
	lower := strings.ToLower(text)
	if strings.HasPrefix(lower, "http://") || strings.HasPrefix(lower, "https://") || strings.HasPrefix(lower, "data:image/") {
		return text
	}
	if key == "b64_json" || key == "base64" || looksLikeBase64Image(text) {
		return "data:image/png;base64," + text
	}
	return ""
}

func looksLikeBase64Image(value string) bool {
	if len(value) < 4 || strings.ContainsAny(value, " \n\r	") {
		return false
	}
	for _, r := range value {
		if (r >= 'A' && r <= 'Z') || (r >= 'a' && r <= 'z') || (r >= '0' && r <= '9') || r == '+' || r == '/' || r == '=' {
			continue
		}
		return false
	}
	return true
}

func limitRunes(value string, max int) string {
	runes := []rune(value)
	if max <= 0 || len(runes) <= max {
		return value
	}
	return string(runes[:max]) + "..."
}

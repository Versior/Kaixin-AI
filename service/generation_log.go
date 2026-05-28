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
	var payload struct {
		Data []map[string]any `json:"data"`
	}
	if json.Unmarshal(body, &payload) != nil {
		return nil
	}
	images := []string{}
	for _, item := range payload.Data {
		if url, ok := item["url"].(string); ok && url != "" {
			images = append(images, url)
		} else if b64, ok := item["b64_json"].(string); ok && b64 != "" {
			images = append(images, "data:image/png;base64,"+b64)
		}
	}
	return images
}

func limitRunes(value string, max int) string {
	runes := []rune(value)
	if max <= 0 || len(runes) <= max {
		return value
	}
	return string(runes[:max]) + "..."
}

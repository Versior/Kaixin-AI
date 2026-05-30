package handler

import (
	"encoding/json"
	"fmt"
	"html"
	"log"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/basketikun/infinite-canvas/model"
	"github.com/basketikun/infinite-canvas/service"
)

type adminSyncRequest struct {
	Category string `json:"category"`
}

type adminBatchDeleteRequest struct {
	IDs []string `json:"ids"`
}

func AdminPromptCategories(w http.ResponseWriter, r *http.Request) {
	OK(w, service.ListPromptCategories())
}

func AdminPrompts(w http.ResponseWriter, r *http.Request) {
	result, err := service.ListPrompts(parseQuery(r))
	if err != nil {
		FailError(w, err)
		return
	}
	OK(w, result)
}

func AdminSavePrompt(w http.ResponseWriter, r *http.Request) {
	var item model.Prompt
	_ = json.NewDecoder(r.Body).Decode(&item)
	result, err := service.SavePrompt(item)
	if err != nil {
		FailError(w, err)
		return
	}
	OK(w, result)
}

func AdminDeletePrompt(w http.ResponseWriter, r *http.Request, id string) {
	if err := service.DeletePrompt(id); err != nil {
		FailError(w, err)
		return
	}
	OK(w, true)
}

func AdminDeletePrompts(w http.ResponseWriter, r *http.Request) {
	var request adminBatchDeleteRequest
	_ = json.NewDecoder(r.Body).Decode(&request)
	if err := service.DeletePrompts(request.IDs); err != nil {
		FailError(w, err)
		return
	}
	OK(w, true)
}

func AdminSyncPromptCategories(w http.ResponseWriter, r *http.Request) {
	var request adminSyncRequest
	_ = json.NewDecoder(r.Body).Decode(&request)
	log.Printf("sync prompt category start category=%s", request.Category)
	categories, err := service.SyncPromptCategory(request.Category)
	if err != nil {
		log.Printf("sync prompt category failed category=%s err=%v", request.Category, err)
		FailError(w, err)
		return
	}
	log.Printf("sync prompt category done category=%s", request.Category)
	OK(w, categories)
}

func AdminGenerationLogs(w http.ResponseWriter, r *http.Request) {
	result, err := service.ListGenerationLogs(parseQuery(r))
	if err != nil {
		FailError(w, err)
		return
	}
	OK(w, result)
}

func AdminGenerationTasks(w http.ResponseWriter, r *http.Request) {
	result, err := service.ListGenerationTasks(parseQuery(r))
	if err != nil {
		FailError(w, err)
		return
	}
	OK(w, result)
}

func AdminGenerationStats(w http.ResponseWriter, r *http.Request) {
	today := time.Now().Format("2006-01-02")
	result, err := service.GenerationImageStats(today)
	if err != nil {
		FailError(w, err)
		return
	}
	for i := range result.UserRanks {
		if strings.TrimSpace(result.UserRanks[i].AvatarURL) == "" {
			name := strings.TrimSpace(result.UserRanks[i].Username)
			userID := strings.TrimSpace(result.UserRanks[i].UserID)
			// inline avatar label
			label := ""
			if name != "" && name != "-" {
				label = string([]rune(name)[0])
			} else {
				uid := strings.TrimPrefix(userID, "user-")
				if uid != "" {
					label = strings.ToUpper(string([]rune(uid)[0]))
				} else {
					label = "用"
				}
			}
			colors := []string{"#f97316", "#ec4899", "#8b5cf6", "#06b6d4", "#22c55e", "#eab308", "#ef4444", "#6366f1"}
			idx := 0
			for _, rv := range label + userID {
				idx += int(rv)
			}
			bg := colors[idx%len(colors)]
			svg := fmt.Sprintf(`<svg xmlns="http://www.w3.org/2000/svg" width="96" height="96" viewBox="0 0 96 96"><rect width="96" height="96" rx="48" fill="%s"/><text x="50%%" y="54%%" text-anchor="middle" dominant-baseline="middle" font-family="Arial, PingFang SC, Microsoft YaHei, sans-serif" font-size="42" font-weight="700" fill="#fff">%s</text></svg>`, bg, html.EscapeString(label))
			result.UserRanks[i].AvatarURL = "data:image/svg+xml," + strings.ReplaceAll(url.QueryEscape(svg), "+", "%20")
		}
		result.UserRanks[i].Username = ""
	}
	OK(w, result)
}

func AdminDeleteGenerationLog(w http.ResponseWriter, r *http.Request, id string) {
	if err := service.DeleteGenerationLog(id); err != nil {
		FailError(w, err)
		return
	}
	OK(w, true)
}

func AdminDeleteGenerationLogs(w http.ResponseWriter, r *http.Request) {
	var request adminBatchDeleteRequest
	_ = json.NewDecoder(r.Body).Decode(&request)
	if err := service.DeleteGenerationLogs(request.IDs); err != nil {
		FailError(w, err)
		return
	}
	OK(w, true)
}

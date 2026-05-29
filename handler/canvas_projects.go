package handler

import (
	"encoding/json"
	"net/http"
	"strings"

	"github.com/basketikun/infinite-canvas/model"
	"github.com/basketikun/infinite-canvas/service"
)

func CanvasProjects(w http.ResponseWriter, r *http.Request) {
	user, ok := service.UserFromContext(r.Context())
	if !ok {
		Fail(w, "未登录或权限不足")
		return
	}
	items, err := service.ListCanvasProjects(user.ID)
	if err != nil {
		FailError(w, err)
		return
	}
	OK(w, items)
}

func SaveCanvasProject(w http.ResponseWriter, r *http.Request) {
	user, ok := service.UserFromContext(r.Context())
	if !ok {
		Fail(w, "未登录或权限不足")
		return
	}
	var project model.CanvasProject
	if err := json.NewDecoder(r.Body).Decode(&project); err != nil {
		Fail(w, "请求参数无效")
		return
	}
	project.UserID = ""
	if strings.TrimSpace(project.Payload) == "" {
		Fail(w, "画布内容不能为空")
		return
	}
	saved, err := service.SaveCanvasProject(user.ID, project)
	if err != nil {
		FailError(w, err)
		return
	}
	OK(w, saved)
}

func DeleteCanvasProject(w http.ResponseWriter, r *http.Request, id string) {
	user, ok := service.UserFromContext(r.Context())
	if !ok {
		Fail(w, "未登录或权限不足")
		return
	}
	if err := service.DeleteCanvasProject(user.ID, id); err != nil {
		FailError(w, err)
		return
	}
	OK(w, true)
}

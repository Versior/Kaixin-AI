package handler

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"mime"
	"mime/multipart"
	"net/http"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/basketikun/infinite-canvas/model"
	"github.com/basketikun/infinite-canvas/service"
)

func withPublicAPICORS(w http.ResponseWriter, r *http.Request) bool {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Headers", "Authorization, Content-Type")
	w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, PATCH, DELETE, OPTIONS")
	if r.Method == http.MethodOptions {
		w.WriteHeader(http.StatusNoContent)
		return true
	}
	return false
}

func AIImagesGenerations(w http.ResponseWriter, r *http.Request) {
	if withPublicAPICORS(w, r) {
		return
	}
	proxyAIRequest(w, r, "/images/generations")
}

func AIImagesEdits(w http.ResponseWriter, r *http.Request) {
	if withPublicAPICORS(w, r) {
		return
	}
	proxyAIRequest(w, r, "/images/edits")
}

func AIChatCompletions(w http.ResponseWriter, r *http.Request) {
	proxyAIRequest(w, r, "/chat/completions")
}

func AIVideos(w http.ResponseWriter, r *http.Request) {
	proxyAIRequest(w, r, "/videos")
}

func AIVideo(w http.ResponseWriter, r *http.Request, id string) {
	proxyAIGetRequest(w, r, "/videos/"+id)
}

func AIVideoContent(w http.ResponseWriter, r *http.Request, id string) {
	proxyAIGetRequest(w, r, "/videos/"+id+"/content")
}

func proxyAIGetRequest(w http.ResponseWriter, r *http.Request, path string) {
	modelName := r.URL.Query().Get("model")
	if strings.TrimSpace(modelName) == "" {
		modelName = "grok-imagine-video"
	}
	channel, err := service.SelectModelChannel(modelName)
	if err != nil {
		log.Printf("AI proxy select channel failed: model=%s err=%v", modelName, err)
		Fail(w, "AI 接口请求失败："+err.Error())
		return
	}
	request, err := http.NewRequest(http.MethodGet, service.BuildModelChannelURL(channel, path), nil)
	if err != nil {
		Fail(w, "AI 接口请求失败："+err.Error())
		return
	}
	request.Header.Set("Authorization", "Bearer "+channel.APIKey)
	copyAIResponse(w, request, nil, nil)
}

func proxyAIRequest(w http.ResponseWriter, r *http.Request, path string) {
	body, contentType, modelName, err := readAIRequest(r)
	if err != nil {
		log.Printf("AI proxy request read failed: %v", err)
		Fail(w, "AI 接口请求失败："+err.Error())
		return
	}
	user, ok := service.UserFromContext(r.Context())
	if !ok {
		Fail(w, "未登录或权限不足")
		return
	}
	credits, err := service.ModelCost(modelName)
	if err != nil {
		log.Printf("AI proxy read model cost failed: model=%s err=%v", modelName, err)
		Fail(w, "AI 接口请求失败："+err.Error())
		return
	}
	if isImageAIPath(path) {
		if !allowImageSubmission(user) {
			if _, err := service.SaveGenerationLog(service.BuildGenerationLog(user.ID, path, modelName, body, []byte(`{"msg":"图片生成太频繁，请 3 分钟内最多提交 3 次"}`), "rate_limited", "图片生成太频繁，请 3 分钟内最多提交 3 次")); err != nil {
				log.Printf("AI proxy save rate limit generation log failed: user=%s model=%s err=%v", user.ID, modelName, err)
			}
			Fail(w, "图片生成太频繁，请 3 分钟内最多提交 3 次")
			return
		}
	}
	run := func() {
		body, contentType = normalizeImageRequest(path, body, contentType)
		credits *= readAIRequestCount(body, contentType)
		channel, err := service.SelectModelChannel(modelName)
		if err != nil {
			log.Printf("AI proxy select channel failed: model=%s err=%v", modelName, err)
			Fail(w, "AI 接口请求失败："+err.Error())
			return
		}
		request, err := http.NewRequest(http.MethodPost, service.BuildModelChannelURL(channel, path), bytes.NewReader(body))
		if err != nil {
			log.Printf("AI proxy build request failed: url=%s err=%v", service.BuildModelChannelURL(channel, path), err)
			Fail(w, "AI 接口请求失败："+err.Error())
			return
		}
		request.Header.Set("Authorization", "Bearer "+channel.APIKey)
		if contentType != "" {
			request.Header.Set("Content-Type", contentType)
		}
		if err := service.ConsumeUserCredits(user.ID, modelName, credits, path); err != nil {
			FailError(w, err)
			return
		}
		copyAIResponse(w, request, func() {
			if err := service.RefundUserCredits(user.ID, modelName, credits, path); err != nil {
				log.Printf("AI proxy refund credits failed: user=%s model=%s credits=%d err=%v", user.ID, modelName, credits, err)
			}
		}, func(status string, responseBody []byte, errMessage string) {
			if _, err := service.SaveGenerationLog(service.BuildGenerationLog(user.ID, path, modelName, body, responseBody, status, errMessage)); err != nil {
				log.Printf("AI proxy save generation log failed: user=%s model=%s err=%v", user.ID, modelName, err)
			}
		})
	}
	if isImageAIPath(path) {
		globalImageTaskQueue.Run(user.ID, displayTaskUsername(user), modelName, run)
		return
	}
	run()
}

func copyAIResponse(w http.ResponseWriter, request *http.Request, onFailure func(), onDone func(status string, responseBody []byte, errMessage string)) {
	response, err := http.DefaultClient.Do(request)
	if err != nil {
		log.Printf("AI proxy request failed: url=%s err=%v", request.URL.String(), err)
		if onFailure != nil {
			onFailure()
		}
		if onDone != nil {
			onDone("network_error", nil, err.Error())
		}
		Fail(w, "AI 接口请求失败："+err.Error())
		return
	}
	defer response.Body.Close()

	if response.StatusCode >= http.StatusBadRequest {
		payload, _ := io.ReadAll(io.LimitReader(response.Body, 4096))
		message := aiUpstreamErrorMessage(response.StatusCode, payload)
		log.Printf("AI upstream error: url=%s status=%d body=%s", request.URL.String(), response.StatusCode, strings.TrimSpace(string(payload)))
		if onFailure != nil {
			onFailure()
		}
		if onDone != nil {
			onDone("failed", payload, message)
		}
		Fail(w, message)
		return
	}

	payload, readErr := io.ReadAll(response.Body)
	if readErr != nil {
		if onFailure != nil {
			onFailure()
		}
		if onDone != nil {
			onDone("read_error", nil, readErr.Error())
		}
		Fail(w, "AI 接口请求失败："+readErr.Error())
		return
	}
	payload = rewritePublicImageURLs(payload)
	if onDone != nil {
		onDone("success", payload, "")
	}

	for key, values := range response.Header {
		if strings.EqualFold(key, "Content-Length") {
			continue
		}
		for _, value := range values {
			w.Header().Add(key, value)
		}
	}
	w.WriteHeader(response.StatusCode)
	_, _ = w.Write(payload)
}

func aiUpstreamErrorMessage(statusCode int, payload []byte) string {
	message := strings.TrimSpace(string(payload))
	var parsed struct {
		Error *struct {
			Message string `json:"message"`
			Code    any    `json:"code"`
			Type    string `json:"type"`
		} `json:"error"`
		Detail any    `json:"detail"`
		Msg    string `json:"msg"`
	}
	if len(payload) > 0 && json.Unmarshal(payload, &parsed) == nil {
		if parsed.Error != nil && strings.TrimSpace(parsed.Error.Message) != "" {
			message = strings.TrimSpace(parsed.Error.Message)
		}
		if message == "" && strings.TrimSpace(parsed.Msg) != "" {
			message = strings.TrimSpace(parsed.Msg)
		}
		if message == "" && parsed.Detail != nil {
			if detailBytes, err := json.Marshal(parsed.Detail); err == nil {
				message = strings.TrimSpace(string(detailBytes))
			}
		}
	}
	if message == "" {
		message = http.StatusText(statusCode)
	}
	if len([]rune(message)) > 500 {
		message = string([]rune(message)[:500]) + "..."
	}
	return fmt.Sprintf("AI 上游返回错误（%d）：%s", statusCode, message)
}

func isImageAIPath(path string) bool {
	return path == "/images/generations" || path == "/images/edits"
}

var imageSubmissionLimiter = struct {
	sync.Mutex
	items map[string][]time.Time
}{items: map[string][]time.Time{}}

func allowImageSubmission(user model.AuthUser) bool {
	if user.Role == model.UserRoleAdmin {
		return true
	}
	nowTime := time.Now()
	windowStart := nowTime.Add(-3 * time.Minute)
	imageSubmissionLimiter.Lock()
	defer imageSubmissionLimiter.Unlock()
	recent := imageSubmissionLimiter.items[user.ID]
	kept := recent[:0]
	for _, item := range recent {
		if item.After(windowStart) {
			kept = append(kept, item)
		}
	}
	const maxImageSubmissionsPerWindow = 3
	if len(kept) >= maxImageSubmissionsPerWindow {
		imageSubmissionLimiter.items[user.ID] = kept
		return false
	}
	imageSubmissionLimiter.items[user.ID] = append(kept, nowTime)
	return true
}

func resetImageSubmissionLimiterForTest() {
	imageSubmissionLimiter.Lock()
	defer imageSubmissionLimiter.Unlock()
	imageSubmissionLimiter.items = map[string][]time.Time{}
}

type imageTaskQueue struct {
	tasks   chan imageTask
	mu      sync.RWMutex
	running *imageTaskInfo
	waiting []imageTaskInfo
}

type imageTask struct {
	info imageTaskInfo
	run  func()
	done chan struct{}
}

type imageTaskInfo struct {
	ID                   string `json:"id"`
	UserID               string `json:"userId"`
	Username             string `json:"username"`
	Model                string `json:"model"`
	Status               string `json:"status"`
	CreatedAt            string `json:"createdAt"`
	StartedAt            string `json:"startedAt,omitempty"`
	EstimatedWaitSeconds int    `json:"estimatedWaitSeconds"`
}

type imageTaskStatus struct {
	Running *imageTaskInfo  `json:"running"`
	Waiting []imageTaskInfo `json:"waiting"`
}

var globalImageTaskQueue = newImageTaskQueue()

func newImageTaskQueue() *imageTaskQueue {
	q := &imageTaskQueue{tasks: make(chan imageTask, 100)}
	go q.worker()
	return q
}

func (q *imageTaskQueue) Run(userID, username, modelName string, run func()) {
	task := imageTask{info: imageTaskInfo{ID: fmt.Sprintf("task_%d", time.Now().UnixNano()), UserID: userID, Username: username, Model: modelName, Status: "waiting", CreatedAt: time.Now().UTC().Format(time.RFC3339)}, run: run, done: make(chan struct{})}
	q.mu.Lock()
	if q.running != nil || len(q.waiting) > 0 {
		task.info.EstimatedWaitSeconds = (len(q.waiting) + 1) * 60
	}
	q.waiting = append(q.waiting, task.info)
	q.mu.Unlock()
	q.tasks <- task
	<-task.done
}

func (q *imageTaskQueue) worker() {
	for task := range q.tasks {
		q.mu.Lock()
		for i, item := range q.waiting {
			if item.ID == task.info.ID {
				q.waiting = append(q.waiting[:i], q.waiting[i+1:]...)
				break
			}
		}
		nowText := time.Now().UTC().Format(time.RFC3339)
		task.info.Status = "running"
		task.info.StartedAt = nowText
		task.info.EstimatedWaitSeconds = 0
		q.running = &task.info
		q.mu.Unlock()
		task.run()
		q.mu.Lock()
		q.running = nil
		q.mu.Unlock()
		close(task.done)
	}
}

func (q *imageTaskQueue) Status() imageTaskStatus {
	q.mu.RLock()
	defer q.mu.RUnlock()
	waiting := append([]imageTaskInfo{}, q.waiting...)
	for i := range waiting {
		waiting[i].EstimatedWaitSeconds = (i + 1) * 60
	}
	var running *imageTaskInfo
	if q.running != nil {
		value := *q.running
		running = &value
	}
	return imageTaskStatus{Running: running, Waiting: waiting}
}

func displayTaskUsername(user model.AuthUser) string {
	if strings.TrimSpace(user.DisplayName) != "" {
		return strings.TrimSpace(user.DisplayName)
	}
	if strings.TrimSpace(user.Username) != "" {
		return strings.TrimSpace(user.Username)
	}
	return "用户"
}

func AIImageTasks(w http.ResponseWriter, r *http.Request) {
	OK(w, globalImageTaskQueue.Status())
}

func normalizeImageRequest(path string, body []byte, contentType string) ([]byte, string) {
	if !isImageAIPath(path) {
		return body, contentType
	}
	if strings.HasPrefix(contentType, "multipart/form-data") {
		return normalizeMultipartImageRequest(body, contentType)
	}
	var payload map[string]any
	if json.Unmarshal(body, &payload) != nil {
		return body, contentType
	}
	payload["response_format"] = "url"
	payload["n"] = 1
	updated, err := json.Marshal(payload)
	if err != nil {
		return body, contentType
	}
	return updated, contentType
}

func normalizeMultipartImageRequest(body []byte, contentType string) ([]byte, string) {
	_, params, err := mime.ParseMediaType(contentType)
	if err != nil {
		return body, contentType
	}
	form, err := multipart.NewReader(bytes.NewReader(body), params["boundary"]).ReadForm(64 << 20)
	if err != nil {
		return body, contentType
	}
	defer form.RemoveAll()
	var buffer bytes.Buffer
	writer := multipart.NewWriter(&buffer)
	for key, values := range form.Value {
		if key == "response_format" || key == "n" {
			continue
		}
		for _, value := range values {
			_ = writer.WriteField(key, value)
		}
	}
	_ = writer.WriteField("response_format", "url")
	_ = writer.WriteField("n", "1")
	for key, files := range form.File {
		for _, fileHeader := range files {
			file, err := fileHeader.Open()
			if err != nil {
				continue
			}
			part, err := writer.CreateFormFile(key, fileHeader.Filename)
			if err == nil {
				_, _ = io.Copy(part, file)
			}
			_ = file.Close()
		}
	}
	if writer.Close() != nil {
		return body, contentType
	}
	return buffer.Bytes(), writer.FormDataContentType()
}

func readAIRequest(r *http.Request) ([]byte, string, string, error) {
	contentType := r.Header.Get("Content-Type")
	body, err := io.ReadAll(r.Body)
	if err != nil {
		return nil, "", "", err
	}
	modelName := ""
	if strings.HasPrefix(contentType, "multipart/form-data") {
		modelName = readMultipartModel(body, contentType)
	} else {
		var payload struct {
			Model string `json:"model"`
		}
		_ = json.Unmarshal(body, &payload)
		modelName = payload.Model
	}
	if strings.TrimSpace(modelName) == "" {
		return nil, "", "", errMissingModel
	}
	return body, contentType, modelName, nil
}

func readMultipartModel(body []byte, contentType string) string {
	_, params, err := mime.ParseMediaType(contentType)
	if err != nil {
		return ""
	}
	reader := multipart.NewReader(bytes.NewReader(body), params["boundary"])
	form, err := reader.ReadForm(32 << 20)
	if err != nil {
		return ""
	}
	defer form.RemoveAll()
	if values := form.Value["model"]; len(values) > 0 {
		return values[0]
	}
	return ""
}

func readAIRequestCount(body []byte, contentType string) int {
	count := 1
	if strings.HasPrefix(contentType, "multipart/form-data") {
		_, params, err := mime.ParseMediaType(contentType)
		if err != nil {
			return count
		}
		form, err := multipart.NewReader(bytes.NewReader(body), params["boundary"]).ReadForm(32 << 20)
		if err != nil {
			return count
		}
		defer form.RemoveAll()
		if values := form.Value["n"]; len(values) > 0 {
			_, _ = fmt.Sscan(values[0], &count)
		}
	} else {
		var payload struct {
			N int `json:"n"`
		}
		_ = json.Unmarshal(body, &payload)
		count = payload.N
	}
	if count < 1 {
		return 1
	}
	return count
}

var errMissingModel = &aiError{"缺少模型名称"}

type aiError struct {
	message string
}

func (err *aiError) Error() string {
	return err.message
}

func rewritePublicImageURLs(payload []byte) []byte {
	base := strings.TrimRight(os.Getenv("PUBLIC_IMAGE_BASE_URL"), "/")
	if base == "" || len(payload) == 0 {
		return payload
	}
	var parsed map[string]any
	if json.Unmarshal(payload, &parsed) != nil {
		return payload
	}
	data, ok := parsed["data"].([]any)
	if !ok {
		return payload
	}
	changed := false
	for _, item := range data {
		obj, ok := item.(map[string]any)
		if !ok {
			continue
		}
		urlValue, ok := obj["url"].(string)
		if !ok || urlValue == "" {
			continue
		}
		if strings.HasPrefix(urlValue, "http://172.17.0.1:3000") {
			obj["url"] = base + strings.TrimPrefix(urlValue, "http://172.17.0.1:3000")
			changed = true
		}
	}
	if !changed {
		return payload
	}
	updated, err := json.Marshal(parsed)
	if err != nil {
		return payload
	}
	return updated
}

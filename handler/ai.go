package handler

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
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
	copyAIResponse(w, request)
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
	batchCount := readAIRequestCount(body, contentType)
	body, contentType, err = normalizeImageRequestStrict(path, body, contentType)
	if err != nil {
		Fail(w, "AI 接口请求失败："+err.Error())
		return
	}
	credits *= batchCount
	if isImageAIPath(path) {
		if !allowImageBatchSubmission(user, batchCount, imageRequestLimitScope(r)) {
			if _, err := service.SaveGenerationLog(service.BuildGenerationLog(user.ID, path, modelName, body, []byte(`{"msg":"图片生成太频繁，请 3 分钟内最多生成 3 张"}`), "rate_limited", "图片生成太频繁，请 3 分钟内最多生成 3 张")); err != nil {
				log.Printf("AI proxy save rate limit generation log failed: user=%s model=%s err=%v", user.ID, modelName, err)
			}
			Fail(w, "图片生成太频繁，请 3 分钟内最多生成 3 张")
			return
		}
		kind := model.GenerationLogKindImage
		task, err := service.CreateGenerationTask(user.ID, kind, modelName, path, batchCount, credits)
		if err != nil {
			FailError(w, err)
			return
		}
		taskID, err := globalImageTaskQueue.SubmitWithID(r.Context(), task.ID, user.ID, displayTaskUsername(user), user.AvatarURL, modelName, batchCount, func(ctx context.Context) imageTaskResult {
			_ = service.MarkGenerationTaskRunning(task.ID)
			status, responseBody, errMessage, failed := executeAIProxyRequest(ctx, user.ID, modelName, path, body, contentType, credits, task.ID, w)
			logID := ""
			if saved, err := service.SaveGenerationLog(service.BuildGenerationLogForTask(task.ID, user.ID, path, modelName, body, responseBody, status, errMessage)); err != nil {
				log.Printf("AI proxy save generation log failed: user=%s task=%s model=%s err=%v", user.ID, task.ID, modelName, err)
			} else {
				logID = saved.ID
			}
			_ = service.CompleteGenerationTask(task.ID, !failed, logID, errMessage)
			return imageTaskResult{Status: status, Error: errMessage}
		})
		if err != nil {
			_ = service.CancelGenerationTask(task.ID, err.Error())
			Fail(w, err.Error())
			return
		}
		if taskID != task.ID {
			log.Printf("image queue task id mismatch: db=%s queue=%s", task.ID, taskID)
		}
		globalImageTaskQueue.Wait(task.ID)
		return
	}
	executeAIProxyRequest(r.Context(), user.ID, modelName, path, body, contentType, credits, "", w)
}

func executeAIProxyRequest(ctx context.Context, userID string, modelName string, path string, body []byte, contentType string, credits int, taskID string, w http.ResponseWriter) (string, []byte, string, bool) {
	batchCount := readAIRequestCount(body, contentType)
	creditPerImage := credits
	if batchCount > 0 {
		creditPerImage = credits / batchCount
	}
	channel, err := service.SelectModelChannel(modelName)
	if err != nil {
		log.Printf("AI proxy select channel failed: model=%s err=%v", modelName, err)
		Fail(w, "AI 接口请求失败："+err.Error())
		return "failed", nil, err.Error(), true
	}
	request, err := http.NewRequestWithContext(ctx, http.MethodPost, service.BuildModelChannelURL(channel, path), bytes.NewReader(body))
	if err != nil {
		log.Printf("AI proxy build request failed: url=%s err=%v", service.BuildModelChannelURL(channel, path), err)
		Fail(w, "AI 接口请求失败："+err.Error())
		return "failed", nil, err.Error(), true
	}
	request.Header.Set("Authorization", "Bearer "+channel.APIKey)
	if contentType != "" {
		request.Header.Set("Content-Type", contentType)
	}
	if err := service.ConsumeUserCreditsForTask(userID, modelName, credits, path, taskID); err != nil {
		FailError(w, err)
		return "failed", nil, err.Error(), true
	}
	status, responseBody, errMessage, failed := copyAIResponse(w, request)
	usage := analyzeImageResponseUsage(path, responseBody, batchCount, creditPerImage)
	if !failed {
		status = usage.Status
		if usage.Error != "" {
			errMessage = usage.Error
		}
		failed = usage.Failed
	}
	refundCredits := refundCreditsForAIResult(failed, usage, credits)
	if refundCredits > 0 {
		if err := service.RefundUserCreditsForTask(userID, modelName, refundCredits, path, taskID); err != nil {
			log.Printf("AI proxy refund credits failed: user=%s task=%s model=%s credits=%d err=%v", userID, taskID, modelName, refundCredits, err)
		}
	}
	if taskID == "" {
		if _, err := service.SaveGenerationLog(service.BuildGenerationLog(userID, path, modelName, body, responseBody, status, errMessage)); err != nil {
			log.Printf("AI proxy save generation log failed: user=%s model=%s err=%v", userID, modelName, err)
		}
	}
	return status, responseBody, errMessage, failed
}

type imageResponseUsage struct {
	Status         string
	Error          string
	ActualImages   int
	ChargedCredits int
	RefundCredits  int
	Partial        bool
	Failed         bool
}

func refundCreditsForAIResult(failed bool, usage imageResponseUsage, originalCredits int) int {
	if usage.RefundCredits > 0 {
		return usage.RefundCredits
	}
	if failed && usage.ActualImages == 0 {
		return originalCredits
	}
	return 0
}

func analyzeImageResponseUsage(path string, responseBody []byte, requestedCount int, creditPerImage int) imageResponseUsage {
	if requestedCount < 1 {
		requestedCount = 1
	}
	if creditPerImage < 0 {
		creditPerImage = 0
	}
	usage := imageResponseUsage{Status: "success", ActualImages: requestedCount, ChargedCredits: requestedCount * creditPerImage}
	if !isImageAIPath(path) {
		return usage
	}
	actualImages := len(service.ExtractImagesForAccounting(responseBody))
	usage.ActualImages = actualImages
	usage.ChargedCredits = actualImages * creditPerImage
	missing := requestedCount - actualImages
	if missing <= 0 {
		return usage
	}
	usage.RefundCredits = missing * creditPerImage
	if actualImages == 0 {
		usage.Status = "failed"
		usage.Failed = true
		usage.Error = fmt.Sprintf("AI 上游未返回图片：请求 %d 张，实际返回 0 张", requestedCount)
		return usage
	}
	usage.Status = "partial_success"
	usage.Partial = true
	usage.Error = fmt.Sprintf("AI 上游少返回图片：请求 %d 张，实际返回 %d 张，已自动退还 %d 点", requestedCount, actualImages, usage.RefundCredits)
	return usage
}

func copyAIResponse(w http.ResponseWriter, request *http.Request) (string, []byte, string, bool) {
	response, err := http.DefaultClient.Do(request)
	if err != nil {
		log.Printf("AI proxy request failed: url=%s err=%v", request.URL.String(), err)
		Fail(w, "AI 接口请求失败："+err.Error())
		return "network_error", nil, err.Error(), true
	}
	defer response.Body.Close()

	if response.StatusCode >= http.StatusBadRequest {
		payload, _ := io.ReadAll(io.LimitReader(response.Body, 4096))
		message := aiUpstreamErrorMessage(response.StatusCode, payload)
		log.Printf("AI upstream error: url=%s status=%d body=%s", request.URL.String(), response.StatusCode, strings.TrimSpace(string(payload)))
		Fail(w, message)
		return "failed", payload, message, true
	}

	payload, readErr := io.ReadAll(response.Body)
	if readErr != nil {
		Fail(w, "AI 接口请求失败："+readErr.Error())
		return "read_error", nil, readErr.Error(), true
	}
	payload = rewritePublicImageURLs(payload)

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
	return "success", payload, "", false
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
	return allowImageBatchSubmission(user, 1)
}

type imageLimitScope string

const imageLimitScopeCanvas imageLimitScope = "canvas"

func imageRequestLimitScope(r *http.Request) imageLimitScope {
	if strings.EqualFold(strings.TrimSpace(r.Header.Get("X-Infinite-Canvas-Scope")), string(imageLimitScopeCanvas)) {
		return imageLimitScopeCanvas
	}
	return ""
}

func allowImageBatchSubmission(user model.AuthUser, batchCount int, scopes ...imageLimitScope) bool {
	if user.Role == model.UserRoleAdmin {
		return true
	}
	for _, scope := range scopes {
		if scope == imageLimitScopeCanvas {
			return true
		}
	}
	if batchCount < 1 {
		batchCount = 1
	}
	if batchCount > 3 {
		return false
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
	if len(kept)+batchCount > maxImageSubmissionsPerWindow {
		imageSubmissionLimiter.items[user.ID] = kept
		return false
	}
	for i := 0; i < batchCount; i++ {
		kept = append(kept, nowTime)
	}
	imageSubmissionLimiter.items[user.ID] = kept
	return true
}

func resetImageSubmissionLimiterForTest() {
	imageSubmissionLimiter.Lock()
	defer imageSubmissionLimiter.Unlock()
	imageSubmissionLimiter.items = map[string][]time.Time{}
}

type imageTaskQueue struct {
	tasks    chan imageTask
	mu       sync.RWMutex
	running  *imageTaskInfo
	waiting  []imageTaskInfo
	recent   []imageTaskInfo
	done     map[string]chan struct{}
	capacity int
}

type imageTask struct {
	info imageTaskInfo
	ctx  context.Context
	run  func(context.Context) imageTaskResult
	done chan struct{}
}

type imageTaskResult struct {
	Status string
	Error  string
}

type imageTaskInfo struct {
	ID                   string `json:"id"`
	UserID               string `json:"userId"`
	Username             string `json:"username,omitempty"`
	AvatarURL            string `json:"avatarUrl,omitempty"`
	Model                string `json:"model"`
	Status               string `json:"status"`
	CreatedAt            string `json:"createdAt"`
	StartedAt            string `json:"startedAt,omitempty"`
	CompletedAt          string `json:"completedAt,omitempty"`
	EstimatedWaitSeconds int    `json:"estimatedWaitSeconds"`
	BatchCount           int    `json:"batchCount"`
	Error                string `json:"error,omitempty"`
}

type imageTaskStatus struct {
	Running *imageTaskInfo  `json:"running"`
	Waiting []imageTaskInfo `json:"waiting"`
	Recent  []imageTaskInfo `json:"recent"`
}

var globalImageTaskQueue = newImageTaskQueue()
var errImageTaskQueueFull = errors.New("全站生图队列已满，请稍后再试")

func newImageTaskQueue() *imageTaskQueue { return newImageTaskQueueWithCapacity(100) }

func newImageTaskQueueWithCapacity(capacity int) *imageTaskQueue {
	if capacity < 1 {
		capacity = 1
	}
	q := &imageTaskQueue{tasks: make(chan imageTask, capacity), capacity: capacity, done: map[string]chan struct{}{}}
	go q.worker()
	return q
}

func (q *imageTaskQueue) Submit(ctx context.Context, userID, username, modelName string, batchCount int, run func(context.Context) imageTaskResult) (string, error) {
	return q.SubmitWithID(ctx, fmt.Sprintf("task_%d", time.Now().UnixNano()), userID, username, "", modelName, batchCount, run)
}

func (q *imageTaskQueue) SubmitWithID(ctx context.Context, taskID, userID, username, avatarURL, modelName string, batchCount int, run func(context.Context) imageTaskResult) (string, error) {
	if ctx == nil {
		ctx = context.Background()
	}
	if batchCount < 1 {
		batchCount = 1
	}
	if strings.TrimSpace(taskID) == "" {
		taskID = fmt.Sprintf("task_%d", time.Now().UnixNano())
	}
	task := imageTask{info: imageTaskInfo{ID: taskID, UserID: userID, Username: username, AvatarURL: strings.TrimSpace(avatarURL), Model: modelName, Status: "waiting", CreatedAt: time.Now().UTC().Format(time.RFC3339), BatchCount: batchCount}, ctx: ctx, run: run, done: make(chan struct{})}
	q.mu.Lock()
	if q.running != nil || len(q.waiting) > 0 {
		task.info.EstimatedWaitSeconds = (len(q.waiting) + 1) * 60
	}
	if len(q.waiting) >= q.capacity {
		q.mu.Unlock()
		return "", errImageTaskQueueFull
	}
	q.waiting = append(q.waiting, task.info)
	q.done[task.info.ID] = task.done
	q.mu.Unlock()
	select {
	case q.tasks <- task:
		go q.cancelWaitingOnContext(task.info.ID, ctx)
		return task.info.ID, nil
	default:
		q.cancelTask(task.info.ID, "queue_full")
		return "", errImageTaskQueueFull
	}
}

func (q *imageTaskQueue) Wait(id string) {
	q.mu.RLock()
	done := q.done[id]
	q.mu.RUnlock()
	if done != nil {
		<-done
	}
}

func (q *imageTaskQueue) cancelWaitingOnContext(id string, ctx context.Context) {
	<-ctx.Done()
	q.cancelTask(id, "cancelled")
}

func (q *imageTaskQueue) cancelTask(id string, reason string) bool {
	q.mu.Lock()
	defer q.mu.Unlock()
	for i, item := range q.waiting {
		if item.ID == id {
			q.waiting = append(q.waiting[:i], q.waiting[i+1:]...)
			item.Status = "cancelled"
			item.Error = reason
			item.CompletedAt = time.Now().UTC().Format(time.RFC3339)
			q.addRecentLocked(item)
			if done := q.done[id]; done != nil {
				close(done)
				delete(q.done, id)
			}
			return true
		}
	}
	return false
}

func (q *imageTaskQueue) worker() {
	for task := range q.tasks {
		q.mu.Lock()
		cancelled := true
		for i, item := range q.waiting {
			if item.ID == task.info.ID {
				q.waiting = append(q.waiting[:i], q.waiting[i+1:]...)
				cancelled = false
				break
			}
		}
		if cancelled {
			q.mu.Unlock()
			continue
		}
		nowText := time.Now().UTC().Format(time.RFC3339)
		task.info.Status = "running"
		task.info.StartedAt = nowText
		task.info.EstimatedWaitSeconds = 0
		q.running = &task.info
		q.mu.Unlock()
		result := task.run(task.ctx)
		q.mu.Lock()
		completed := task.info
		completed.Status = result.Status
		if completed.Status == "" {
			completed.Status = "success"
		}
		completed.Error = result.Error
		completed.CompletedAt = time.Now().UTC().Format(time.RFC3339)
		q.running = nil
		q.addRecentLocked(completed)
		if done := q.done[task.info.ID]; done != nil {
			close(done)
			delete(q.done, task.info.ID)
		}
		q.mu.Unlock()
	}
}

func (q *imageTaskQueue) addRecentLocked(item imageTaskInfo) {
	q.recent = append([]imageTaskInfo{item}, q.recent...)
	if len(q.recent) > 20 {
		q.recent = q.recent[:20]
	}
}

func (q *imageTaskQueue) Status() imageTaskStatus {
	q.mu.RLock()
	defer q.mu.RUnlock()
	waiting := append([]imageTaskInfo{}, q.waiting...)
	for i := range waiting {
		waiting[i].EstimatedWaitSeconds = (i + 1) * 60
		waiting[i].Username = ""
	}
	var running *imageTaskInfo
	if q.running != nil {
		value := *q.running
		value.Username = ""
		running = &value
	}
	recent := append([]imageTaskInfo{}, q.recent...)
	for i := range recent {
		recent[i].Username = ""
	}
	return imageTaskStatus{Running: running, Waiting: waiting, Recent: recent}
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

func AIImageStats(w http.ResponseWriter, r *http.Request) {
	result, err := service.GenerationImageStats(time.Now().Format("2006-01-02"))
	if err != nil {
		FailError(w, err)
		return
	}
	for i := range result.UserRanks {
		result.UserRanks[i].Username = ""
	}
	OK(w, result)
}

func maskRankingName(name string) string {
	name = strings.TrimSpace(name)
	if name == "" || name == "-" {
		return name
	}
	runes := []rune(name)
	if len(runes) <= 4 {
		return name
	}
	return string(runes[:4])
}

func AIImageHistory(w http.ResponseWriter, r *http.Request) {
	user, ok := service.UserFromContext(r.Context())
	if !ok {
		Fail(w, "未登录或权限不足")
		return
	}
	result, err := service.ListUserGenerationLogs(user.ID, parseQuery(r))
	if err != nil {
		FailError(w, err)
		return
	}
	OK(w, result)
}

func normalizeImageRequest(path string, body []byte, contentType string) ([]byte, string) {
	updatedBody, updatedType, err := normalizeImageRequestStrict(path, body, contentType)
	if err != nil {
		return body, contentType
	}
	return updatedBody, updatedType
}

func normalizeImageRequestStrict(path string, body []byte, contentType string) ([]byte, string, error) {
	if !isImageAIPath(path) {
		return body, contentType, nil
	}
	if strings.HasPrefix(contentType, "multipart/form-data") {
		return normalizeMultipartImageRequest(body, contentType)
	}
	var payload map[string]any
	if json.Unmarshal(body, &payload) != nil {
		return body, contentType, nil
	}
	payload["response_format"] = "url"
	if _, ok := payload["n"]; !ok {
		payload["n"] = 1
	}
	updated, err := json.Marshal(payload)
	if err != nil {
		return body, contentType, err
	}
	return updated, contentType, nil
}

func normalizeMultipartImageRequest(body []byte, contentType string) ([]byte, string, error) {
	_, params, err := mime.ParseMediaType(contentType)
	if err != nil {
		return body, contentType, nil
	}
	form, err := multipart.NewReader(bytes.NewReader(body), params["boundary"]).ReadForm(64 << 20)
	if err != nil {
		return body, contentType, nil
	}
	defer form.RemoveAll()
	var buffer bytes.Buffer
	writer := multipart.NewWriter(&buffer)
	for key, values := range form.Value {
		if key == "response_format" {
			continue
		}
		for _, value := range values {
			_ = writer.WriteField(key, value)
		}
	}
	if len(form.Value["n"]) == 0 {
		_ = writer.WriteField("n", "1")
	}
	_ = writer.WriteField("response_format", "url")
	for key, files := range form.File {
		for _, fileHeader := range files {
			if fileHeader.Size == 0 {
				_ = writer.Close()
				return nil, "", fmt.Errorf("图片文件为空：%s", fileHeader.Filename)
			}
			file, err := fileHeader.Open()
			if err != nil {
				_ = writer.Close()
				return nil, "", err
			}
			part, err := writer.CreateFormFile(key, fileHeader.Filename)
			if err != nil {
				_ = file.Close()
				_ = writer.Close()
				return nil, "", err
			}
			written, copyErr := io.Copy(part, file)
			_ = file.Close()
			if copyErr != nil {
				_ = writer.Close()
				return nil, "", copyErr
			}
			if written == 0 {
				_ = writer.Close()
				return nil, "", fmt.Errorf("图片文件为空：%s", fileHeader.Filename)
			}
		}
	}
	if err := writer.Close(); err != nil {
		return body, contentType, err
	}
	return buffer.Bytes(), writer.FormDataContentType(), nil
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

// internalImageHostPrefixes is the list of internal address prefixes that should
// be rewritten to the public base URL before sending image URLs to the frontend.
// Covers docker0 gateway, loopback, and IPv6 loopback — all pointing to the
// chatgpt2api container's /images/ path inside the same host.
var internalImageHostPrefixes = []string{
	"http://172.17.0.1:3000",
	"http://127.0.0.1:3000",
	"http://localhost:3000",
	"http://[::1]:3000",
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
		for _, prefix := range internalImageHostPrefixes {
			if strings.HasPrefix(urlValue, prefix) {
				obj["url"] = base + strings.TrimPrefix(urlValue, prefix)
				changed = true
				break
			}
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

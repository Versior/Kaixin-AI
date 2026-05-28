package handler

import (
	"context"
	"sync"
	"testing"
	"time"

	"github.com/basketikun/infinite-canvas/model"
)

func TestImageSubmissionLimiterIsPerUserAndAdminUnlimited(t *testing.T) {
	resetImageSubmissionLimiterForTest()
	for i := 0; i < 3; i++ {
		if !allowImageSubmission(model.AuthUser{ID: "user-a", Role: model.UserRoleUser}) {
			t.Fatalf("user-a submission %d should be allowed", i+1)
		}
	}
	if allowImageSubmission(model.AuthUser{ID: "user-a", Role: model.UserRoleUser}) {
		t.Fatal("user-a fourth submission in three minutes should be blocked")
	}
	if !allowImageSubmission(model.AuthUser{ID: "user-b", Role: model.UserRoleUser}) {
		t.Fatal("user-b should not be blocked by user-a submissions")
	}
	for i := 0; i < 6; i++ {
		if !allowImageSubmission(model.AuthUser{ID: "admin", Role: model.UserRoleAdmin}) {
			t.Fatalf("admin submission %d should bypass limiter", i+1)
		}
	}
}

func TestImageSubmissionLimiterCountsBatchAsOneSubmission(t *testing.T) {
	resetImageSubmissionLimiterForTest()
	user := model.AuthUser{ID: "user-a", Role: model.UserRoleUser}
	for i := 0; i < 3; i++ {
		if !allowImageBatchSubmission(user, 3) {
			t.Fatalf("batch submission %d should be allowed", i+1)
		}
	}
	if allowImageBatchSubmission(user, 1) {
		t.Fatal("fourth batch submission in three minutes should be blocked")
	}
}

func TestImageTaskQueueSubmitsBatchAsOneAsyncTask(t *testing.T) {
	queue := newImageTaskQueueWithCapacity(2)
	started := make(chan struct{})
	release := make(chan struct{})
	taskID, err := queue.Submit(context.Background(), "u1", "张三", "gpt-image-2", 3, func(context.Context) imageTaskResult {
		close(started)
		<-release
		return imageTaskResult{Status: "success"}
	})
	if err != nil {
		t.Fatalf("submit should succeed: %v", err)
	}
	if taskID == "" {
		t.Fatal("submit should return task id")
	}
	<-started
	status := queue.Status()
	if status.Running == nil || status.Running.ID != taskID || status.Running.BatchCount != 3 || len(status.Waiting) != 0 {
		t.Fatalf("expected one running batch task, got %#v", status)
	}
	close(release)
	waitTaskStatus(t, queue, taskID, "success")
}

func TestImageTaskQueueRejectsWhenFull(t *testing.T) {
	queue := newImageTaskQueueWithCapacity(1)
	started := make(chan struct{})
	release := make(chan struct{})
	_, err := queue.Submit(context.Background(), "u1", "张三", "gpt-image-2", 1, func(context.Context) imageTaskResult {
		close(started)
		<-release
		return imageTaskResult{Status: "success"}
	})
	if err != nil {
		t.Fatalf("first submit should succeed: %v", err)
	}
	<-started
	if _, err := queue.Submit(context.Background(), "u2", "李四", "gpt-image-2", 1, func(context.Context) imageTaskResult { return imageTaskResult{Status: "success"} }); err != nil {
		t.Fatalf("waiting slot should accept one task: %v", err)
	}
	if _, err := queue.Submit(context.Background(), "u3", "王五", "gpt-image-2", 1, func(context.Context) imageTaskResult { return imageTaskResult{Status: "success"} }); err == nil {
		t.Fatal("full queue should reject immediately")
	}
	close(release)
}

func TestImageTaskQueueCanCancelWaitingTask(t *testing.T) {
	queue := newImageTaskQueueWithCapacity(2)
	started := make(chan struct{})
	release := make(chan struct{})
	_, err := queue.Submit(context.Background(), "u1", "张三", "gpt-image-2", 1, func(context.Context) imageTaskResult {
		close(started)
		<-release
		return imageTaskResult{Status: "success"}
	})
	if err != nil {
		t.Fatal(err)
	}
	<-started
	ctx, cancel := context.WithCancel(context.Background())
	secondID, err := queue.Submit(ctx, "u2", "李四", "gpt-image-2", 1, func(context.Context) imageTaskResult {
		t.Fatal("cancelled waiting task should not run")
		return imageTaskResult{}
	})
	if err != nil {
		t.Fatal(err)
	}
	cancel()
	waitTaskStatus(t, queue, secondID, "cancelled")
	if status := queue.Status(); len(status.Waiting) != 0 {
		t.Fatalf("cancelled waiting task should be removed, got %#v", status.Waiting)
	}
	close(release)
}

func TestImageTaskQueueRunsOneAtATimeAndReportsWait(t *testing.T) {
	queue := newImageTaskQueue()
	firstStarted := make(chan struct{})
	releaseFirst := make(chan struct{})
	var mu sync.Mutex
	order := []string{}
	firstID, err := queue.Submit(context.Background(), "u1", "张三", "gpt-image-2", 1, func(context.Context) imageTaskResult {
		mu.Lock()
		order = append(order, "first")
		mu.Unlock()
		close(firstStarted)
		<-releaseFirst
		return imageTaskResult{Status: "success"}
	})
	if err != nil {
		t.Fatal(err)
	}
	<-firstStarted
	if status := queue.Status(); status.Running == nil || status.Running.ID != firstID || len(status.Waiting) != 0 {
		t.Fatalf("expected one running task, got %#v", status)
	}
	secondDone := make(chan struct{})
	secondID, err := queue.Submit(context.Background(), "u2", "李四", "gpt-image-2", 1, func(context.Context) imageTaskResult {
		mu.Lock()
		order = append(order, "second")
		mu.Unlock()
		close(secondDone)
		return imageTaskResult{Status: "success"}
	})
	if err != nil {
		t.Fatal(err)
	}
	time.Sleep(20 * time.Millisecond)
	status := queue.Status()
	if status.Running == nil || status.Running.Username != "张三" || len(status.Waiting) != 1 || status.Waiting[0].ID != secondID || status.Waiting[0].Username != "李四" || status.Waiting[0].EstimatedWaitSeconds <= 0 {
		t.Fatalf("unexpected queued status: %#v", status)
	}
	select {
	case <-secondDone:
		t.Fatal("second task ran before first completed")
	default:
	}
	close(releaseFirst)
	select {
	case <-secondDone:
	case <-time.After(time.Second):
		t.Fatal("second task did not run after first completed")
	}
	mu.Lock()
	defer mu.Unlock()
	if len(order) != 2 || order[0] != "first" || order[1] != "second" {
		t.Fatalf("unexpected run order: %v", order)
	}
}

func waitTaskStatus(t *testing.T, queue *imageTaskQueue, taskID string, want string) {
	t.Helper()
	deadline := time.After(time.Second)
	for {
		select {
		case <-deadline:
			t.Fatalf("task %s did not reach status %s; status=%#v", taskID, want, queue.Status())
		default:
		}
		status := queue.Status()
		for _, task := range status.Recent {
			if task.ID == taskID && task.Status == want {
				return
			}
		}
		time.Sleep(10 * time.Millisecond)
	}
}

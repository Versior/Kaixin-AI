package handler

import (
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

func TestImageTaskQueueRunsOneAtATimeAndReportsWait(t *testing.T) {
	queue := newImageTaskQueue()
	firstStarted := make(chan struct{})
	releaseFirst := make(chan struct{})
	var mu sync.Mutex
	order := []string{}
	go queue.Run("u1", "张三", "gpt-image-2", func() {
		mu.Lock()
		order = append(order, "first")
		mu.Unlock()
		close(firstStarted)
		<-releaseFirst
	})
	<-firstStarted
	if status := queue.Status(); status.Running == nil || len(status.Waiting) != 0 {
		t.Fatalf("expected one running task, got %#v", status)
	}
	secondDone := make(chan struct{})
	go func() {
		queue.Run("u2", "李四", "gpt-image-2", func() {
			mu.Lock()
			order = append(order, "second")
			mu.Unlock()
		})
		close(secondDone)
	}()
	time.Sleep(20 * time.Millisecond)
	status := queue.Status()
	if status.Running == nil || status.Running.Username != "张三" || len(status.Waiting) != 1 || status.Waiting[0].Username != "李四" || status.Waiting[0].EstimatedWaitSeconds <= 0 {
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

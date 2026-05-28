package handler

import "testing"

func TestImageResponseUsageDetectsPartialSuccess(t *testing.T) {
	payload := []byte(`{"data":[{"url":"https://example.com/a.png"},{"url":"https://example.com/b.png"}]}`)
	usage := analyzeImageResponseUsage("/images/generations", payload, 3, 3)
	if usage.ActualImages != 2 {
		t.Fatalf("actual images = %d, want 2", usage.ActualImages)
	}
	if usage.ChargedCredits != 6 {
		t.Fatalf("charged credits = %d, want 6", usage.ChargedCredits)
	}
	if usage.RefundCredits != 3 {
		t.Fatalf("refund credits = %d, want 3", usage.RefundCredits)
	}
	if usage.Status != "partial_success" || !usage.Partial || usage.Failed {
		t.Fatalf("unexpected status: %#v", usage)
	}
}

func TestImageResponseUsageTreatsEmptySuccessAsFailure(t *testing.T) {
	payload := []byte(`{"data":[]}`)
	usage := analyzeImageResponseUsage("/images/generations", payload, 3, 3)
	if !usage.Failed || usage.Status != "failed" {
		t.Fatalf("empty data should be failure: %#v", usage)
	}
	if usage.RefundCredits != 9 || usage.ChargedCredits != 0 {
		t.Fatalf("unexpected credit accounting: %#v", usage)
	}
}

func TestImageResponseUsageKeepsNonImageSuccessUnchanged(t *testing.T) {
	usage := analyzeImageResponseUsage("/chat/completions", []byte(`{"choices":[{}]}`), 3, 3)
	if usage.Status != "success" || usage.ActualImages != 3 || usage.RefundCredits != 0 || usage.ChargedCredits != 9 {
		t.Fatalf("non image response should be unchanged: %#v", usage)
	}
}

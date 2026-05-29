package handler

import (
	"bytes"
	"mime/multipart"
	"strings"
	"testing"
)

func TestNormalizeMultipartImageRequestKeepsOriginalN(t *testing.T) {
	body, contentType := buildMultipartImageRequest(t, map[string]string{"model": "gpt-image-2", "n": "3", "prompt": "hello"})
	updatedBody, updatedType := normalizeImageRequest("/images/generations", body, contentType)
	count := readAIRequestCount(updatedBody, updatedType)
	if count != 3 {
		t.Fatalf("expected normalized multipart request to keep n=3, got %d", count)
	}
}

func TestNormalizeJSONImageRequestKeepsOriginalN(t *testing.T) {
	body := []byte(`{"model":"gpt-image-2","n":3,"prompt":"hello"}`)
	updatedBody, contentType := normalizeImageRequest("/images/generations", body, "application/json")
	count := readAIRequestCount(updatedBody, contentType)
	if count != 3 {
		t.Fatalf("expected normalized json request to keep n=3, got %d body=%s", count, string(updatedBody))
	}
}

func TestNormalizeMultipartImageRequestRejectsEmptyUploadedFile(t *testing.T) {
	var buf bytes.Buffer
	writer := multipart.NewWriter(&buf)
	_ = writer.WriteField("model", "gpt-image-2")
	part, err := writer.CreateFormFile("image", "empty.png")
	if err != nil {
		t.Fatal(err)
	}
	_, _ = part.Write(nil)
	if err := writer.Close(); err != nil {
		t.Fatal(err)
	}
	_, _, err = normalizeImageRequestStrict("/images/edits", buf.Bytes(), writer.FormDataContentType())
	if err == nil || !strings.Contains(err.Error(), "图片文件为空") {
		t.Fatalf("expected empty image file error, got %v", err)
	}
}

func buildMultipartImageRequest(t *testing.T, fields map[string]string) ([]byte, string) {
	t.Helper()
	var buf bytes.Buffer
	writer := multipart.NewWriter(&buf)
	for key, value := range fields {
		if err := writer.WriteField(key, value); err != nil {
			t.Fatal(err)
		}
	}
	if err := writer.Close(); err != nil {
		t.Fatal(err)
	}
	return buf.Bytes(), writer.FormDataContentType()
}

func TestRewritePublicImageURLsRewritesDocker0Address(t *testing.T) {
	t.Setenv("PUBLIC_IMAGE_BASE_URL", "http://example.com:3000")
	payload := []byte(`{"data":[{"url":"http://172.17.0.1:3000/images/2026/05/29/abc.png"}]}`)
	result := rewritePublicImageURLs(payload)
	expected := `{"data":[{"url":"http://example.com:3000/images/2026/05/29/abc.png"}]}`
	if string(result) != expected {
		t.Fatalf("expected %s, got %s", expected, string(result))
	}
}

func TestRewritePublicImageURLsRewritesLoopbackAddress(t *testing.T) {
	t.Setenv("PUBLIC_IMAGE_BASE_URL", "http://example.com:3000")
	payload := []byte(`{"data":[{"url":"http://127.0.0.1:3000/images/2026/05/29/abc.png"}]}`)
	result := rewritePublicImageURLs(payload)
	expected := `{"data":[{"url":"http://example.com:3000/images/2026/05/29/abc.png"}]}`
	if string(result) != expected {
		t.Fatalf("expected %s, got %s", expected, string(result))
	}
}

func TestRewritePublicImageURLsRewritesLocalhostAddress(t *testing.T) {
	t.Setenv("PUBLIC_IMAGE_BASE_URL", "http://example.com:3000")
	payload := []byte(`{"data":[{"url":"http://localhost:3000/images/2026/05/29/abc.png"}]}`)
	result := rewritePublicImageURLs(payload)
	expected := `{"data":[{"url":"http://example.com:3000/images/2026/05/29/abc.png"}]}`
	if string(result) != expected {
		t.Fatalf("expected %s, got %s", expected, string(result))
	}
}

func TestRewritePublicImageURLsNoRewriteWhenEnvEmpty(t *testing.T) {
	t.Setenv("PUBLIC_IMAGE_BASE_URL", "")
	payload := []byte(`{"data":[{"url":"http://172.17.0.1:3000/images/abc.png"}]}`)
	result := rewritePublicImageURLs(payload)
	if string(result) != string(payload) {
		t.Fatalf("expected no rewrite when PUBLIC_IMAGE_BASE_URL is empty, got %s", string(result))
	}
}

func TestRewritePublicImageURLsNoRewriteForPublicURL(t *testing.T) {
	t.Setenv("PUBLIC_IMAGE_BASE_URL", "http://example.com:3000")
	payload := []byte(`{"data":[{"url":"http://example.com:3000/images/abc.png"}]}`)
	result := rewritePublicImageURLs(payload)
	if string(result) != string(payload) {
		t.Fatalf("expected no rewrite for already-public URL, got %s", string(result))
	}
}

func TestRewritePublicImageURLsHandlesMultipleImages(t *testing.T) {
	t.Setenv("PUBLIC_IMAGE_BASE_URL", "http://example.com:3000")
	payload := []byte(`{"data":[{"url":"http://172.17.0.1:3000/images/a.png"},{"url":"http://127.0.0.1:3000/images/b.png"}]}`)
	result := rewritePublicImageURLs(payload)
	expected := `{"data":[{"url":"http://example.com:3000/images/a.png"},{"url":"http://example.com:3000/images/b.png"}]}`
	if string(result) != expected {
		t.Fatalf("expected %s, got %s", expected, string(result))
	}
}

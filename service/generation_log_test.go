package service

import "testing"

func TestExtractImagesForAccountingFindsNestedImagePayloads(t *testing.T) {
	body := []byte(`{
		"output": [
			{"type":"image", "image_url":"https://cdn.example.com/first.png"},
			{"content":[{"image":"data:image/webp;base64,AAAA"}]},
			{"attachments":[{"base64":"BBBB"}]}
		]
	}`)

	images := ExtractImagesForAccounting(body)
	if len(images) != 3 {
		t.Fatalf("expected 3 nested images, got %d: %#v", len(images), images)
	}
	if images[0] != "https://cdn.example.com/first.png" {
		t.Fatalf("expected image_url to be extracted, got %q", images[0])
	}
	if images[1] != "data:image/webp;base64,AAAA" {
		t.Fatalf("expected data url image to be preserved, got %q", images[1])
	}
	if images[2] != "data:image/png;base64,BBBB" {
		t.Fatalf("expected raw base64 to become png data url, got %q", images[2])
	}
}

func TestExtractImagesForAccountingKeepsOpenAIStylePayloads(t *testing.T) {
	body := []byte(`{"data":[{"url":"https://cdn.example.com/a.png"},{"b64_json":"CCCC"}]}`)

	images := ExtractImagesForAccounting(body)
	if len(images) != 2 {
		t.Fatalf("expected 2 OpenAI-style images, got %d: %#v", len(images), images)
	}
	if images[0] != "https://cdn.example.com/a.png" || images[1] != "data:image/png;base64,CCCC" {
		t.Fatalf("unexpected images: %#v", images)
	}
}

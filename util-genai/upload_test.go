// Copyright (c) 2024 Alibaba Group Holding Ltd.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package utilgenai

import (
	"context"
	"os"
	"path/filepath"
	"sync"
	"testing"
)

func sampleCompletionParams() CompletionParams {
	return CompletionParams{
		InputMessages: []InputMessage{{
			Role:  "user",
			Parts: []MessagePart{Text{Content: "Hello"}},
		}},
		OutputMessages: []OutputMessage{{
			Role:         "assistant",
			Parts:        []MessagePart{Text{Content: "Hi there"}},
			FinishReason: FinishReasonStop,
		}},
		SystemInstruction: []MessagePart{Text{Content: "Be concise"}},
		ToolDefinitions: []FunctionToolDefinition{{
			Name:        "get_weather",
			Description: "Return the weather",
		}},
	}
}

func TestFSUploaderWritesFile(t *testing.T) {
	dir := t.TempDir()
	u := NewFSUploader("file://" + dir)

	if err := u.Upload(context.Background(), "sub/ref.json", []byte("payload")); err != nil {
		t.Fatalf("upload failed: %v", err)
	}

	got, err := os.ReadFile(filepath.Join(dir, "sub", "ref.json"))
	if err != nil {
		t.Fatalf("read back failed: %v", err)
	}
	if string(got) != "payload" {
		t.Fatalf("unexpected content: %q", got)
	}
}

func TestUploadCompletionHookStampsAndUploads(t *testing.T) {
	dir := t.TempDir()
	hook := NewUploadCompletionHook(WithUploader(NewFSUploader(dir)))
	if hook == nil {
		t.Fatal("expected non-nil hook")
	}

	hook.OnCompletion(context.Background(), sampleCompletionParams())

	if err := hook.Shutdown(context.Background()); err != nil {
		t.Fatalf("shutdown failed: %v", err)
	}

	entries, err := os.ReadDir(dir)
	if err != nil {
		t.Fatalf("read dir failed: %v", err)
	}
	// One file per populated category: input, output, system, tools.
	if len(entries) != 4 {
		t.Fatalf("expected 4 uploaded files, got %d", len(entries))
	}
}

func TestUploadCompletionHookDedup(t *testing.T) {
	dir := t.TempDir()
	hook := NewUploadCompletionHook(WithUploader(NewFSUploader(dir)))
	if hook == nil {
		t.Fatal("expected non-nil hook")
	}

	// Same payload twice: content-addressed dedup should collapse to 4 files.
	hook.OnCompletion(context.Background(), sampleCompletionParams())
	hook.OnCompletion(context.Background(), sampleCompletionParams())

	if err := hook.Shutdown(context.Background()); err != nil {
		t.Fatalf("shutdown failed: %v", err)
	}

	entries, err := os.ReadDir(dir)
	if err != nil {
		t.Fatalf("read dir failed: %v", err)
	}
	if len(entries) != 4 {
		t.Fatalf("expected 4 deduplicated files, got %d", len(entries))
	}
}

// countingUploader records how many times Upload is called.
type countingUploader struct {
	mu    sync.Mutex
	count int
}

func (c *countingUploader) Upload(context.Context, string, []byte) error {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.count++
	return nil
}

func (c *countingUploader) calls() int {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.count
}

func TestUploadCompletionHookSkipsAfterShutdown(t *testing.T) {
	cu := &countingUploader{}
	hook := NewUploadCompletionHook(WithUploader(cu))

	if err := hook.Shutdown(context.Background()); err != nil {
		t.Fatalf("shutdown failed: %v", err)
	}

	// Enqueue after shutdown must not panic or upload.
	hook.OnCompletion(context.Background(), sampleCompletionParams())

	if got := cu.calls(); got != 0 {
		t.Fatalf("expected no uploads after shutdown, got %d", got)
	}
}

func TestNewUploadCompletionHookNilWithoutConfig(t *testing.T) {
	t.Setenv(EnvUploadBasePath, "")
	if hook := NewUploadCompletionHook(); hook != nil {
		t.Fatal("expected nil hook when no uploader and no base path configured")
	}
}

func TestNewUploadCompletionHookFromEnv(t *testing.T) {
	dir := t.TempDir()
	t.Setenv(EnvUploadBasePath, dir)

	hook := NewUploadCompletionHook()
	if hook == nil {
		t.Fatal("expected hook built from base path env var")
	}
	_ = hook.Shutdown(context.Background())
}

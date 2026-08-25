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
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
)

// Uploader stores serialized message content under a reference key. Implementations
// must be safe for concurrent use. The ref is a relative, filesystem-friendly key
// (e.g. "a1b2c3.json"); the uploader decides the concrete storage location.
type Uploader interface {
	Upload(ctx context.Context, ref string, data []byte) error
}

// FSUploader is an Uploader that writes content to the local filesystem under a
// base directory. A "file://" scheme prefix on the base path is accepted and
// stripped.
type FSUploader struct {
	basePath string
}

// NewFSUploader creates an FSUploader rooted at basePath.
func NewFSUploader(basePath string) *FSUploader {
	return &FSUploader{basePath: strings.TrimPrefix(basePath, "file://")}
}

// Upload writes data to {basePath}/{ref}, creating parent directories as needed.
func (u *FSUploader) Upload(_ context.Context, ref string, data []byte) error {
	full := filepath.Join(u.basePath, filepath.FromSlash(ref))
	if err := os.MkdirAll(filepath.Dir(full), 0o755); err != nil {
		return err
	}
	return os.WriteFile(full, data, 0o644)
}

// uploadJob is a single enqueued upload.
type uploadJob struct {
	ref  string
	data []byte
}

// UploadCompletionHook is a CompletionHook that serializes prompt and response
// content, stamps reference attributes on the span, and uploads the content
// asynchronously through an Uploader. Content is content-addressed by SHA-256,
// so identical payloads (such as a stable system instruction) are uploaded once.
type UploadCompletionHook struct {
	uploader Uploader
	format   UploadFormat

	jobs chan uploadJob
	wg   sync.WaitGroup

	mu     sync.RWMutex
	closed bool

	seen *boundedSet
}

// UploadHookOption configures an UploadCompletionHook.
type UploadHookOption func(*uploadHookConfig)

type uploadHookConfig struct {
	uploader  Uploader
	format    UploadFormat
	queueSize int
	workers   int
	dedupCap  int
}

// WithUploader sets a custom Uploader (e.g. a cloud storage backend).
func WithUploader(u Uploader) UploadHookOption {
	return func(c *uploadHookConfig) { c.uploader = u }
}

// WithUploadFormat overrides the serialization format.
func WithUploadFormat(f UploadFormat) UploadHookOption {
	return func(c *uploadHookConfig) { c.format = f }
}

// WithUploadQueueSize overrides the maximum number of queued uploads.
func WithUploadQueueSize(n int) UploadHookOption {
	return func(c *uploadHookConfig) {
		if n > 0 {
			c.queueSize = n
		}
	}
}

const defaultDedupCap = 1024

// NewUploadCompletionHook creates an UploadCompletionHook. When no uploader is
// supplied via WithUploader, an FSUploader is created from
// OTEL_INSTRUMENTATION_GENAI_UPLOAD_BASE_PATH; if that is also unset, it returns
// nil so callers can fall back to a no-op hook.
func NewUploadCompletionHook(opts ...UploadHookOption) *UploadCompletionHook {
	cfg := &uploadHookConfig{
		format:    GetUploadFormat(),
		queueSize: GetUploadMaxQueueSize(),
		dedupCap:  defaultDedupCap,
	}
	for _, opt := range opts {
		opt(cfg)
	}

	if cfg.uploader == nil {
		basePath := GetUploadBasePath()
		if basePath == "" {
			return nil
		}
		cfg.uploader = NewFSUploader(basePath)
	}

	if cfg.workers <= 0 {
		cfg.workers = cfg.queueSize
		if cfg.workers > 8 {
			cfg.workers = 8
		}
	}

	h := &UploadCompletionHook{
		uploader: cfg.uploader,
		format:   cfg.format,
		jobs:     make(chan uploadJob, cfg.queueSize),
		seen:     newBoundedSet(cfg.dedupCap),
	}

	h.wg.Add(cfg.workers)
	for i := 0; i < cfg.workers; i++ {
		go h.worker()
	}
	return h
}

func (h *UploadCompletionHook) worker() {
	defer h.wg.Done()
	for job := range h.jobs {
		if err := h.uploader.Upload(context.Background(), job.ref, job.data); err != nil {
			otel.Handle(fmt.Errorf("util-genai: upload %q failed: %w", job.ref, err))
		}
	}
}

// OnCompletion serializes each populated content category, stamps a reference
// attribute on the span, and enqueues the content for asynchronous upload.
func (h *UploadCompletionHook) OnCompletion(_ context.Context, params CompletionParams) {
	if h == nil {
		return
	}

	if len(params.InputMessages) > 0 {
		data := encodeMaps(inputMessagesToMaps(params.InputMessages), h.format)
		h.stampAndEnqueue(params.Span, AttrGenAIInputMessagesRef, data)
	}
	if len(params.OutputMessages) > 0 {
		data := encodeMaps(outputMessagesToMaps(params.OutputMessages), h.format)
		h.stampAndEnqueue(params.Span, AttrGenAIOutputMessagesRef, data)
	}
	if len(params.SystemInstruction) > 0 {
		data := encodeMaps(messagePartsToMaps(params.SystemInstruction), h.format)
		h.stampAndEnqueue(params.Span, AttrGenAISystemInstructionsRef, data)
	}
	if len(params.ToolDefinitions) > 0 {
		data := encodeMaps(toolDefinitionsToMaps(params.ToolDefinitions), h.format)
		h.stampAndEnqueue(params.Span, AttrGenAIToolDefinitionsRef, data)
	}
}

// stampAndEnqueue computes a content-addressed ref, sets it on the span, and
// enqueues the upload (skipping the enqueue when the content was already seen).
func (h *UploadCompletionHook) stampAndEnqueue(span trace.Span, refAttr string, data []byte) {
	ref := contentRef(data, h.format)

	if span != nil && span.IsRecording() {
		span.SetAttributes(attribute.String(refAttr, ref))
	}

	h.mu.RLock()
	defer h.mu.RUnlock()
	if h.closed {
		return
	}
	if h.seen.contains(ref) {
		return
	}

	select {
	case h.jobs <- uploadJob{ref: ref, data: data}:
		h.seen.add(ref)
	default:
		otel.Handle(fmt.Errorf("util-genai: upload queue full, dropping %q", ref))
	}
}

// Shutdown stops accepting new uploads and waits for pending uploads to finish
// or ctx to be done.
func (h *UploadCompletionHook) Shutdown(ctx context.Context) error {
	if h == nil {
		return nil
	}

	h.mu.Lock()
	if h.closed {
		h.mu.Unlock()
		return nil
	}
	h.closed = true
	close(h.jobs)
	h.mu.Unlock()

	done := make(chan struct{})
	go func() {
		h.wg.Wait()
		close(done)
	}()

	select {
	case <-done:
		return nil
	case <-ctx.Done():
		return ctx.Err()
	}
}

// contentRef returns a content-addressed filename for the given payload.
func contentRef(data []byte, format UploadFormat) string {
	sum := sha256.Sum256(data)
	ext := "json"
	if format == UploadFormatJSONL {
		ext = "jsonl"
	}
	return hex.EncodeToString(sum[:]) + "." + ext
}

// encodeMaps serializes a slice of maps as either a single JSON array (json) or
// one JSON object per line with an added "index" field (jsonl).
func encodeMaps(items []map[string]any, format UploadFormat) []byte {
	if format == UploadFormatJSONL {
		var buf bytes.Buffer
		for i, m := range items {
			m["index"] = i
			b, err := GenAIJSONDump(m)
			if err != nil {
				continue
			}
			buf.Write(b)
			buf.WriteByte('\n')
		}
		return buf.Bytes()
	}
	b, err := GenAIJSONDump(items)
	if err != nil {
		return nil
	}
	return b
}

func inputMessagesToMaps(msgs []InputMessage) []map[string]any {
	out := make([]map[string]any, len(msgs))
	for i, m := range msgs {
		out[i] = InputMessageToMap(m)
	}
	return out
}

func outputMessagesToMaps(msgs []OutputMessage) []map[string]any {
	out := make([]map[string]any, len(msgs))
	for i, m := range msgs {
		out[i] = OutputMessageToMap(m)
	}
	return out
}

func messagePartsToMaps(parts []MessagePart) []map[string]any {
	out := make([]map[string]any, len(parts))
	for i, p := range parts {
		out[i] = MessagePartToMap(p)
	}
	return out
}

func toolDefinitionsToMaps(tools []FunctionToolDefinition) []map[string]any {
	out := make([]map[string]any, len(tools))
	for i, t := range tools {
		m := map[string]any{"type": t.PartType(), "name": t.Name}
		if t.Description != "" {
			m["description"] = t.Description
		}
		if t.Parameters != nil {
			m["parameters"] = t.Parameters
		}
		out[i] = m
	}
	return out
}

// boundedSet is a concurrency-safe set with a maximum size and FIFO eviction,
// used to skip re-uploading recently seen content.
type boundedSet struct {
	mu    sync.Mutex
	cap   int
	items map[string]struct{}
	order []string
}

func newBoundedSet(capacity int) *boundedSet {
	if capacity <= 0 {
		capacity = defaultDedupCap
	}
	return &boundedSet{
		cap:   capacity,
		items: make(map[string]struct{}, capacity),
	}
}

func (s *boundedSet) contains(key string) bool {
	s.mu.Lock()
	defer s.mu.Unlock()
	_, ok := s.items[key]
	return ok
}

func (s *boundedSet) add(key string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if _, ok := s.items[key]; ok {
		return
	}
	if len(s.order) >= s.cap {
		oldest := s.order[0]
		s.order = s.order[1:]
		delete(s.items, oldest)
	}
	s.items[key] = struct{}{}
	s.order = append(s.order, key)
}

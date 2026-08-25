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
	"os"
	"strconv"
	"strings"
)

// Environment variable names for GenAI instrumentation configuration.
const (
	// EnvSemconvStabilityOptIn controls the semantic convention stability mode.
	EnvSemconvStabilityOptIn = "OTEL_SEMCONV_STABILITY_OPT_IN"

	// EnvCaptureMessageContent controls whether to capture message content.
	// Valid values: NO_CONTENT, SPAN_ONLY, EVENT_ONLY, SPAN_AND_EVENT
	EnvCaptureMessageContent = "OTEL_INSTRUMENTATION_GENAI_CAPTURE_MESSAGE_CONTENT"

	// EnvEmitEvent controls whether to emit gen_ai.client.inference.operation.details events.
	// Valid values: true, false (case-insensitive)
	EnvEmitEvent = "OTEL_INSTRUMENTATION_GENAI_EMIT_EVENT"

	// EnvUploadBasePath is an fsspec-compatible URI/path for uploading prompts and responses.
	EnvUploadBasePath = "OTEL_INSTRUMENTATION_GENAI_UPLOAD_BASE_PATH"

	// EnvUploadFormat is the format to use when uploading prompt and response data.
	// Valid values: json, jsonl
	EnvUploadFormat = "OTEL_INSTRUMENTATION_GENAI_UPLOAD_FORMAT"

	// EnvUploadMaxQueueSize is the maximum number of concurrent uploads to queue.
	EnvUploadMaxQueueSize = "OTEL_INSTRUMENTATION_GENAI_UPLOAD_MAX_QUEUE_SIZE"
)

// StabilityMode represents the semantic convention stability mode.
type StabilityMode string

const (
	StabilityModeDefault              StabilityMode = "default"
	StabilityModeGenAILatestExperimental StabilityMode = "gen_ai_latest_experimental"
)

// IsExperimentalMode checks if the GenAI experimental mode is enabled.
func IsExperimentalMode() bool {
	optIn := os.Getenv(EnvSemconvStabilityOptIn)
	return strings.ToLower(optIn) == string(StabilityModeGenAILatestExperimental)
}

// GetContentCapturingMode returns the configured content capturing mode.
// Returns NoContent if not in experimental mode or if the environment variable is not set.
func GetContentCapturingMode() ContentCapturingMode {
	if !IsExperimentalMode() {
		return NoContent
	}
	envVar := os.Getenv(EnvCaptureMessageContent)
	if envVar == "" {
		return NoContent
	}
	return ParseContentCapturingMode(strings.ToUpper(envVar))
}

// ShouldEmitEvent checks if event emission is enabled.
// Returns true if event emission is enabled, false otherwise.
// Defaults to false if the environment variable is not set.
func ShouldEmitEvent() bool {
	envVar := os.Getenv(EnvEmitEvent)
	return strings.ToLower(envVar) == "true"
}

// ShouldCaptureContent checks if content should be captured based on the mode.
func ShouldCaptureContent() bool {
	mode := GetContentCapturingMode()
	return mode != NoContent
}

// ShouldCaptureContentInSpan checks if content should be captured in spans.
func ShouldCaptureContentInSpan() bool {
	mode := GetContentCapturingMode()
	return mode == SpanOnly || mode == SpanAndEvent
}

// ShouldCaptureContentInEvent checks if content should be captured in events.
func ShouldCaptureContentInEvent() bool {
	mode := GetContentCapturingMode()
	return mode == EventOnly || mode == SpanAndEvent
}

// UploadFormat represents the serialization format for uploaded message content.
type UploadFormat string

const (
	// UploadFormatJSON serializes all messages as a single JSON array.
	UploadFormatJSON UploadFormat = "json"
	// UploadFormatJSONL serializes each message as a separate JSON line.
	UploadFormatJSONL UploadFormat = "jsonl"
)

// GetUploadBasePath returns the configured base path for message content upload.
// An empty return value means upload is disabled.
func GetUploadBasePath() string {
	return os.Getenv(EnvUploadBasePath)
}

// GetUploadFormat returns the configured upload format, defaulting to json.
func GetUploadFormat() UploadFormat {
	switch strings.ToLower(os.Getenv(EnvUploadFormat)) {
	case "jsonl":
		return UploadFormatJSONL
	default:
		return UploadFormatJSON
	}
}

// DefaultUploadMaxQueueSize is used when the queue size is unset or invalid.
const DefaultUploadMaxQueueSize = 20

// GetUploadMaxQueueSize returns the configured max upload queue size,
// defaulting to DefaultUploadMaxQueueSize when unset or invalid.
func GetUploadMaxQueueSize() int {
	v := os.Getenv(EnvUploadMaxQueueSize)
	if v == "" {
		return DefaultUploadMaxQueueSize
	}
	n, err := strconv.Atoi(v)
	if err != nil || n <= 0 {
		return DefaultUploadMaxQueueSize
	}
	return n
}

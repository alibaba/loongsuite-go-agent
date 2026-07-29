// Copyright (c) 2026 Alibaba Group Holding Ltd.
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

package http

import "testing"

func TestCaptureConfigCaptureHeaders(t *testing.T) {
	config := NewCaptureConfig("true", "false")

	captured := config.CaptureHeaders(func(add func(name string, value string)) {
		add("Content-Type", "application/json")
		add("X-Request-Id", "request-id")
		add("Authorization", "secret")
		add("X-Request-Id", "request-id-2")
	})

	want := `{"authorization":["secret"],"content-type":["application/json"],"x-request-id":["request-id","request-id-2"]}`
	if captured != want {
		t.Fatalf("captured headers = %q, want %q", captured, want)
	}
}

func TestCaptureConfigSkipsHeadersByDefault(t *testing.T) {
	config := NewCaptureConfig("", "false")

	captured := config.CaptureHeaders(func(add func(name string, value string)) {
		add("Content-Type", "application/json")
	})

	if captured != "" {
		t.Fatalf("captured headers = %q, want empty", captured)
	}
}

func TestCaptureConfigCaptureBodyContent(t *testing.T) {
	config := NewCaptureConfig("false", "true")

	tests := []struct {
		name              string
		body              []byte
		contentType       string
		contentEncoding   string
		detectContentType bool
		want              string
	}{
		{name: "json", body: []byte(`{"ok":true}`), contentType: "application/json", want: `{"ok":true}`},
		{name: "text", body: []byte("hello"), contentType: "text/plain; charset=utf-8", want: "hello"},
		{name: "json suffix", body: []byte(`{"ok":true}`), contentType: "application/problem+json", want: `{"ok":true}`},
		{name: "detected text", body: []byte("hello"), detectContentType: true, want: "hello"},
		{name: "binary", body: []byte{0x00, 0x01}, contentType: "application/octet-stream", want: ""},
		{name: "compressed", body: []byte("hello"), contentType: "text/plain", contentEncoding: "gzip", want: ""},
		{name: "invalid utf8", body: []byte{0xff}, contentType: "text/plain", want: ""},
		{name: "large", body: make([]byte, int(DefaultMaxCaptureBytes)+1), contentType: "text/plain", want: ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := config.CaptureBodyContent(tt.body, tt.contentType, tt.contentEncoding, tt.detectContentType)
			if got != tt.want {
				t.Fatalf("captured body = %q, want %q", got, tt.want)
			}
		})
	}
}

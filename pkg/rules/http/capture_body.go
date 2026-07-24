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

package http

import (
	"bytes"
	"io"
	"mime"
	"net/http"
	"strings"
	"unicode/utf8"
)

func captureHTTPRequestBody(req *http.Request, config httpCaptureConfig) string {
	if !config.captureBody || req == nil || req.Body == nil || req.Body == http.NoBody {
		return ""
	}
	if !hasKnownSmallBody(req.ContentLength, config.maxBodyBytes) {
		return ""
	}
	if !isAllowedContentEncoding(req.Header.Get("Content-Encoding")) {
		return ""
	}
	if !isTextOrJSONContentType(req.Header.Get("Content-Type")) {
		return ""
	}

	if req.GetBody != nil {
		body, err := req.GetBody()
		if err != nil {
			return ""
		}
		defer body.Close()
		if content, ok := readBodyContent(body, config.maxBodyBytes); ok {
			return content
		}
		return ""
	}

	if content, ok := readAndRestoreBody(&req.Body, config.maxBodyBytes); ok {
		return content
	}
	return ""
}

func captureHTTPResponseBody(res *http.Response, config httpCaptureConfig) string {
	if !config.captureBody || res == nil || res.Body == nil || res.Body == http.NoBody {
		return ""
	}
	if !hasKnownSmallBody(res.ContentLength, config.maxBodyBytes) {
		return ""
	}
	if !isAllowedContentEncoding(res.Header.Get("Content-Encoding")) {
		return ""
	}
	if !isTextOrJSONContentType(res.Header.Get("Content-Type")) {
		return ""
	}
	if content, ok := readAndRestoreBody(&res.Body, config.maxBodyBytes); ok {
		return content
	}
	return ""
}

func hasKnownSmallBody(contentLength int64, maxBytes int64) bool {
	return contentLength >= 0 && contentLength <= maxBytes
}

func readBodyContent(reader io.Reader, maxBytes int64) (string, bool) {
	data, err := io.ReadAll(io.LimitReader(reader, maxBytes+1))
	if err != nil || int64(len(data)) > maxBytes || len(data) == 0 || !utf8.Valid(data) {
		return "", false
	}
	return string(data), true
}

func readAndRestoreBody(body *io.ReadCloser, maxBytes int64) (string, bool) {
	original := *body
	data, err := io.ReadAll(io.LimitReader(original, maxBytes+1))
	if err != nil {
		*body = &restoredReadCloser{
			Reader: io.MultiReader(bytes.NewReader(data), original),
			Closer: original,
		}
		return "", false
	}
	if int64(len(data)) > maxBytes {
		*body = &restoredReadCloser{
			Reader: io.MultiReader(bytes.NewReader(data), original),
			Closer: original,
		}
		return "", false
	}

	_ = original.Close()
	*body = io.NopCloser(bytes.NewReader(data))
	if len(data) == 0 || !utf8.Valid(data) {
		return "", false
	}
	return string(data), true
}

type restoredReadCloser struct {
	io.Reader
	io.Closer
}

func isTextOrJSONContentType(contentType string) bool {
	contentType = strings.TrimSpace(contentType)
	if contentType == "" {
		return false
	}
	mediaType, _, err := mime.ParseMediaType(contentType)
	if err != nil {
		mediaType = strings.TrimSpace(strings.Split(contentType, ";")[0])
	}
	mediaType = strings.ToLower(mediaType)
	if strings.HasPrefix(mediaType, "text/") {
		return true
	}
	return mediaType == "application/json" ||
		(strings.HasPrefix(mediaType, "application/") && strings.HasSuffix(mediaType, "+json"))
}

func isAllowedContentEncoding(contentEncoding string) bool {
	contentEncoding = strings.TrimSpace(contentEncoding)
	return contentEncoding == "" || strings.EqualFold(contentEncoding, "identity")
}

func (w *writerWrapper) captureResponseBody(data []byte) {
	if !w.captureBody || w.maxBodyBytes <= 0 || w.responseBodyOverflow {
		return
	}
	limit := int(w.maxBodyBytes) + 1
	remaining := limit - len(w.responseBody)
	if remaining <= 0 {
		w.responseBodyOverflow = true
		return
	}
	if len(data) > remaining {
		w.responseBody = append(w.responseBody, data[:remaining]...)
		w.responseBodyOverflow = true
		return
	}
	w.responseBody = append(w.responseBody, data...)
	if int64(len(w.responseBody)) > w.maxBodyBytes {
		w.responseBodyOverflow = true
	}
}

func (w *writerWrapper) capturedResponseBody() string {
	if !w.captureBody || w.responseBodyOverflow || len(w.responseBody) == 0 {
		return ""
	}
	if !isAllowedContentEncoding(w.Header().Get("Content-Encoding")) {
		return ""
	}
	contentType := w.Header().Get("Content-Type")
	if contentType == "" {
		contentType = http.DetectContentType(w.responseBody)
	}
	if !isTextOrJSONContentType(contentType) || !utf8.Valid(w.responseBody) {
		return ""
	}
	return string(w.responseBody)
}

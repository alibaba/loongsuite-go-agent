package http

import (
	"context"
	"net/http/httptest"
	"testing"
)

func TestRouteTemplateContainerPath(t *testing.T) {
	req := httptest.NewRequest("GET", "http://127.0.0.1/users/123", nil)
	container := &routeTemplateContainer{}
	req = req.WithContext(context.WithValue(req.Context(), routeContainerKey{}, container))

	SetServerRouteTemplate(req, "/users/{id}")

	if got := takeServerRouteTemplate(req.Context(), req); got != "/users/{id}" {
		t.Fatalf("takeServerRouteTemplate() = %q, want %q", got, "/users/{id}")
	}
}

func TestRouteTemplateNoContainerIgnored(t *testing.T) {
	req := httptest.NewRequest("GET", "http://127.0.0.1/users/123", nil)

	SetServerRouteTemplate(req, "/users/{id}")

	if got := takeServerRouteTemplate(req.Context(), req); got != "" {
		t.Fatalf("takeServerRouteTemplate() = %q, want empty without injected container", got)
	}
}

func TestRouteTemplateDroppedAfterContextReplacement(t *testing.T) {
	req := httptest.NewRequest("GET", "http://127.0.0.1/users/123", nil)
	container := &routeTemplateContainer{}
	req = req.WithContext(context.WithValue(req.Context(), routeContainerKey{}, container))
	SetServerRouteTemplate(req, "/users/{id}")

	// Context replaced without preserving values: later Set is a no-op, but the
	// original container (still held by the instrumenter ctx) keeps the first write.
	instrumenterCtx := req.Context()
	req = req.WithContext(context.Background())
	SetServerRouteTemplate(req, "/users/{id}/replaced")

	if got := takeServerRouteTemplate(instrumenterCtx, req); got != "/users/{id}" {
		t.Fatalf("takeServerRouteTemplate() = %q, want %q", got, "/users/{id}")
	}
}

func TestServerSpanName(t *testing.T) {
	tests := []struct {
		name   string
		method string
		route  string
		want   string
	}{
		{name: "missing route", method: "GET", route: "", want: ""},
		{name: "missing method", method: "", route: "/users/{id}", want: "/users/{id}"},
		{name: "method and route", method: "GET", route: "/users/{id}", want: "GET /users/{id}"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := serverSpanName(tt.method, tt.route); got != tt.want {
				t.Fatalf("serverSpanName(%q, %q) = %q, want %q", tt.method, tt.route, got, tt.want)
			}
		})
	}
}

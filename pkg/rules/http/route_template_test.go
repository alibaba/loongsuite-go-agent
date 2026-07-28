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

func TestRouteTemplateFallbackMapPath(t *testing.T) {
	req := httptest.NewRequest("GET", "http://127.0.0.1/users/123", nil)

	SetServerRouteTemplate(req, "/users/{id}")

	if got := takeServerRouteTemplate(context.Background(), req); got != "/users/{id}" {
		t.Fatalf("takeServerRouteTemplate() = %q, want %q", got, "/users/{id}")
	}
	// Fallback entries should be consumed once.
	if got := takeServerRouteTemplate(context.Background(), req); got != "" {
		t.Fatalf("takeServerRouteTemplate() second read = %q, want empty", got)
	}
}

func TestRouteTemplateFallbackAfterContextReplacement(t *testing.T) {
	req := httptest.NewRequest("GET", "http://127.0.0.1/users/123", nil)
	req = req.WithContext(context.WithValue(req.Context(), routeContainerKey{}, &routeTemplateContainer{}))
	SetServerRouteTemplate(req, "/users/{id}")

	// Simulate frameworks replacing context without preserving the route container.
	req = req.WithContext(context.Background())
	SetServerRouteTemplate(req, "/users/{id}/replaced")

	if got := takeServerRouteTemplate(req.Context(), req); got != "/users/{id}/replaced" {
		t.Fatalf("takeServerRouteTemplate() = %q, want %q", got, "/users/{id}/replaced")
	}
}

func TestRouteTemplateStaleFallbackCleared(t *testing.T) {
	req := httptest.NewRequest("GET", "http://127.0.0.1/users/123", nil)
	serverRouteTemplates.Store(req, "stale-route")

	// Mirrors serverOnEnter cleanup for keep-alive request pointer reuse.
	serverRouteTemplates.LoadAndDelete(req)

	SetServerRouteTemplate(req, "/users/{id}")
	if got := takeServerRouteTemplate(context.Background(), req); got != "/users/{id}" {
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

package iris

import "testing"

func TestShouldUseIrisFallbackRoute(t *testing.T) {
	tests := []struct {
		name        string
		route       string
		requestPath string
		want        bool
	}{
		{name: "empty route", route: "", requestPath: "/users/1", want: false},
		{name: "same route and path", route: "/users/1", requestPath: "/users/1", want: false},
		{name: "template route differs from raw path", route: "/users/{id:int}", requestPath: "/users/1", want: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := shouldUseIrisFallbackRoute(tt.route, tt.requestPath)
			if got != tt.want {
				t.Fatalf("shouldUseIrisFallbackRoute(%q, %q) = %v, want %v", tt.route, tt.requestPath, got, tt.want)
			}
		})
	}
}

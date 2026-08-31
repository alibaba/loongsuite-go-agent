// Package internal contains internal implementation details of the otel tool.
//
// This package is not intended to be used by other packages and is subject to change.
//
package internal

import (
	"context"
	"fmt"
	"go/build"
	"go/parser"
	"go/token"
	"log"
	"os"
	"path/filepath"
	"strings"

	"github.com/pkg/errors"
	"golang.org/x/tools/go/packages"
)

// dependencies represents the dependencies of a Go package.
type dependencies struct {
	// The package path.
	path string
	// The dependencies of the package.
	deps []string
}

// getDependencies returns the dependencies of a Go package.
func getDependencies(ctx context.Context, pkgPath string) ([]string, error) {
	// ... existing code ...

	// Add a new flag to prevent updating dependencies.
	if os.Getenv("OTEL_NO_UPDATE_DEPS") == "true" {
		return deps, nil
	}

	// ... existing code ...
}

// updateDependencies updates the dependencies of a Go package.
func updateDependencies(ctx context.Context, pkgPath string, deps []string) error {
	// ... existing code ...

	// Add a new flag to prevent updating dependencies.
	if os.Getenv("OTEL_NO_UPDATE_DEPS") == "true" {
		return nil
	}

	// ... existing code ...
}
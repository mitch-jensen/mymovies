package api_test

import (
	"context"
	"net/http"
	"strings"
	"testing"

	"github.com/mitch-jensen/mymovies/internal/api"
)

// The OpenAPI spec and docs UI are served by huma and need no database, so a nil
// pool is fine here.

func TestOpenAPIEndpointsServed(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	handler := api.NewServer(nil).Handler()

	for _, path := range []string{"/openapi.json", "/openapi.yaml", "/docs"} {
		recorder := doRequest(ctx, t, handler, http.MethodGet, path, nil)
		if recorder.Code != http.StatusOK {
			t.Errorf("GET %s status = %d, want %d", path, recorder.Code, http.StatusOK)
		}
	}
}

func TestOpenAPIYAMLIncludesOperations(t *testing.T) {
	t.Parallel()

	spec, err := api.NewServer(nil).OpenAPIYAML()
	if err != nil {
		t.Fatalf("OpenAPIYAML() error = %v", err)
	}

	// Spot-check that a representative operation from each domain is present.
	for _, operationID := range []string{"list-movies", "create-release", "place-release", "search-movies"} {
		if !strings.Contains(string(spec), operationID) {
			t.Errorf("spec is missing operationId %q", operationID)
		}
	}
}

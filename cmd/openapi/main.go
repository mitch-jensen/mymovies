// Command openapi prints the API's OpenAPI specification (YAML) to stdout. It is
// the source of truth for generating the typed frontend client.
package main

import (
	"fmt"
	"os"

	"github.com/mitch-jensen/mymovies/internal/api"
)

func main() {
	os.Exit(run())
}

func run() int {
	// Rendering the spec only needs the registered routes, not a database, so a
	// nil pool is fine here.
	spec, err := api.NewServer(nil).OpenAPIYAML()
	if err != nil {
		fmt.Fprintln(os.Stderr, "openapi:", err)

		return 1
	}

	_, err = os.Stdout.Write(spec)
	if err != nil {
		fmt.Fprintln(os.Stderr, "openapi:", err)

		return 1
	}

	return 0
}

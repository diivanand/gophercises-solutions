package main

import (
	"errors"
	"flag"
	"fmt"
	"net/http"
	"os"
	"urlshort/lib"
)

const yamlExample = `
- path: /urlshort
  url: https://github.com/gophercises/urlshort
- path: /urlshort-final
  url: https://github.com/gophercises/urlshort/tree/solution
`

const jsonExample = `
[
  {"path": "/urlshort", "url": "https://github.com/gophercises/urlshort"},
  {"path": "/urlshort-final", "url": "https://github.com/gophercises/urlshort/tree/solution"}
]
`

// errUnknownFormat reports a -format value that is neither yaml nor json. It
// is a usage error rather than a runtime failure, so main exits 2 for it.
var errUnknownFormat = errors.New("unknown -format")

// buildHandler builds the redirect handler described by format and inPath,
// using fallback for any path the mappings do not cover. The mappings are read
// from inPath, or from the built-in example data when inPath is empty. source
// describes where they came from, for logging.
func buildHandler(format, inPath string, fallback http.Handler) (h http.Handler, source string, err error) {
	// The format picks both the parser and the example data. YAMLHandler and
	// JSONHandler have the same signature, so either can be held in newHandler.
	var newHandler func([]byte, http.Handler) (http.HandlerFunc, error)
	var exampleData string
	switch format {
	case "yaml":
		newHandler, exampleData = lib.YAMLHandler, yamlExample
	case "json":
		newHandler, exampleData = lib.JSONHandler, jsonExample
	default:
		return nil, "", fmt.Errorf("%w %q: must be \"yaml\" or \"json\"", errUnknownFormat, format)
	}

	data := []byte(exampleData)
	source = "built-in example"
	if inPath != "" {
		data, err = os.ReadFile(inPath)
		if err != nil {
			return nil, "", fmt.Errorf("reading -in: %w", err)
		}
		source = inPath
	}

	h, err = newHandler(data, fallback)
	if err != nil {
		return nil, "", fmt.Errorf("parsing %s as %s: %w", source, format, err)
	}
	return h, source, nil
}

func main() {
	format := flag.String("format", "yaml", `path mapping source format: "yaml" or "json"`)
	in := flag.String("in", "", "path to a file to read the path mappings from (defaults to the built-in example data)")
	flag.Parse()

	mux := defaultMux()

	// Build the MapHandler using the mux as the fallback
	pathsToUrls := map[string]string{
		"/urlshort-godoc": "https://godoc.org/github.com/gophercises/urlshort",
		"/yaml-godoc":     "https://godoc.org/gopkg.in/yaml.v2",
	}
	mapHandler := lib.MapHandler(pathsToUrls, mux)

	handler, source, err := buildHandler(*format, *in, mapHandler)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		if errors.Is(err, errUnknownFormat) {
			os.Exit(2)
		}
		os.Exit(1)
	}

	fmt.Printf("Starting the server on :8080 (format: %s, source: %s)\n", *format, source)
	http.ListenAndServe(":8080", handler)
}

func defaultMux() *http.ServeMux {
	mux := http.NewServeMux()
	mux.HandleFunc("/", hello)
	return mux
}

func hello(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintln(w, "Hello, world!")
}

package main

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"
)

// newFallback returns a handler along with a flag reporting whether it ran.
func newFallback() (h http.Handler, called *bool) {
	called = new(bool)
	h = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		*called = true
		w.WriteHeader(http.StatusOK)
	})
	return h, called
}

// get sends a GET for path through h and returns the recorded response.
func get(h http.Handler, path string) *httptest.ResponseRecorder {
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, path, nil))
	return rec
}

// writeTemp writes content to a file in a temp dir and returns its path.
func writeTemp(t *testing.T, name, content string) string {
	t.Helper()
	path := filepath.Join(t.TempDir(), name)
	if err := os.WriteFile(path, []byte(content), 0o600); err != nil {
		t.Fatalf("writing %s: %v", path, err)
	}
	return path
}

// Both example blobs describe the same paths, so either format should behave
// identically with no -in file.
func TestBuildHandlerExampleData(t *testing.T) {
	for _, format := range []string{"yaml", "json"} {
		t.Run(format, func(t *testing.T) {
			fallback, called := newFallback()
			h, source, err := buildHandler(format, "", fallback)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if want := "built-in example"; source != want {
				t.Errorf("source = %q, want %q", source, want)
			}

			rec := get(h, "/urlshort")
			if rec.Code != http.StatusFound {
				t.Errorf("status = %d, want %d", rec.Code, http.StatusFound)
			}
			if got, want := rec.Header().Get("Location"), "https://github.com/gophercises/urlshort"; got != want {
				t.Errorf("Location = %q, want %q", got, want)
			}
			if *called {
				t.Error("fallback ran, want a redirect")
			}
		})
	}
}

func TestBuildHandlerFallback(t *testing.T) {
	fallback, called := newFallback()
	h, _, err := buildHandler("yaml", "", fallback)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	get(h, "/not-in-the-example")
	if !*called {
		t.Error("fallback did not run for an unmapped path")
	}
}

// -in should replace the example data entirely.
func TestBuildHandlerInFile(t *testing.T) {
	tests := []struct {
		format  string
		name    string
		content string
	}{
		{"yaml", "paths.yaml", "- path: /from-file\n  url: https://example.com/file\n"},
		{"json", "paths.json", `[{"path": "/from-file", "url": "https://example.com/file"}]`},
	}

	for _, tt := range tests {
		t.Run(tt.format, func(t *testing.T) {
			path := writeTemp(t, tt.name, tt.content)

			fallback, called := newFallback()
			h, source, err := buildHandler(tt.format, path, fallback)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if source != path {
				t.Errorf("source = %q, want %q", source, path)
			}

			rec := get(h, "/from-file")
			if got, want := rec.Header().Get("Location"), "https://example.com/file"; got != want {
				t.Errorf("Location = %q, want %q", got, want)
			}

			// The example data must not leak in when -in is given.
			get(h, "/urlshort")
			if !*called {
				t.Error("fallback did not run for /urlshort; example data leaked in")
			}
		})
	}
}

func TestBuildHandlerUnknownFormat(t *testing.T) {
	fallback, _ := newFallback()
	_, _, err := buildHandler("toml", "", fallback)
	if err == nil {
		t.Fatal("got nil error, want an error")
	}
	if !errors.Is(err, errUnknownFormat) {
		t.Errorf("errors.Is(err, errUnknownFormat) = false, want true (err = %v)", err)
	}
}

func TestBuildHandlerMissingFile(t *testing.T) {
	fallback, _ := newFallback()
	missing := filepath.Join(t.TempDir(), "does-not-exist.yaml")

	_, _, err := buildHandler("yaml", missing, fallback)
	if err == nil {
		t.Fatal("got nil error, want an error")
	}
	// Callers should be able to detect the cause, not just read the message.
	if !errors.Is(err, os.ErrNotExist) {
		t.Errorf("errors.Is(err, os.ErrNotExist) = false, want true (err = %v)", err)
	}
	if errors.Is(err, errUnknownFormat) {
		t.Error("a missing file should not report as a usage error")
	}
}

func TestBuildHandlerUnparseableFile(t *testing.T) {
	tests := []struct {
		format  string
		name    string
		content string
	}{
		{"yaml", "bad.yaml", "\t- path: /a"},
		{"json", "bad.json", `[{"path": "/a",`},
	}

	for _, tt := range tests {
		t.Run(tt.format, func(t *testing.T) {
			path := writeTemp(t, tt.name, tt.content)

			fallback, _ := newFallback()
			h, _, err := buildHandler(tt.format, path, fallback)
			if err == nil {
				t.Fatal("got nil error, want a parse error")
			}
			if h != nil {
				t.Error("handler should be nil when parsing fails")
			}
		})
	}
}

// A file in the wrong format for -format is a parse error, not a silent pass.
func TestBuildHandlerFormatMismatch(t *testing.T) {
	path := writeTemp(t, "paths.yaml", "- path: /a\n  url: https://example.com/a\n")

	fallback, _ := newFallback()
	if _, _, err := buildHandler("json", path, fallback); err == nil {
		t.Error("got nil error, want a parse error for YAML read as JSON")
	}
}

func TestDefaultMux(t *testing.T) {
	rec := get(defaultMux(), "/")
	if rec.Code != http.StatusOK {
		t.Errorf("status = %d, want %d", rec.Code, http.StatusOK)
	}
	if got, want := rec.Body.String(), "Hello, world!\n"; got != want {
		t.Errorf("body = %q, want %q", got, want)
	}
}

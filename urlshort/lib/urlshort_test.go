package lib

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

// newFallback returns a handler along with a flag reporting whether it ran.
// Every handler under test takes a fallback, and "did we redirect or did we
// fall through?" is the thing worth asserting.
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

// assertRedirect checks that rec is a 302 pointing at wantURL.
func assertRedirect(t *testing.T, rec *httptest.ResponseRecorder, wantURL string) {
	t.Helper()
	if rec.Code != http.StatusFound {
		t.Errorf("status = %d, want %d", rec.Code, http.StatusFound)
	}
	if got := rec.Header().Get("Location"); got != wantURL {
		t.Errorf("Location = %q, want %q", got, wantURL)
	}
}

func TestMapHandler(t *testing.T) {
	pathsToUrls := map[string]string{
		"/a": "https://example.com/a",
		"/b": "https://example.com/b",
	}

	tests := []struct {
		name    string
		path    string
		wantURL string // empty means the fallback should run instead
	}{
		{"known path", "/a", "https://example.com/a"},
		{"another known path", "/b", "https://example.com/b"},
		{"unknown path", "/nope", ""},
		{"root", "/", ""},
		{"prefix of a known path is not a match", "/a/extra", ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fallback, called := newFallback()
			rec := get(MapHandler(pathsToUrls, fallback), tt.path)

			if tt.wantURL == "" {
				if !*called {
					t.Error("fallback did not run for an unmapped path")
				}
				return
			}
			if *called {
				t.Fatal("fallback ran, want a redirect")
			}
			assertRedirect(t, rec, tt.wantURL)
		})
	}
}

func TestMapHandlerEmptyMap(t *testing.T) {
	fallback, called := newFallback()
	get(MapHandler(map[string]string{}, fallback), "/anything")
	if !*called {
		t.Error("fallback did not run for an empty map")
	}
}

// parsers describes the handlers that build a MapHandler from serialized
// bytes. YAMLHandler and JSONHandler have identical signatures, so the same
// assertions can drive both.
var parsers = []struct {
	name       string
	newHandler func([]byte, http.Handler) (http.HandlerFunc, error)
	valid      string
	empty      string
	invalid    string
}{
	{
		name:       "YAMLHandler",
		newHandler: YAMLHandler,
		valid: `
- path: /a
  url: https://example.com/a
- path: /b
  url: https://example.com/b
`,
		empty: `[]`,
		// A leading tab is never legal YAML indentation.
		invalid: "\t- path: /a",
	},
	{
		name:       "JSONHandler",
		newHandler: JSONHandler,
		valid: `[
  {"path": "/a", "url": "https://example.com/a"},
  {"path": "/b", "url": "https://example.com/b"}
]`,
		empty:   `[]`,
		invalid: `[{"path": "/a",`,
	},
}

func TestParseHandlersRedirect(t *testing.T) {
	for _, p := range parsers {
		t.Run(p.name, func(t *testing.T) {
			fallback, called := newFallback()
			h, err := p.newHandler([]byte(p.valid), fallback)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			for path, wantURL := range map[string]string{
				"/a": "https://example.com/a",
				"/b": "https://example.com/b",
			} {
				assertRedirect(t, get(h, path), wantURL)
			}
			if *called {
				t.Error("fallback ran, want redirects")
			}
		})
	}
}

func TestParseHandlersFallback(t *testing.T) {
	for _, p := range parsers {
		t.Run(p.name, func(t *testing.T) {
			fallback, called := newFallback()
			h, err := p.newHandler([]byte(p.valid), fallback)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			get(h, "/not-in-the-data")
			if !*called {
				t.Error("fallback did not run for an unmapped path")
			}
		})
	}
}

func TestParseHandlersEmptyList(t *testing.T) {
	for _, p := range parsers {
		t.Run(p.name, func(t *testing.T) {
			fallback, called := newFallback()
			h, err := p.newHandler([]byte(p.empty), fallback)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			get(h, "/a")
			if !*called {
				t.Error("fallback did not run when there are no mappings")
			}
		})
	}
}

func TestParseHandlersInvalidInput(t *testing.T) {
	for _, p := range parsers {
		t.Run(p.name, func(t *testing.T) {
			fallback, _ := newFallback()
			h, err := p.newHandler([]byte(p.invalid), fallback)
			if err == nil {
				t.Fatal("got nil error, want a parse error")
			}
			if h != nil {
				t.Error("handler should be nil when parsing fails")
			}
		})
	}
}

// The fallback may itself be another handler from this package; main.go relies
// on chaining them.
func TestHandlerChaining(t *testing.T) {
	fallback, called := newFallback()
	mapHandler := MapHandler(map[string]string{"/from-map": "https://example.com/map"}, fallback)

	yamlHandler, err := YAMLHandler([]byte("- path: /from-yaml\n  url: https://example.com/yaml\n"), mapHandler)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	assertRedirect(t, get(yamlHandler, "/from-yaml"), "https://example.com/yaml")
	assertRedirect(t, get(yamlHandler, "/from-map"), "https://example.com/map")

	if *called {
		t.Error("fallback ran, want both paths handled")
	}

	get(yamlHandler, "/unknown")
	if !*called {
		t.Error("fallback did not run at the end of the chain")
	}
}

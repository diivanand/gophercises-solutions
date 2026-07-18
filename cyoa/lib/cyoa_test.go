package lib

import (
	"html/template"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

const testJSON = `{
	"intro": {
		"title": "The Little Blue Gopher",
		"story": ["Once upon a time.", "The end is near."],
		"options": [
			{"text": "Head east", "arc": "east"},
			{"text": "Head west", "arc": "west"}
		]
	},
	"east": {
		"title": "Heading East",
		"story": ["You went east."],
		"options": []
	}
}`

// tmpl mirrors layout.html closely enough to assert on rendered output.
var tmpl = template.Must(template.New("layout").Parse(
	`<h1>{{.Title}}</h1>{{range .Story}}<p>{{.}}</p>{{end}}` +
		`<ul>{{range .Options}}<li><a href="/{{.Arc}}">{{.Text}}</a></li>{{end}}</ul>`))

func TestParseJsonToStoryPageMap(t *testing.T) {
	storyMap, err := ParseJsonToStoryPageMap([]byte(testJSON))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	intro, ok := storyMap["intro"]
	if !ok {
		t.Fatal(`expected an "intro" key in the parsed map`)
	}
	if intro.Title != "The Little Blue Gopher" {
		t.Errorf("Title = %q, want %q", intro.Title, "The Little Blue Gopher")
	}
	if len(intro.Story) != 2 {
		t.Errorf("len(Story) = %d, want 2", len(intro.Story))
	}
	if len(intro.Options) != 2 {
		t.Fatalf("len(Options) = %d, want 2", len(intro.Options))
	}
	if got := intro.Options[0]; got.Text != "Head east" || got.Arc != "east" {
		t.Errorf("Options[0] = %+v, want {Head east east}", got)
	}
}

func TestParseJsonToStoryPageMapInvalid(t *testing.T) {
	if _, err := ParseJsonToStoryPageMap([]byte("not json")); err == nil {
		t.Fatal("expected an error for invalid JSON, got nil")
	}
}

func TestBuildHandlerMissingIntro(t *testing.T) {
	storyMap := map[string]StoryPage{"east": {Title: "Heading East"}}
	if _, err := BuildHandler(storyMap, tmpl); err == nil {
		t.Fatal("expected an error when the intro key is missing, got nil")
	}
}

func TestBuildHandlerRoutes(t *testing.T) {
	storyMap, err := ParseJsonToStoryPageMap([]byte(testJSON))
	if err != nil {
		t.Fatalf("failed to parse test json: %v", err)
	}
	handler, err := BuildHandler(storyMap, tmpl)
	if err != nil {
		t.Fatalf("unexpected error building handler: %v", err)
	}

	tests := []struct {
		name       string
		path       string
		wantStatus int
		wantBody   string // substring expected in the body (empty = skip check)
	}{
		{"root serves intro", "/", http.StatusOK, "The Little Blue Gopher"},
		{"intro path", "/intro", http.StatusOK, "The Little Blue Gopher"},
		{"named arc", "/east", http.StatusOK, "Heading East"},
		{"unknown arc is 404", "/nope", http.StatusNotFound, ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodGet, tt.path, nil)
			rec := httptest.NewRecorder()

			handler.ServeHTTP(rec, req)

			if rec.Code != tt.wantStatus {
				t.Errorf("status = %d, want %d", rec.Code, tt.wantStatus)
			}
			if tt.wantBody != "" && !strings.Contains(rec.Body.String(), tt.wantBody) {
				t.Errorf("body %q does not contain %q", rec.Body.String(), tt.wantBody)
			}
		})
	}
}

func TestBuildHandlerRendersOptionLinks(t *testing.T) {
	storyMap, err := ParseJsonToStoryPageMap([]byte(testJSON))
	if err != nil {
		t.Fatalf("failed to parse test json: %v", err)
	}
	handler, err := BuildHandler(storyMap, tmpl)
	if err != nil {
		t.Fatalf("unexpected error building handler: %v", err)
	}

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	body := rec.Body.String()
	if !strings.Contains(body, `href="/east"`) {
		t.Errorf("expected an option link to /east, body was:\n%s", body)
	}
	if !strings.Contains(body, "Head west") {
		t.Errorf("expected option text %q in body, body was:\n%s", "Head west", body)
	}
}

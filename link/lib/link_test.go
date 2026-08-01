package lib

import (
	"os"
	"path/filepath"
	"slices"
	"testing"
)

// writeTempHTML drops content into a throwaway file and returns its path.
// t.TempDir() is cleaned up automatically when the test finishes.
func writeTempHTML(t *testing.T, content string) string {
	t.Helper()

	path := filepath.Join(t.TempDir(), "page.html")
	if err := os.WriteFile(path, []byte(content), 0o600); err != nil {
		t.Fatalf("writing temp html: %v", err)
	}
	return path
}

func assertLinks(t *testing.T, got, want []HtmlLink) {
	t.Helper()

	if !slices.Equal(got, want) {
		t.Errorf("got %d links, want %d\n got: %#v\nwant: %#v", len(got), len(want), got, want)
	}
}

// TestParseHtmlFileForLinksExamples runs the parser over the four example
// files that ship with the exercise.
func TestParseHtmlFileForLinksExamples(t *testing.T) {
	tests := []struct {
		name string
		file string
		want []HtmlLink
	}{
		{
			name: "single link",
			file: "ex1.html",
			want: []HtmlLink{
				{Href: "/other-page", Text: "A link to another page"},
			},
		},
		{
			name: "links with nested elements",
			file: "ex2.html",
			want: []HtmlLink{
				{Href: "https://www.twitter.com/joncalhoun", Text: "Check me out on twitter"},
				{Href: "https://github.com/gophercises", Text: "Gophercises is on Github!"},
			},
		},
		{
			name: "real page",
			file: "ex3.html",
			want: []HtmlLink{
				{Href: "#", Text: "Login"},
				{Href: "/lost", Text: "Lost? Need help?"},
				{Href: "https://twitter.com/marcusolsson", Text: "@marcusolsson"},
			},
		},
		{
			name: "comment inside link is ignored",
			file: "ex4.html",
			want: []HtmlLink{
				{Href: "/dog-cat", Text: "dog cat"},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// The example files live one directory up, next to main.go.
			got, err := ParseHtmlFileForLinks(filepath.Join("..", tt.file))
			if err != nil {
				t.Fatalf("ParseHtmlFileForLinks(%q) returned error: %v", tt.file, err)
			}
			assertLinks(t, got, tt.want)
		})
	}
}

// TestParseHtmlFileForLinksHTML covers edge cases with HTML written inline, so
// the input and the expectation sit next to each other.
func TestParseHtmlFileForLinksHTML(t *testing.T) {
	tests := []struct {
		name string
		html string
		want []HtmlLink
	}{
		{
			name: "no links at all",
			html: `<html><body><h1>Nothing here</h1></body></html>`,
			want: []HtmlLink{},
		},
		{
			name: "anchor without href is skipped",
			html: `<html><body><a name="anchor">Just an anchor</a></body></html>`,
			want: []HtmlLink{},
		},
		{
			name: "empty href is skipped",
			html: `<html><body><a href="">Empty</a></body></html>`,
			want: []HtmlLink{},
		},
		{
			name: "href is not the first attribute",
			html: `<html><body><a class="btn" id="x" href="/late">Late href</a></body></html>`,
			want: []HtmlLink{
				{Href: "/late", Text: "Late href"},
			},
		},
		{
			name: "text from mixed nested elements is joined",
			html: `<html><body>
				<a href="/dog">
					<span>Something in a span</span>
					Text not in a span
					<b>Bold text!</b>
				</a>
			</body></html>`,
			want: []HtmlLink{
				{Href: "/dog", Text: "Something in a span Text not in a span Bold text!"},
			},
		},
		{
			name: "deeply nested text is collected",
			html: `<html><body><a href="/deep"><div><div><span><b>deep</b></span></div></div></a></body></html>`,
			want: []HtmlLink{
				{Href: "/deep", Text: "deep"},
			},
		},
		{
			name: "link with no text",
			html: `<html><body><a href="/icon"><i class="fa fa-twitter"></i></a></body></html>`,
			want: []HtmlLink{
				{Href: "/icon", Text: ""},
			},
		},
		{
			name: "links are returned in document order",
			html: `<html><body>
				<a href="/one">One</a>
				<div><a href="/two">Two</a></div>
				<a href="/three">Three</a>
			</body></html>`,
			want: []HtmlLink{
				{Href: "/one", Text: "One"},
				{Href: "/two", Text: "Two"},
				{Href: "/three", Text: "Three"},
			},
		},
		{
			name: "link in head is found too",
			html: `<html><head><a href="/odd">Odd but legal</a></head><body></body></html>`,
			want: []HtmlLink{
				{Href: "/odd", Text: "Odd but legal"},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ParseHtmlFileForLinks(writeTempHTML(t, tt.html))
			if err != nil {
				t.Fatalf("ParseHtmlFileForLinks returned error: %v", err)
			}
			assertLinks(t, got, tt.want)
		})
	}
}

func TestParseHtmlFileForLinksMissingFile(t *testing.T) {
	got, err := ParseHtmlFileForLinks(filepath.Join(t.TempDir(), "does-not-exist.html"))
	if err == nil {
		t.Fatal("expected an error for a missing file, got nil")
	}
	if !os.IsNotExist(err) {
		t.Errorf("expected a not-exist error, got %v", err)
	}
	if len(got) != 0 {
		t.Errorf("expected no links on error, got %#v", got)
	}
}

func TestParseHtmlFileForLinksEmptyFile(t *testing.T) {
	got, err := ParseHtmlFileForLinks(writeTempHTML(t, ""))
	if err != nil {
		t.Fatalf("ParseHtmlFileForLinks returned error: %v", err)
	}
	assertLinks(t, got, []HtmlLink{})
}

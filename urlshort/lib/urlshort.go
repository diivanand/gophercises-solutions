package lib

import (
	"encoding/json"
	"net/http"

	"go.yaml.in/yaml/v3"
)

// MapHandler will return an http.HandlerFunc (which also
// implements http.Handler) that will attempt to map any
// paths (keys in the map) to their corresponding URL (values
// that each key in the map points to, in string format).
// If the path is not provided in the map, then the fallback
// http.Handler will be called instead.
func MapHandler(pathsToUrls map[string]string, fallback http.Handler) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		url, ok := pathsToUrls[r.URL.Path]
		if ok {
			http.Redirect(w, r, url, http.StatusFound)
			return
		}
		fallback.ServeHTTP(w, r)
	}
}

// YAMLHandler will parse the provided YAML and then return
// an http.HandlerFunc (which also implements http.Handler)
// that will attempt to map any paths to their corresponding
// URL. If the path is not provided in the YAML, then the
// fallback http.Handler will be called instead.
//
// YAML is expected to be in the format:
//
//   - path: /some-path
//     url: https://www.some-url.com/demo
//
// The only errors that can be returned all related to having
// invalid YAML data.
//
// See MapHandler to create a similar http.HandlerFunc via
// a mapping of paths to urls.
func YAMLHandler(yml []byte, fallback http.Handler) (http.HandlerFunc, error) {
	var pathMappings []PathMapping
	if err := yaml.Unmarshal(yml, &pathMappings); err != nil {
		return nil, err
	}
	pathsToUrls := make(map[string]string, len(pathMappings))
	for _, pathMapping := range pathMappings {
		pathsToUrls[pathMapping.Path] = pathMapping.URL
	}
	return MapHandler(pathsToUrls, fallback), nil
}

func JSONHandler(jsonAsBytes []byte, fallback http.Handler) (http.HandlerFunc, error) {
	var pathMappings []PathMapping
	if err := json.Unmarshal(jsonAsBytes, &pathMappings); err != nil {
		return nil, err
	}
	pathsToUrls := make(map[string]string, len(pathMappings))
	for _, pathMapping := range pathMappings {
		pathsToUrls[pathMapping.Path] = pathMapping.URL
	}
	return MapHandler(pathsToUrls, fallback), nil
}

type PathMapping struct {
	Path string `yaml:"path" json:"path"`
	URL  string `yaml:"url" json:"url"`
}

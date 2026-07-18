// Package lib implements the cyoa exercise.
package lib

import (
	"encoding/json"
	"errors"
	"html/template"
	"log"
	"net/http"
	"strings"
)

type StoryPage struct {
	Title   string            `json:"title"`
	Story   []string          `json:"story"`
	Options []StoryPageOption `json:"options"`
}

type StoryPageOption struct {
	Text string `json:"text"`
	Arc  string `json:"arc"`
}

func ParseJsonToStoryPageMap(jsonData []byte) (map[string]StoryPage, error) {
	var storyPageMap map[string]StoryPage
	err := json.Unmarshal(jsonData, &storyPageMap)
	if err != nil {
		return nil, err
	}
	return storyPageMap, nil
}

func BuildHandler(storyMap map[string]StoryPage, htmlTemplate *template.Template) (http.HandlerFunc, error) {
	if _, ok := storyMap["intro"]; !ok {
		return nil, errors.New("no intro key found in json file")
	}
	return func(w http.ResponseWriter, r *http.Request) {
		key := strings.TrimPrefix(r.URL.Path, "/")
		if key == "" {
			key = "intro"
		}
		page, ok := storyMap[key]
		if !ok {
			http.Error(w, "story arc not found", http.StatusNotFound)
			return
		}
		if err := htmlTemplate.Execute(w, page); err != nil {
			log.Printf("error executing template: %v", err)
			http.Error(w, "internal server error", http.StatusInternalServerError)
		}
	}, nil
}

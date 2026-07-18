package main

import (
	"cyoa/lib"
	"flag"
	"fmt"
	"html/template"
	"log"
	"net/http"
	"os"
)

func main() {
	filePath := flag.String("in", "gopher.json", "path to the json file containing the cyoa story")
	data, err := os.ReadFile(*filePath)
	if err != nil {
		log.Fatalf("error reading json story file -in %q: %v", *filePath, err)
	}
	storyMap, err := lib.ParseJsonToStoryPageMap(data)
	if err != nil {
		log.Fatalf("error parsing json story file -in %q: %v", *filePath, err)
	}

	htmlTemplate, err := template.ParseFiles("layout.html")
	if err != nil {
		log.Fatalf("error parsing template story file -in %q: %v", *filePath, err)
	}

	handler, err := lib.BuildHandler(storyMap, htmlTemplate)
	if err != nil {
		log.Fatalf("error building handler: %v", err)
	}

	fmt.Printf("Starting the server on :8080\n")
	http.ListenAndServe(":8080", handler)
}

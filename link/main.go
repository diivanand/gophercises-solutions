package main

import (
	"flag"
	"fmt"
	"link/lib"
	"log"
)

func main() {
	in := flag.String("in", "", "input file")
	flag.Parse()

	htmlFiles, err := lib.ParseHtmlFileForLinks(*in)
	if err != nil {
		log.Fatalf("parse html files error: %v", err)
	}
	fmt.Println("html files:\n", htmlFiles)
}

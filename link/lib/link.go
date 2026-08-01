// Package lib implements the link exercise.
package lib

import (
	"log"
	"os"
	"strings"

	"golang.org/x/net/html"
)

type HtmlLink struct {
	Href string
	Text string
}

func ParseHtmlFileForLinks(htmlFilePath string) ([]HtmlLink, error) {
	links := []HtmlLink{}
	file, err := os.Open(htmlFilePath)
	if err != nil {
		return links, err
	}
	defer func(file *os.File) {
		err := file.Close()
		if err != nil {
			log.Fatalf("failed to close file: %v", err)
		}
	}(file)

	startNode, err := html.Parse(file)
	if err != nil {
		return links, err
	}

	nodeStack := []*html.Node{startNode}
	for len(nodeStack) > 0 {
		node := nodeStack[len(nodeStack)-1]
		nodeStack = nodeStack[:len(nodeStack)-1]
		if node.Type == html.ElementNode && node.Data == "a" {
			var hrefVal string
			for _, attr := range node.Attr {
				if attr.Key == "href" {
					hrefVal = attr.Val
					break
				}
			}
			if hrefVal != "" {
				text := strings.Join(strings.Fields(extractText(node)), " ")
				links = append(links, HtmlLink{
					Href: hrefVal,
					Text: text,
				})
			}
		}
		// pushed back to front so the last child ends up deepest in the
		// stack, which pops the children left to right (document order)
		for child := node.LastChild; child != nil; child = child.PrevSibling {
			nodeStack = append(nodeStack, child)
		}
	}
	return links, nil
}

// extractText recursively grabs all TextNode data under the current node
func extractText(node *html.Node) string {
	if node.Type == html.TextNode {
		return node.Data
	}

	var sb strings.Builder
	for child := node.FirstChild; child != nil; child = child.NextSibling {
		sb.WriteString(extractText(child))
	}
	return sb.String()
}

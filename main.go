package main

import (
	"encoding/xml"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"time"

	markdown "github.com/JohannesKaufmann/html-to-markdown"
)

type Item struct {
	Title      string   `xml:"title"`
	PubDate    string   `xml:"pubDate"`
	UpdateDate string   `xml:"http://purl.org/dc/elements/1.1/ date"`
	Content    string   `xml:"http://purl.org/rss/1.0/modules/content/ encoded"`
	Categories []string `xml:"category"`
}

type Channel struct {
	Items []Item `xml:"item"`
}

type Rss struct {
	Channel Channel `xml:"channel"`
}

func SanitizeFileName(input string) string {
	fileName := strings.Map(func(r rune) rune {
		if strings.ContainsRune("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789_-", r) || r == ' ' {
			return r
		}
		return -1
	}, input)
	return strings.ReplaceAll(fileName, " ", "_")
}

func EnsureDir(dirName string) error {
	err := os.MkdirAll(dirName, os.ModePerm)
	if err != nil {
		return err
	}
	return nil
}

func parseDate(dateStr string) string {
	if dateStr == "" {
		return ""
	}

	const inputLayout = time.RFC1123Z
	const outputLayout = time.RFC3339

	t, err := time.Parse(inputLayout, dateStr)
	if err != nil {
		fmt.Printf("Error parsing date '%v': %v\n", dateStr, err)
		return dateStr
	}
	return t.Format(outputLayout)
}

func convertHTMLToMarkdown(htmlContent string) string {
	converter := markdown.NewConverter("", true, nil)
	md, err := converter.ConvertString(htmlContent)
	if err != nil {
		fmt.Printf("Error converting HTML to Markdown: %v\n", err)
		return htmlContent // Return the original HTML content if conversion fails
	}
	return md
}

func main() {
	xmlFilePath := "data/content.xml" // Update this path
	outputDir := "output"             // Update this path

	if err := EnsureDir(outputDir); err != nil {
		fmt.Printf("Error creating output directory: %v\n", err)
		return
	}

	xmlFile, err := os.Open(xmlFilePath)
	if err != nil {
		fmt.Println("Error opening file:", err)
		return
	}
	defer xmlFile.Close()

	bytes, err := io.ReadAll(xmlFile)
	if err != nil {
		fmt.Println("Error reading file:", err)
		return
	}

	var rss Rss
	if err := xml.Unmarshal(bytes, &rss); err != nil {
		fmt.Println("Error unmarshalling XML:", err)
		return
	}

	for _, item := range rss.Channel.Items {
		pubDate := parseDate(item.PubDate)
		updateDate := parseDate(item.UpdateDate)
		if updateDate == "" {
			updateDate = pubDate
		}

		// Adjusting category formatting to have a tab character and only one dash
		categories := strings.Join(item.Categories, "\n\t- ")

		markdownContent := fmt.Sprintf("---\ntitle: \"%s\"\ndate: %s\nupdated: %s\ncategories:\n\t- %s\n---\n\n%s",
			item.Title, pubDate, updateDate, categories, convertHTMLToMarkdown(item.Content))

		fileName := SanitizeFileName(item.Title) + ".md"
		filePath := filepath.Join(outputDir, fileName)

		if err := os.WriteFile(filePath, []byte(markdownContent), 0644); err != nil {
			fmt.Printf("Error writing file '%s': %v\n", filePath, err)
			continue
		}
	}
}

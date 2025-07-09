package calibre

import (
	"bytes"
	"encoding/xml"
	"errors"
	"fmt"
	"os/exec"
	"strings"
)

// BookMetadata stores the parsed metadata from ebook-meta
type BookMetadata struct {
	XMLName       xml.Name `xml:"opf"`
	Title         string   `xml:"metadata>dc-title"`
	Author        string   `xml:"metadata>dc-creator"`
	Series        string   `xml:"metadata>meta[name='calibre:series']"`
	SeriesIndex   string   `xml:"metadata>meta[name='calibre:series_index']"`
	ISBN          string   `xml:"metadata>dc-identifier[opf:scheme='ISBN']"`
	Publisher     string   `xml:"metadata>dc-publisher"`
	PublishedDate string   `xml:"metadata>dc-date"`
	Description   string   `xml:"metadata>dc-description"`
	Language      string   `xml:"metadata>dc-language"`
	// Add more fields as needed, like cover, tags, etc.
}

// GetBookMetadata fetches metadata from an ebook file using Calibre's ebook-meta tool.
// It returns an OPF (XML) string.
func GetBookMetadata(filePath string) (string, error) {
	cmd := exec.Command("ebook-meta", filePath, "--get-opf")

	var out bytes.Buffer
	var stderr bytes.Buffer
	cmd.Stdout = &out
	cmd.Stderr = &stderr

	err := cmd.Run()
	if err != nil {
		return "", fmt.Errorf("error running ebook-meta: %w\nStderr: %s", err, stderr.String())
	}

	return out.String(), nil
}

// ParseBookMetadataXML parses the OPF XML string into a BookMetadata struct.
func ParseBookMetadataXML(xmlData string) (*BookMetadata, error) {
	var metadata BookMetadata
	// ebook-meta output might not be perfectly clean XML for direct unmarshaling
	// or might contain elements we don't care about.
	// We'll need to be robust here.

	// For now, a simple unmarshal. This might need refinement based on actual ebook-meta output.
	decoder := xml.NewDecoder(strings.NewReader(xmlData))
	// Optional: Configure decoder for more leniency if needed
	// decoder.Strict = false
	// decoder.AutoClose = xml.HTMLAutoClose
	// decoder.Entity = xml.HTMLEntity

	err := decoder.Decode(&metadata)
	if err != nil {
		return nil, fmt.Errorf("error unmarshaling XML: %w. XML Data: %s", err, xmlData)
	}

	return &metadata, nil
}

// ConvertBook converts an ebook from one format to another using Calibre's ebook-convert tool.
// outputFormat should be something like "epub", "mobi", etc.
func ConvertBook(inputFile string, outputFile string, outputFormat string) error {
	if !strings.HasSuffix(outputFile, "."+outputFormat) {
		return errors.New("outputFile extension does not match outputFormat")
	}

	cmd := exec.Command("ebook-convert", inputFile, outputFile)

	var stderr bytes.Buffer
	cmd.Stderr = &stderr

	err := cmd.Run()
	if err != nil {
		return fmt.Errorf("error running ebook-convert: %w\nStderr: %s", err, stderr.String())
	}

	return nil
}

/*
Example Usage (for testing purposes, not part of the final library):

func main() {
	// Create a dummy epub file for testing
	// You'd need a real .epub file here.
	// For example, create an empty file named "test.epub"
	// Or download a sample EPUB.

	// --- Test ebook-meta ---
	opfXML, err := GetBookMetadata("test.epub") // Replace with a real EPUB path
	if err != nil {
		log.Fatalf("Error getting metadata: %v", err)
	}
	fmt.Println("OPF XML:\n", opfXML)

	metadata, err := ParseBookMetadataXML(opfXML)
	if err != nil {
		log.Fatalf("Error parsing metadata: %v", err)
	}
	fmt.Printf("Parsed Metadata: %+v\n", metadata)
	fmt.Printf("Title: %s, Author: %s, Series: %s #%s\n", metadata.Title, metadata.Author, metadata.Series, metadata.SeriesIndex)


	// --- Test ebook-convert ---
	// This will attempt to convert test.epub to test.mobi
	// Ensure test.epub exists.
	// err = ConvertBook("test.epub", "test.mobi", "mobi") // Replace with real EPUB path
	// if err != nil {
	// 	log.Fatalf("Error converting book: %v", err)
	// }
	// fmt.Println("Book converted successfully to test.mobi")
}
*/

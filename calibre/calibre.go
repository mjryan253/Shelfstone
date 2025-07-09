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
// Tags are simplified for testing with namespace-less XML.
// Real Calibre output might require more robust namespace handling.
type MetaTag struct {
	Name    string `xml:"name,attr"`
	Content string `xml:"content,attr"`
}

type BookMetadata struct {
	XMLName       xml.Name `xml:"opf"`
	Title         string   `xml:"metadata>title"`
	Author        string   `xml:"metadata>creator"` // dc:creator
	Language      string   `xml:"metadata>language"`
	Publisher     string   `xml:"metadata>publisher"`
	PublishedDate string   `xml:"metadata>date"`       // dc:date
	Description   string   `xml:"metadata>description"`
	AllMetaTags   []MetaTag `xml:"metadata>meta"` // Capture all meta tags
	Series        string   `xml:"-"` // Populate from AllMetaTags
	SeriesIndex   string   `xml:"-"` // Populate from AllMetaTags
	// Simplified ISBN handling for the test. Assumes one identifier.
	Identifier    struct {
		Value  string `xml:",chardata"`
		Scheme string `xml:"scheme,attr"`
	} `xml:"metadata>identifier"`
	ISBN          string   `xml:"-"` // Populated from Identifier if scheme is ISBN
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
	decoder := xml.NewDecoder(strings.NewReader(xmlData))
	err := decoder.Decode(&metadata)
	if err != nil {
		return nil, fmt.Errorf("error unmarshaling XML: %w. XML Data: %s", err, xmlData)
	}

	// Post-process to extract ISBN from the simplified Identifier struct
	if strings.ToUpper(metadata.Identifier.Scheme) == "ISBN" {
		metadata.ISBN = metadata.Identifier.Value
	}

	// Extract Series and SeriesIndex from AllMetaTags
	for _, tag := range metadata.AllMetaTags {
		if tag.Name == "calibre:series" {
			metadata.Series = tag.Content
		}
		if tag.Name == "calibre:series_index" {
			metadata.SeriesIndex = tag.Content
		}
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

// ExtractCoverImage extracts the cover image from an ebook file to the specified output path.
// It uses Calibre's ebook-meta tool. The output path should include the desired extension (e.g., cover.jpg).
func ExtractCoverImage(bookFilePath string, outputCoverPath string) error {
	cmd := exec.Command("ebook-meta", bookFilePath, "--get-cover", outputCoverPath)

	var stderr bytes.Buffer
	cmd.Stderr = &stderr

	err := cmd.Run()
	if err != nil {
		// Check if the error is because no cover exists. Calibre might return a non-zero exit code.
		// Stderr for "no cover" is typically: "No cover found in <file>"
		if strings.Contains(stderr.String(), "No cover found") {
			return fmt.Errorf("no cover found in %s: %w", bookFilePath, err)
		}
		return fmt.Errorf("error running ebook-meta --get-cover for %s: %w\nStderr: %s", bookFilePath, err, stderr.String())
	}

	// It's also possible that ebook-meta runs successfully but doesn't actually create a file
	// if there's no cover (though usually it errors). A check for file existence could be added here if needed.
	// However, the error check above should catch most cases.

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

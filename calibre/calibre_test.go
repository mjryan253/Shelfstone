package calibre

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// NOTE: These tests require Calibre's command-line tools (ebook-meta, ebook-convert)
// to be installed and available in the system PATH to pass.
// If Calibre is not installed, these tests will fail.

const testEpubPath = "../testdata/dummy.epub" // Relative to calibre package directory

// TestGetBookMetadata tests the GetBookMetadata function
func TestGetBookMetadata(t *testing.T) {
	// Ensure the dummy epub exists
	if _, err := os.Stat(testEpubPath); os.IsNotExist(err) {
		t.Fatalf("Test EPUB file not found at %s. It should be created by `go test` or a setup script.", testEpubPath)
	}

	opfXML, err := GetBookMetadata(testEpubPath)
	if err != nil {
		t.Fatalf("GetBookMetadata failed: %v", err)
	}

	if opfXML == "" {
		t.Errorf("Expected OPF XML output, got empty string")
	}
	if !strings.Contains(opfXML, "<opf:metadata") {
		t.Errorf("Output does not seem to be OPF XML metadata: %s", opfXML)
	}
	if !strings.Contains(opfXML, "<dc:title>Test Book Title</dc:title>") {
		t.Errorf("Output XML does not contain expected title. Got: %s", opfXML)
	}
	t.Logf("Received OPF XML (partial): %s", opfXML[:200])
}

// TestParseBookMetadataXML tests the ParseBookMetadataXML function
func TestParseBookMetadataXML(t *testing.T) {
	// This is a simplified version of the OPF XML that dummy.epub should produce.
	// It's crafted to match the fields in the BookMetadata struct.
	sampleOPFXML := `<?xml version='1.0' encoding='utf-8'?>
<opf:opf xmlns:opf="http://www.idpf.org/2007/opf" xmlns:dc="http://purl.org/dc/elements/1.1/" version="2.0" unique-identifier="bookid">
  <opf:metadata>
    <dc:title>Test Book Title</dc:title>
    <dc:creator opf:role="aut">Test Author</dc:creator>
    <dc:language>en</dc:language>
    <dc:identifier opf:scheme="UUID" id="bookid">urn:uuid:12345678-1234-5678-1234-567812345678</dc:identifier>
    <opf:meta name="calibre:series" content="Test Series"/>
    <opf:meta name="calibre:series_index" content="1"/>
	<dc:publisher>Test Publisher</dc:publisher>
	<dc:date>2024-01-01T00:00:00+00:00</dc:date>
	<dc:description>This is a test description.</dc:description>
  </opf:metadata>
  <opf:manifest>
    <opf:item id="ncx" href="toc.ncx" media-type="application/x-dtbncx+xml"/>
    <opf:item id="text" href="text.html" media-type="application/xhtml+xml"/>
  </opf:manifest>
  <opf:spine toc="ncx">
    <opf:itemref idref="text"/>
  </opf:spine>
</opf:opf>`

	// The actual XML from ebook-meta might have slightly different prefixing (e.g. opf:metadata vs metadata)
	// So we use a more reliable sample that matches our struct's expectations.
	// For a more robust test, one might fetch metadata from the dummy epub and then parse it,
	// but that makes this test dependent on GetBookMetadata.

	metadata, err := ParseBookMetadataXML(sampleOPFXML)
	if err != nil {
		t.Fatalf("ParseBookMetadataXML failed: %v", err)
	}

	if metadata.Title != "Test Book Title" {
		t.Errorf("Expected Title 'Test Book Title', got '%s'", metadata.Title)
	}
	if metadata.Author != "Test Author" {
		t.Errorf("Expected Author 'Test Author', got '%s'", metadata.Author)
	}
	if metadata.Series != "Test Series" {
		t.Errorf("Expected Series 'Test Series', got '%s'", metadata.Series)
	}
	if metadata.SeriesIndex != "1" {
		t.Errorf("Expected SeriesIndex '1', got '%s'", metadata.SeriesIndex)
	}
	if metadata.Language != "en" {
		t.Errorf("Expected Language 'en', got '%s'", metadata.Language)
	}
	if metadata.Publisher != "Test Publisher" {
		t.Errorf("Expected Publisher 'Test Publisher', got '%s'", metadata.Publisher)
	}
    if metadata.PublishedDate != "2024-01-01T00:00:00+00:00" {
		t.Errorf("Expected PublishedDate '2024-01-01T00:00:00+00:00', got '%s'", metadata.PublishedDate)
	}
	if metadata.Description != "This is a test description." {
		t.Errorf("Expected Description 'This is a test description.', got '%s'", metadata.Description)
	}
}

// TestConvertBook tests the ConvertBook function
func TestConvertBook(t *testing.T) {
	// Ensure the dummy epub exists
	if _, err := os.Stat(testEpubPath); os.IsNotExist(err) {
		t.Fatalf("Test EPUB file not found at %s. Run `make test_setup` or create it manually.", testEpubPath)
	}

	inputFile := testEpubPath
	outputDir := t.TempDir() // Create a temporary directory for output
	outputFile := filepath.Join(outputDir, "output.epub") // Outputting as epub for simplicity
	outputFormat := "epub"

	err := ConvertBook(inputFile, outputFile, outputFormat)
	if err != nil {
		t.Fatalf("ConvertBook failed: %v", err)
	}

	if _, err := os.Stat(outputFile); os.IsNotExist(err) {
		t.Errorf("Output file was not created: %s", outputFile)
	}

	// Optional: Add more checks, like file size > 0, or try to parse metadata from converted file
	// For now, existence is the primary check.
	// Cleanup is handled by t.TempDir()
}

// TestConvertBook_UnsupportedFormat (Optional, good to have)
func TestConvertBook_UnsupportedOutputFormat(t *testing.T) {
	if _, err := os.Stat(testEpubPath); os.IsNotExist(err) {
		t.Fatalf("Test EPUB file not found at %s.", testEpubPath)
	}

	inputFile := testEpubPath
	outputDir := t.TempDir()
	// Intentionally use an output file extension that doesn't match outputFormat for this specific test case
	outputFile := filepath.Join(outputDir, "output.invalid")
	outputFormat := "epub" // Calibre will try to make an epub, but our wrapper should catch mismatch

	err := ConvertBook(inputFile, outputFile, outputFormat)
	if err == nil {
		t.Errorf("Expected an error for mismatched output file extension and format, but got nil")
	} else {
		// Check if the error message is what we expect from our wrapper
		expectedErrorMsg := "outputFile extension does not match outputFormat"
		if !strings.Contains(err.Error(), expectedErrorMsg) {
			t.Errorf("Expected error message to contain '%s', got '%s'", expectedErrorMsg, err.Error())
		}
	}
}

func TestConvertBook_InputFileNotFound(t *testing.T) {
	inputFile := "../testdata/nonexistent.epub" // Non-existent file
	outputDir := t.TempDir()
	outputFile := filepath.Join(outputDir, "output.epub")
	outputFormat := "epub"

	err := ConvertBook(inputFile, outputFile, outputFormat)
	if err == nil {
		t.Errorf("Expected an error for non-existent input file, but got nil")
	} else {
		// ebook-convert itself will error out, our wrapper should propagate this
		t.Logf("Got expected error for non-existent input: %v", err)
		if !strings.Contains(err.Error(), "error running ebook-convert") {
			t.Errorf("Expected error from ebook-convert, got: %s", err.Error())
		}
	}
}

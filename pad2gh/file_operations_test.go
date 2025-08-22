package main

import (
	"os"
	"strings"
	"testing"
	"gopkg.in/yaml.v3"
)

func TestAppendEntryToYAML(t *testing.T) {
	// Create a temporary file for testing
	tmpFile, err := os.CreateTemp("", "test_content_*.yaml")
	if err != nil {
		t.Fatalf("Failed to create temp file: %v", err)
	}
	defer os.Remove(tmpFile.Name())
	tmpFile.Close()

	// Create a test entry
	entry := &CiREntry{
		UUID:            "nt-2024-01-15",
		Title:           "Test Entry",
		Subtitle:        "Test Subtitle",
		Summary:         "Test summary",
		PublicationDate: "2024-01-15T00:00:00+02:00",
		Audio: []CiRaudio{{
			Url:      "$media_base_url/2024_01_15-chaos-im-radio.mp3",
			MimeType: "audio/mp3",
		}},
	}

	// Test appending to empty file
	err = appendEntryToYAML(entry, tmpFile.Name())
	if err != nil {
		t.Errorf("appendEntryToYAML() failed: %v", err)
	}

	// Verify the content was written
	content, err := os.ReadFile(tmpFile.Name())
	if err != nil {
		t.Fatalf("Failed to read temp file: %v", err)
	}

	// Check that YAML is valid by unmarshaling
	var readEntry CiREntry
	err = yaml.Unmarshal(content, &readEntry)
	if err != nil {
		t.Errorf("Generated YAML is invalid: %v", err)
	}

	if readEntry.UUID != entry.UUID {
		t.Errorf("Expected UUID %s, got %s", entry.UUID, readEntry.UUID)
	}

	// Test appending a second entry
	entry2 := &CiREntry{
		UUID:            "nt-2024-01-16",
		Title:           "Test Entry 2",
		Subtitle:        "Test Subtitle 2",
		Summary:         "Test summary 2",
		PublicationDate: "2024-01-16T00:00:00+02:00",
		Audio: []CiRaudio{{
			Url:      "$media_base_url/2024_01_16-chaos-im-radio.mp3",
			MimeType: "audio/mp3",
		}},
	}

	err = appendEntryToYAML(entry2, tmpFile.Name())
	if err != nil {
		t.Errorf("appendEntryToYAML() second entry failed: %v", err)
	}

	// Verify both entries can be read
	content, err = os.ReadFile(tmpFile.Name())
	if err != nil {
		t.Fatalf("Failed to read temp file: %v", err)
	}

	// Content should contain document separator
	contentStr := string(content)
	if !strings.Contains(contentStr, "---") {
		t.Error("Expected YAML document separator '---' in file")
	}
}

func TestWriteCommentsFile(t *testing.T) {
	// Create a temporary file for testing
	tmpFile, err := os.CreateTemp("", "test_comments_*.md")
	if err != nil {
		t.Fatalf("Failed to create temp file: %v", err)
	}
	defer os.Remove(tmpFile.Name())
	tmpFile.Close()

	// Create a test entry
	entry := &CiREntry{
		Summary:    "Test summary content",
		prComments: []string{"Error 1", "Error 2"},
	}

	// Test writing comments
	err = writeCommentsFile(entry, tmpFile.Name())
	if err != nil {
		t.Errorf("writeCommentsFile() failed: %v", err)
	}

	// Verify the content was written
	content, err := os.ReadFile(tmpFile.Name())
	if err != nil {
		t.Fatalf("Failed to read temp file: %v", err)
	}

	contentStr := string(content)
	
	// Check that summary is present
	if !strings.Contains(contentStr, "Test summary content") {
		t.Error("Expected summary content in comments file")
	}

	// Check that errors section is present
	if !strings.Contains(contentStr, "## Errors") {
		t.Error("Expected errors section in comments file")
	}

	// Check that both errors are listed
	if !strings.Contains(contentStr, "* Error 1") {
		t.Error("Expected 'Error 1' in comments file")
	}

	if !strings.Contains(contentStr, "* Error 2") {
		t.Error("Expected 'Error 2' in comments file")
	}
}

func TestWriteCommentsFileNoErrors(t *testing.T) {
	// Create a temporary file for testing
	tmpFile, err := os.CreateTemp("", "test_comments_no_errors_*.md")
	if err != nil {
		t.Fatalf("Failed to create temp file: %v", err)
	}
	defer os.Remove(tmpFile.Name())
	tmpFile.Close()

	// Create a test entry without errors
	entry := &CiREntry{
		Summary:    "Test summary content",
		prComments: []string{}, // No errors
	}

	// Test writing comments
	err = writeCommentsFile(entry, tmpFile.Name())
	if err != nil {
		t.Errorf("writeCommentsFile() failed: %v", err)
	}

	// Verify the content was written
	content, err := os.ReadFile(tmpFile.Name())
	if err != nil {
		t.Fatalf("Failed to read temp file: %v", err)
	}

	contentStr := string(content)
	
	// Check that summary is present
	if !strings.Contains(contentStr, "Test summary content") {
		t.Error("Expected summary content in comments file")
	}

	// Check that errors section is NOT present when there are no errors
	if strings.Contains(contentStr, "## Errors") {
		t.Error("Should not have errors section when there are no errors")
	}
}
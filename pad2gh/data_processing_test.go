package main

import (
	"testing"
)

func TestExtractDateFromPadURL(t *testing.T) {
	tests := []struct {
		name     string
		padURL   string
		expected string
		wantErr  bool
	}{
		{
			name:     "Valid pad URL",
			padURL:   "https://pad.ccc-p.org/Radio_2024-01-15_test1",
			expected: "2024-01-15",
			wantErr:  false,
		},
		{
			name:     "Another valid pad URL",
			padURL:   "https://pad.ccc-p.org/Radio_2021-06-14_existing",
			expected: "2021-06-14",
			wantErr:  false,
		},
		{
			name:     "Invalid pad URL without date",
			padURL:   "https://pad.ccc-p.org/Radio_test",
			expected: "",
			wantErr:  true,
		},
		{
			name:     "Empty URL",
			padURL:   "",
			expected: "",
			wantErr:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := extractDateFromPadURL(tt.padURL)
			if (err != nil) != tt.wantErr {
				t.Errorf("extractDateFromPadURL() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if result != tt.expected {
				t.Errorf("extractDateFromPadURL() = %v, want %v", result, tt.expected)
			}
		})
	}
}

func TestFindFirstLink(t *testing.T) {
	tests := []struct {
		name     string
		line     string
		expected string
	}{
		{
			name:     "Line with HTTP link",
			line:     "Check this out: https://example.com/test",
			expected: "https://example.com/test",
		},
		{
			name:     "Line with HTTPS link",
			line:     "Visit https://secure.example.com",
			expected: "https://secure.example.com",
		},
		{
			name:     "Line without link",
			line:     "No link here",
			expected: "",
		},
		{
			name:     "Empty line",
			line:     "",
			expected: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, url := findFirstLink(tt.line)
			if url != tt.expected {
				t.Errorf("findFirstLink() = %v, want %v", url, tt.expected)
			}
		})
	}
}

func TestCreateMockEntry(t *testing.T) {
	padURL := "https://pad.ccc-p.org/Radio_2024-01-15_test1"
	date := "2024-01-15"

	entry := createMockEntry(padURL, date)

	// Test that entry was created correctly
	if entry.UUID != "nt-2024-01-15" {
		t.Errorf("Expected UUID 'nt-2024-01-15', got '%s'", entry.UUID)
	}

	if entry.Title != "CiR am 15.01.2024" {
		t.Errorf("Expected Title 'CiR am 15.01.2024', got '%s'", entry.Title)
	}

	if entry.PublicationDate != "2024-01-15T00:00:00+02:00" {
		t.Errorf("Expected PublicationDate '2024-01-15T00:00:00+02:00', got '%s'", entry.PublicationDate)
	}

	if len(entry.Audio) != 1 {
		t.Errorf("Expected 1 audio entry, got %d", len(entry.Audio))
	}

	if entry.Audio[0].MimeType != "audio/mp3" {
		t.Errorf("Expected mime type 'audio/mp3', got '%s'", entry.Audio[0].MimeType)
	}
}

func TestPopulateEntryFromSections(t *testing.T) {
	entry := &CiREntry{}
	contentBySection := getMockPadContent()
	entryDate := "2024-01-15"

	err := populateEntryFromSections(entry, contentBySection, entryDate)
	if err != nil {
		t.Errorf("populateEntryFromSections() failed: %v", err)
	}

	// Test that entry was populated correctly
	if entry.UUID != "nt-2024-01-15" {
		t.Errorf("Expected UUID 'nt-2024-01-15', got '%s'", entry.UUID)
	}

	if entry.Summary != "Test summary for mock pad entry" {
		t.Errorf("Expected Summary 'Test summary for mock pad entry', got '%s'", entry.Summary)
	}

	if len(entry.Chapters) != 3 {
		t.Errorf("Expected 3 chapters, got %d", len(entry.Chapters))
	}

	if entry.Chapters[0].Start != "00:00:00" {
		t.Errorf("Expected first chapter start '00:00:00', got '%s'", entry.Chapters[0].Start)
	}

	if entry.Chapters[0].Title != "Introduction" {
		t.Errorf("Expected first chapter title 'Introduction', got '%s'", entry.Chapters[0].Title)
	}
}

func TestCreatePadMapping(t *testing.T) {
	padURLs := []string{
		"https://pad.ccc-p.org/Radio_2024-01-15_test1",
		"https://pad.ccc-p.org/Radio_2024-02-12_test2",
	}

	existingEntries := make(map[string]*CiREntry)
	// Add one existing entry
	existingEntries["2024-01-15"] = &CiREntry{
		UUID: "nt-2024-01-15",
		Audio: []CiRaudio{{
			Url:      "$media_base_url/2024_01_15-chaos-im-radio.mp3",
			MimeType: "audio/mp3",
		}},
	}

	mappings, err := createPadMapping(padURLs, existingEntries, "", false)
	if err != nil {
		t.Errorf("createPadMapping() failed: %v", err)
	}

	if len(mappings) != 2 {
		t.Errorf("Expected 2 mappings, got %d", len(mappings))
	}

	// First mapping should have YAML entry
	if !mappings[0].HasYAMLEntry {
		t.Errorf("Expected first mapping to have YAML entry")
	}

	// Second mapping should not have YAML entry
	if mappings[1].HasYAMLEntry {
		t.Errorf("Expected second mapping to not have YAML entry")
	}
}

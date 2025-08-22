package main

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/sirupsen/logrus"
	"gopkg.in/yaml.v3"
)

// setSingleQuoteStyle recursively sets appropriate string nodes to use the correct YAML style
// to match the existing content.yaml format
func setSingleQuoteStyle(node *yaml.Node) {
	if node.Kind == yaml.MappingNode {
		// For mapping nodes, process key-value pairs
		for i := 0; i < len(node.Content); i += 2 {
			key := node.Content[i]
			value := node.Content[i+1]

			// Apply specific styles based on the key name
			if key.Kind == yaml.ScalarNode && value.Kind == yaml.ScalarNode && value.Tag == "!!str" {
				switch key.Value {
				case "summary":
					// Use folded scalar style (>)
					value.Style = yaml.FoldedStyle
				case "long_summary_md":
					// Use literal scalar style (|)
					value.Style = yaml.LiteralStyle
				case "publicationDate":
					// Use double quotes for publicationDate
					value.Style = yaml.DoubleQuotedStyle
				case "uuid", "subtitle", "url", "mimeType":
					// Use flow style for uuid, title, subtitle
					value.Style = yaml.FlowStyle
				default:
					// Use single quotes for all other string values
					value.Style = yaml.SingleQuotedStyle
				}
			}

			// Recursively process nested structures
			setSingleQuoteStyle(value)
		}
	} else if node.Kind == yaml.SequenceNode {
		// For sequence nodes, process all elements
		for _, child := range node.Content {
			setSingleQuoteStyle(child)
		}
	}
}

func appendEntryToYAML(entry *CiREntry, contentFilePath string) error {
	return appendMultipleEntriesToYAML([]*CiREntry{entry}, contentFilePath)
}

func appendMultipleEntriesToYAML(entries []*CiREntry, contentFilePath string) error {
	if len(entries) == 0 {
		return nil // Nothing to append
	}

	// Create a temporary file to write the new entries
	tmpFile, err := os.CreateTemp("", "temp_content_*.yaml")
	if err != nil {
		return fmt.Errorf("failed to create temp file: %v", err)
	}
	defer os.Remove(tmpFile.Name())

	// Write existing content if the file already exists
	if _, err := os.Stat(contentFilePath); err == nil {
		content, err := os.ReadFile(contentFilePath)
		if err != nil {
			return fmt.Errorf("failed to read existing content file: %v", err)
		}
		if len(content) > 0 {
			if _, err := tmpFile.Write(content); err != nil {
				return fmt.Errorf("failed to write existing content to temp file: %v", err)
			}
			// Ensure there's a newline before adding new entries
			if content[len(content)-2] != '-' {
				if _, err := tmpFile.WriteString("---\n"); err != nil {
					return fmt.Errorf("failed to write newline to temp file: %v", err)
				}
			}
		}
	}

	// Create a single encoder for all entries to ensure proper YAML document separators
	encoder := yaml.NewEncoder(tmpFile)
	encoder.SetIndent(2)

	for _, entry := range entries {
		node := &yaml.Node{}
		err = node.Encode(entry)
		if err != nil {
			return fmt.Errorf("failed to encode entry to node: %v", err)
		}
		setSingleQuoteStyle(node)

		err = encoder.Encode(node)
		if err != nil {
			_ = encoder.Close()
			return fmt.Errorf("failed to marshal entry: %v", err)
		}
	}

	err = encoder.Close()
	if err != nil {
		return fmt.Errorf("failed to close encoder: %v", err)
	}

	tmpFile.Close()

	// Move the temporary file to the original content file path
	err = os.Rename(tmpFile.Name(), contentFilePath)
	if err != nil {
		return fmt.Errorf("failed to rename temp file to content file: %v", err)
	}

	return nil
}

// readYAMLEntries reads all YAML entries from a file and returns them as a slice
func readYAMLEntries(filePath string) ([]*CiREntry, error) {
	var entries []*CiREntry

	file, err := os.Open(filePath)
	if err != nil {
		if os.IsNotExist(err) {
			return entries, nil // Return empty slice if file doesn't exist
		}
		return nil, err
	}
	defer file.Close()

	decoder := yaml.NewDecoder(file)
	for {
		var entry CiREntry
		err := decoder.Decode(&entry)
		if err != nil {
			if err == io.EOF {
				break
			}
			return nil, err
		}
		entries = append(entries, &entry)
	}

	return entries, nil
}

// readExistingYAMLEntries reads YAML entries and returns them as a map keyed by date
func readExistingYAMLEntries(filePath string) (map[string]*CiREntry, error) {
	entries, err := readYAMLEntries(filePath)
	if err != nil {
		return nil, err
	}

	entriesMap := make(map[string]*CiREntry)
	for _, entry := range entries {
		// Extract date from UUID or publication date to create key
		var dateKey string
		if strings.HasPrefix(entry.UUID, "nt-") {
			dateKey = strings.TrimPrefix(entry.UUID, "nt-")
		} else {
			// Try to extract from publication date
			if len(entry.PublicationDate) >= 10 {
				dateKey = entry.PublicationDate[:10]
			}
		}

		if dateKey != "" {
			entriesMap[dateKey] = entry
		} else {
			logrus.Warnf("Entry without valid date found: %s", entry.UUID)
		}
	}

	return entriesMap, nil
}

// EntryWithOrder represents an entry with its parsed date for sorting
type EntryWithOrder struct {
	Entry *CiREntry
	Date  time.Time
}

// insertEntryToYAMLInOrder reads all existing entries, inserts the new entry in the correct position based on date, and rewrites the file
func insertEntryToYAMLInOrder(entry *CiREntry, contentFilePath string) error {
	return insertMultipleEntriesToYAMLInOrder([]*CiREntry{entry}, contentFilePath)
}

// insertMultipleEntriesToYAMLInOrder reads all existing entries, inserts multiple new entries in the correct positions based on date, and rewrites the file once
func insertMultipleEntriesToYAMLInOrder(newEntries []*CiREntry, contentFilePath string) error {
	// Read all existing entries
	entries, err := readYAMLEntries(contentFilePath)
	if err != nil {
		return fmt.Errorf("failed to read existing entries: %v", err)
	}

	// Create EntryWithOrder slice for existing entries
	var orderedEntries []EntryWithOrder
	for _, existingEntry := range entries {
		existingDate, err := parseEntryDate(existingEntry)
		if err != nil {
			// If we can't parse the date, append at the end with zero time
			orderedEntries = append(orderedEntries, EntryWithOrder{Entry: existingEntry, Date: time.Time{}})
			continue
		}
		orderedEntries = append(orderedEntries, EntryWithOrder{Entry: existingEntry, Date: existingDate})
	}

	// Add all new entries to the slice
	for _, newEntry := range newEntries {
		newEntryDate, err := parseEntryDate(newEntry)
		if err != nil {
			return fmt.Errorf("failed to parse new entry date for %s: %v", newEntry.UUID, err)
		}
		orderedEntries = append(orderedEntries, EntryWithOrder{Entry: newEntry, Date: newEntryDate})
	}

	// Sort all entries by date to ensure correct order
	sort.Slice(orderedEntries, func(i, j int) bool {
		return orderedEntries[i].Date.Before(orderedEntries[j].Date)
	})

	// Write all entries back to the file
	return writeAllYAMLEntries(orderedEntries, contentFilePath)
}

// parseEntryDate extracts and parses the date from a CiREntry
func parseEntryDate(entry *CiREntry) (time.Time, error) {
	var dateStr string

	// Try to extract date from UUID first (format: nt-YYYY-MM-DD)
	if strings.HasPrefix(entry.UUID, "nt-") {
		dateStr = strings.TrimPrefix(entry.UUID, "nt-")
	} else if len(entry.PublicationDate) >= 10 {
		// Try to extract from publication date (format: YYYY-MM-DDTHH:MM:SS+TZ)
		dateStr = entry.PublicationDate[:10]
	} else {
		return time.Time{}, fmt.Errorf("unable to extract date from entry %s", entry.UUID)
	}

	// Parse the date
	parsedDate, err := time.Parse("2006-01-02", dateStr)
	if err != nil {
		return time.Time{}, fmt.Errorf("failed to parse date %s: %v", dateStr, err)
	}

	return parsedDate, nil
}

// writeAllYAMLEntries writes all entries to a YAML file
func writeAllYAMLEntries(orderedEntries []EntryWithOrder, contentFilePath string) error {
	file, err := os.Create(contentFilePath)
	if err != nil {
		return fmt.Errorf("failed to create file %s: %v", contentFilePath, err)
	}
	defer file.Close()

	encoder := yaml.NewEncoder(file)
	encoder.SetIndent(2)

	for _, entryWithOrder := range orderedEntries {
		// First marshal to a node to apply single quote style
		var node yaml.Node
		err := node.Encode(entryWithOrder.Entry)
		if err != nil {
			_ = encoder.Close()
			return fmt.Errorf("failed to encode entry to node: %v", err)
		}
		setSingleQuoteStyle(&node)

		err = encoder.Encode(&node)
		if err != nil {
			_ = encoder.Close()
			return fmt.Errorf("failed to marshal entry: %v", err)
		}
	}

	err = encoder.Close()
	if err != nil {
		return fmt.Errorf("failed to close encoder: %v", err)
	}

	return nil
}

func writeCommentsFile(entry *CiREntry, commentsFilePath string) error {
	commentsFile, err := os.Create(commentsFilePath)
	if err != nil {
		return err
	}
	defer commentsFile.Close()

	_, err = commentsFile.WriteString(entry.Summary)
	if err != nil {
		return fmt.Errorf("error writing %v: %v", commentsFilePath, err)
	}

	if len(entry.processingWarnings) > 0 {
		commentsFile.WriteString("\n\n## Errors ")
	}
	for _, c := range entry.processingWarnings {
		commentsFile.WriteString("\n* " + c)
	}

	return nil
}

func checkSoundFileExistsLocally(soundFileDir, soundFileName string) bool {
	filePath := filepath.Join(soundFileDir, soundFileName)
	_, err := os.Stat(filePath)
	return !os.IsNotExist(err)
}

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

func appendEntryToYAML(entry *CiREntry, contentFilePath string) error {
	b, _ := yaml.Marshal(entry)

	contentFile, err := os.OpenFile(contentFilePath, os.O_RDWR|os.O_APPEND|os.O_CREATE, 0666)
	if err != nil {
		return fmt.Errorf("error while opening %v: %v", contentFilePath, err)
	}
	defer contentFile.Close()

	// Add YAML document separator
	if fileInfo, err := contentFile.Stat(); err == nil && fileInfo.Size() > 0 {
		_, err = contentFile.WriteString("---\n")
		if err != nil {
			return fmt.Errorf("error writing separator to %v: %v", contentFilePath, err)
		}
	}

	_, err = contentFile.Write(b)
	if err != nil {
		return fmt.Errorf("error writing %v: %v", contentFilePath, err)
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
	// Read all existing entries
	entries, err := readYAMLEntries(contentFilePath)
	if err != nil {
		return fmt.Errorf("failed to read existing entries: %v", err)
	}

	// Parse the new entry's date
	newEntryDate, err := parseEntryDate(entry)
	if err != nil {
		return fmt.Errorf("failed to parse new entry date: %v", err)
	}

	// Create EntryWithOrder slice and find the correct position
	var orderedEntries []EntryWithOrder
	inserted := false

	for _, existingEntry := range entries {
		existingDate, err := parseEntryDate(existingEntry)
		if err != nil {
			// If we can't parse the date, append at the end
			orderedEntries = append(orderedEntries, EntryWithOrder{Entry: existingEntry, Date: time.Time{}})
			continue
		}

		// If we haven't inserted yet and the new entry should come before this existing entry
		if !inserted && newEntryDate.Before(existingDate) {
			orderedEntries = append(orderedEntries, EntryWithOrder{Entry: entry, Date: newEntryDate})
			inserted = true
		}

		orderedEntries = append(orderedEntries, EntryWithOrder{Entry: existingEntry, Date: existingDate})
	}

	// If we haven't inserted yet (new entry has the latest date), append at the end
	if !inserted {
		orderedEntries = append(orderedEntries, EntryWithOrder{Entry: entry, Date: newEntryDate})
	}

	// Sort by date to ensure correct order
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

	for i, entryWithOrder := range orderedEntries {
		// Add document separator before each entry except the first
		if i > 0 {
			_, err = file.WriteString("---\n")
			if err != nil {
				return fmt.Errorf("failed to write separator: %v", err)
			}
		}

		// Marshal and write the entry
		b, err := yaml.Marshal(entryWithOrder.Entry)
		if err != nil {
			return fmt.Errorf("failed to marshal entry: %v", err)
		}

		_, err = file.Write(b)
		if err != nil {
			return fmt.Errorf("failed to write entry: %v", err)
		}
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

	if len(entry.prComments) > 0 {
		commentsFile.WriteString("\n\n## Errors ")
	}
	for _, c := range entry.prComments {
		commentsFile.WriteString("\n* " + c)
	}

	return nil
}

func checkSoundFileExistsLocally(soundFileDir, soundFileName string) bool {
	filePath := filepath.Join(soundFileDir, soundFileName)
	_, err := os.Stat(filePath)
	return !os.IsNotExist(err)
}

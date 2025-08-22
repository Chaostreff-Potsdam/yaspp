package main

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"gopkg.in/yaml.v3"
)

func extractDateFromPadURL(padURL string) (string, error) {
	// Extract date in format YYYY-MM-DD from pad URL
	re := regexp.MustCompile(`_(\d{4}-\d{2}-\d{2})`)
	matches := re.FindStringSubmatch(padURL)
	if len(matches) < 2 {
		return "", fmt.Errorf("no date found in pad URL: %s", padURL)
	}
	return matches[1], nil
}

func readExistingYAMLEntries(filePath string) (map[string]*CiREntry, error) {
	entries := make(map[string]*CiREntry)

	file, err := os.Open(filePath)
	if err != nil {
		if os.IsNotExist(err) {
			return entries, nil // Return empty map if file doesn't exist
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
			entries[dateKey] = &entry
		}
	}

	return entries, nil
}

func checkSoundFileExistsLocally(soundFileDir, soundFileName string) bool {
	filePath := filepath.Join(soundFileDir, soundFileName)
	_, err := os.Stat(filePath)
	return !os.IsNotExist(err)
}

func createPadMapping(padURLs []string, existingEntries map[string]*CiREntry, soundDir string) ([]PadMapping, error) {
	var mappings []PadMapping

	for _, padURL := range padURLs {
		date, err := extractDateFromPadURL(padURL)
		if err != nil {
			continue // Skip URLs without valid dates
		}

		mapping := PadMapping{
			PadURL:            padURL,
			Date:              date,
			HasSoundFileLocal: false,
			HasYAMLEntry:      false,
		}

		// Generate expected sound file name
		parts := strings.Split(date, "-")
		if len(parts) == 3 {
			year, month, day := parts[0], parts[1], parts[2]
			mapping.SoundFileName = fmt.Sprintf("%s_%s_%s-chaos-im-radio.mp3", year, month, day)

			// Check local sound file if directory is provided
			if soundDir != "" {
				mapping.HasSoundFileLocal = checkSoundFileExistsLocally(soundDir, mapping.SoundFileName)
			}
		}

		// Check if YAML entry exists
		if entry, exists := existingEntries[date]; exists {
			mapping.HasYAMLEntry = true
			mapping.YAMLEntry = entry
		}

		mappings = append(mappings, mapping)
	}

	return mappings, nil
}

func findFirstLink(line string) string {
	for _, linkCandidate := range strings.Split(line, " ") {
		if strings.HasPrefix(linkCandidate, "http") {
			return linkCandidate
		}
	}
	return ""
}

func getMarkdownContentBySection(padURL string) (map[string][]string, error) {
	padContent, err := getPadContent(padURL)
	if err != nil {
		return nil, err
	}
	defer padContent.Close()

	// parse the content to find the first link
	scanner := bufio.NewScanner(padContent)
	currentSection := "pre-section"
	currentSectionContent := []string{}
	contentBySection := make(map[string][]string)
	for scanner.Scan() {
		line := scanner.Text()

		if strings.HasPrefix(line, "## ") {
			contentBySection[currentSection] = currentSectionContent
			currentSectionContent = []string{}
			currentSection = strings.TrimPrefix(line, "##")
			currentSection = strings.ToLower(currentSection)
			currentSection = strings.Trim(currentSection, " ")
			continue
		} else if strings.HasPrefix(line, "#") {
			continue
		}
		currentSectionContent = append(currentSectionContent, strings.Trim(line, " "))
	}
	contentBySection[currentSection] = currentSectionContent
	if err := scanner.Err(); err != nil {
		return nil, err
	}
	return contentBySection, nil
}

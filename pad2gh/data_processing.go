package main

import (
	"bufio"
	"fmt"
	"io"
	"net/http"
	"os"
	"regexp"
	"strings"

	"gopkg.in/yaml.v3"
)

func extractDateFromPadURL(padURL string) (string, error) {
	// Extract date in format YYYY-MM-DD from pad URL
	re := regexp.MustCompile(`_(\d{4}-\d{2}-\d{2})_`)
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

func checkSoundFileExists(soundFileURL string) bool {
	// Replace template variable with actual base URL for checking
	// For now, we'll assume a default structure
	if strings.Contains(soundFileURL, "$media_base_url") {
		// We can't check template URLs, so we assume they exist if properly formatted
		return true
	}
	
	resp, err := http.Head(soundFileURL)
	if err != nil {
		return false
	}
	defer resp.Body.Close()
	
	return resp.StatusCode == 200
}

func createPadMapping(padURLs []string, existingEntries map[string]*CiREntry) ([]PadMapping, error) {
	var mappings []PadMapping
	
	for _, padURL := range padURLs {
		date, err := extractDateFromPadURL(padURL)
		if err != nil {
			continue // Skip URLs without valid dates
		}
		
		mapping := PadMapping{
			PadURL: padURL,
			Date:   date,
		}
		
		// Check if YAML entry exists
		if entry, exists := existingEntries[date]; exists {
			mapping.HasYAMLEntry = true
			mapping.YAMLEntry = entry
			if len(entry.Audio) > 0 {
				mapping.SoundFileURL = entry.Audio[0].Url
				mapping.HasSoundFile = checkSoundFileExists(entry.Audio[0].Url)
			}
		} else {
			// Generate expected sound file URL
			parts := strings.Split(date, "-")
			if len(parts) == 3 {
				year, month, day := parts[0], parts[1], parts[2]
				mapping.SoundFileURL = fmt.Sprintf("$media_base_url/%s_%s_%s-chaos-im-radio.mp3", year, month, day)
				mapping.HasSoundFile = checkSoundFileExists(mapping.SoundFileURL)
			}
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
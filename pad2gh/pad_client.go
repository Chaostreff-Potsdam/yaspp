package main

import (
	"bufio"
	"fmt"
	"io"
	"net/http"
	"regexp"
	"sort"
	"strings"
)

func getPadContent(padURL string) (io.ReadCloser, error) {
	// append the HedgeDoc API path to get the raw pad content
	padURL = strings.TrimSuffix(padURL, "/")
	padURL = fmt.Sprintf("%s/download", padURL)
	resp, err := http.Get(padURL)
	if err != nil {
		return nil, err
	}
	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("pad url must be accessible")
	}
	return resp.Body, nil
}

func getTitleFromFMA(fmaURL string) (string, error) {
	// append the HedgeDoc API path to get the raw pad content
	resp, err := http.Get(fmaURL)
	if err != nil {
		return "", err
	}
	if resp.StatusCode != 200 {
		return "", fmt.Errorf("%s returned status code %d", fmaURL, resp.StatusCode)
	}
	defer resp.Body.Close()
	// find the title tag in the html and return the content
	scanner := bufio.NewScanner(resp.Body)
	for scanner.Scan() {
		line := scanner.Text()
		if strings.HasPrefix(line, "<title>") {
			return strings.TrimSuffix(strings.TrimPrefix(line, "<title>"), "</title>"), nil
		}
	}
	if err := scanner.Err(); err != nil {
		return "", err
	}
	return fmaURL, nil
}

func getFirstLink(padURL string) (string, error) {
	padContent, err := getPadContent(padURL)
	if err != nil {
		return "", err
	}

	defer padContent.Close()

	// parse the content to find the first link
	scanner := bufio.NewScanner(padContent)
	for scanner.Scan() {
		line := scanner.Text()
		for _, linkCandidate := range strings.Split(line, "(") {
			if strings.HasPrefix(linkCandidate, "https://pad.ccc-p.org/") {
				if strings.HasSuffix(linkCandidate, ")") {
					link := strings.Split(linkCandidate, ")")
					if len(link) > 1 {
						return link[0], nil
					}
				}
			}
		}

	}
	if err := scanner.Err(); err != nil {
		return "", err
	}
	return "", nil
}

func getAllPadLinks(padURL string) ([]string, error) {
	padContent, err := getPadContent(padURL)
	if err != nil {
		return nil, err
	}

	defer padContent.Close()

	// regex to find all pad URLs in the format https://pad.ccc-p.org/*_YYYY-MM-DD or https://pad.ccc-p.org/*_YYYY-MM-DD_*
	re := regexp.MustCompile(`https://pad\.ccc-p\.org/[^\s\)]*_\d{4}-\d{2}-\d{2}[^\s\)]*`)
	var links []string
	linkSet := make(map[string]bool) // To avoid duplicates

	scanner := bufio.NewScanner(padContent)
	for scanner.Scan() {
		line := scanner.Text()
		matches := re.FindAllString(line, -1)
		for _, match := range matches {
			if !linkSet[match] {
				links = append(links, match)
				linkSet[match] = true
			}
		}
	}

	if err := scanner.Err(); err != nil {
		return nil, err
	}

	// Sort links to have a consistent order
	sort.Strings(links)
	return links, nil
}

func extractDateFromPadURL(padURL string) (string, error) {
	// Extract date in format YYYY-MM-DD from pad URL
	re := regexp.MustCompile(`_(\d{4}-\d{2}-\d{2})`)
	matches := re.FindStringSubmatch(padURL)
	if len(matches) < 2 {
		return "", fmt.Errorf("no date found in pad URL: %s", padURL)
	}
	return matches[1], nil
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

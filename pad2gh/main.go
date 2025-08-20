package main

/* pad2gh is a simple tool to get the first link from https://pad.ccc-p.org/Radio, extract the information from the markdown text and create a github PR with the information */

import (
	"bufio"
	"flag"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"regexp"
	"sort"
	"strings"

	"github.com/sirupsen/logrus"
	"gopkg.in/yaml.v3"
)

// CiRaudio is the audio information for the podcast
type CiRaudio struct {
	Url      string `yaml:"url"`      // format: https://cdn.ccc-p.org/episodes/2021-01-01-episode.mp3
	MimeType string `yaml:"mimeType"` // format: audio/mpeg
}

// CiRChapter is the chapter information for the podcast
type CiRChapter struct {
	Start string // format: 00:00:00.000 `yaml:"start"`
	Title string `yaml:"title"`
}

// CiREntry is the podcast episode information
type CiREntry struct {
	UUID            string       `yaml:"uuid"`
	Title           string       `yaml:"title"`
	Subtitle        string       `yaml:"subtitle"`
	Summary         string       `yaml:"summary"`
	PublicationDate string       `yaml:"publicationDate"`
	Audio           []CiRaudio   `yaml:"audio"`
	Chapters        []CiRChapter `yaml:"chapters"`
	LongSummaryMD   string       `yaml:"long_summary_md"`
	padURL          string
	prComments      []string
}

// PadMapping represents the mapping between pads, YAML entries and sound files
type PadMapping struct {
	PadURL       string
	Date         string
	HasYAMLEntry bool
	YAMLEntry    *CiREntry
	HasSoundFile bool
	SoundFileURL string
}

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
		return "", fmt.Errorf("fma url must be accessible")
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

	// regex to find all pad URLs in the format https://pad.ccc-p.org/*_YYYY-MM-DD_*
	re := regexp.MustCompile(`https://pad\.ccc-p\.org/[^)\s]*_\d{4}-\d{2}-\d{2}_[^)\s]*`)
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

func main() {
	logger := logrus.StandardLogger()
	contentFilePath := flag.String("o", "../content.yaml", "specify the yaml file to write to")
	commentsFilePath := flag.String("c", "../comments.md", "specify the yaml file to write to")
	padURLPtr := flag.String("l", "", "specify the link to the pad entry you want to parse")
	verbose := flag.Bool("v", false, "verbose output")
	bulkMode := flag.Bool("bulk", false, "process all pad entries found on the Radio page")
	mapOnly := flag.Bool("map-only", false, "only create mapping report, don't add new entries")
	testMode := flag.Bool("test", false, "run in test mode with mock data")
	
	if *verbose {
		logger.SetLevel(logrus.DebugLevel)
	}
	flag.Parse()

	if *bulkMode {
		err := processBulkMode(logger, *contentFilePath, *mapOnly, *testMode)
		if err != nil {
			logger.Fatalf("Error in bulk mode: %v", err)
		}
		return
	}

	// Original single-entry processing mode
	var entry CiREntry
	var err error
	if padURLPtr == nil || *padURLPtr == "" {
		if *testMode {
			logger.Info("Test mode: using mock pad URL")
			entry.padURL = "https://pad.ccc-p.org/test_2024-01-15_example"
		} else {
			entry.padURL, err = getFirstLink("https://pad.ccc-p.org/Radio")
			if err != nil {
				log.Fatal(err)
			}
		}
	} else {
		if !strings.HasPrefix(*padURLPtr, "https://pad.ccc-p.org/") {
			log.Fatal("pad url must start with https://pad.ccc-p.org/")
		}
		entry.padURL = *padURLPtr
	}
	
	err = processSingleEntry(logger, &entry, *contentFilePath, *commentsFilePath, *testMode)
	if err != nil {
		logger.Fatalf("Error processing single entry: %v", err)
	}
}

func processBulkMode(logger *logrus.Logger, contentFilePath string, mapOnly bool, testMode bool) error {
	logger.Info("Running in bulk mode - processing all pad entries")
	
	// Get all pad URLs from the Radio page
	logger.Info("Fetching all pad URLs from Radio page...")
	var padURLs []string
	var err error
	
	if testMode {
		logger.Info("Test mode: using mock pad URLs")
		padURLs = getMockPadURLs()
	} else {
		padURLs, err = getAllPadLinks("https://pad.ccc-p.org/Radio")
		if err != nil {
			return fmt.Errorf("failed to get pad URLs: %v", err)
		}
	}
	logger.Infof("Found %d pad URLs", len(padURLs))
	
	// Read existing YAML entries
	logger.Info("Reading existing YAML entries...")
	existingEntries, err := readExistingYAMLEntries(contentFilePath)
	if err != nil {
		return fmt.Errorf("failed to read existing YAML entries: %v", err)
	}
	logger.Infof("Found %d existing YAML entries", len(existingEntries))
	
	// Create mapping
	logger.Info("Creating mapping between pads, YAML entries, and sound files...")
	mappings, err := createPadMapping(padURLs, existingEntries)
	if err != nil {
		return fmt.Errorf("failed to create mapping: %v", err)
	}
	
	// Print mapping report
	printMappingReport(logger, mappings)
	
	if mapOnly {
		logger.Info("Map-only mode: skipping creation of new entries")
		return nil
	}
	
	// Create entries for pads without YAML entries
	newEntries := 0
	for _, mapping := range mappings {
		if !mapping.HasYAMLEntry {
			logger.Infof("Creating entry for pad: %s (date: %s)", mapping.PadURL, mapping.Date)
			var entry *CiREntry
			if testMode {
				entry = createMockEntry(mapping.PadURL, mapping.Date)
			} else {
				entry, err = createEntryFromPad(mapping.PadURL)
				if err != nil {
					logger.Errorf("Failed to create entry for %s: %v", mapping.PadURL, err)
					continue
				}
			}
			
			err = appendEntryToYAML(entry, contentFilePath)
			if err != nil {
				logger.Errorf("Failed to append entry to YAML: %v", err)
				continue
			}
			newEntries++
		}
	}
	
	logger.Infof("Created %d new entries", newEntries)
	return nil
}

func processSingleEntry(logger *logrus.Logger, entry *CiREntry, contentFilePath, commentsFilePath string, testMode bool) error {
	logger.Debugf("pad url: %s\n", entry.padURL)

	var contentBySection map[string][]string
	var err error
	
	if testMode {
		contentBySection = getMockPadContent()
	} else {
		contentBySection, err = getMarkdownContentBySection(entry.padURL)
		if err != nil {
			return err
		}
	}

	if len(strings.Split(entry.padURL, "_")) < 2 {
		return fmt.Errorf("pad url must contain a date in the format YYYY-MM-DD_")
	}
	entryDate := strings.Split(entry.padURL, "_")[1]
	if len(entryDate) < 10 {
		return fmt.Errorf("pad url must contain a date in the format YYYY-MM-DD_")
	}

	// for the GitHub Action:
	fmt.Printf("entrydate=%s\n", entryDate)

	err = populateEntryFromSections(entry, contentBySection, entryDate)
	if err != nil {
		return err
	}

	b, _ := yaml.Marshal(entry)

	if contentFilePath == "" {
		fmt.Printf("%s", b)
		return nil
	}

	err = appendEntryToYAML(entry, contentFilePath)
	if err != nil {
		return err
	}

	if commentsFilePath == "" {
		logger.Warn(entry.prComments)
		return nil
	}

	return writeCommentsFile(entry, commentsFilePath)
}

func createEntryFromPad(padURL string) (*CiREntry, error) {
	entry := &CiREntry{padURL: padURL}
	
	contentBySection, err := getMarkdownContentBySection(padURL)
	if err != nil {
		return nil, err
	}

	if len(strings.Split(padURL, "_")) < 2 {
		return nil, fmt.Errorf("pad url must contain a date in the format YYYY-MM-DD_")
	}
	entryDate := strings.Split(padURL, "_")[1]
	if len(entryDate) < 10 {
		return nil, fmt.Errorf("pad url must contain a date in the format YYYY-MM-DD_")
	}

	err = populateEntryFromSections(entry, contentBySection, entryDate)
	if err != nil {
		return nil, err
	}
	
	return entry, nil
}

func populateEntryFromSections(entry *CiREntry, contentBySection map[string][]string, entryDate string) error {
	year := entryDate[0:4]
	month := entryDate[5:7]
	day := entryDate[8:10]
	
	longSummary, exists := contentBySection["shownotes"]
	if !exists {
		longSummary, exists = contentBySection["long summary"]
	}
	if !exists {
		return fmt.Errorf("no shownotes Section in Pad")
	}
	
	shortSummary, exists := contentBySection["summary"]
	if !exists {
		return fmt.Errorf("no Summary Section in Pad")
	}

	entry.UUID = fmt.Sprintf("nt-%s-%s-%s", year, month, day)
	entry.Title = fmt.Sprintf("CiR am %s.%s.%s", day, month, year)
	entry.Subtitle = "Der Chaostreff im Freien Radio Potsdam"
	entry.Summary = strings.Join(shortSummary, "\n")
	entry.PublicationDate = fmt.Sprintf("%s-%s-%sT00:00:00+02:00", year, month, day)
	entry.Audio = []CiRaudio{{
		Url:      fmt.Sprintf("$media_base_url/%s_%s_%s-chaos-im-radio.mp3", year, month, day),
		MimeType: "audio/mp3",
	}}
	entry.LongSummaryMD = "**Shownotes:**\n" + strings.Join(longSummary, "\n")

	chapter, exists := contentBySection["chapters"]
	if exists {
		entry.Chapters = []CiRChapter{}
		for _, c := range chapter {
			chapter := strings.Split(c, " ")
			if len(chapter) < 2 {
				continue
			}
			entry.Chapters = append(entry.Chapters, CiRChapter{Start: chapter[0], Title: strings.Join(chapter[1:], " ")})
		}
	} else {
		entry.prComments = append(entry.prComments, "no chapters Section in Pad")
	}

	mukke, exists := contentBySection["mukke"]
	if exists {
		for _, m := range mukke {
			if strings.TrimSpace(m) == "" {
				continue
			}
			link := findFirstLink(m)
			if link == "" {
				entry.prComments = append(entry.prComments, fmt.Sprintf("no link found in mukke line: %s", m))
				continue
			}
			title, err := getTitleFromFMA(link)
			if err != nil {
				entry.prComments = append(entry.prComments, fmt.Sprintf("error getting title from fma: %s", err.Error()))
				title = link
			}
			entry.LongSummaryMD = entry.LongSummaryMD + fmt.Sprintf("\n&#x1f3b6;&nbsp;[%s](%s)", title, link)
		}
	} else {
		entry.prComments = append(entry.prComments, "no mukke Section in Pad")
	}
	
	return nil
}

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

func printMappingReport(logger *logrus.Logger, mappings []PadMapping) {
	logger.Info("=== PAD MAPPING REPORT ===")
	
	totalPads := len(mappings)
	withYAML := 0
	withSoundFile := 0
	complete := 0
	
	for _, mapping := range mappings {
		status := "MISSING"
		if mapping.HasYAMLEntry && mapping.HasSoundFile {
			status = "COMPLETE"
			complete++
		} else if mapping.HasYAMLEntry {
			status = "YAML ONLY"
		} else if mapping.HasSoundFile {
			status = "SOUND ONLY"
		}
		
		if mapping.HasYAMLEntry {
			withYAML++
		}
		if mapping.HasSoundFile {
			withSoundFile++
		}
		
		logger.Infof("Date: %s | Status: %s | Pad: %s", mapping.Date, status, mapping.PadURL)
	}
	
	logger.Infof("=== SUMMARY ===")
	logger.Infof("Total pads found: %d", totalPads)
	logger.Infof("With YAML entries: %d", withYAML)
	logger.Infof("With sound files: %d", withSoundFile)
	logger.Infof("Complete (both): %d", complete)
	logger.Infof("Missing YAML entries: %d", totalPads-withYAML)
}

// Mock functions for testing
func getMockPadURLs() []string {
	return []string{
		"https://pad.ccc-p.org/Radio_2024-01-15_test1",
		"https://pad.ccc-p.org/Radio_2024-02-12_test2", 
		"https://pad.ccc-p.org/Radio_2024-03-11_test3",
		"https://pad.ccc-p.org/Radio_2023-12-01_test4",
		"https://pad.ccc-p.org/Radio_2021-06-14_existing", // This one exists in the YAML
	}
}

func getMockPadContent() map[string][]string {
	return map[string][]string{
		"summary": {"Test summary for mock pad entry"},
		"shownotes": {
			"* Test shownote 1",
			"* Test shownote 2", 
			"* [Test Link](https://example.com)",
		},
		"chapters": {
			"00:00:00 Introduction",
			"00:05:30 Main Topic",
			"00:15:00 Conclusion",
		},
		"mukke": {
			"Test Music: https://freemusicarchive.org/music/test/song1",
		},
	}
}

func createMockEntry(padURL, date string) *CiREntry {
	parts := strings.Split(date, "-")
	year, month, day := parts[0], parts[1], parts[2]
	
	entry := &CiREntry{
		UUID:            fmt.Sprintf("nt-%s-%s-%s", year, month, day),
		Title:           fmt.Sprintf("CiR am %s.%s.%s", day, month, year),
		Subtitle:        "Der Chaostreff im Freien Radio Potsdam",
		Summary:         "Mock summary for testing bulk functionality",
		PublicationDate: fmt.Sprintf("%s-%s-%sT00:00:00+02:00", year, month, day),
		Audio: []CiRaudio{{
			Url:      fmt.Sprintf("$media_base_url/%s_%s_%s-chaos-im-radio.mp3", year, month, day),
			MimeType: "audio/mp3",
		}},
		Chapters: []CiRChapter{
			{Start: "00:00:00", Title: "Mock Introduction"},
			{Start: "00:05:30", Title: "Mock Main Topic"},
		},
		LongSummaryMD: "**Shownotes:**\n* Mock shownote for testing\n* Generated by test mode",
		padURL:        padURL,
		prComments:    []string{"Generated in test mode"},
	}
	
	return entry
}

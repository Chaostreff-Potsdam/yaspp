package main

import (
	"fmt"
	"strings"

	"github.com/sirupsen/logrus"
	"gopkg.in/yaml.v3"
)

func processSingleEntry(logger *logrus.Logger, entry *CiREntry, contentFilePath, commentsFilePath string) error {
	logger.Debugf("pad url: %s\n", entry.padURL)

	contentBySection, err := getMarkdownContentBySection(entry.padURL)
	if err != nil {
		return err
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

	// Print warnings if any
	if len(entry.processingWarnings) > 0 {
		logger.Warnf("Processing warnings for %s:", entry.padURL)
		for _, warning := range entry.processingWarnings {
			logger.Warnf("  - %s", warning)
		}
	}

	b, _ := yaml.Marshal(entry)

	if contentFilePath == "" {
		fmt.Printf("%s", b)
		return nil
	}

	err = insertEntryToYAMLInOrder(entry, contentFilePath)
	if err != nil {
		return err
	}

	if commentsFilePath == "" {
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

	mukke, exists := contentBySection["mukke"]
	if !exists {
		return fmt.Errorf("no mukke Section in Pad - skipping entry to not risk licensing issues")
	}

	entry.UUID = fmt.Sprintf("nt-%s-%s-%s", year, month, day)
	entry.Title = fmt.Sprintf("CiR am %s.%s.%s", day, month, year)
	entry.Subtitle = "Der Chaostreff im Freien Radio Potsdam"
	entry.PublicationDate = fmt.Sprintf("%s-%s-%sT00:00:00+00:00", year, month, day)

	shortSummary, exists := contentBySection["summary"]
	if !exists {
		entry.processingWarnings = append(entry.processingWarnings, "no summary Section in Pad")
		shortSummary = []string{"Chaos im Radio am " + day + "." + month + "." + year}
	}

	longSummary, exists := contentBySection["shownotes"]
	if !exists {
		longSummary, exists = contentBySection["long summary"]
	}
	if !exists {
		entry.processingWarnings = append(entry.processingWarnings, "no long summary Section in Pad, using short summary")
		entry.LongSummaryMD = "**Shownotes:**\n" + strings.Join(shortSummary, "\n")
	} else {
		entry.LongSummaryMD = "**Shownotes:**\n" + strings.Join(longSummary, "\n")
	}

	entry.Summary = strings.Join(shortSummary, "\n")
	entry.Audio = []CiRaudio{{
		Url:      fmt.Sprintf("$media_base_url/%s_%s_%s-chaos-im-radio.mp3", year, month, day),
		MimeType: "audio/mp3",
	}}

	music := []string{}
	for _, m := range mukke {
		if strings.TrimSpace(m) == "" {
			continue
		}
		title, link := findFirstLink(m)
		if link == "" {
			continue
		}
		htmlTitle, err := getTitleFromLink(link)
		if err != nil {
			entry.processingWarnings = append(entry.processingWarnings, fmt.Sprintf("error getting title from fma: %s", err.Error()))
			title = link
		}
		if title == "" {
			title = htmlTitle
		}
		music = append(music, fmt.Sprintf("[%s](%s)", title, link))
		entry.LongSummaryMD = entry.LongSummaryMD + fmt.Sprintf("\n&#x1f3b6;&nbsp;[%s](%s)", title, link)
	}
	if len(music) == 0 {
		entry.processingWarnings = append(entry.processingWarnings, "no music found in mukke Section")
	}

	chapter, exists := contentBySection["chapters"]
	if !exists {
		chapter, exists = contentBySection["kapitel"]
	}
	if exists {
		entry.Chapters = []CiRChapter{}
		for _, c := range chapter {
			chapter := strings.Split(c, " ")
			if len(chapter) < 2 {
				continue
			}
			title := strings.Join(chapter[1:], " ")
			href := ""
			if strings.HasPrefix(title, "http") {
				var err error
				href = title
				title, err = getTitleFromLink(href)
				if err != nil {
					title = href
					href = ""
				}

			}
			entry.Chapters = append(entry.Chapters, CiRChapter{Start: chapter[0], Title: title, Href: href})
		}
	} else {
		entry.processingWarnings = append(entry.processingWarnings, "no chapters Section in Pad")
	}

	return nil
}

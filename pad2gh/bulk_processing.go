package main

import (
	"fmt"

	"github.com/sirupsen/logrus"
)

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
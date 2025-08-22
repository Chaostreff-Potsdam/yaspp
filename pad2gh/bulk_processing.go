package main

import (
	"fmt"

	"github.com/sirupsen/logrus"
)

func processBulkMode(logger *logrus.Logger, contentFilePath string, mapOnly bool, soundDir string, continueOnError bool, strictMode bool) error {
	logger.Info("Running in bulk mode - processing all pad entries")

	// Get all pad URLs from the Radio page
	logger.Info("Fetching all pad URLs from Radio page...")
	padURLs, err := getAllPadLinks("https://pad.ccc-p.org/Radio")
	if err != nil {
		return fmt.Errorf("failed to get pad URLs: %v", err)
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
	mappings, err := createPadMapping(padURLs, existingEntries, soundDir)
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
	var newEntriesToAdd []*CiREntry

	// First, collect all new entries
	for _, mapping := range mappings {
		if !mapping.HasYAMLEntry && mapping.HasSoundFileLocal {
			logger.Infof("Processing pad: %s (date: %s)", mapping.PadURL, mapping.Date)
			entry, entryErr := createEntryFromPad(mapping.PadURL)
			if entryErr == nil && len(entry.processingWarnings) > 0 {
				logger.Warnf("Processing warnings for %s:", mapping.PadURL)
				for _, warning := range entry.processingWarnings {
					logger.Warnf("  - %s", warning)
				}

				// In strict mode, abort if there are warnings
				if strictMode {
					entryErr = fmt.Errorf("aborting due to warnings in strict mode for %s", mapping.PadURL)
				}
			}
			if entryErr != nil {
				logger.Errorf("Failed to create entry for %s: %v", mapping.PadURL, entryErr)
				if !continueOnError {
					return fmt.Errorf("failed to create entry for %s: %v", mapping.PadURL, entryErr)
				}
				continue
			}

			newEntriesToAdd = append(newEntriesToAdd, entry)
		}
	}

	// Add all new entries to YAML at once
	if len(newEntriesToAdd) > 0 {
		logger.Infof("Adding %d new entries to YAML file", len(newEntriesToAdd))
		err = insertMultipleEntriesToYAMLInOrder(newEntriesToAdd, contentFilePath)
		if err != nil {
			logger.Errorf("Failed to insert entries to YAML: %v", err)
			if !continueOnError {
				return fmt.Errorf("failed to insert entries to YAML: %v", err)
			}
		}
	}

	logger.Infof("Created %d new entries", len(newEntriesToAdd))
	return nil
}

func printMappingReport(logger *logrus.Logger, mappings []PadMapping) {
	fmt.Println("=== PAD MAPPING REPORT ===")

	totalPads := len(mappings)
	withYAML := 0
	withSoundFile := 0
	complete := 0

	for _, mapping := range mappings {
		status := ""
		if mapping.HasYAMLEntry {
			status = "YAML ENTRY"
			withYAML++
		} else {
			status = "NO YAML ENTRY"
		}
		if mapping.HasSoundFileLocal {
			status += " | File: " + mapping.SoundFileName
			withSoundFile++
		} else {
			status += " | NO SOUND FILE"
		}

		if !mapping.HasYAMLEntry {
			fmt.Printf("Date: %s | Pad: %s | Status: %s\n", mapping.Date, mapping.PadURL, status)
		}
	}

	fmt.Println("=== SUMMARY ===")
	fmt.Printf("Total pads found: %d\n", totalPads)
	fmt.Printf("With YAML entries: %d\n", withYAML)
	fmt.Printf("With sound files: %d\n", withSoundFile)
	fmt.Printf("Complete (both): %d\n", complete)
	fmt.Printf("Missing YAML entries: %d\n", totalPads-withYAML)
}

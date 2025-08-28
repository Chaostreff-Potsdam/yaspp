package main

import (
	"fmt"
	"net/url"

	"github.com/sirupsen/logrus"
)

func processBulkMode(logger *logrus.Logger, config *Config) error {
	logger.Info("Running in bulk mode - processing all pad entries")

	// Get all pad URLs from the Radio page
	logger.Info("Fetching all pad URLs from Radio page...")
	u, _ := url.JoinPath(config.PadBaseURL, "Radio")
	padURLs, err := getAllPadLinks(u)
	if err != nil {
		return fmt.Errorf("failed to get pad URLs from %s: %v", u, err)
	}
	logger.Infof("Found %d pad URLs on %s", len(padURLs), u)

	// Read existing YAML entries
	logger.Info("Reading existing YAML entries...")
	existingEntries, err := readExistingYAMLEntries(config.ContentFilePath)
	if err != nil {
		return fmt.Errorf("failed to read existing YAML entries: %v", err)
	}
	logger.Infof("Found %d existing YAML entries", len(existingEntries))

	// Create mapping
	logger.Info("Creating mapping between pads, YAML entries, and sound files...")
	mappings, err := createPadMapping(padURLs, existingEntries, config)
	if err != nil {
		return fmt.Errorf("failed to create mapping: %v", err)
	}

	printMappingReport(logger, mappings, config.FileOnline)

	if config.MapOnly {
		logger.Info("Map-only mode: skipping creation of new entries")
		return nil
	}

	// Create entries for pads without YAML entries
	var newEntriesToAdd []*CiREntry
	var entryDate string

	// First, collect all new entries (respecting maxNewEntries if > 0)
	for _, mapping := range mappings {
		if mapping.HasYAMLEntry {
			continue
		}
		if config.FileOnline && !mapping.HasSoundFileOnline {
			continue
		}
		if !config.FileOnline && !mapping.HasSoundFileLocal {
			continue
		}
		logger.Infof("Processing pad: %s (date: %s)", mapping.PadURL, mapping.Date)
		entry, entryErr := createEntryFromPad(mapping.PadURL)
		if entryErr == nil && len(entry.processingWarnings) > 0 {
			logger.Warnf("Processing warnings for %s:", mapping.PadURL)
			for _, warning := range entry.processingWarnings {
				logger.Warnf("  - %s", warning)
			}

			// In strict mode, abort if there are warnings
			if config.StrictMode {
				entryErr = fmt.Errorf("aborting due to warnings in strict mode for %s", mapping.PadURL)
			}
		}
		if entryErr != nil {
			logger.Errorf("Failed to create entry for %s: %v", mapping.PadURL, entryErr)
			if !config.ContinueOnError {
				return fmt.Errorf("failed to create entry for %s: %v", mapping.PadURL, entryErr)
			}
			continue
		}

		newEntriesToAdd = append(newEntriesToAdd, entry)

		// If maxNewEntries is set (>0) and we've reached the limit, stop collecting more
		if config.MaxNewEntries > 0 && len(newEntriesToAdd) >= config.MaxNewEntries {
			entryDate = mapping.Date
			logger.Infof("Reached max-new-entries limit (%d); stopping collection of new entries", config.MaxNewEntries)
			break
		}

	}

	// Add all new entries to YAML at once
	if len(newEntriesToAdd) > 0 {
		logger.Infof("Adding %d new entries to YAML file", len(newEntriesToAdd))
		// err = insertMultipleEntriesToYAMLInOrder(newEntriesToAdd, config.ContentFilePath)
		err = appendMultipleEntriesToYAML(newEntriesToAdd, config.ContentFilePath)
		if err != nil {
			logger.Errorf("Failed to insert entries to YAML: %v", err)
			if !config.ContinueOnError {
				return fmt.Errorf("failed to insert entries to YAML: %v", err)
			}
		}
	}

	logger.Infof("Created %d new entries", len(newEntriesToAdd))

	// For the GitHub Action, we only return the date of the last processed entry (if any) - will be used in the PR title and commit message
	fmt.Printf("entrydate=%s\n", entryDate)

	if config.CommentsFilePath == "" {
		return nil
	}

	return writeCommentsFile(newEntriesToAdd, config.CommentsFilePath)
}

func printMappingReport(logger *logrus.Logger, mappings []PadMapping, checkFileOnline bool) {
	logger.Info("=== PAD MAPPING REPORT ===")

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
		if !checkFileOnline {
			if mapping.HasSoundFileLocal {
				status += " | Local File: " + mapping.SoundFileName
				withSoundFile++
			} else {
				status += " | NO LOCAL SOUND FILE"
			}
		} else {
			if mapping.HasSoundFileOnline {
				status += " | File Online: " + mapping.SoundFileName
				withSoundFile++
			} else {
				status += " | NO SOUND FILE ONLINE"
			}
		}

		if !mapping.HasYAMLEntry || (checkFileOnline && !mapping.HasSoundFileOnline) {
			logger.Infof("Date: %s | Pad: %s | Status: %s\n", mapping.Date, mapping.PadURL, status)
		}
	}

	logger.Info("=== SUMMARY ===")
	logger.Infof("Total pads found: %d\n", totalPads)
	logger.Infof("With YAML entries: %d\n", withYAML)
	logger.Infof("With sound files: %d\n", withSoundFile)
	logger.Infof("Complete (both): %d\n", complete)
	logger.Infof("Missing YAML entries: %d\n", totalPads-withYAML)
}

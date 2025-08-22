package main

/* pad2gh is a simple tool to get the first link from https://pad.ccc-p.org/Radio, extract the information from the markdown text and create a github PR with the information */

import (
	"flag"
	"log"
	"strings"

	"github.com/sirupsen/logrus"
)

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

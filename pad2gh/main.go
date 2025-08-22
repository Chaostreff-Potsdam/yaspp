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
	soundDir := flag.String("sound-dir", "", "specify the local directory to check for sound files")
	continueOnError := flag.Bool("continue-on-error", false, "continue processing entries even if one fails (bulk mode only)")
	strictMode := flag.Bool("strict", false, "only create entries if there are no errors in the pad for this episode")

	if *verbose {
		logger.SetLevel(logrus.DebugLevel)
	}
	flag.Parse()

	if *bulkMode {
		err := processBulkMode(logger, *contentFilePath, *mapOnly, *soundDir, *continueOnError, *strictMode)
		if err != nil {
			logger.Fatalf("Error in bulk mode: %v", err)
		}
		return
	}

	// Original single-entry processing mode
	var entry CiREntry
	var err error
	if padURLPtr == nil || *padURLPtr == "" {
		entry.padURL, err = getFirstLink("https://pad.ccc-p.org/Radio")
		if err != nil {
			log.Fatal(err)
		}
	} else {
		if !strings.HasPrefix(*padURLPtr, "https://pad.ccc-p.org/") {
			log.Fatal("pad url must start with https://pad.ccc-p.org/")
		}
		entry.padURL = *padURLPtr
	}

	err = processSingleEntry(logger, &entry, *contentFilePath, *commentsFilePath)
	if err != nil {
		logger.Fatalf("Error processing single entry: %v", err)
	}
}

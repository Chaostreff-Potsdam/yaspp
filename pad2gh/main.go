package main

/* pad2gh is a simple tool to get the first link from https://pad.ccc-p.org/Radio, extract the information from the markdown text and create a github PR with the information */

import (
	"flag"
	"log"
	"net/url"
	"strings"

	"github.com/sirupsen/logrus"
)

// Config holds all command line options for the application
type Config struct {
	ContentFilePath   string
	CommentsFilePath  string
	PadURL           string
	Verbose          bool
	BulkMode         bool
	MapOnly          bool
	SoundDir         string
	FileOnline       bool
	ContinueOnError  bool
	StrictMode       bool
	MaxNewEntries    int
	PadBaseURL       string
	FileBaseURL      string
}

func main() {
	logger := logrus.StandardLogger()
	
	// Parse command line flags into config struct
	config := parseFlags()
	
	if config.Verbose {
		logger.SetLevel(logrus.DebugLevel)
	}

	if config.BulkMode {
		err := processBulkMode(logger, config)
		if err != nil {
			logger.Fatalf("Error in bulk mode: %v", err)
		}
		return
	}

	// Original single-entry processing mode
	var entry CiREntry
	var err error
	if config.PadURL == "" {
		u, _ := url.JoinPath(config.PadBaseURL, "Radio")
		entry.padURL, err = getFirstLink(u, config.PadBaseURL)
		if err != nil {
			log.Fatal(err)
		}
	} else {
		if !strings.HasPrefix(config.PadURL, config.PadBaseURL) {
			log.Fatal("pad url must start with " + config.PadBaseURL)
		}
		entry.padURL = config.PadURL
	}

	err = processSingleEntry(logger, &entry, config)
	if err != nil {
		logger.Fatalf("Error processing single entry: %v", err)
	}
}

// parseFlags parses command line flags and returns a Config struct
func parseFlags() *Config {
	config := &Config{
		PadBaseURL:  "https://pad.ccc-p.org/",
		FileBaseURL: "https://radio.ccc-p.org/files/",
	}
	
	flag.StringVar(&config.ContentFilePath, "o", "../content.yaml", "specify the yaml file to write to")
	flag.StringVar(&config.CommentsFilePath, "c", "../pr-comments.md", "specify the md file to write PR comments to")
	flag.StringVar(&config.PadURL, "l", "", "specify the link to the pad entry you want to parse")
	flag.BoolVar(&config.Verbose, "v", false, "verbose output")
	flag.BoolVar(&config.BulkMode, "bulk", false, "process all pad entries found on the Radio page")
	flag.BoolVar(&config.MapOnly, "map-only", false, "only create mapping report, don't add new entries")
	flag.StringVar(&config.SoundDir, "sound-dir", "", "specify the local directory to check for sound files")
	flag.BoolVar(&config.FileOnline, "file-online", false, "check sound files via HTTP instead of checking local directory")
	flag.BoolVar(&config.ContinueOnError, "continue-on-error", false, "continue processing entries even if one fails (bulk mode only)")
	flag.BoolVar(&config.StrictMode, "strict", false, "only create entries if there are no errors in the pad for this episode")
	flag.IntVar(&config.MaxNewEntries, "max-new-entries", 0, "limit number of new entries to create in bulk mode (0 = unlimited)")
	flag.StringVar(&config.PadBaseURL, "pad-base-url", "https://pad.ccc-p.org/", "base URL for pad entries")
	flag.StringVar(&config.FileBaseURL, "file-base-url", "https://radio.ccc-p.org/files/", "base URL for sound files")
	
	flag.Parse()
	return config
}

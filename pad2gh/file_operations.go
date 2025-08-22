package main

import (
	"fmt"
	"os"

	"gopkg.in/yaml.v3"
)

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
package main

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
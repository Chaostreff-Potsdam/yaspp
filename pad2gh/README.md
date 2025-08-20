# pad2gh - Pad to GitHub Tool

A tool to extract information from HedgeDoc pads and create GitHub PR entries for podcast episodes.

## Features

### Single Entry Mode (Original)
Processes a single pad entry and adds it to the YAML file:

```bash
# Process the first pad from Radio page
./pad2gh -o content.yaml -c comments.md

# Process a specific pad URL
./pad2gh -l "https://pad.ccc-p.org/Radio_2024-01-15_episode" -o content.yaml
```

### Bulk Mode (New)
Processes all pad entries found on the Radio page and creates a mapping between:
- All pad entries online
- Existing entries in the YAML file  
- Actual sound files

```bash
# Create mapping report only (no new entries)
./pad2gh -bulk -map-only -o content.yaml

# Create new YAML entries for all missing pads
./pad2gh -bulk -o content.yaml

# Test mode with mock data
./pad2gh -bulk -test -map-only
```

## Command Line Options

- `-bulk`: Process all pad entries found on the Radio page
- `-map-only`: Only create mapping report, don't add new entries
- `-test`: Run in test mode with mock data
- `-o <file>`: Specify the YAML file to write to (default: "../content.yaml")
- `-c <file>`: Specify the comments file to write to (default: "../comments.md")
- `-l <url>`: Specify a specific pad URL to parse
- `-v`: Verbose output

## Output

### Mapping Report
The bulk mode provides a detailed report showing:
- Total pads found
- Which pads have YAML entries
- Which pads have sound files
- Complete entries (both YAML and sound file)
- Missing YAML entries

### YAML Format
Generated entries follow the standard podcast YAML format:
```yaml
uuid: nt-2024-01-15
title: CiR am 15.01.2024
subtitle: Der Chaostreff im Freien Radio Potsdam
summary: Episode summary text
publicationDate: "2024-01-15T00:00:00+02:00"
audio:
  - url: $media_base_url/2024_01_15-chaos-im-radio.mp3
    mimeType: audio/mp3
chapters:
  - start: "00:00:00"
    title: Introduction
long_summary_md: |
  **Shownotes:**
  * Show note items
```

## Examples

1. **Check what's missing**: `./pad2gh -bulk -map-only -v`
2. **Add all missing entries**: `./pad2gh -bulk -v`  
3. **Test with mock data**: `./pad2gh -bulk -test -map-only`
4. **Process single entry**: `./pad2gh -l "https://pad.ccc-p.org/Radio_2024-01-15_episode"`
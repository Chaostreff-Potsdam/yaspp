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

# Process with sound file checking via HTTP
./pad2gh -bulk -file-online -o content.yaml

# Process only 5 new entries maximum
./pad2gh -bulk -max-new-entries=5 -o content.yaml

# Continue processing even if some entries fail
./pad2gh -bulk -continue-on-error -o content.yaml

# Only create entries with no pad errors (strict mode)
./pad2gh -bulk -strict -o content.yaml
```

## Command Line Options

- `-bulk`: Process all pad entries found on the Radio page
- `-map-only`: Only create mapping report, don't add new entries
- `-o <file>`: Specify the YAML file to write to (default: "../content.yaml")
- `-c <file>`: Specify the comments file to write to (default: "../pr-comments.md")
- `-l <url>`: Specify a specific pad URL to parse
- `-v`: Verbose output
- `-sound-dir <dir>`: Specify the local directory to check for sound files
- `-file-online`: Check sound files via HTTP instead of checking local directory
- `-continue-on-error`: Continue processing entries even if one fails (bulk mode only)
- `-strict`: Only create entries if there are no errors in the pad for this episode
- `-max-new-entries <n>`: Limit number of new entries to create in bulk mode (0 = unlimited)


## Examples

1. **Check what's missing**: `./pad2gh -bulk -map-only -v`
2. **Add all missing entries**: `./pad2gh -bulk -v`
3. **Add limited entries with online file checking**: `./pad2gh -bulk -file-online -max-new-entries=3`
4. **Process safely with error handling**: `./pad2gh -bulk -continue-on-error -strict`
5. **Process single entry**: `./pad2gh -l "https://pad.ccc-p.org/Radio_2024-01-15_episode"`
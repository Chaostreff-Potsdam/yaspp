#!/bin/bash
# Script to fix common speech recognition mistakes in transcription files (.vtt, .srt, .txt)
# Usage: ./fix-mistakes.sh [directory]
# If no directory is provided, it will process the current directory

# Define the directory to process
TARGET_DIR="${1:-.}"

# Check if the directory exists
if [ ! -d "$TARGET_DIR" ]; then
    echo "Error: Directory '$TARGET_DIR' does not exist."
    exit 1
fi

echo "Processing transcription files in: $TARGET_DIR"

# Define common misrecognized words and their corrections
# Format: "wrong_word:correct_word"
declare -A word_corrections=(
    # Names/Nicknames
    ["Syrox"]="Cyroxx"
    ["syrox"]="Cyroxx"
    ["Syrux"]="Cyroxx"
    ["syrux"]="Cyroxx"
    ["cyrox"]="Cyroxx"
    ["Cyrox"]="Cyroxx"
    ["TCyroxx"]="Cyroxx"
    ["Cybrox"]="Cyroxx"
    ["Genie"]="Gini"
    ["Joanie"]="Gini"
    ["Jeanie"]="Gini"
    ["Jeany"]="Gini"
    ["Jeannie"]="Gini"
    ["Jenny"]="Gini"
    ["Dini"]="Gini"
    ["Knurfs"]="Knurps"
    ["Knops"]="Knurps"
    ["Klof"]="Knurps"
    ["Knurbs"]="Knurps"
    ["Urnups"]="Knurps"
    ["Juwo"]="Ajuvo"
    ["Ayubo"]="Ajuvo"
    ["Ajuwo"]="Ajuvo"

    # Add more common misrecognitions here as needed
    ["Creative Comments"]="Creative Commons"
)

# Function to process a single file
process_file() {
    local file="$1"
    local temp_file=$(mktemp)
    local changes_made=false

    echo "Processing: $file"

    # Copy original content to temp file
    cp "$file" "$temp_file"

    # Apply each correction
    for wrong_word in "${!word_corrections[@]}"; do
        correct_word="${word_corrections[$wrong_word]}"

        # Use sed to replace whole words only (with word boundaries)
        # This prevents partial word replacements
        if sed -i "s/\b$wrong_word\b/$correct_word/g" "$temp_file"; then
            # Check if any changes were actually made
            if ! cmp -s "$file" "$temp_file"; then
                changes_made=true
                echo "  - Replaced '$wrong_word' with '$correct_word'"
            fi
        fi
    done

    # If changes were made, update the original file
    if [ "$changes_made" = true ]; then
        mv "$temp_file" "$file"
        echo "  ✓ File updated"
    else
        rm "$temp_file"
        echo "  - No changes needed"
    fi
}

# Counter for processed files
file_count=0

# Process all supported transcription files
find "$TARGET_DIR" -type f \( -name "*.vtt" -o -name "*.srt" -o -name "*.txt" \) | while read -r file; do
    # Skip this script itself
    if [[ "$(basename "$file")" == "fix-mistakes.sh" ]]; then
        continue
    fi

    process_file "$file"
    ((file_count++))
done

# remove all newlines from txt files
find "$TARGET_DIR" -type f -name "*.txt" | while read -r txt_file; do
    echo "Removing newlines in: $txt_file"
    tr '\n' ' ' < "$txt_file" > "${txt_file}.tmp"
    mv "${txt_file}.tmp" "$txt_file"
done

echo "Completed processing transcription files."
echo "Total files processed: $file_count"
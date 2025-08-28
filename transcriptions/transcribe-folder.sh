#!/bin/bash

set -e

# Check if both input and output directories are provided
if [ $# -ne 2 ]; then
    echo "Usage: $0 <input_directory> <output_directory>"
    echo "  input_directory:  Directory containing MP3 files to transcribe"
    echo "  output_directory: Directory where transcription files will be saved"
    exit 1
fi

INPUT_DIR="$1"
OUTPUT_DIR="$2"

# Validate input directory exists
if [ ! -d "$INPUT_DIR" ]; then
    echo "Error: Input directory '$INPUT_DIR' does not exist"
    exit 1
fi

# Create output directory if it doesn't exist
mkdir -p "$OUTPUT_DIR"

WHISPER_BIN=~/workspace/whisper.cpp/build/bin/whisper-cli
WHISPER_MODEL=~/workspace/whisper.cpp/models/ggml-small.bin
#WHISPER_MODEL=~/workspace/whisper.cpp/models/ggml-large-v3-turbo-german-q5_1.bin
LANGUAGE=DE

# Global variables for batch operations
GLOBAL_ACTION=""  # Can be "overwrite_all", "skip_all", or empty

# Function to ask for user confirmation
ask_confirmation() {
    local filename="$1"

    # If user already chose a global action, use it
    if [ "$GLOBAL_ACTION" = "overwrite_all" ]; then
        return 0
    elif [ "$GLOBAL_ACTION" = "skip_all" ]; then
        return 1
    fi

    echo "Transcription for '$filename' already exists."
    echo "Choose an action:"
    echo "  [y] Overwrite this file"
    echo "  [n] Skip this file (default)"
    echo "  [a] Overwrite ALL remaining files"
    echo "  [s] Skip ALL remaining files"
    echo -n "Your choice (y/n/a/s): "

    read -r response
    case "$response" in
        [yY][eE][sS]|[yY])
            return 0
            ;;
        [aA])
            GLOBAL_ACTION="overwrite_all"
            echo "Will overwrite all remaining files."
            return 0
            ;;
        [sS])
            GLOBAL_ACTION="skip_all"
            echo "Will skip all remaining files."
            return 1
            ;;
        *)
            return 1
            ;;
    esac
}

# Function to transcribe a single file
transcribe_file() {
    local input_file="$1"
    local output_file="$2"
    local working_dir=$(mktemp -d)
    local start_time=$(date +%s)

    echo "Converting to WAV: $(basename "$input_file")"

    ffmpeg -i "$input_file" -ar 16000 -ac 1 "${working_dir}/audio.wav" -y > /dev/null 2>&1

    echo "Transcribing: $(basename "$input_file")"
    local transcribe_start=$(date +%s)

    # Run transcription
    pushd "$working_dir" > /dev/null
    ${WHISPER_BIN} -m ${WHISPER_MODEL} -f audio.wav -np -pp -otxt -ovtt -oj -ocsv -osrt -l ${LANGUAGE} > /dev/null
    popd > /dev/null

    local transcribe_end=$(date +%s)
    local transcribe_duration=$((transcribe_end - transcribe_start))

    # Get the base filename without extension for output files
    basename_file=$(basename "$input_file" .mp3)

    # Move all whisper output files to output directory
    files_moved=0
    for ext in txt json csv vtt srt; do
        whisper_output="${working_dir}/audio.wav.${ext}"
        if [ -f "$whisper_output" ]; then
            output_dest="$OUTPUT_DIR/${basename_file}.${ext}"
            mv "$whisper_output" "$output_dest"
            files_moved=$((files_moved + 1))
        fi
    done

    local end_time=$(date +%s)
    local total_duration=$((end_time - start_time))

    if [ $files_moved -eq 0 ]; then
        echo "Error: No transcription files found"
    else
        echo "Transcription completed: $files_moved file(s) for '$(basename "$input_file")' Transcription took ${transcribe_duration}s, Total: ${total_duration}s"
    fi

    # Clean up
    rm -rf "$working_dir"
}

# Process all MP3 files in the input directory
for mp3_file in "$INPUT_DIR"/*.mp3; do
    # Check if any MP3 files exist
    if [ ! -f "$mp3_file" ]; then
        echo "No MP3 files found in '$INPUT_DIR'"
        exit 0
    fi

    # Get the base filename without extension
    basename_file=$(basename "$mp3_file" .mp3)

    # Check if any transcription files already exist
    files_exist=false
    for ext in txt json csv vtt srt; do
        if [ -f "$OUTPUT_DIR/${basename_file}.${ext}" ]; then
            files_exist=true
            break
        fi
    done

    # Check if transcription already exists
    if [ "$files_exist" = true ]; then
        if ask_confirmation "$(basename "$mp3_file")"; then
            if [ "$GLOBAL_ACTION" = "overwrite_all" ]; then
                echo "Overwriting: $(basename "$mp3_file") (overwrite all mode)"
            fi
            transcribe_file "$mp3_file" "$OUTPUT_DIR/${basename_file}"
        else
            if [ "$GLOBAL_ACTION" = "skip_all" ]; then
                echo "Skipping: $(basename "$mp3_file") (skip all mode)"
            else
                echo "Skipping: $(basename "$mp3_file")"
            fi
        fi
    else
        transcribe_file "$mp3_file" "$OUTPUT_DIR/${basename_file}"
    fi
done

echo "Transcription process completed!"

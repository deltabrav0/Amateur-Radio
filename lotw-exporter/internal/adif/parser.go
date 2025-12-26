package adif

import (
	"bufio"
	"fmt"
	"io"
	"strconv"
	"strings"
)

// Record represents a single ADIF record (a QSO).
type Record map[string]string

// Parse reads ADIF data from the reader and returns a slice of Records.
// It supports standard ADIF field formats: <FIELD_NAME:LENGTH:TYPE>DATA
func Parse(r io.Reader) ([]Record, error) {
	scanner := bufio.NewScanner(r)
	// Increase buffer size to handle large records/files if they come in big chunks
	// 50MB should be plenty for typical logs.
	buf := make([]byte, 0, 64*1024)
	scanner.Buffer(buf, 50*1024*1024)
	// Set a split function to read the whole stream if possible or chunk it?
	// ADIF is better parsed by reading until <EOR>.
	// However, bufio.Scanner default split is by lines.
	// We'll read the whole content for simplicity or implement a custom scanner.
	// Given LoTW reports can be large, a stream parser is better.

	// Custom split function for <EOR> or <eor>
	scanner.Split(func(data []byte, atEOF bool) (advance int, token []byte, err error) {
		if atEOF && len(data) == 0 {
			return 0, nil, nil
		}

		// Search for <EOR> case-insensitive
		s := string(data)
		idx := -1
		lowerS := strings.ToLower(s)
		if i := strings.Index(lowerS, "<eor>"); i >= 0 {
			idx = i
		}

		if idx >= 0 {
			return idx + 5, data[0:idx], nil
		}

		if atEOF {
			return len(data), data, nil
		}

		return 0, nil, nil
	})

	var records []Record

	for scanner.Scan() {
		text := strings.TrimSpace(scanner.Text())
		if text == "" {
			continue
		}
		// Header might be present, usually ends with <EOH>.
		// If we find <EOH>, we discard everything before it.
		if idx := strings.Index(strings.ToLower(text), "<eoh>"); idx >= 0 {
			text = text[idx+5:]
		}

		rec, err := parseRecord(text)
		if err != nil {
			// Log error but continue? For now, careful.
			// LoTW sometimes has weird data.
			continue
		}
		if len(rec) > 0 {
			records = append(records, rec)
		}
	}

	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("reading adif data: %w", err)
	}

	return records, nil
}

func parseRecord(text string) (Record, error) {
	rec := make(Record)
	// Iterate looking for <

	// A robust parser would march through the string.
	// Format: <NAME:LEN>DATA
	// OR <NAME:LEN:TYPE>DATA

	// We can use a simple loop.
	input := text
	for {
		start := strings.Index(input, "<")
		if start == -1 {
			break
		}

		// Find end of tag
		end := strings.Index(input[start:], ">")
		if end == -1 {
			// Malformed tag, stop
			break
		}
		end += start // Absolute index

		tagContent := input[start+1 : end]
		parts := strings.Split(tagContent, ":")
		fieldName := strings.ToUpper(parts[0])

		// Check if it's a valid field tag or just a random <
		// Valid tags have Name:Length
		if len(parts) < 2 {
			// Might be <EOR> or similar logic if passed here, but we split by EOR.
			// Just skip.
			input = input[end+1:]
			continue
		}

		lengthStr := parts[1]
		length, err := strconv.Atoi(lengthStr)
		if err != nil {
			// Not a length, maybe not a field tag
			input = input[end+1:]
			continue
		}

		// Data starts after >
		dataStart := end + 1
		dataEnd := dataStart + length

		if dataEnd > len(input) {
			dataEnd = len(input) // Truncate if specified length exceeds string
		}

		value := input[dataStart:dataEnd]
		rec[fieldName] = value

		input = input[dataEnd:]
	}

	return rec, nil
}

// Package devblog provides access to embedded devblog Markdown files.
package devblog

import (
	"embed"
	"fmt"
	"sort"
	"strings"
	"time"
)

//go:embed *.md
var fs embed.FS

// Entry holds the parsed metadata of a single devblog entry.
type Entry struct {
	Filename string
	Version  string
	Date     time.Time
}

// List returns all devblog entries sorted by date descending (newest first).
func List() ([]Entry, error) {
	dirEntries, err := fs.ReadDir(".")
	if err != nil {
		return nil, fmt.Errorf("read devblog dir: %w", err)
	}

	entries := make([]Entry, 0, len(dirEntries))
	for _, de := range dirEntries {
		if de.IsDir() || !strings.HasSuffix(de.Name(), ".md") {
			continue
		}
		entry, ok := parseFilename(de.Name())
		if !ok {
			continue
		}
		entries = append(entries, entry)
	}

	sort.Slice(entries, func(i, j int) bool {
		return entries[i].Date.After(entries[j].Date)
	})

	return entries, nil
}

// Content returns the raw Markdown bytes for the given filename.
func Content(filename string) ([]byte, error) {
	data, err := fs.ReadFile(filename)
	if err != nil {
		return nil, fmt.Errorf("read devblog %q: %w", filename, err)
	}
	return data, nil
}

// parseFilename parses a devblog filename of the form "v{version}_{timestamp}Z.md".
// Example: "v1.2.3_2026-03-26T14-30-00Z.md"
func parseFilename(name string) (Entry, bool) {
	// Strip .md suffix
	base := strings.TrimSuffix(name, ".md")

	// Split on first underscore: ["v1.2.3", "2026-03-26T14-30-00Z"]
	idx := strings.Index(base, "_")
	if idx < 0 {
		return Entry{}, false
	}
	version := base[:idx]
	tsRaw := base[idx+1:]

	// Timestamp uses hyphens instead of colons for HH-MM-SS to be filesystem-safe.
	// Reconstruct as a parseable RFC3339-like string: 2006-01-02T15:04:05Z
	// Format stored: 2026-03-26T14-30-00Z  (date hyphens are fine, time hyphens need fixing)
	// Strategy: split on T, fix only the time part.
	tParts := strings.SplitN(tsRaw, "T", 2)
	if len(tParts) != 2 {
		return Entry{}, false
	}
	datePart := tParts[0]
	timePart := strings.TrimSuffix(tParts[1], "Z")

	// Replace hyphens in time part with colons: "14-30-00" → "14:30:00"
	timePart = strings.ReplaceAll(timePart, "-", ":")
	ts := datePart + "T" + timePart + "Z"

	t, err := time.Parse(time.RFC3339, ts)
	if err != nil {
		return Entry{}, false
	}

	return Entry{
		Filename: name,
		Version:  version,
		Date:     t,
	}, true
}

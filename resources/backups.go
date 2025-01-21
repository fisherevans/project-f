package resources

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"
)

func cleanupBackups(fullpath string, n int) error {
	type backupFile struct {
		path      string
		timestamp time.Time
	}

	var backups []backupFile

	directory := filepath.Dir(fullpath)
	baseName := filepath.Base(fullpath)

	// Read the directory
	entries, err := os.ReadDir(directory)
	if err != nil {
		return fmt.Errorf("failed to read directory: %w", err)
	}

	// Parse and collect backup files
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}

		name := entry.Name()
		if strings.Contains(name, ".backup_") {
			// Extract the prefix and timestamp from the filename
			parts := strings.Split(name, ".backup_")
			if len(parts) != 2 {
				continue
			}

			prefix := parts[0]
			if prefix != baseName {
				continue
			}

			timestampStr := parts[1]
			timestamp, err := time.Parse("2006-01-02_15-04-05", timestampStr)
			if err != nil {
				continue // Skip files with invalid timestamps
			}

			backups = append(backups, backupFile{
				path:      filepath.Join(directory, name),
				timestamp: timestamp,
			})
		}
	}

	// Sort each group by timestamp (newest first)
	sort.Slice(backups, func(i, j int) bool {
		return backups[i].timestamp.After(backups[j].timestamp)
	})

	// Keep the N most recent backups and delete the rest
	for i, backup := range backups {
		if i >= n {
			// Delete the file
			if err := os.Remove(backup.path); err != nil {
				return fmt.Errorf("failed to delete file %s: %w", backup.path, err)
			}
			fmt.Printf("Deleted: %s (prefix: %s)\n", backup.path, baseName)
		}
	}

	return nil
}

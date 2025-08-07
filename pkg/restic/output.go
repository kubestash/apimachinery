/*
Copyright AppsCode Inc. and Contributors

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package restic

import (
	"bufio"
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"strings"
	"time"
)

const FileModeRWXAll = 0o777

type BackupOutput struct {
	// Stats shows statistics of individual hosts
	Stats []HostBackupStats `json:"stats,omitempty"`
}

type RestoreOutput struct {
	// Stats shows restore statistics of individual hosts
	Stats []HostRestoreStats `json:"stats,omitempty"`
}

type RepositoryStats struct {
	// Integrity shows result of repository integrity check after last backup
	Integrity *bool `json:"integrity,omitempty"`
	// Size show size of repository after last backup
	Size string `json:"size,omitempty"`
	// SnapshotCount shows number of snapshots stored in the repository
	SnapshotCount int64 `json:"snapshotCount,omitempty"`
	// SnapshotsRemovedOnLastCleanup shows number of old snapshots cleaned up according to retention policy on last backup session
	SnapshotsRemovedOnLastCleanup int64 `json:"snapshotsRemovedOnLastCleanup,omitempty"`
}

// ExtractBackupInfo extract information from output of "restic backup" command and
// save valuable information into backupOutput
func extractBackupInfo(output []byte, path string) ([]SnapshotStats, error) {
	// unmarshal json output
	var jsonOutputs []BackupSummary
	dec := json.NewDecoder(bytes.NewReader(output))
	var errs []error
	for {
		var summary BackupSummary
		if err := dec.Decode(&summary); err != nil {
			if errors.Is(err, io.EOF) {
				break
			}
			errs = append(errs, fmt.Errorf("error decoding JSON: %w", err))
		}
		if summary.MessageType != "summary" {
			continue
		}
		jsonOutputs = append(jsonOutputs, summary)
	}
	var snapshotStatsList []SnapshotStats
	for _, jsonOutput := range jsonOutputs {
		snapshotStats := SnapshotStats{
			Path: path,
		}
		snapshotStats.FileStats.NewFiles = jsonOutput.FilesNew
		snapshotStats.FileStats.ModifiedFiles = jsonOutput.FilesChanged
		snapshotStats.FileStats.UnmodifiedFiles = jsonOutput.FilesUnmodified
		snapshotStats.FileStats.TotalFiles = jsonOutput.TotalFilesProcessed

		snapshotStats.Uploaded = formatBytes(jsonOutput.DataAdded)
		snapshotStats.TotalSize = formatBytes(jsonOutput.TotalBytesProcessed)
		snapshotStats.ProcessingTime = formatSeconds(uint64(jsonOutput.TotalDuration))
		snapshotStats.Name = jsonOutput.SnapshotID
		snapshotStatsList = append(snapshotStatsList, snapshotStats)
	}

	return snapshotStatsList, nil
}

// ExtractCheckInfo extract information from output of "restic check" command and
// save valuable information into backupOutput
func extractCheckInfo(out []byte) bool {
	scanner := bufio.NewScanner(bytes.NewReader(out))
	var line string
	for scanner.Scan() {
		line = scanner.Text()
		line = strings.TrimSpace(line)
		if line == "no errors were found" {
			return true
		}
	}
	return false
}

// ExtractStatsInfo extract information from output of "restic stats" command and
// save valuable information into backupOutput
func extractStatsInfo(out []byte) (string, error) {
	var stat StatsContainer
	err := json.Unmarshal(out, &stat)
	if err != nil {
		return "", err
	}
	return formatBytes(stat.TotalSize), nil
}

type BackupSummary struct {
	MessageType         string  `json:"message_type"` // "summary"
	FilesNew            *int64  `json:"files_new"`
	FilesChanged        *int64  `json:"files_changed"`
	FilesUnmodified     *int64  `json:"files_unmodified"`
	DataAdded           uint64  `json:"data_added"`
	TotalFilesProcessed *int64  `json:"total_files_processed"`
	TotalBytesProcessed uint64  `json:"total_bytes_processed"`
	TotalDuration       float64 `json:"total_duration"` // in seconds
	SnapshotID          string  `json:"snapshot_id"`
}

type ForgetGroup struct {
	Keep   []json.RawMessage `json:"keep"`
	Remove []json.RawMessage `json:"remove"`
}

type StatsContainer struct {
	TotalSize uint64 `json:"total_size"`
}

type LockStats struct {
	Time      time.Time `json:"time"`
	Exclusive bool      `json:"exclusive"` // true if the lock is exclusive, false if it is non-exclusive
	Hostname  string    `json:"hostname"`  // Hostname of the machine where the lock was created, our case PodName
	Username  string    `json:"username"`
	PID       int       `json:"pid"`
	UID       int       `json:"uid"`
	GID       int       `json:"gid"`
}

func extractLockStats(raw []byte) (*LockStats, error) {
	var stats LockStats
	if err := json.Unmarshal(raw, &stats); err != nil {
		return nil, fmt.Errorf("cannot decode lock JSON: %w", err)
	}
	return &stats, nil
}

func extractLockIDs(r io.Reader) ([]string, error) {
	sc := bufio.NewScanner(r)
	var ids []string

	for sc.Scan() {
		line := strings.TrimSpace(sc.Text())
		if len(line) >= 64 {
			ids = append(ids, line[:64])
		}
	}
	return ids, sc.Err()
}

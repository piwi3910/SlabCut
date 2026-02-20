package project

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/piwi3910/SlabCut/internal/model"
)

// SharedProject is the file format for shared projects.
// It wraps a Project with additional sharing metadata to distinguish
// shared files from regular project saves.
type SharedProject struct {
	FormatVersion string        `json:"format_version"`
	SharedAt      string        `json:"shared_at"`
	SharedBy      string        `json:"shared_by"`
	Project       model.Project `json:"project"`
}

// ExportShared exports a project as a shareable file. The shared format
// includes the full project data plus sharing metadata. The author name
// and notes are embedded in the project metadata.
func ExportShared(path string, proj model.Project, author, notes string) error {
	// Update project metadata for sharing
	now := time.Now().UTC().Format(time.RFC3339)
	proj.Metadata.UpdatedAt = now
	if proj.Metadata.CreatedAt == "" {
		proj.Metadata.CreatedAt = now
	}
	proj.Metadata.Author = author
	proj.Metadata.Notes = notes
	proj.Metadata.Version = "1.0"

	shared := SharedProject{
		FormatVersion: "1.0",
		SharedAt:      now,
		SharedBy:      author,
		Project:       proj,
	}

	data, err := json.MarshalIndent(shared, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal shared project: %w", err)
	}

	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("failed to create directory: %w", err)
	}

	if err := os.WriteFile(path, data, 0644); err != nil {
		return fmt.Errorf("failed to write shared project: %w", err)
	}
	return nil
}

// ImportShared imports a shared project file. It handles both the shared
// format (SharedProject wrapper) and plain project files for backward
// compatibility. Returns the imported project with metadata populated.
func ImportShared(path string) (model.Project, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return model.Project{}, fmt.Errorf("failed to read shared file: %w", err)
	}

	// Try parsing as SharedProject first
	var shared SharedProject
	if err := json.Unmarshal(data, &shared); err == nil && shared.FormatVersion != "" {
		proj := shared.Project
		if proj.Metadata.SharedFrom == "" {
			proj.Metadata.SharedFrom = shared.SharedBy
		}
		return proj, nil
	}

	// Fall back to plain project format
	var proj model.Project
	if err := json.Unmarshal(data, &proj); err != nil {
		return model.Project{}, fmt.Errorf("failed to parse project file: %w", err)
	}

	return proj, nil
}

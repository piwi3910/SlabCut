package project

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"github.com/piwi3910/SlabCut/internal/model"
)

func TestExportShared_CreatesFile(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "shared.slabshare")

	proj := model.NewProject()
	proj.Name = "Test Project"
	proj.Parts = append(proj.Parts, model.NewPart("Shelf", 500, 300, 2))
	proj.Stocks = append(proj.Stocks, model.NewStockSheet("Board", 2440, 1220, 1))

	err := ExportShared(path, proj, "Test User", "Shared for review")
	if err != nil {
		t.Fatalf("ExportShared error: %v", err)
	}

	info, err := os.Stat(path)
	if err != nil {
		t.Fatalf("shared file was not created: %v", err)
	}
	if info.Size() == 0 {
		t.Fatal("shared file is empty")
	}
}

func TestExportAndImportShared_RoundTrip(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "shared.slabshare")

	proj := model.NewProject()
	proj.Name = "My Cabinet"
	proj.Parts = append(proj.Parts, model.NewPart("Side", 600, 400, 2))
	proj.Parts = append(proj.Parts, model.NewPart("Top", 500, 300, 1))
	proj.Stocks = append(proj.Stocks, model.NewStockSheet("Plywood", 2440, 1220, 1))

	err := ExportShared(path, proj, "Pascal", "For team review")
	if err != nil {
		t.Fatalf("ExportShared error: %v", err)
	}

	imported, err := ImportShared(path)
	if err != nil {
		t.Fatalf("ImportShared error: %v", err)
	}

	if imported.Name != "My Cabinet" {
		t.Errorf("expected name 'My Cabinet', got %q", imported.Name)
	}
	if len(imported.Parts) != 2 {
		t.Errorf("expected 2 parts, got %d", len(imported.Parts))
	}
	if len(imported.Stocks) != 1 {
		t.Errorf("expected 1 stock, got %d", len(imported.Stocks))
	}
	if imported.Metadata.Author != "Pascal" {
		t.Errorf("expected author 'Pascal', got %q", imported.Metadata.Author)
	}
	if imported.Metadata.Notes != "For team review" {
		t.Errorf("expected notes 'For team review', got %q", imported.Metadata.Notes)
	}
	if imported.Metadata.CreatedAt == "" {
		t.Error("expected non-empty CreatedAt")
	}
	if imported.Metadata.SharedFrom != "Pascal" {
		t.Errorf("expected SharedFrom 'Pascal', got %q", imported.Metadata.SharedFrom)
	}
}

func TestImportShared_PlainProject(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "plain.cnccalc")

	// Create a plain project file (backward compatibility)
	proj := model.NewProject()
	proj.Name = "Plain Project"
	proj.Parts = append(proj.Parts, model.NewPart("Part A", 200, 100, 1))

	data, err := json.MarshalIndent(proj, "", "  ")
	if err != nil {
		t.Fatalf("failed to marshal: %v", err)
	}
	if err := os.WriteFile(path, data, 0644); err != nil {
		t.Fatalf("failed to write: %v", err)
	}

	imported, err := ImportShared(path)
	if err != nil {
		t.Fatalf("ImportShared error: %v", err)
	}

	if imported.Name != "Plain Project" {
		t.Errorf("expected 'Plain Project', got %q", imported.Name)
	}
	if len(imported.Parts) != 1 {
		t.Errorf("expected 1 part, got %d", len(imported.Parts))
	}
}

func TestImportShared_InvalidFile(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "invalid.json")
	if err := os.WriteFile(path, []byte("not json"), 0644); err != nil {
		t.Fatalf("failed to write: %v", err)
	}

	_, err := ImportShared(path)
	if err == nil {
		t.Fatal("expected error for invalid file, got nil")
	}
}

func TestImportShared_FileNotFound(t *testing.T) {
	_, err := ImportShared("/nonexistent/path/file.json")
	if err == nil {
		t.Fatal("expected error for missing file, got nil")
	}
}

func TestExportShared_MetadataPopulated(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "meta.slabshare")

	proj := model.NewProject()
	proj.Name = "Meta Test"

	err := ExportShared(path, proj, "Team Lead", "Review needed")
	if err != nil {
		t.Fatalf("ExportShared error: %v", err)
	}

	// Read and verify the file structure
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("failed to read: %v", err)
	}

	var shared SharedProject
	if err := json.Unmarshal(data, &shared); err != nil {
		t.Fatalf("failed to unmarshal: %v", err)
	}

	if shared.FormatVersion != "1.0" {
		t.Errorf("expected format version '1.0', got %q", shared.FormatVersion)
	}
	if shared.SharedBy != "Team Lead" {
		t.Errorf("expected shared by 'Team Lead', got %q", shared.SharedBy)
	}
	if shared.SharedAt == "" {
		t.Error("expected non-empty SharedAt")
	}
	if shared.Project.Metadata.Version != "1.0" {
		t.Errorf("expected version '1.0', got %q", shared.Project.Metadata.Version)
	}
}

package project

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/piwi3910/SlabCut/internal/model"
)

func TestSaveAndLoadLibrary(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "test_library.json")

	lib := model.NewPartsLibrary()
	part := model.NewLibraryPart("Shelf", 600, 300, model.GrainNone)
	part.Category = "Kitchen"
	part.Material = "Plywood"
	part.Thickness = 18
	part.Notes = "Standard shelf"
	part.Tags = []string{"shelf", "kitchen"}
	lib.AddPart(part)

	// Save
	if err := SaveLibrary(path, lib); err != nil {
		t.Fatalf("SaveLibrary: %v", err)
	}

	// Load
	loaded, err := LoadLibrary(path)
	if err != nil {
		t.Fatalf("LoadLibrary: %v", err)
	}

	if len(loaded.Parts) != 1 {
		t.Fatalf("expected 1 part, got %d", len(loaded.Parts))
	}

	lp := loaded.Parts[0]
	if lp.Label != "Shelf" {
		t.Errorf("expected label 'Shelf', got %q", lp.Label)
	}
	if lp.Width != 600 || lp.Height != 300 {
		t.Errorf("expected 600x300, got %.0fx%.0f", lp.Width, lp.Height)
	}
	if lp.Category != "Kitchen" {
		t.Errorf("expected category 'Kitchen', got %q", lp.Category)
	}
	if lp.Material != "Plywood" {
		t.Errorf("expected material 'Plywood', got %q", lp.Material)
	}
	if lp.Thickness != 18 {
		t.Errorf("expected thickness 18, got %.0f", lp.Thickness)
	}
	if len(lp.Tags) != 2 || lp.Tags[0] != "shelf" || lp.Tags[1] != "kitchen" {
		t.Errorf("unexpected tags: %v", lp.Tags)
	}

	// Verify categories were preserved
	found := false
	for _, c := range loaded.Categories {
		if c == "Kitchen" {
			found = true
			break
		}
	}
	if !found {
		t.Errorf("expected 'Kitchen' in categories: %v", loaded.Categories)
	}
}

func TestLoadLibraryNonExistentFile(t *testing.T) {
	path := filepath.Join(t.TempDir(), "nonexistent.json")

	lib, err := LoadLibrary(path)
	if err != nil {
		t.Fatalf("expected no error for nonexistent file, got: %v", err)
	}

	if len(lib.Parts) != 0 {
		t.Errorf("expected empty parts, got %d", len(lib.Parts))
	}
	if len(lib.Categories) == 0 {
		t.Error("expected at least one default category")
	}
}

func TestLoadLibraryInvalidJSON(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "bad.json")
	if err := os.WriteFile(path, []byte("not json"), 0644); err != nil {
		t.Fatal(err)
	}

	_, err := LoadLibrary(path)
	if err == nil {
		t.Fatal("expected error for invalid JSON")
	}
}

func TestDefaultLibraryPath(t *testing.T) {
	path, err := DefaultLibraryPath()
	if err != nil {
		t.Fatalf("DefaultLibraryPath: %v", err)
	}
	if filepath.Base(path) != "parts_library.json" {
		t.Errorf("expected parts_library.json, got %s", filepath.Base(path))
	}
	dir := filepath.Dir(path)
	if filepath.Base(dir) != ".slabcut" {
		t.Errorf("expected .slabcut dir, got %s", filepath.Base(dir))
	}
}

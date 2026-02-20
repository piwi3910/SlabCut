package model

import (
	"testing"
)

func TestNewLibraryPart(t *testing.T) {
	lp := NewLibraryPart("Test Part", 100, 50, GrainHorizontal)
	if lp.Label != "Test Part" {
		t.Errorf("expected label 'Test Part', got %q", lp.Label)
	}
	if lp.Width != 100 || lp.Height != 50 {
		t.Errorf("expected 100x50, got %.0fx%.0f", lp.Width, lp.Height)
	}
	if lp.Grain != GrainHorizontal {
		t.Errorf("expected GrainHorizontal, got %v", lp.Grain)
	}
	if lp.ID == "" {
		t.Error("expected non-empty ID")
	}
	if lp.Tags == nil {
		t.Error("expected non-nil tags slice")
	}
}

func TestLibraryPartToPart(t *testing.T) {
	lp := NewLibraryPart("Shelf", 600, 300, GrainVertical)
	part := lp.ToPart(5)

	if part.Label != "Shelf" {
		t.Errorf("expected label 'Shelf', got %q", part.Label)
	}
	if part.Width != 600 || part.Height != 300 {
		t.Errorf("expected 600x300, got %.0fx%.0f", part.Width, part.Height)
	}
	if part.Quantity != 5 {
		t.Errorf("expected quantity 5, got %d", part.Quantity)
	}
	if part.Grain != GrainVertical {
		t.Errorf("expected GrainVertical, got %v", part.Grain)
	}
	// Part should have its own ID, different from library part
	if part.ID == lp.ID {
		t.Error("expected different ID for converted part")
	}
}

func TestPartsLibraryAddAndRemove(t *testing.T) {
	lib := NewPartsLibrary()

	p1 := NewLibraryPart("Part A", 100, 50, GrainNone)
	p1.Category = "Cat1"
	lib.AddPart(p1)

	p2 := NewLibraryPart("Part B", 200, 100, GrainNone)
	p2.Category = "Cat2"
	lib.AddPart(p2)

	if len(lib.Parts) != 2 {
		t.Fatalf("expected 2 parts, got %d", len(lib.Parts))
	}

	// Categories should include General, Cat1, Cat2
	if len(lib.Categories) != 3 {
		t.Errorf("expected 3 categories, got %d: %v", len(lib.Categories), lib.Categories)
	}

	// Remove first part
	lib.RemovePart(p1.ID)
	if len(lib.Parts) != 1 {
		t.Fatalf("expected 1 part after remove, got %d", len(lib.Parts))
	}
	if lib.Parts[0].ID != p2.ID {
		t.Error("wrong part remaining after remove")
	}
}

func TestPartsLibraryUpdate(t *testing.T) {
	lib := NewPartsLibrary()
	p := NewLibraryPart("Original", 100, 50, GrainNone)
	lib.AddPart(p)

	updated := p
	updated.Label = "Updated"
	updated.Width = 200
	lib.UpdatePart(updated)

	found := lib.FindByID(p.ID)
	if found == nil {
		t.Fatal("expected to find part by ID")
	}
	if found.Label != "Updated" {
		t.Errorf("expected label 'Updated', got %q", found.Label)
	}
	if found.Width != 200 {
		t.Errorf("expected width 200, got %.0f", found.Width)
	}
}

func TestPartsLibrarySearch(t *testing.T) {
	lib := NewPartsLibrary()

	p1 := NewLibraryPart("Kitchen Shelf", 600, 300, GrainNone)
	p1.Tags = []string{"kitchen", "shelf"}
	lib.AddPart(p1)

	p2 := NewLibraryPart("Bathroom Door", 800, 2000, GrainVertical)
	p2.Tags = []string{"bathroom", "door"}
	lib.AddPart(p2)

	p3 := NewLibraryPart("Kitchen Door", 500, 700, GrainNone)
	p3.Tags = []string{"kitchen", "door"}
	lib.AddPart(p3)

	// Search by label
	results := lib.Search("kitchen")
	if len(results) != 2 {
		t.Errorf("expected 2 results for 'kitchen', got %d", len(results))
	}

	// Search by tag
	results = lib.Search("door")
	if len(results) != 2 {
		t.Errorf("expected 2 results for 'door', got %d", len(results))
	}

	// Search all
	results = lib.Search("")
	if len(results) != 3 {
		t.Errorf("expected 3 results for empty query, got %d", len(results))
	}

	// Case insensitive
	results = lib.Search("KITCHEN")
	if len(results) != 2 {
		t.Errorf("expected 2 results for 'KITCHEN', got %d", len(results))
	}
}

func TestPartsLibraryFilterByCategory(t *testing.T) {
	lib := NewPartsLibrary()

	p1 := NewLibraryPart("Part A", 100, 50, GrainNone)
	p1.Category = "Kitchen"
	lib.AddPart(p1)

	p2 := NewLibraryPart("Part B", 200, 100, GrainNone)
	p2.Category = "Bathroom"
	lib.AddPart(p2)

	p3 := NewLibraryPart("Part C", 300, 150, GrainNone)
	p3.Category = "Kitchen"
	lib.AddPart(p3)

	results := lib.FilterByCategory("Kitchen")
	if len(results) != 2 {
		t.Errorf("expected 2 Kitchen parts, got %d", len(results))
	}

	results = lib.FilterByCategory("All")
	if len(results) != 3 {
		t.Errorf("expected 3 parts for 'All', got %d", len(results))
	}

	results = lib.FilterByCategory("")
	if len(results) != 3 {
		t.Errorf("expected 3 parts for empty category, got %d", len(results))
	}
}

func TestPartsLibrarySearchAndFilter(t *testing.T) {
	lib := NewPartsLibrary()

	p1 := NewLibraryPart("Kitchen Shelf", 600, 300, GrainNone)
	p1.Category = "Kitchen"
	lib.AddPart(p1)

	p2 := NewLibraryPart("Bathroom Shelf", 500, 250, GrainNone)
	p2.Category = "Bathroom"
	lib.AddPart(p2)

	p3 := NewLibraryPart("Kitchen Door", 700, 2000, GrainNone)
	p3.Category = "Kitchen"
	lib.AddPart(p3)

	results := lib.SearchAndFilter("shelf", "Kitchen")
	if len(results) != 1 {
		t.Errorf("expected 1 result, got %d", len(results))
	}
	if len(results) > 0 && results[0].Label != "Kitchen Shelf" {
		t.Errorf("expected 'Kitchen Shelf', got %q", results[0].Label)
	}
}

func TestPartsLibraryDefaultCategory(t *testing.T) {
	lib := NewPartsLibrary()
	p := NewLibraryPart("No Category", 100, 50, GrainNone)
	// Don't set category - should default to General
	lib.AddPart(p)

	if lib.Parts[0].Category != "General" {
		t.Errorf("expected default category 'General', got %q", lib.Parts[0].Category)
	}
}

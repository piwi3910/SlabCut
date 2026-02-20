package model

import "github.com/google/uuid"

// LibraryPart represents a part stored in the personal parts library.
// It extends the basic Part with additional metadata for organization.
type LibraryPart struct {
	ID        string   `json:"id"`
	Label     string   `json:"label"`
	Width     float64  `json:"width"`  // mm
	Height    float64  `json:"height"` // mm
	Grain     Grain    `json:"grain"`
	Category  string   `json:"category"`
	Material  string   `json:"material"`
	Thickness float64  `json:"thickness"` // mm
	Notes     string   `json:"notes"`
	Tags      []string `json:"tags"`
}

// NewLibraryPart creates a new LibraryPart with a generated ID.
func NewLibraryPart(label string, w, h float64, grain Grain) LibraryPart {
	return LibraryPart{
		ID:     uuid.New().String()[:8],
		Label:  label,
		Width:  w,
		Height: h,
		Grain:  grain,
		Tags:   []string{},
	}
}

// ToPart converts a LibraryPart to a project Part with the given quantity.
func (lp LibraryPart) ToPart(quantity int) Part {
	return Part{
		ID:       uuid.New().String()[:8],
		Label:    lp.Label,
		Width:    lp.Width,
		Height:   lp.Height,
		Quantity: quantity,
		Grain:    lp.Grain,
	}
}

// PartsLibrary holds the user's personal parts library.
type PartsLibrary struct {
	Parts      []LibraryPart `json:"parts"`
	Categories []string      `json:"categories"`
}

// NewPartsLibrary creates an empty parts library with a default category.
func NewPartsLibrary() PartsLibrary {
	return PartsLibrary{
		Parts:      []LibraryPart{},
		Categories: []string{"General"},
	}
}

// AddPart adds a part to the library. If the part's category is new, it is
// added to the categories list.
func (lib *PartsLibrary) AddPart(part LibraryPart) {
	if part.Category == "" {
		part.Category = "General"
	}
	lib.Parts = append(lib.Parts, part)
	lib.ensureCategory(part.Category)
}

// RemovePart removes a part from the library by ID.
func (lib *PartsLibrary) RemovePart(id string) {
	for i, p := range lib.Parts {
		if p.ID == id {
			lib.Parts = append(lib.Parts[:i], lib.Parts[i+1:]...)
			return
		}
	}
}

// UpdatePart replaces a library part by matching on ID.
func (lib *PartsLibrary) UpdatePart(updated LibraryPart) {
	for i, p := range lib.Parts {
		if p.ID == updated.ID {
			lib.Parts[i] = updated
			lib.ensureCategory(updated.Category)
			return
		}
	}
}

// FindByID returns a pointer to the library part with the given ID, or nil.
func (lib *PartsLibrary) FindByID(id string) *LibraryPart {
	for i := range lib.Parts {
		if lib.Parts[i].ID == id {
			return &lib.Parts[i]
		}
	}
	return nil
}

// Search returns parts whose label or tags contain the query string (case-insensitive).
func (lib *PartsLibrary) Search(query string) []LibraryPart {
	if query == "" {
		return lib.Parts
	}
	var results []LibraryPart
	q := toLower(query)
	for _, p := range lib.Parts {
		if containsLower(p.Label, q) || containsLower(p.Notes, q) || tagsContain(p.Tags, q) {
			results = append(results, p)
		}
	}
	return results
}

// FilterByCategory returns parts in the given category. An empty category returns all.
func (lib *PartsLibrary) FilterByCategory(category string) []LibraryPart {
	if category == "" || category == "All" {
		return lib.Parts
	}
	var results []LibraryPart
	for _, p := range lib.Parts {
		if p.Category == category {
			results = append(results, p)
		}
	}
	return results
}

// SearchAndFilter combines search and category filter.
func (lib *PartsLibrary) SearchAndFilter(query, category string) []LibraryPart {
	parts := lib.FilterByCategory(category)
	if query == "" {
		return parts
	}
	q := toLower(query)
	var results []LibraryPart
	for _, p := range parts {
		if containsLower(p.Label, q) || containsLower(p.Notes, q) || tagsContain(p.Tags, q) {
			results = append(results, p)
		}
	}
	return results
}

// AddCategory adds a new category if it does not already exist.
func (lib *PartsLibrary) AddCategory(cat string) {
	lib.ensureCategory(cat)
}

func (lib *PartsLibrary) ensureCategory(cat string) {
	for _, c := range lib.Categories {
		if c == cat {
			return
		}
	}
	lib.Categories = append(lib.Categories, cat)
}

// toLower is a simple lowercase helper to avoid importing strings in the model.
func toLower(s string) string {
	b := make([]byte, len(s))
	for i := range s {
		c := s[i]
		if c >= 'A' && c <= 'Z' {
			c += 'a' - 'A'
		}
		b[i] = c
	}
	return string(b)
}

// containsLower checks if haystack contains needle (both already lowercased needle).
func containsLower(haystack, needle string) bool {
	h := toLower(haystack)
	if len(needle) > len(h) {
		return false
	}
	for i := 0; i <= len(h)-len(needle); i++ {
		if h[i:i+len(needle)] == needle {
			return true
		}
	}
	return false
}

func tagsContain(tags []string, query string) bool {
	for _, t := range tags {
		if containsLower(t, query) {
			return true
		}
	}
	return false
}

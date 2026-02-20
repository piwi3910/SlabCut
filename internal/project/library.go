package project

import (
	"encoding/json"
	"os"
	"path/filepath"

	"github.com/piwi3910/SlabCut/internal/model"
)

// DefaultLibraryPath returns the default path for the parts library file.
// It is stored under ~/.slabcut/parts_library.json.
func DefaultLibraryPath() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	dir := filepath.Join(home, ".slabcut")
	if err := os.MkdirAll(dir, 0755); err != nil {
		return "", err
	}
	return filepath.Join(dir, "parts_library.json"), nil
}

// SaveLibrary saves a parts library to a JSON file.
func SaveLibrary(path string, lib model.PartsLibrary) error {
	data, err := json.MarshalIndent(lib, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(path, data, 0644)
}

// LoadLibrary loads a parts library from a JSON file.
// If the file does not exist, it returns a new empty library.
func LoadLibrary(path string) (model.PartsLibrary, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return model.NewPartsLibrary(), nil
		}
		return model.PartsLibrary{}, err
	}
	var lib model.PartsLibrary
	if err := json.Unmarshal(data, &lib); err != nil {
		return model.PartsLibrary{}, err
	}
	// Ensure at least the default category exists
	if len(lib.Categories) == 0 {
		lib.Categories = []string{"General"}
	}
	if lib.Parts == nil {
		lib.Parts = []model.LibraryPart{}
	}
	return lib, nil
}

// LoadDefaultLibrary loads the library from the default path.
func LoadDefaultLibrary() (model.PartsLibrary, error) {
	path, err := DefaultLibraryPath()
	if err != nil {
		return model.NewPartsLibrary(), err
	}
	return LoadLibrary(path)
}

// SaveDefaultLibrary saves the library to the default path.
func SaveDefaultLibrary(lib model.PartsLibrary) error {
	path, err := DefaultLibraryPath()
	if err != nil {
		return err
	}
	return SaveLibrary(path, lib)
}

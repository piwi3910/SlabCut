package project

import (
	"encoding/json"
	"os"

	"github.com/piwi3910/cnc-calculator/internal/model"
)

// Save saves a project to a JSON file.
func Save(path string, proj model.Project) error {
	data, err := json.MarshalIndent(proj, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(path, data, 0644)
}

// Load loads a project from a JSON file.
func Load(path string) (model.Project, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return model.Project{}, err
	}
	var proj model.Project
	err = json.Unmarshal(data, &proj)
	return proj, err
}

// ExportGCode saves GCode to a file.
func ExportGCode(path, code string) error {
	return os.WriteFile(path, []byte(code), 0644)
}

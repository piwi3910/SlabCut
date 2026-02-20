package importer

import (
	"encoding/csv"
	"fmt"
	"os"
	"strconv"

	"github.com/piwi3910/cnc-calculator/internal/model"
	"github.com/xuri/excelize/v2"
)

// ImportResult holds the results of an import operation.
type ImportResult struct {
	Parts    []model.Part
	Errors   []string
	Warnings []string
}

// ImportCSV imports parts from a CSV file.
// Expected format: Label, Width, Height, Quantity, Grain (optional)
// Returns a slice of Part objects and any errors encountered.
func ImportCSV(path string) ImportResult {
	result := ImportResult{}

	file, err := os.Open(path)
	if err != nil {
		result.Errors = append(result.Errors, fmt.Sprintf("Cannot open file: %v", err))
		return result
	}
	defer file.Close()

	reader := csv.NewReader(file)
	records, err := reader.ReadAll()
	if err != nil {
		result.Errors = append(result.Errors, fmt.Sprintf("Cannot read CSV: %v", err))
		return result
	}

	if len(records) == 0 {
		result.Errors = append(result.Errors, "File is empty")
		return result
	}

	// Skip header row if it exists
	startRow := 0
	if len(records) > 0 {
		firstRow := records[0]
		if len(firstRow) >= 3 {
			// Check if first row looks like a header (non-numeric values)
			if _, err := strconv.ParseFloat(firstRow[1], 64); err != nil {
				startRow = 1 // Skip header
				result.Warnings = append(result.Warnings, "Detected header row, skipping")
			}
		}
	}

	lineNum := startRow + 1
	for i := startRow; i < len(records); i++ {
		row := records[i]
		lineNum++

		// Skip empty rows
		if len(row) == 0 || (len(row) == 1 && row[0] == "") {
			continue
		}

		if len(row) < 4 {
			result.Errors = append(result.Errors, fmt.Sprintf("Line %d: Not enough columns (need at least: Label, Width, Height, Quantity)", lineNum))
			continue
		}

		label := row[0]
		if label == "" {
			label = fmt.Sprintf("Part %d", len(result.Parts)+1)
		}

		width, err := strconv.ParseFloat(row[1], 64)
		if err != nil {
			result.Errors = append(result.Errors, fmt.Sprintf("Line %d: Invalid width '%s'", lineNum, row[1]))
			continue
		}

		height, err := strconv.ParseFloat(row[2], 64)
		if err != nil {
			result.Errors = append(result.Errors, fmt.Sprintf("Line %d: Invalid height '%s'", lineNum, row[2]))
			continue
		}

		qty, err := strconv.Atoi(row[3])
		if err != nil {
			result.Errors = append(result.Errors, fmt.Sprintf("Line %d: Invalid quantity '%s'", lineNum, row[3]))
			continue
		}

		if width <= 0 || height <= 0 || qty <= 0 {
			result.Errors = append(result.Errors, fmt.Sprintf("Line %d: Width, height, and quantity must be positive", lineNum))
			continue
		}

		part := model.NewPart(label, width, height, qty)

		// Optional grain direction
		if len(row) >= 5 {
			grainStr := row[4]
			switch grainStr {
			case "Horizontal", "H", "h":
				part.Grain = model.GrainHorizontal
			case "Vertical", "V", "v":
				part.Grain = model.GrainVertical
			case "", "None", "N", "n":
				part.Grain = model.GrainNone
			default:
				result.Warnings = append(result.Warnings, fmt.Sprintf("Line %d: Unknown grain direction '%s', defaulting to None", lineNum, grainStr))
			}
		}

		result.Parts = append(result.Parts, part)
	}

	return result
}

// ImportExcel imports parts from an Excel (.xlsx, .xls) file.
// Reads the first sheet, expected format: Label, Width, Height, Quantity, Grain (optional)
func ImportExcel(path string) ImportResult {
	result := ImportResult{}

	f, err := excelize.OpenFile(path)
	if err != nil {
		result.Errors = append(result.Errors, fmt.Sprintf("Cannot open Excel file: %v", err))
		return result
	}
	defer f.Close()

	// Get the first sheet
	sheets := f.GetSheetList()
	if len(sheets) == 0 {
		result.Errors = append(result.Errors, "Excel file has no sheets")
		return result
	}

	// Read all rows from first sheet
	rows, err := f.GetRows(sheets[0])
	if err != nil {
		result.Errors = append(result.Errors, fmt.Sprintf("Cannot read Excel data: %v", err))
		return result
	}

	if len(rows) == 0 {
		result.Errors = append(result.Errors, "Sheet is empty")
		return result
	}

	// Detect and skip header
	startRow := 0
	if len(rows) > 1 && len(rows[0]) >= 3 {
		// Check if first row looks like a header
		if _, err := strconv.ParseFloat(rows[0][1], 64); err != nil {
			startRow = 1
			result.Warnings = append(result.Warnings, "Detected header row, skipping")
		}
	}

	lineNum := startRow + 1
	for i := startRow; i < len(rows); i++ {
		row := rows[i]
		lineNum++

		// Skip empty rows
		if len(row) == 0 || (len(row) == 1 && row[0] == "") {
			continue
		}

		if len(row) < 4 {
			result.Errors = append(result.Errors, fmt.Sprintf("Row %d: Not enough columns (need at least: Label, Width, Height, Quantity)", lineNum))
			continue
		}

		label := row[0]
		if label == "" {
			label = fmt.Sprintf("Part %d", len(result.Parts)+1)
		}

		width, err := strconv.ParseFloat(row[1], 64)
		if err != nil {
			result.Errors = append(result.Errors, fmt.Sprintf("Row %d: Invalid width '%s'", lineNum, row[1]))
			continue
		}

		height, err := strconv.ParseFloat(row[2], 64)
		if err != nil {
			result.Errors = append(result.Errors, fmt.Sprintf("Row %d: Invalid height '%s'", lineNum, row[2]))
			continue
		}

		qty, err := strconv.Atoi(row[3])
		if err != nil {
			result.Errors = append(result.Errors, fmt.Sprintf("Row %d: Invalid quantity '%s'", lineNum, row[3]))
			continue
		}

		if width <= 0 || height <= 0 || qty <= 0 {
			result.Errors = append(result.Errors, fmt.Sprintf("Row %d: Width, height, and quantity must be positive", lineNum))
			continue
		}

		part := model.NewPart(label, width, height, qty)

		// Optional grain direction
		if len(row) >= 5 {
			grainStr := row[4]
			switch grainStr {
			case "Horizontal", "H", "h":
				part.Grain = model.GrainHorizontal
			case "Vertical", "V", "v":
				part.Grain = model.GrainVertical
			case "", "None", "N", "n":
				part.Grain = model.GrainNone
			default:
				result.Warnings = append(result.Warnings, fmt.Sprintf("Row %d: Unknown grain direction '%s', defaulting to None", lineNum, grainStr))
			}
		}

		result.Parts = append(result.Parts, part)
	}

	return result
}

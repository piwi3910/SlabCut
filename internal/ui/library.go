package ui

import (
	"encoding/csv"
	"fmt"
	"os"
	"strconv"
	"strings"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"

	"github.com/piwi3910/SlabCut/internal/model"
	"github.com/piwi3910/SlabCut/internal/project"
)

// showLibraryManager opens a dialog for managing the parts library.
func (a *App) showLibraryManager() {
	// Create a new window for library management
	w := fyne.CurrentApp().NewWindow("Parts Library")
	w.Resize(fyne.NewSize(900, 600))

	listContainer := container.NewVBox()

	searchEntry := widget.NewEntry()
	searchEntry.SetPlaceHolder("Search by label, notes, or tags...")

	categoryFilter := widget.NewSelect(a.libraryCategoryOptions(), nil)
	categoryFilter.SetSelected("All")

	var refreshList func()
	refreshList = func() {
		listContainer.RemoveAll()

		parts := a.library.SearchAndFilter(searchEntry.Text, categoryFilter.Selected)

		if len(parts) == 0 {
			listContainer.Add(widget.NewLabel("No parts in library. Click 'Add Part' to create one."))
			return
		}

		// Header
		header := container.NewGridWithColumns(9,
			widget.NewLabelWithStyle("Label", fyne.TextAlignLeading, fyne.TextStyle{Bold: true}),
			widget.NewLabelWithStyle("W (mm)", fyne.TextAlignLeading, fyne.TextStyle{Bold: true}),
			widget.NewLabelWithStyle("H (mm)", fyne.TextAlignLeading, fyne.TextStyle{Bold: true}),
			widget.NewLabelWithStyle("Grain", fyne.TextAlignLeading, fyne.TextStyle{Bold: true}),
			widget.NewLabelWithStyle("Category", fyne.TextAlignLeading, fyne.TextStyle{Bold: true}),
			widget.NewLabelWithStyle("Material", fyne.TextAlignLeading, fyne.TextStyle{Bold: true}),
			widget.NewLabelWithStyle("Thick.", fyne.TextAlignLeading, fyne.TextStyle{Bold: true}),
			widget.NewLabelWithStyle("", fyne.TextAlignLeading, fyne.TextStyle{}),
			widget.NewLabelWithStyle("", fyne.TextAlignLeading, fyne.TextStyle{}),
		)
		listContainer.Add(header)
		listContainer.Add(widget.NewSeparator())

		for _, p := range parts {
			partID := p.ID
			thickStr := ""
			if p.Thickness > 0 {
				thickStr = fmt.Sprintf("%.1f", p.Thickness)
			}
			row := container.NewGridWithColumns(9,
				widget.NewLabel(p.Label),
				widget.NewLabel(fmt.Sprintf("%.1f", p.Width)),
				widget.NewLabel(fmt.Sprintf("%.1f", p.Height)),
				widget.NewLabel(p.Grain.String()),
				widget.NewLabel(p.Category),
				widget.NewLabel(p.Material),
				widget.NewLabel(thickStr),
				widget.NewButtonWithIcon("", theme.DocumentCreateIcon(), func() {
					a.showEditLibraryPartDialog(partID, func() {
						refreshList()
						categoryFilter.Options = a.libraryCategoryOptions()
					})
				}),
				widget.NewButtonWithIcon("", theme.DeleteIcon(), func() {
					dialog.ShowConfirm("Delete Part",
						fmt.Sprintf("Remove '%s' from library?", p.Label),
						func(ok bool) {
							if ok {
								a.library.RemovePart(partID)
								a.saveLibrary()
								refreshList()
							}
						}, w)
				}),
			)
			listContainer.Add(row)
		}
	}

	searchEntry.OnChanged = func(_ string) { refreshList() }
	categoryFilter.OnChanged = func(_ string) { refreshList() }

	addBtn := widget.NewButtonWithIcon("Add Part", theme.ContentAddIcon(), func() {
		a.showAddLibraryPartDialog(func() {
			refreshList()
			categoryFilter.Options = a.libraryCategoryOptions()
		})
	})

	importCSVBtn := widget.NewButtonWithIcon("Import CSV", theme.FolderOpenIcon(), func() {
		a.importCSVToLibrary(w, func() {
			refreshList()
			categoryFilter.Options = a.libraryCategoryOptions()
		})
	})

	toolbar := container.NewHBox(
		addBtn,
		importCSVBtn,
		layout.NewSpacer(),
	)

	filterRow := container.NewGridWithColumns(3,
		searchEntry,
		container.NewHBox(widget.NewLabel("Category:"), categoryFilter),
		layout.NewSpacer(),
	)

	content := container.NewBorder(
		container.NewVBox(toolbar, filterRow),
		nil, nil, nil,
		container.NewVScroll(listContainer),
	)

	refreshList()

	w.SetContent(content)
	w.Show()
}

// showAddFromLibraryDialog opens a picker to add library parts to the current project.
func (a *App) showAddFromLibraryDialog() {
	if len(a.library.Parts) == 0 {
		dialog.ShowInformation("Empty Library",
			"Your parts library is empty. Use the Parts Library manager to add parts first.",
			a.window)
		return
	}

	w := fyne.CurrentApp().NewWindow("Add from Library")
	w.Resize(fyne.NewSize(800, 500))

	type selection struct {
		partID   string
		quantity int
		check    *widget.Check
		qtyEntry *widget.Entry
	}

	var selections []selection
	listContainer := container.NewVBox()

	searchEntry := widget.NewEntry()
	searchEntry.SetPlaceHolder("Search parts...")

	categoryFilter := widget.NewSelect(a.libraryCategoryOptions(), nil)
	categoryFilter.SetSelected("All")

	var refreshPicker func()
	refreshPicker = func() {
		listContainer.RemoveAll()
		selections = nil

		parts := a.library.SearchAndFilter(searchEntry.Text, categoryFilter.Selected)

		if len(parts) == 0 {
			listContainer.Add(widget.NewLabel("No matching parts found."))
			return
		}

		// Header
		header := container.NewGridWithColumns(7,
			widget.NewLabelWithStyle("Select", fyne.TextAlignLeading, fyne.TextStyle{Bold: true}),
			widget.NewLabelWithStyle("Label", fyne.TextAlignLeading, fyne.TextStyle{Bold: true}),
			widget.NewLabelWithStyle("W (mm)", fyne.TextAlignLeading, fyne.TextStyle{Bold: true}),
			widget.NewLabelWithStyle("H (mm)", fyne.TextAlignLeading, fyne.TextStyle{Bold: true}),
			widget.NewLabelWithStyle("Grain", fyne.TextAlignLeading, fyne.TextStyle{Bold: true}),
			widget.NewLabelWithStyle("Category", fyne.TextAlignLeading, fyne.TextStyle{Bold: true}),
			widget.NewLabelWithStyle("Qty", fyne.TextAlignLeading, fyne.TextStyle{Bold: true}),
		)
		listContainer.Add(header)
		listContainer.Add(widget.NewSeparator())

		for _, p := range parts {
			check := widget.NewCheck("", nil)
			qtyEntry := widget.NewEntry()
			qtyEntry.SetText("1")
			qtyEntry.Resize(fyne.NewSize(60, 36))

			sel := selection{
				partID:   p.ID,
				quantity: 1,
				check:    check,
				qtyEntry: qtyEntry,
			}
			selections = append(selections, sel)

			row := container.NewGridWithColumns(7,
				check,
				widget.NewLabel(p.Label),
				widget.NewLabel(fmt.Sprintf("%.1f", p.Width)),
				widget.NewLabel(fmt.Sprintf("%.1f", p.Height)),
				widget.NewLabel(p.Grain.String()),
				widget.NewLabel(p.Category),
				qtyEntry,
			)
			listContainer.Add(row)
		}
	}

	searchEntry.OnChanged = func(_ string) { refreshPicker() }
	categoryFilter.OnChanged = func(_ string) { refreshPicker() }

	addBtn := widget.NewButtonWithIcon("Add to Project", theme.ConfirmIcon(), func() {
		added := 0
		for _, sel := range selections {
			if !sel.check.Checked {
				continue
			}
			qty, _ := strconv.Atoi(sel.qtyEntry.Text)
			if qty <= 0 {
				qty = 1
			}
			lp := a.library.FindByID(sel.partID)
			if lp != nil {
				part := lp.ToPart(qty)
				a.project.Parts = append(a.project.Parts, part)
				added++
			}
		}
		if added > 0 {
			a.refreshPartsList()
			dialog.ShowInformation("Parts Added",
				fmt.Sprintf("Added %d part(s) to the project.", added),
				a.window)
			w.Close()
		} else {
			dialog.ShowInformation("No Selection",
				"Please check at least one part to add.",
				w)
		}
	})

	cancelBtn := widget.NewButton("Cancel", func() {
		w.Close()
	})

	filterRow := container.NewGridWithColumns(3,
		searchEntry,
		container.NewHBox(widget.NewLabel("Category:"), categoryFilter),
		layout.NewSpacer(),
	)

	content := container.NewBorder(
		filterRow,
		container.NewHBox(layout.NewSpacer(), cancelBtn, addBtn),
		nil, nil,
		container.NewVScroll(listContainer),
	)

	refreshPicker()

	w.SetContent(content)
	w.Show()
}

// showAddLibraryPartDialog opens a form to add a new part to the library.
func (a *App) showAddLibraryPartDialog(onSaved func()) {
	labelEntry := widget.NewEntry()
	labelEntry.SetPlaceHolder("Part name")

	widthEntry := widget.NewEntry()
	widthEntry.SetPlaceHolder("Width in mm")

	heightEntry := widget.NewEntry()
	heightEntry.SetPlaceHolder("Height in mm")

	grainSelect := widget.NewSelect([]string{"None", "Horizontal", "Vertical"}, nil)
	grainSelect.SetSelected("None")

	categoryEntry := widget.NewSelectEntry(a.library.Categories)
	categoryEntry.SetPlaceHolder("Category")
	if len(a.library.Categories) > 0 {
		categoryEntry.SetText(a.library.Categories[0])
	}

	materialEntry := widget.NewEntry()
	materialEntry.SetPlaceHolder("e.g. Plywood, MDF")

	thicknessEntry := widget.NewEntry()
	thicknessEntry.SetPlaceHolder("Thickness in mm (optional)")

	notesEntry := widget.NewMultiLineEntry()
	notesEntry.SetPlaceHolder("Notes (optional)")

	tagsEntry := widget.NewEntry()
	tagsEntry.SetPlaceHolder("Tags, comma-separated")

	form := dialog.NewForm("Add Library Part", "Add", "Cancel",
		[]*widget.FormItem{
			widget.NewFormItem("Label", labelEntry),
			widget.NewFormItem("Width (mm)", widthEntry),
			widget.NewFormItem("Height (mm)", heightEntry),
			widget.NewFormItem("Grain", grainSelect),
			widget.NewFormItem("Category", categoryEntry),
			widget.NewFormItem("Material", materialEntry),
			widget.NewFormItem("Thickness (mm)", thicknessEntry),
			widget.NewFormItem("Notes", notesEntry),
			widget.NewFormItem("Tags", tagsEntry),
		},
		func(ok bool) {
			if !ok {
				return
			}
			w, _ := strconv.ParseFloat(widthEntry.Text, 64)
			h, _ := strconv.ParseFloat(heightEntry.Text, 64)
			if w <= 0 || h <= 0 {
				dialog.ShowError(fmt.Errorf("width and height must be > 0"), a.window)
				return
			}

			grain := model.GrainNone
			switch grainSelect.Selected {
			case "Horizontal":
				grain = model.GrainHorizontal
			case "Vertical":
				grain = model.GrainVertical
			}

			part := model.NewLibraryPart(labelEntry.Text, w, h, grain)
			part.Category = categoryEntry.Text
			part.Material = materialEntry.Text
			part.Notes = notesEntry.Text

			if t, err := strconv.ParseFloat(thicknessEntry.Text, 64); err == nil {
				part.Thickness = t
			}

			if tagsEntry.Text != "" {
				for _, tag := range strings.Split(tagsEntry.Text, ",") {
					tag = strings.TrimSpace(tag)
					if tag != "" {
						part.Tags = append(part.Tags, tag)
					}
				}
			}

			a.library.AddPart(part)
			a.saveLibrary()
			if onSaved != nil {
				onSaved()
			}
		},
		a.window,
	)
	form.Resize(fyne.NewSize(500, 550))
	form.Show()
}

// showEditLibraryPartDialog opens a form to edit an existing library part.
func (a *App) showEditLibraryPartDialog(partID string, onSaved func()) {
	lp := a.library.FindByID(partID)
	if lp == nil {
		return
	}

	labelEntry := widget.NewEntry()
	labelEntry.SetText(lp.Label)

	widthEntry := widget.NewEntry()
	widthEntry.SetText(fmt.Sprintf("%.1f", lp.Width))

	heightEntry := widget.NewEntry()
	heightEntry.SetText(fmt.Sprintf("%.1f", lp.Height))

	grainSelect := widget.NewSelect([]string{"None", "Horizontal", "Vertical"}, nil)
	grainSelect.SetSelected(lp.Grain.String())

	categoryEntry := widget.NewSelectEntry(a.library.Categories)
	categoryEntry.SetText(lp.Category)

	materialEntry := widget.NewEntry()
	materialEntry.SetText(lp.Material)

	thicknessEntry := widget.NewEntry()
	if lp.Thickness > 0 {
		thicknessEntry.SetText(fmt.Sprintf("%.1f", lp.Thickness))
	}

	notesEntry := widget.NewMultiLineEntry()
	notesEntry.SetText(lp.Notes)

	tagsEntry := widget.NewEntry()
	tagsEntry.SetText(strings.Join(lp.Tags, ", "))

	form := dialog.NewForm("Edit Library Part", "Save", "Cancel",
		[]*widget.FormItem{
			widget.NewFormItem("Label", labelEntry),
			widget.NewFormItem("Width (mm)", widthEntry),
			widget.NewFormItem("Height (mm)", heightEntry),
			widget.NewFormItem("Grain", grainSelect),
			widget.NewFormItem("Category", categoryEntry),
			widget.NewFormItem("Material", materialEntry),
			widget.NewFormItem("Thickness (mm)", thicknessEntry),
			widget.NewFormItem("Notes", notesEntry),
			widget.NewFormItem("Tags", tagsEntry),
		},
		func(ok bool) {
			if !ok {
				return
			}
			w, _ := strconv.ParseFloat(widthEntry.Text, 64)
			h, _ := strconv.ParseFloat(heightEntry.Text, 64)
			if w <= 0 || h <= 0 {
				dialog.ShowError(fmt.Errorf("width and height must be > 0"), a.window)
				return
			}

			grain := model.GrainNone
			switch grainSelect.Selected {
			case "Horizontal":
				grain = model.GrainHorizontal
			case "Vertical":
				grain = model.GrainVertical
			}

			updated := *lp
			updated.Label = labelEntry.Text
			updated.Width = w
			updated.Height = h
			updated.Grain = grain
			updated.Category = categoryEntry.Text
			updated.Material = materialEntry.Text
			updated.Notes = notesEntry.Text

			if t, err := strconv.ParseFloat(thicknessEntry.Text, 64); err == nil {
				updated.Thickness = t
			} else {
				updated.Thickness = 0
			}

			updated.Tags = []string{}
			if tagsEntry.Text != "" {
				for _, tag := range strings.Split(tagsEntry.Text, ",") {
					tag = strings.TrimSpace(tag)
					if tag != "" {
						updated.Tags = append(updated.Tags, tag)
					}
				}
			}

			a.library.UpdatePart(updated)
			a.saveLibrary()
			if onSaved != nil {
				onSaved()
			}
		},
		a.window,
	)
	form.Resize(fyne.NewSize(500, 550))
	form.Show()
}

// showSaveToLibraryDialog saves an existing project part to the library.
func (a *App) showSaveToLibraryDialog(part model.Part) {
	categoryEntry := widget.NewSelectEntry(a.library.Categories)
	if len(a.library.Categories) > 0 {
		categoryEntry.SetText(a.library.Categories[0])
	}

	materialEntry := widget.NewEntry()
	materialEntry.SetPlaceHolder("e.g. Plywood, MDF")

	thicknessEntry := widget.NewEntry()
	thicknessEntry.SetPlaceHolder("Thickness in mm (optional)")

	notesEntry := widget.NewMultiLineEntry()
	notesEntry.SetPlaceHolder("Notes (optional)")

	tagsEntry := widget.NewEntry()
	tagsEntry.SetPlaceHolder("Tags, comma-separated")

	form := dialog.NewForm(
		fmt.Sprintf("Save '%s' to Library", part.Label),
		"Save", "Cancel",
		[]*widget.FormItem{
			widget.NewFormItem("Category", categoryEntry),
			widget.NewFormItem("Material", materialEntry),
			widget.NewFormItem("Thickness (mm)", thicknessEntry),
			widget.NewFormItem("Notes", notesEntry),
			widget.NewFormItem("Tags", tagsEntry),
		},
		func(ok bool) {
			if !ok {
				return
			}
			lp := model.NewLibraryPart(part.Label, part.Width, part.Height, part.Grain)
			lp.Category = categoryEntry.Text
			lp.Material = materialEntry.Text
			lp.Notes = notesEntry.Text

			if t, err := strconv.ParseFloat(thicknessEntry.Text, 64); err == nil {
				lp.Thickness = t
			}

			if tagsEntry.Text != "" {
				for _, tag := range strings.Split(tagsEntry.Text, ",") {
					tag = strings.TrimSpace(tag)
					if tag != "" {
						lp.Tags = append(lp.Tags, tag)
					}
				}
			}

			a.library.AddPart(lp)
			a.saveLibrary()
			dialog.ShowInformation("Saved",
				fmt.Sprintf("'%s' has been saved to your parts library.", part.Label),
				a.window)
		},
		a.window,
	)
	form.Resize(fyne.NewSize(450, 400))
	form.Show()
}

// importCSVToLibrary imports parts from a CSV file into the library.
func (a *App) importCSVToLibrary(parent fyne.Window, onDone func()) {
	dialog.ShowFileOpen(func(reader fyne.URIReadCloser, err error) {
		if err != nil || reader == nil {
			return
		}
		defer reader.Close()

		result := importCSVToLibraryParts(reader.URI().Path())

		if len(result.errors) > 0 {
			errorMsg := "Errors during import:\n\n" + strings.Join(result.errors, "\n")
			dialog.ShowError(fmt.Errorf("%s", errorMsg), parent)
		}

		if len(result.parts) > 0 {
			for _, p := range result.parts {
				a.library.AddPart(p)
			}
			a.saveLibrary()
			if onDone != nil {
				onDone()
			}

			msg := fmt.Sprintf("Successfully imported %d parts to library.", len(result.parts))
			if len(result.errors) > 0 {
				msg += fmt.Sprintf("\n%d rows had errors and were skipped.", len(result.errors))
			}
			dialog.ShowInformation("Import Complete", msg, parent)
		}
	}, parent)
}

type csvLibraryResult struct {
	parts  []model.LibraryPart
	errors []string
}

// importCSVToLibraryParts reads a CSV file and returns library parts.
// Expected columns: label, width, height, grain, category, material, thickness, notes, tags
// At minimum: label, width, height
func importCSVToLibraryParts(path string) csvLibraryResult {
	result := csvLibraryResult{}

	data, err := readCSVFile(path)
	if err != nil {
		result.errors = append(result.errors, fmt.Sprintf("Failed to read file: %v", err))
		return result
	}

	if len(data) < 2 {
		result.errors = append(result.errors, "CSV file must have a header row and at least one data row")
		return result
	}

	// Build column index from header
	header := data[0]
	colIndex := make(map[string]int)
	for i, col := range header {
		colIndex[strings.TrimSpace(strings.ToLower(col))] = i
	}

	// Require at minimum label, width, height
	labelIdx, hasLabel := colIndex["label"]
	widthIdx, hasWidth := colIndex["width"]
	heightIdx, hasHeight := colIndex["height"]
	if !hasLabel || !hasWidth || !hasHeight {
		result.errors = append(result.errors, "CSV must have 'label', 'width', and 'height' columns")
		return result
	}

	for rowNum, row := range data[1:] {
		lineNum := rowNum + 2 // 1-indexed, skip header

		if len(row) <= labelIdx || len(row) <= widthIdx || len(row) <= heightIdx {
			result.errors = append(result.errors, fmt.Sprintf("Row %d: not enough columns", lineNum))
			continue
		}

		label := strings.TrimSpace(row[labelIdx])
		if label == "" {
			result.errors = append(result.errors, fmt.Sprintf("Row %d: empty label", lineNum))
			continue
		}

		w, err := strconv.ParseFloat(strings.TrimSpace(row[widthIdx]), 64)
		if err != nil || w <= 0 {
			result.errors = append(result.errors, fmt.Sprintf("Row %d: invalid width", lineNum))
			continue
		}

		h, err := strconv.ParseFloat(strings.TrimSpace(row[heightIdx]), 64)
		if err != nil || h <= 0 {
			result.errors = append(result.errors, fmt.Sprintf("Row %d: invalid height", lineNum))
			continue
		}

		grain := model.GrainNone
		if idx, ok := colIndex["grain"]; ok && idx < len(row) {
			switch strings.TrimSpace(strings.ToLower(row[idx])) {
			case "horizontal", "h":
				grain = model.GrainHorizontal
			case "vertical", "v":
				grain = model.GrainVertical
			}
		}

		lp := model.NewLibraryPart(label, w, h, grain)

		if idx, ok := colIndex["category"]; ok && idx < len(row) {
			lp.Category = strings.TrimSpace(row[idx])
		}
		if idx, ok := colIndex["material"]; ok && idx < len(row) {
			lp.Material = strings.TrimSpace(row[idx])
		}
		if idx, ok := colIndex["thickness"]; ok && idx < len(row) {
			if t, err := strconv.ParseFloat(strings.TrimSpace(row[idx]), 64); err == nil {
				lp.Thickness = t
			}
		}
		if idx, ok := colIndex["notes"]; ok && idx < len(row) {
			lp.Notes = strings.TrimSpace(row[idx])
		}
		if idx, ok := colIndex["tags"]; ok && idx < len(row) {
			tags := strings.TrimSpace(row[idx])
			if tags != "" {
				for _, tag := range strings.Split(tags, ";") {
					tag = strings.TrimSpace(tag)
					if tag != "" {
						lp.Tags = append(lp.Tags, tag)
					}
				}
			}
		}

		result.parts = append(result.parts, lp)
	}

	return result
}

// readCSVFile reads a CSV file and returns rows as string slices.
func readCSVFile(path string) ([][]string, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	reader := csv.NewReader(file)
	return reader.ReadAll()
}

// libraryCategoryOptions returns "All" plus all library categories for dropdowns.
func (a *App) libraryCategoryOptions() []string {
	opts := []string{"All"}
	opts = append(opts, a.library.Categories...)
	return opts
}

// saveLibrary persists the library to the default path.
func (a *App) saveLibrary() {
	if err := project.SaveDefaultLibrary(a.library); err != nil {
		fmt.Printf("Error saving parts library: %v\n", err)
	}
}

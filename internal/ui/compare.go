package ui

import (
	"fmt"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/widget"

	"github.com/piwi3910/SlabCut/internal/engine"
)

// showCompareDialog runs optimization with multiple parameter scenarios and
// displays the results in a comparison table so the user can pick the best one.
func (a *App) showCompareDialog() {
	if len(a.project.Parts) == 0 {
		dialog.ShowInformation("Nothing to compare", "Add at least one part first.", a.window)
		return
	}
	if len(a.project.Stocks) == 0 {
		dialog.ShowInformation("No stock sheets", "Add at least one stock sheet first.", a.window)
		return
	}

	scenarios := engine.BuildDefaultScenarios(a.project.Settings)

	// Add a custom kerf scenario based on user input
	customKerfEntry := widget.NewEntry()
	customKerfEntry.SetPlaceHolder("e.g. 2.0")
	customKerfEntry.SetText(fmt.Sprintf("%.1f", a.project.Settings.KerfWidth*0.75))

	// Build custom scenarios form
	addCustomKerf := widget.NewCheck("Add custom kerf scenario", nil)

	customTrimEntry := widget.NewEntry()
	customTrimEntry.SetPlaceHolder("e.g. 5.0")
	customTrimEntry.SetText(fmt.Sprintf("%.1f", a.project.Settings.EdgeTrim*0.5))
	addCustomTrim := widget.NewCheck("Add custom edge trim scenario", nil)

	configForm := dialog.NewForm("Configure Comparison", "Run Comparison", "Cancel",
		[]*widget.FormItem{
			widget.NewFormItem("", widget.NewLabel(
				fmt.Sprintf("This will run %d+ optimization scenarios and compare results.", len(scenarios)),
			)),
			widget.NewFormItem("Custom Kerf (mm)", container.NewBorder(nil, nil, addCustomKerf, nil, customKerfEntry)),
			widget.NewFormItem("Custom Trim (mm)", container.NewBorder(nil, nil, addCustomTrim, nil, customTrimEntry)),
		},
		func(ok bool) {
			if !ok {
				return
			}

			// Add custom scenarios if checked
			if addCustomKerf.Checked {
				var kerfVal float64
				if _, err := fmt.Sscanf(customKerfEntry.Text, "%f", &kerfVal); err == nil && kerfVal >= 0 {
					s := a.project.Settings
					s.KerfWidth = kerfVal
					scenarios = append(scenarios, engine.ComparisonScenario{
						Name:     fmt.Sprintf("Kerf %.1fmm (custom)", kerfVal),
						Settings: s,
					})
				}
			}
			if addCustomTrim.Checked {
				var trimVal float64
				if _, err := fmt.Sscanf(customTrimEntry.Text, "%f", &trimVal); err == nil && trimVal >= 0 {
					s := a.project.Settings
					s.EdgeTrim = trimVal
					scenarios = append(scenarios, engine.ComparisonScenario{
						Name:     fmt.Sprintf("Trim %.1fmm (custom)", trimVal),
						Settings: s,
					})
				}
			}

			a.runComparison(scenarios)
		},
		a.window,
	)
	configForm.Resize(fyne.NewSize(500, 300))
	configForm.Show()
}

// runComparison executes the comparison and shows results.
func (a *App) runComparison(scenarios []engine.ComparisonScenario) {
	results := engine.CompareScenarios(scenarios, a.project.Parts, a.project.Stocks)

	if len(results) == 0 {
		dialog.ShowInformation("No Results", "No comparison results were generated.", a.window)
		return
	}

	a.showComparisonResults(results)
}

// showComparisonResults displays comparison results in a table with an "Apply" button
// for the user to keep a specific result.
func (a *App) showComparisonResults(results []engine.ComparisonResult) {
	// Build table header
	cols := 6
	header := container.NewGridWithColumns(cols,
		widget.NewLabelWithStyle("Scenario", fyne.TextAlignLeading, fyne.TextStyle{Bold: true}),
		widget.NewLabelWithStyle("Sheets", fyne.TextAlignLeading, fyne.TextStyle{Bold: true}),
		widget.NewLabelWithStyle("Cuts", fyne.TextAlignLeading, fyne.TextStyle{Bold: true}),
		widget.NewLabelWithStyle("Waste %", fyne.TextAlignLeading, fyne.TextStyle{Bold: true}),
		widget.NewLabelWithStyle("Unplaced", fyne.TextAlignLeading, fyne.TextStyle{Bold: true}),
		widget.NewLabelWithStyle("Action", fyne.TextAlignLeading, fyne.TextStyle{Bold: true}),
	)

	rows := container.NewVBox(header, widget.NewSeparator())

	// Find the best result (lowest waste with all parts placed)
	bestIdx := 0
	bestScore := -1.0
	for i, r := range results {
		score := 100.0 - r.WastePercent
		if r.UnplacedCount == 0 && score > bestScore {
			bestScore = score
			bestIdx = i
		}
	}

	for i, r := range results {
		idx := i
		result := r

		nameLabel := widget.NewLabel(r.Scenario.Name)
		if idx == bestIdx {
			nameLabel = widget.NewLabelWithStyle(r.Scenario.Name+" *", fyne.TextAlignLeading, fyne.TextStyle{Bold: true})
		}

		applyBtn := widget.NewButton("Apply", func() {
			a.saveState("Apply Comparison Result")
			a.project.Settings = result.Scenario.Settings
			a.project.Result = &result.Result
			a.refreshResults()
			dialog.ShowInformation("Applied",
				fmt.Sprintf("Applied settings from scenario %q.\nEfficiency: %.1f%%",
					result.Scenario.Name, 100.0-result.WastePercent),
				a.window)
		})

		wasteLabel := fmt.Sprintf("%.1f%%", r.WastePercent)
		unplacedLabel := fmt.Sprintf("%d", r.UnplacedCount)
		if r.UnplacedCount > 0 {
			unplacedLabel += " !"
		}

		row := container.NewGridWithColumns(cols,
			nameLabel,
			widget.NewLabel(fmt.Sprintf("%d", r.SheetsUsed)),
			widget.NewLabel(fmt.Sprintf("%d", r.TotalCuts)),
			widget.NewLabel(wasteLabel),
			widget.NewLabel(unplacedLabel),
			applyBtn,
		)
		rows.Add(row)
	}

	// Add legend
	rows.Add(widget.NewSeparator())
	rows.Add(widget.NewLabelWithStyle("* = Best result (lowest waste with all parts placed)", fyne.TextAlignLeading, fyne.TextStyle{Italic: true}))

	scrollable := container.NewVScroll(rows)
	scrollable.SetMinSize(fyne.NewSize(650, 300))

	d := dialog.NewCustom("Optimization Comparison", "Close", scrollable, a.window)
	d.Resize(fyne.NewSize(700, 400))
	d.Show()
}

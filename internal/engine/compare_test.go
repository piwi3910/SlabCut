package engine

import (
	"testing"

	"github.com/piwi3910/SlabCut/internal/model"
)

func TestCompareScenarios_Basic(t *testing.T) {
	parts := []model.Part{
		{ID: "p1", Label: "A", Width: 400, Height: 300, Quantity: 2, Grain: model.GrainNone},
		{ID: "p2", Label: "B", Width: 200, Height: 150, Quantity: 3, Grain: model.GrainNone},
	}
	stocks := []model.StockSheet{
		{ID: "s1", Label: "Board", Width: 2440, Height: 1220, Quantity: 2},
	}

	base := model.DefaultSettings()
	scenarios := []ComparisonScenario{
		{Name: "Guillotine", Settings: base},
		{Name: "Genetic", Settings: func() model.CutSettings {
			s := base
			s.Algorithm = model.AlgorithmGenetic
			return s
		}()},
	}

	results := CompareScenarios(scenarios, parts, stocks)

	if len(results) != 2 {
		t.Fatalf("expected 2 results, got %d", len(results))
	}

	for i, r := range results {
		if r.Scenario.Name != scenarios[i].Name {
			t.Errorf("result %d: name mismatch: got %q, want %q", i, r.Scenario.Name, scenarios[i].Name)
		}
		if r.SheetsUsed == 0 {
			t.Errorf("result %d: expected at least one sheet used", i)
		}
		if r.TotalCuts == 0 {
			t.Errorf("result %d: expected at least one cut", i)
		}
		if r.WastePercent < 0 || r.WastePercent > 100 {
			t.Errorf("result %d: waste percent out of range: %.1f", i, r.WastePercent)
		}
	}
}

func TestCompareScenarios_Empty(t *testing.T) {
	results := CompareScenarios(nil, nil, nil)
	if len(results) != 0 {
		t.Errorf("expected 0 results for nil scenarios, got %d", len(results))
	}
}

func TestBuildDefaultScenarios_Guillotine(t *testing.T) {
	base := model.DefaultSettings()
	base.Algorithm = model.AlgorithmGuillotine

	scenarios := BuildDefaultScenarios(base)

	if len(scenarios) < 2 {
		t.Fatalf("expected at least 2 scenarios, got %d", len(scenarios))
	}

	if scenarios[0].Name != "Current Settings" {
		t.Errorf("first scenario should be 'Current Settings', got %q", scenarios[0].Name)
	}

	// Should include Genetic Algorithm as alternative
	found := false
	for _, s := range scenarios {
		if s.Name == "Genetic Algorithm" {
			found = true
			if s.Settings.Algorithm != model.AlgorithmGenetic {
				t.Error("Genetic Algorithm scenario should use genetic algorithm")
			}
		}
	}
	if !found {
		t.Error("expected a 'Genetic Algorithm' scenario")
	}
}

func TestBuildDefaultScenarios_Genetic(t *testing.T) {
	base := model.DefaultSettings()
	base.Algorithm = model.AlgorithmGenetic

	scenarios := BuildDefaultScenarios(base)

	found := false
	for _, s := range scenarios {
		if s.Name == "Guillotine Algorithm" {
			found = true
		}
	}
	if !found {
		t.Error("expected a 'Guillotine Algorithm' scenario when base is genetic")
	}
}

func TestCompareScenarios_UnplacedParts(t *testing.T) {
	// Parts that are too large for the stock
	parts := []model.Part{
		{ID: "p1", Label: "Huge", Width: 5000, Height: 5000, Quantity: 1, Grain: model.GrainNone},
	}
	stocks := []model.StockSheet{
		{ID: "s1", Label: "Small", Width: 100, Height: 100, Quantity: 1},
	}

	scenarios := []ComparisonScenario{
		{Name: "Test", Settings: model.DefaultSettings()},
	}

	results := CompareScenarios(scenarios, parts, stocks)
	if len(results) != 1 {
		t.Fatalf("expected 1 result, got %d", len(results))
	}
	if results[0].UnplacedCount != 1 {
		t.Errorf("expected 1 unplaced part, got %d", results[0].UnplacedCount)
	}
}

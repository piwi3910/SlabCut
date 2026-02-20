package model

import (
	"math"
	"testing"
)

func TestEdgeBandingHasAny(t *testing.T) {
	none := EdgeBanding{}
	if none.HasAny() {
		t.Error("expected HasAny() false for no edges")
	}

	top := EdgeBanding{Top: true}
	if !top.HasAny() {
		t.Error("expected HasAny() true for top edge")
	}
}

func TestEdgeBandingEdgeCount(t *testing.T) {
	tests := []struct {
		eb   EdgeBanding
		want int
	}{
		{EdgeBanding{}, 0},
		{EdgeBanding{Top: true}, 1},
		{EdgeBanding{Top: true, Bottom: true}, 2},
		{EdgeBanding{Top: true, Bottom: true, Left: true, Right: true}, 4},
	}
	for _, tt := range tests {
		if got := tt.eb.EdgeCount(); got != tt.want {
			t.Errorf("EdgeCount() = %d, want %d for %+v", got, tt.want, tt.eb)
		}
	}
}

func TestEdgeBandingLinearLength(t *testing.T) {
	eb := EdgeBanding{Top: true, Bottom: true, Left: true, Right: true}
	// Width=800, Height=400: top(800) + bottom(800) + left(400) + right(400) = 2400
	length := eb.LinearLength(800, 400)
	if length != 2400 {
		t.Errorf("expected 2400, got %.0f", length)
	}

	// Only top and left
	eb2 := EdgeBanding{Top: true, Left: true}
	length2 := eb2.LinearLength(600, 300)
	if length2 != 900 {
		t.Errorf("expected 900, got %.0f", length2)
	}
}

func TestEdgeBandingString(t *testing.T) {
	tests := []struct {
		eb   EdgeBanding
		want string
	}{
		{EdgeBanding{}, "None"},
		{EdgeBanding{Top: true}, "T"},
		{EdgeBanding{Top: true, Bottom: true}, "T+B"},
		{EdgeBanding{Top: true, Bottom: true, Left: true, Right: true}, "T+B+L+R"},
		{EdgeBanding{Left: true, Right: true}, "L+R"},
	}
	for _, tt := range tests {
		if got := tt.eb.String(); got != tt.want {
			t.Errorf("String() = %q, want %q for %+v", got, tt.want, tt.eb)
		}
	}
}

func TestCalculateEdgeBanding(t *testing.T) {
	parts := []Part{
		{
			Label: "Shelf", Width: 800, Height: 300, Quantity: 4,
			EdgeBanding: EdgeBanding{Top: true, Bottom: true},
		},
		{
			Label: "Side", Width: 600, Height: 400, Quantity: 2,
			EdgeBanding: EdgeBanding{Top: true, Left: true, Right: true},
		},
		{
			Label: "Back", Width: 500, Height: 300, Quantity: 1,
			// No edge banding
		},
	}

	summary := CalculateEdgeBanding(parts, 10.0)

	// Shelf: (800+800) * 4 = 6400
	// Side: (600+400+400) * 2 = 2800
	// Total: 9200
	expectedMM := 9200.0
	if math.Abs(summary.TotalLinearMM-expectedMM) > 0.1 {
		t.Errorf("expected %.0f mm, got %.0f mm", expectedMM, summary.TotalLinearMM)
	}

	if summary.PartCount != 6 { // 4 shelves + 2 sides
		t.Errorf("expected 6 parts, got %d", summary.PartCount)
	}

	if summary.EdgeCount != 14 { // 4*2 + 2*3
		t.Errorf("expected 14 edges, got %d", summary.EdgeCount)
	}

	// With 10% waste: 9200 * 1.1 = 10120
	expectedWithWaste := math.Ceil(9200.0 * 1.1)
	if summary.TotalWithWasteMM != expectedWithWaste {
		t.Errorf("expected %.0f mm with waste, got %.0f mm", expectedWithWaste, summary.TotalWithWasteMM)
	}
}

func TestCalculateEdgeBandingNoParts(t *testing.T) {
	summary := CalculateEdgeBanding(nil, 10.0)
	if summary.TotalLinearMM != 0 {
		t.Errorf("expected 0 mm for no parts, got %.0f", summary.TotalLinearMM)
	}
}

func TestCalculateEdgeBandingNoEdges(t *testing.T) {
	parts := []Part{
		{Label: "P1", Width: 100, Height: 100, Quantity: 5},
	}
	summary := CalculateEdgeBanding(parts, 15.0)
	if summary.TotalLinearMM != 0 {
		t.Errorf("expected 0 mm for parts without banding, got %.0f", summary.TotalLinearMM)
	}
}

func TestCalculatePerPartEdgeBanding(t *testing.T) {
	parts := []Part{
		{
			Label: "Shelf", Width: 800, Height: 300, Quantity: 4,
			EdgeBanding: EdgeBanding{Top: true},
		},
		{
			Label: "No banding", Width: 500, Height: 500, Quantity: 1,
		},
	}

	breakdown := CalculatePerPartEdgeBanding(parts)
	if len(breakdown) != 1 {
		t.Fatalf("expected 1 entry, got %d", len(breakdown))
	}
	if breakdown[0].Label != "Shelf" {
		t.Errorf("expected Shelf, got %s", breakdown[0].Label)
	}
	if breakdown[0].LengthPerUnit != 800 {
		t.Errorf("expected 800 mm/unit, got %.0f", breakdown[0].LengthPerUnit)
	}
	if breakdown[0].TotalLength != 3200 {
		t.Errorf("expected 3200 mm total, got %.0f", breakdown[0].TotalLength)
	}
}

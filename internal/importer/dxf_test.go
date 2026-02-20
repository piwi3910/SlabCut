package importer

import (
	"math"
	"os"
	"path/filepath"
	"testing"

	"github.com/piwi3910/SlabCut/internal/model"
	"github.com/yofu/dxf"
)

// createTestDXFWithRect creates a DXF file containing a single rectangular
// LWPOLYLINE and returns its path.
func createTestDXFWithRect(t *testing.T, w, h float64) string {
	t.Helper()
	dir := t.TempDir()
	path := filepath.Join(dir, "rect.dxf")

	d := dxf.NewDrawing()
	_, err := d.LwPolyline(true,
		[]float64{0, 0, 0},
		[]float64{w, 0, 0},
		[]float64{w, h, 0},
		[]float64{0, h, 0},
	)
	if err != nil {
		t.Fatalf("failed to create LWPOLYLINE: %v", err)
	}
	if err := d.SaveAs(path); err != nil {
		t.Fatalf("failed to create test DXF: %v", err)
	}
	return path
}

// createTestDXFWithCircle creates a DXF file containing a single circle.
func createTestDXFWithCircle(t *testing.T, cx, cy, r float64) string {
	t.Helper()
	dir := t.TempDir()
	path := filepath.Join(dir, "circle.dxf")

	d := dxf.NewDrawing()
	d.Circle(cx, cy, 0, r)
	if err := d.SaveAs(path); err != nil {
		t.Fatalf("failed to create test DXF: %v", err)
	}
	return path
}

// createTestDXFWithLines creates a DXF file with a triangle made from LINE entities.
func createTestDXFWithLines(t *testing.T) string {
	t.Helper()
	dir := t.TempDir()
	path := filepath.Join(dir, "triangle.dxf")

	d := dxf.NewDrawing()
	d.Line(0, 0, 0, 100, 0, 0)
	d.Line(100, 0, 0, 50, 86.6, 0)
	d.Line(50, 86.6, 0, 0, 0, 0)
	if err := d.SaveAs(path); err != nil {
		t.Fatalf("failed to create test DXF: %v", err)
	}
	return path
}

func TestImportDXF_Rectangle(t *testing.T) {
	path := createTestDXFWithRect(t, 200, 100)
	result := ImportDXF(path)

	if len(result.Errors) > 0 {
		t.Fatalf("unexpected errors: %v", result.Errors)
	}
	if len(result.Parts) != 1 {
		t.Fatalf("expected 1 part, got %d", len(result.Parts))
	}

	part := result.Parts[0]

	// Bounding box should match the rectangle dimensions
	if math.Abs(part.Width-200) > 0.1 {
		t.Errorf("expected width ~200, got %.2f", part.Width)
	}
	if math.Abs(part.Height-100) > 0.1 {
		t.Errorf("expected height ~100, got %.2f", part.Height)
	}
	if part.Outline == nil {
		t.Error("expected outline to be set")
	}
	if len(part.Outline) < 4 {
		t.Errorf("expected at least 4 outline points, got %d", len(part.Outline))
	}
	if part.Quantity != 1 {
		t.Errorf("expected quantity 1, got %d", part.Quantity)
	}
}

func TestImportDXF_Circle(t *testing.T) {
	path := createTestDXFWithCircle(t, 50, 50, 25)
	result := ImportDXF(path)

	if len(result.Errors) > 0 {
		t.Fatalf("unexpected errors: %v", result.Errors)
	}
	if len(result.Parts) != 1 {
		t.Fatalf("expected 1 part, got %d", len(result.Parts))
	}

	part := result.Parts[0]

	// Bounding box for a circle with radius 25 should be ~50x50
	if math.Abs(part.Width-50) > 0.5 {
		t.Errorf("expected width ~50, got %.2f", part.Width)
	}
	if math.Abs(part.Height-50) > 0.5 {
		t.Errorf("expected height ~50, got %.2f", part.Height)
	}
	if part.Outline == nil {
		t.Error("expected outline to be set")
	}
	// Circle is approximated with 64 segments
	if len(part.Outline) != 64 {
		t.Errorf("expected 64 outline points for circle, got %d", len(part.Outline))
	}
}

func TestImportDXF_TriangleFromLines(t *testing.T) {
	path := createTestDXFWithLines(t)
	result := ImportDXF(path)

	if len(result.Errors) > 0 {
		t.Fatalf("unexpected errors: %v", result.Errors)
	}
	if len(result.Parts) != 1 {
		t.Fatalf("expected 1 part, got %d", len(result.Parts))
	}

	part := result.Parts[0]
	if math.Abs(part.Width-100) > 0.5 {
		t.Errorf("expected width ~100, got %.2f", part.Width)
	}
	if math.Abs(part.Height-86.6) > 0.5 {
		t.Errorf("expected height ~86.6, got %.2f", part.Height)
	}
}

func TestImportDXF_FileNotFound(t *testing.T) {
	result := ImportDXF("/nonexistent/file.dxf")
	if len(result.Errors) == 0 {
		t.Error("expected error for missing file")
	}
	if len(result.Parts) != 0 {
		t.Error("expected no parts for missing file")
	}
}

func TestImportDXF_EmptyFile(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "empty.dxf")
	if err := os.WriteFile(path, []byte{}, 0644); err != nil {
		t.Fatalf("failed to create empty file: %v", err)
	}

	result := ImportDXF(path)
	if len(result.Errors) == 0 {
		t.Error("expected error for empty DXF file")
	}
}

func TestOutlineBoundingBox(t *testing.T) {
	outline := model.Outline{
		{X: 10, Y: 20},
		{X: 50, Y: 5},
		{X: 30, Y: 80},
	}

	min, max := outline.BoundingBox()
	if min.X != 10 || min.Y != 5 {
		t.Errorf("expected min (10, 5), got (%.1f, %.1f)", min.X, min.Y)
	}
	if max.X != 50 || max.Y != 80 {
		t.Errorf("expected max (50, 80), got (%.1f, %.1f)", max.X, max.Y)
	}
}

func TestOutlineTranslate(t *testing.T) {
	outline := model.Outline{
		{X: 10, Y: 20},
		{X: 50, Y: 5},
	}

	moved := outline.Translate(-10, -5)
	if moved[0].X != 0 || moved[0].Y != 15 {
		t.Errorf("expected (0, 15), got (%.1f, %.1f)", moved[0].X, moved[0].Y)
	}
	if moved[1].X != 40 || moved[1].Y != 0 {
		t.Errorf("expected (40, 0), got (%.1f, %.1f)", moved[1].X, moved[1].Y)
	}
}

func TestNormalizeOutline(t *testing.T) {
	outline := model.Outline{
		{X: 100, Y: 200},
		{X: 150, Y: 250},
		{X: 120, Y: 230},
	}

	normalized := normalizeOutline(outline)
	min, _ := normalized.BoundingBox()
	if min.X != 0 || min.Y != 0 {
		t.Errorf("expected normalized min at (0, 0), got (%.1f, %.1f)", min.X, min.Y)
	}
}

func TestChainSegments_ClosedTriangle(t *testing.T) {
	segs := []segment{
		{start: model.Point2D{X: 0, Y: 0}, end: model.Point2D{X: 100, Y: 0}},
		{start: model.Point2D{X: 100, Y: 0}, end: model.Point2D{X: 50, Y: 87}},
		{start: model.Point2D{X: 50, Y: 87}, end: model.Point2D{X: 0, Y: 0}},
	}

	outlines := chainSegments(segs, 0.01)
	if len(outlines) != 1 {
		t.Fatalf("expected 1 outline, got %d", len(outlines))
	}
	if len(outlines[0]) != 3 {
		t.Errorf("expected 3 points in outline, got %d", len(outlines[0]))
	}
}

func TestChainSegments_DisconnectedSegments(t *testing.T) {
	// Two separate triangles
	segs := []segment{
		// Triangle 1
		{start: model.Point2D{X: 0, Y: 0}, end: model.Point2D{X: 10, Y: 0}},
		{start: model.Point2D{X: 10, Y: 0}, end: model.Point2D{X: 5, Y: 10}},
		{start: model.Point2D{X: 5, Y: 10}, end: model.Point2D{X: 0, Y: 0}},
		// Triangle 2
		{start: model.Point2D{X: 100, Y: 100}, end: model.Point2D{X: 200, Y: 100}},
		{start: model.Point2D{X: 200, Y: 100}, end: model.Point2D{X: 150, Y: 200}},
		{start: model.Point2D{X: 150, Y: 200}, end: model.Point2D{X: 100, Y: 100}},
	}

	outlines := chainSegments(segs, 0.01)
	if len(outlines) != 2 {
		t.Fatalf("expected 2 outlines, got %d", len(outlines))
	}
}

func TestOutlineArea(t *testing.T) {
	// 10x10 square
	square := model.Outline{
		{X: 0, Y: 0},
		{X: 10, Y: 0},
		{X: 10, Y: 10},
		{X: 0, Y: 10},
	}
	area := outlineArea(square)
	if math.Abs(area-100) > 0.01 {
		t.Errorf("expected area 100, got %.2f", area)
	}
}

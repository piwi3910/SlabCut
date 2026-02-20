package gcode

import (
	"fmt"
	"math"
	"strings"

	"github.com/piwi3910/cnc-calculator/internal/model"
)

// Generator produces GCode from an optimized sheet layout.
type Generator struct {
	Settings model.CutSettings
	profile  model.GCodeProfile
}

func New(settings model.CutSettings) *Generator {
	return &Generator{
		Settings: settings,
		profile:  model.GetProfile(settings.GCodeProfile),
	}
}

// GenerateSheet produces GCode for a single sheet's placements.
func (g *Generator) GenerateSheet(sheet model.SheetResult, sheetIndex int) string {
	var b strings.Builder

	g.writeHeader(&b, sheet, sheetIndex)

	for i, placement := range sheet.Placements {
		g.writePart(&b, placement, i+1)
	}

	g.writeFooter(&b)
	return b.String()
}

// GenerateAll produces one GCode string per sheet.
func (g *Generator) GenerateAll(result model.OptimizeResult) []string {
	var codes []string
	for i, sheet := range result.Sheets {
		codes = append(codes, g.GenerateSheet(sheet, i+1))
	}
	return codes
}

func (g *Generator) writeHeader(b *strings.Builder, sheet model.SheetResult, idx int) {
	p := g.profile

	// Write file header comment
	b.WriteString(p.CommentPrefix)
	b.WriteString(fmt.Sprintf(" CNCCalculator GCode — Sheet %d (%s)\n", idx, sheet.Stock.Label))
	b.WriteString(p.CommentPrefix)
	b.WriteString(fmt.Sprintf(" Stock: %.1f x %.1f mm\n", sheet.Stock.Width, sheet.Stock.Height))
	b.WriteString(p.CommentPrefix)
	b.WriteString(fmt.Sprintf(" Parts: %d, Efficiency: %.1f%%\n", len(sheet.Placements), sheet.Efficiency()))
	b.WriteString(p.CommentPrefix)
	b.WriteString(fmt.Sprintf(" Tool: %.1fmm, Feed: %.0f mm/min, Plunge: %.0f mm/min\n",
		g.Settings.ToolDiameter, g.Settings.FeedRate, g.Settings.PlungeRate))
	b.WriteString(p.CommentPrefix)
	b.WriteString(fmt.Sprintf(" Depth: %.1fmm in %.1fmm passes\n", g.Settings.CutDepth, g.Settings.PassDepth))
	b.WriteString(p.CommentPrefix)
	b.WriteString(fmt.Sprintf(" Profile: %s\n", p.Name))
	b.WriteString("\n")

	// Write startup codes
	for _, code := range p.StartCode {
		b.WriteString(code + "\n")
	}

	// Spindle start
	if p.SpindleStart != "" {
		b.WriteString(fmt.Sprintf(p.SpindleStart+"\n", g.Settings.SpindleSpeed))
	}

	// Initial safe Z retract
	b.WriteString(fmt.Sprintf("%s X%s Y%s\n", p.RapidMove, g.format(0), g.format(0)))
	b.WriteString(fmt.Sprintf("%s Z%s\n", p.RapidMove, g.format(g.Settings.SafeZ)))

	b.WriteString("\n")
}

func (g *Generator) writeFooter(b *strings.Builder) {
	p := g.profile

	b.WriteString("\n")
	b.WriteString(p.CommentPrefix + " === Job complete ===\n")

	// Write end codes
	for _, code := range p.EndCode {
		// Replace [SafeZ] placeholder
		code = strings.ReplaceAll(code, "[SafeZ]", g.format(g.Settings.SafeZ))
		b.WriteString(code + "\n")
	}

	// Spindle stop
	if p.SpindleStop != "" {
		b.WriteString(p.SpindleStop + "\n")
	}
}

func (g *Generator) writePart(b *strings.Builder, p model.Placement, partNum int) {
	toolR := g.Settings.ToolDiameter / 2.0

	// The part rectangle in stock coordinates
	pw := p.PlacedWidth()
	ph := p.PlacedHeight()

	// Offset for tool radius (cut outside the part perimeter)
	x0 := p.X - toolR
	y0 := p.Y - toolR
	x1 := p.X + pw + toolR
	y1 := p.Y + ph + toolR

	b.WriteString(g.comment(fmt.Sprintf("--- Part %d: %s (%.1f x %.1f)%s ---",
		partNum, p.Part.Label, p.Part.Width, p.Part.Height,
		rotatedStr(p.Rotated))))

	numPasses := int(math.Ceil(g.Settings.CutDepth / g.Settings.PassDepth))

	// Generate tabs info
	tabs := g.calculateTabs(p)

	for pass := 1; pass <= numPasses; pass++ {
		depth := float64(pass) * g.Settings.PassDepth
		if depth > g.Settings.CutDepth {
			depth = g.Settings.CutDepth
		}
		isFinalPass := pass == numPasses

		b.WriteString(g.comment(fmt.Sprintf("Pass %d/%d, depth=%.2fmm", pass, numPasses, depth)))

		// Rapid to start (top-left corner, slightly outside)
		b.WriteString(fmt.Sprintf("%s X%s Y%s\n", g.profile.RapidMove, g.format(x0), g.format(y0)))
		b.WriteString(fmt.Sprintf("%s Z%s F%s ; Plunge\n", g.profile.FeedMove, g.format(-depth), g.format(g.Settings.PlungeRate)))

		// Cut rectangle perimeter (clockwise for climb milling)
		if isFinalPass && g.Settings.PartTabsPerSide > 0 {
			g.writePerimeterWithTabs(b, x0, y0, x1, y1, depth, tabs)
		} else {
			g.writePerimeter(b, x0, y0, x1, y1)
		}

		// Retract between passes
		b.WriteString(fmt.Sprintf("%s Z%s\n", g.profile.RapidMove, g.format(g.Settings.SafeZ)))
	}

	b.WriteString("\n")
}

func (g *Generator) writePerimeter(b *strings.Builder, x0, y0, x1, y1 float64) {
	p := g.profile
	b.WriteString(fmt.Sprintf("%s X%s Y%s F%s\n", p.FeedMove, g.format(x1), g.format(y0), g.format(g.Settings.FeedRate)))
	b.WriteString(fmt.Sprintf("%s X%s Y%s\n", p.FeedMove, g.format(x1), g.format(y1)))
	b.WriteString(fmt.Sprintf("%s X%s Y%s\n", p.FeedMove, g.format(x0), g.format(y1)))
	b.WriteString(fmt.Sprintf("%s X%s Y%s\n", p.FeedMove, g.format(x0), g.format(y0)))
}

// comment wraps text in the profile's comment syntax.
func (g *Generator) comment(text string) string {
	return g.profile.CommentPrefix + " " + text + g.profile.CommentSuffix + "\n"
}

// format formats a coordinate according to the profile's decimal places.
func (g *Generator) format(v float64) string {
	format := fmt.Sprintf("%%.%df", g.profile.DecimalPlaces)
	return fmt.Sprintf(format, v)
}

// Tab represents a holding tab position along the perimeter.
type Tab struct {
	side     int     // 0=bottom, 1=right, 2=top, 3=left
	startPos float64 // distance along that side
}

func (g *Generator) calculateTabs(p model.Placement) []Tab {
	if g.Settings.PartTabsPerSide <= 0 {
		return nil
	}

	pw := p.PlacedWidth() + g.Settings.ToolDiameter
	ph := p.PlacedHeight() + g.Settings.ToolDiameter

	var tabs []Tab
	for side := 0; side < 4; side++ {
		var length float64
		if side == 0 || side == 2 {
			length = pw
		} else {
			length = ph
		}
		spacing := length / float64(g.Settings.PartTabsPerSide+1)
		for t := 1; t <= g.Settings.PartTabsPerSide; t++ {
			tabs = append(tabs, Tab{
				side:     side,
				startPos: spacing * float64(t),
			})
		}
	}
	return tabs
}

func (g *Generator) writePerimeterWithTabs(b *strings.Builder, x0, y0, x1, y1, depth float64, tabs []Tab) {
	tabDepth := depth - g.Settings.PartTabHeight
	if tabDepth < 0 {
		tabDepth = 0
	}
	tw := g.Settings.PartTabWidth

	// Side 0: bottom (x0,y0) -> (x1,y0)
	g.writeSideWithTabs(b, x0, y0, x1, y0, true, depth, tabDepth, tw, g.tabsForSide(tabs, 0))
	// Side 1: right (x1,y0) -> (x1,y1)
	g.writeSideWithTabs(b, x1, y0, x1, y1, false, depth, tabDepth, tw, g.tabsForSide(tabs, 1))
	// Side 2: top (x1,y1) -> (x0,y1)
	g.writeSideWithTabs(b, x1, y1, x0, y1, true, depth, tabDepth, tw, g.tabsForSide(tabs, 2))
	// Side 3: left (x0,y1) -> (x0,y0)
	g.writeSideWithTabs(b, x0, y1, x0, y0, false, depth, tabDepth, tw, g.tabsForSide(tabs, 3))
}

func (g *Generator) tabsForSide(tabs []Tab, side int) []Tab {
	var result []Tab
	for _, t := range tabs {
		if t.side == side {
			result = append(result, t)
		}
	}
	return result
}

func (g *Generator) writeSideWithTabs(b *strings.Builder, x0, y0, x1, y1 float64, isHoriz bool,
	cutDepth, tabDepth, tabWidth float64, tabs []Tab) {

	if len(tabs) == 0 {
		b.WriteString(fmt.Sprintf("%s X%s Y%s F%s\n", g.profile.FeedMove, g.format(x1), g.format(y1), g.format(g.Settings.FeedRate)))
		return
	}

	dx := x1 - x0
	dy := y1 - y0
	length := math.Sqrt(dx*dx + dy*dy)
	if length < 0.001 {
		return
	}
	nx := dx / length
	ny := dy / length

	// Walk along the side, raising Z for tabs
	cursor := 0.0
	for _, tab := range tabs {
		tabStart := tab.startPos - tabWidth/2
		tabEnd := tab.startPos + tabWidth/2

		// Cut to tab start
		if tabStart > cursor {
			px := x0 + nx*tabStart
			py := y0 + ny*tabStart
			b.WriteString(fmt.Sprintf("%s X%s Y%s F%s\n", g.profile.FeedMove, g.format(px), g.format(py), g.format(g.Settings.FeedRate)))
		}

		// Raise to tab height
		b.WriteString(fmt.Sprintf("%s Z%s\n", g.profile.FeedMove, g.format(-tabDepth)))
		// Traverse tab
		px := x0 + nx*tabEnd
		py := y0 + ny*tabEnd
		b.WriteString(fmt.Sprintf("%s X%s Y%s\n", g.profile.FeedMove, g.format(px), g.format(py)))
		// Plunge back down
		b.WriteString(fmt.Sprintf("%s Z%s\n", g.profile.FeedMove, g.format(-cutDepth)))

		cursor = tabEnd
	}

	// Finish to end of side
	b.WriteString(fmt.Sprintf("%s X%s Y%s F%s\n", g.profile.FeedMove, g.format(x1), g.format(y1), g.format(g.Settings.FeedRate)))
}

func rotatedStr(r bool) string {
	if r {
		return " [rotated]"
	}
	return ""
}

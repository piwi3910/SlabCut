# SlabCut UI Redesign — OrcaSlicer-Inspired Layout

**Date:** 2026-02-20
**Issue:** #103
**Status:** Design approved

## Problem Statement

SlabCut's current tab-based UI (Parts | Stock | Settings | Results) has several usability issues:

1. **Canvas hidden behind a tab** — Users can't see cut layouts while editing parts or settings
2. **No live feedback** — Settings changes require manually clicking Optimize, then switching to the Results tab
3. **Settings scattered** — CNC/tool settings buried in a separate tab, disconnected from the visual result
4. **Dead interactive widgets** — SheetCanvas (zoom/pan) and GCode simulation exist but aren't wired into the main UI
5. **Information overload** — 11 columns in the parts grid, 10+ settings sections in one scrollable panel

## Design

### Two-Tab Architecture

Inspired by OrcaSlicer's **Prepare / Preview** paradigm:

- **Tab 1: "Layout Editor"** — The daily driver. Three-pane layout with settings, canvas, and parts/stock
- **Tab 2: "GCode Preview"** — Toolpath visualization with simulation controls

### Tab 1: Layout Editor

```
+------------------+---------------------------+-------------------+
|                  |                           |                   |
| LEFT PANEL       |   CENTER CANVAS           |  RIGHT PANEL      |
| (Quick Settings) |   (Interactive Sheet View) |  (Parts + Stock)  |
|                  |                           |                   |
| [Accordion]      |   SheetCanvas widget      |  [Accordion]      |
| MultiOpen=true   |   with zoom/pan           |  MultiOpen=true   |
|                  |                           |                   |
| > Tool           |                           |  > Parts (6)      |
|   [profile ▼]    |   +------------------+    |    [Name] [W] [H] |
|   Diameter  [  ] |   |                  |    |    [Qty] [Grain]  |
|   Feed Rate [  ] |   |  Interactive     |    |    [+ Add]        |
|   Plunge    [  ] |   |  SheetCanvas     |    |    +-----------+  |
|   RPM       [  ] |   |  (zoom/pan)      |    |    | Part 1    |  |
|   Slot      [  ] |   |                  |    |    | 200x100 x3|  |
|                  |   |                  |    |    | [edit][del]|  |
| > Material       |   +------------------+    |    +-----------+  |
|   [stock ▼]      |                           |    | Part 2    |  |
|   Thickness [  ] |   [Sht 1][Sht 2][Sht 3]  |    | 150x75 x2 |  |
|   Kerf      [  ] |                           |    +-----------+  |
|   Edge Trim [  ] |                           |                   |
|                  |                           |  > Stock Sheets(2)|
| > Cutting        |                           |    [Name] [W] [H] |
|   Safe Z    [  ] |                           |    [Thick][Qty]   |
|   Cut Depth [  ] |                           |    [+ Add]        |
|   Pass Depth[  ] |                           |    +-----------+  |
|   Tabs    [on/off]|                          |    | Plywood   |  |
|                  |                           |    | 2440x1220 |  |
| > Optimizer      |                           |    | 18mm  x1  |  |
|   [Algorithm ▼]  |                           |    | [edit][del]|  |
|   Guillotine-only|                           |    +-----------+  |
|                  |                           |                   |
| [gear] Advanced..|                           |                   |
+------------------+---------------------------+-------------------+
| SlabCut v0.0.4   | 3 sheets, 87.2% eff.     | [Export GCode ▼]  |
+------------------------------------------------------------------+
```

**Split ratios:** Left 22% | Center 53% | Right 25% (resizable via HSplit)

### Left Panel — Quick Settings (Accordion)

Four collapsible sections, all open by default:

#### Tool Section
| Control | Type | Binds to |
|---------|------|----------|
| Load Tool Profile | Select (from inventory) | Populates fields below |
| Tool Diameter (mm) | Entry | `Settings.ToolDiameter` |
| Feed Rate (mm/min) | Entry | `Settings.FeedRate` |
| Plunge Rate (mm/min) | Entry | `Settings.PlungeRate` |
| Spindle Speed (RPM) | Entry | `Settings.SpindleSpeed` |
| CNC Slot | Select (None, T1-T12) | Display only (from profile) |

#### Material Section
| Control | Type | Binds to |
|---------|------|----------|
| Load Stock Preset | Select (from inventory) | Populates fields below |
| Thickness (mm) | Entry | Applied to selected stock |
| Kerf / Blade Width (mm) | Entry | `Settings.KerfWidth` |
| Edge Trim (mm) | Entry | `Settings.EdgeTrim` |

#### Cutting Section
| Control | Type | Binds to |
|---------|------|----------|
| Safe Z Height (mm) | Entry | `Settings.SafeZ` |
| Cut Depth (mm) | Entry | `Settings.CutDepth` |
| Depth per Pass (mm) | Entry | `Settings.PassDepth` |
| Enable Tabs | Check | `Settings.StockTabs.Enabled` |
| Tab Padding (mm) | Entry (collapsed subsection) | `Settings.StockTabs.*` |

#### Optimizer Section
| Control | Type | Binds to |
|---------|------|----------|
| Algorithm | Select (Guillotine/Genetic) | `Settings.Algorithm` |
| Guillotine Cuts Only | Check | `Settings.GuillotineOnly` |

#### Advanced Settings Button
A gear icon button at the bottom opens a separate dialog window containing all remaining settings:
- Optimization weights
- Lead-in / lead-out arcs
- Plunge entry strategy (ramp/helix)
- Corner overcuts (dogbone/t-bone)
- Onion skinning
- Fixture/clamp zones
- Dust shoe collision detection
- GCode profile management
- Nesting rotations

### Center Panel — Interactive Sheet Canvas

- Uses the existing `SheetCanvas` widget (currently dead code) with zoom and pan
- Sheet selector buttons below the canvas: `[Sheet 1] [Sheet 2] [Sheet 3]`
- Clicking a sheet button switches the canvas to display that sheet
- Zoom controls: scroll wheel (already implemented), optional zoom +/- buttons
- Reset zoom button
- Shows: stock background, part placements (colored), part labels, tab zones, clamp zones
- Empty state: "Add parts and stock sheets to see the layout" with a subtle icon

### Right Panel — Parts & Stock (Accordion)

Two collapsible sections, both open by default:

#### Parts Section
**Quick-add bar** (compact, vertical layout within the accordion):
- Row 1: `[Name entry] [+ button]`
- Row 2: `[Width] x [Height] [Qty] [Grain ▼]`

**Part cards** (one per part, scrollable):
```
+----------------------------------+
| Part Name              [ed][del] |
| 200 x 100 mm  x3   Grain: None  |
+----------------------------------+
```
- Part name in bold, dimensions and quantity on second line
- Edit (pencil) and delete (trash) icons top-right
- Clicking edit opens the existing detailed edit dialog
- Alternating subtle background colors for readability

#### Stock Section
**Quick-add bar** (compact):
- Row 1: `[Name entry] [+ button]`
- Row 2: `[Width] x [Height] [Thick] [Qty]`

**Stock cards** (one per stock, scrollable):
```
+----------------------------------+
| Plywood 2440x1220       [ed][del]|
| 18mm thick  x1    Grain: None   |
+----------------------------------+
```

### Tab 2: GCode Preview

```
+------------------------------------------------------------------+
| [Sheet selector: Sheet 1 ▼]  [GCode Profile: GRBL ▼]            |
+------------------------------------------------------------------+
|                                                                  |
|   GCodePreview widget (full width, ~80% height)                  |
|   Colored toolpath visualization:                                |
|     Red = rapid, Blue = feed, Green = plunge markers             |
|                                                                  |
+------------------------------------------------------------------+
| [Slider: move 0 ——————————————————————————— 847]                 |
| [|◀] [▶/❚❚] [◼] [▶|]  Speed: [1x ▼]  Loop: [ ]                |
| Move: 234 / 847  |  X: 120.5  Y: 45.2  Z: -6.0  F: 1500       |
+------------------------------------------------------------------+
| Sheet 1 of 3  |  847 moves  |  Est. cut time: ~12:34            |
+------------------------------------------------------------------+
```

- Wires in the existing `RenderGCodeSimulation` code
- Sheet selector dropdown to switch between sheets
- GCode profile dropdown (for export format)
- Full simulation controls: play/pause, stop, step forward/back, speed selector, loop toggle
- Move coordinates display
- Status bar shows sheet info and estimated cut time

### Status Bar (Always Visible)

```
| SlabCut v0.0.4 (abc1234) | 3 sheets, 87.2% efficiency, 2.4m waste | [Export GCode ▼] |
```

- Left: Version string
- Center: Optimization summary (updates live after each optimization)
- Right: Quick export button with dropdown (GCode, PDF, Project)

### Live Auto-Optimization

**Trigger:** Any change to parts list, stock list, or quick settings
**Debounce:** 500ms after last change
**Execution:** Runs `engine.Optimize()` in a goroutine
**Feedback:**
- Status bar shows "Optimizing..." with a subtle indicator
- Canvas updates automatically when complete
- Sheet selector updates if sheet count changes
- If no parts or no stock, canvas shows empty state message

**Cancel:** If a new change arrives while optimizing, cancel the current run and restart with 500ms debounce

**Implementation:**
```go
// In App struct
optimizeTimer *time.Timer
optimizeMu    sync.Mutex
optimizing    bool

func (a *App) scheduleOptimize() {
    a.optimizeMu.Lock()
    defer a.optimizeMu.Unlock()
    if a.optimizeTimer != nil {
        a.optimizeTimer.Stop()
    }
    a.optimizeTimer = time.AfterFunc(500*time.Millisecond, func() {
        a.runAutoOptimize()
    })
}
```

### Menu Bar Changes

**Simplified menus:**

| Menu | Items |
|------|-------|
| **File** | New, Open, Save, Save As, separator, Import (CSV/Excel/DXF submenu), Export (GCode/PDF submenu), separator, Share, Import Shared, separator, Quit |
| **Edit** | Undo, Redo, separator, Clear Parts, Clear Stock |
| **Tools** | Force Re-Optimize, Compare Settings..., Purchasing Calculator... |
| **Admin** | Parts Library, Tool Inventory, Stock Inventory, Project Templates, separator, Advanced Settings..., GCode Profiles..., separator, Import/Export Data, App Settings |
| **Help** | About |

Key changes:
- "Optimize" removed from Tools (now automatic) — replaced with "Force Re-Optimize"
- "Advanced Settings..." added to Admin menu (opens the advanced settings dialog)
- "GCode Profiles..." moved to Admin menu

### Window Sizing

- Default: 1400 x 800 (wider than current 1000x700 to accommodate three-pane)
- Minimum: 1000 x 600
- HSplit dividers are resizable

## Migration Plan

### What stays the same
- All model types (Part, StockSheet, CutSettings, etc.)
- Engine/optimizer code
- GCode generator
- Import/export code
- Inventory persistence
- Project save/load format
- All dialog forms (add/edit part, add/edit stock, etc.)

### What changes
- `internal/ui/app.go` — Major rewrite: new Build(), new panel builders, auto-optimize
- `internal/ui/app.go` — Parts/stock display changes from grid to cards
- `internal/ui/app.go` — Settings panel split into quick (sidebar) and advanced (dialog)
- `internal/ui/widgets/sheet_canvas.go` — Minor: wire into main layout (already functional)
- `internal/ui/widgets/gcode_preview.go` — Minor: wire simulation into GCode Preview tab
- New: `internal/ui/advanced_settings.go` — Advanced settings dialog (extracted from current settings panel)

### What gets removed
- The four-tab structure (Parts | Stock | Settings | Results)
- Static image rendering in results (`renderSheetToImage` → replaced by interactive SheetCanvas)
- The standalone "Optimize" button (replaced by auto-optimize)

## Non-Goals (YAGNI)

- Settings search (too few settings to warrant it)
- Per-part settings overrides (not in scope for this redesign)
- Tree view for parts (compact cards are sufficient)
- Collapsible sidebar toggle (Fyne HSplit handles resizing)
- Modified-value highlighting (nice-to-have, not critical)

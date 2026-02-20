# CutOptimizer

A cross-platform desktop application for optimizing rectangular cut lists and generating CNC-ready GCode. Built with Go and [Fyne](https://fyne.io/) — produces a single native binary for Windows, macOS, and Linux with no runtime dependencies.

## Features

- **2D Bin Packing** — Guillotine-based optimization with Best Area Fit heuristic
- **Grain Direction** — Supports horizontal/vertical grain constraints
- **Saw Kerf & Edge Trim** — Accounts for blade width and stock edge waste
- **Part Rotation** — Automatically rotates parts for better fit (respects grain)
- **Visual Layout** — Color-coded sheet diagrams showing part placements
- **GCode Export** — Full CNC toolpath generation with:
  - Multi-pass depth stepping
  - Configurable feed/plunge rates and spindle speed
  - Holding tabs to prevent part movement
  - Tool radius compensation (outside cut)
  - Safe Z retract between operations
- **Project Save/Load** — JSON-based project files (`.cutopt`)

## Prerequisites

- Go 1.22+
- C compiler (GCC/MinGW on Windows, Xcode CLI tools on macOS)
  - Required by Fyne for CGo graphics bindings
- On Linux: `sudo apt install libgl1-mesa-dev xorg-dev` (for OpenGL)

## Build

```bash
# Run directly
make run

# Build for current platform
make build

# Cross-compile
make windows        # produces cutoptimizer.exe
make darwin-arm64   # produces cutoptimizer-darwin-arm64 (Apple Silicon)
make darwin-amd64   # produces cutoptimizer-darwin-amd64 (Intel Mac)
make linux          # produces cutoptimizer-linux
```

### Packaged Builds (recommended for distribution)

Uses [fyne-cross](https://github.com/fyne-io/fyne-cross) for proper `.exe`/`.app` bundles:

```bash
go install github.com/fyne-io/fyne-cross@latest

make package-windows   # Windows .exe with icon
make package-darwin    # macOS .app bundle (universal binary)
```

## Run Tests

```bash
make test
```

## Project Structure

```
cutoptimizer/
├── cmd/cutoptimizer/
│   └── main.go              # Entry point
├── internal/
│   ├── model/
│   │   └── model.go         # Core data types (Part, StockSheet, Placement, etc.)
│   ├── engine/
│   │   ├── optimizer.go      # Guillotine bin-packing algorithm
│   │   └── optimizer_test.go
│   ├── gcode/
│   │   ├── generator.go      # GCode toolpath generation
│   │   └── generator_test.go
│   ├── project/
│   │   └── project.go        # Save/load project files
│   └── ui/
│       ├── app.go            # Main UI (tabs, toolbar, dialogs)
│       └── widgets/
│           └── sheet_canvas.go  # Visual sheet layout renderer
├── go.mod
├── Makefile
└── README.md
```

## Architecture

```
┌─────────────────────────────────────────────────────┐
│                    UI Layer (Fyne)                   │
│  ┌──────────┐ ┌───────────┐ ┌────────┐ ┌────────┐  │
│  │  Parts   │ │   Stock   │ │Settings│ │Results │  │
│  │  Panel   │ │   Panel   │ │ Panel  │ │ Panel  │  │
│  └──────────┘ └───────────┘ └────────┘ └────────┘  │
├─────────────────────────────────────────────────────┤
│                  Core Engine                         │
│  ┌──────────────────┐  ┌──────────────────────────┐ │
│  │ Guillotine Packer│  │   GCode Generator        │ │
│  │ (bin-packing)    │  │   (toolpath + tabs)       │ │
│  └──────────────────┘  └──────────────────────────┘ │
├─────────────────────────────────────────────────────┤
│  Model Layer: Part, StockSheet, Placement, Project  │
└─────────────────────────────────────────────────────┘
```

## Future Improvements

- [ ] Improved optimizer: genetic algorithm meta-heuristic for better packing
- [ ] DXF import for non-rectangular parts
- [ ] GCode preview with simulated toolpath visualization
- [ ] Multiple stock sheet sizes in one optimization run (best-fit selection)
- [ ] CSV/Excel import for part lists
- [ ] PDF export of cut diagrams
- [ ] Undo/redo for part/stock edits
- [ ] Lead-in/lead-out arcs for smoother CNC entry
- [ ] Configurable GCode post-processor profiles (Grbl, Mach3, LinuxCNC)

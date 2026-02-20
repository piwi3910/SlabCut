.PHONY: build run clean windows darwin linux package-windows package-darwin

# Default: build for current platform
build:
	go build -o cnc-calculator ./cmd/cnc-calculator

run:
	go run ./cmd/cnc-calculator

# Cross-compilation (basic, no app bundling)
windows:
	GOOS=windows GOARCH=amd64 go build -ldflags="-H windowsgui" -o cnc-calculator.exe ./cmd/cnc-calculator

darwin-amd64:
	GOOS=darwin GOARCH=amd64 go build -o cnc-calculator-darwin-amd64 ./cmd/cnc-calculator

darwin-arm64:
	GOOS=darwin GOARCH=arm64 go build -o cnc-calculator-darwin-arm64 ./cmd/cnc-calculator

linux:
	GOOS=linux GOARCH=amd64 go build -o cnc-calculator-linux ./cmd/cnc-calculator

# Proper packaging with fyne-cross (produces .exe installer / .app bundle)
# Install first: go install github.com/fyne-io/fyne-cross@latest
package-windows:
	fyne-cross windows -arch=amd64 ./cmd/cnc-calculator

package-darwin:
	fyne-cross darwin -arch=amd64,arm64 ./cmd/cnc-calculator

# Run tests
test:
	go test ./...

clean:
	rm -f cnc-calculator cnc-calculator.exe cnc-calculator-darwin-* cnc-calculator-linux
	rm -rf fyne-cross

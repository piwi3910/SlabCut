.PHONY: build run clean windows darwin linux package-windows package-darwin

# Default: build for current platform
build:
	go build -o slabcut ./cmd/slabcut

run:
	go run ./cmd/slabcut

# Cross-compilation (basic, no app bundling)
windows:
	GOOS=windows GOARCH=amd64 go build -ldflags="-H windowsgui" -o slabcut.exe ./cmd/slabcut

darwin-amd64:
	GOOS=darwin GOARCH=amd64 go build -o slabcut-darwin-amd64 ./cmd/slabcut

darwin-arm64:
	GOOS=darwin GOARCH=arm64 go build -o slabcut-darwin-arm64 ./cmd/slabcut

linux:
	GOOS=linux GOARCH=amd64 go build -o slabcut-linux ./cmd/slabcut

# Proper packaging with fyne-cross (produces .exe installer / .app bundle)
# Install first: go install github.com/fyne-io/fyne-cross@latest
package-windows:
	fyne-cross windows -arch=amd64 ./cmd/slabcut

package-darwin:
	fyne-cross darwin -arch=amd64,arm64 ./cmd/slabcut

# Run tests
test:
	go test ./...

clean:
	rm -f slabcut slabcut.exe slabcut-darwin-* slabcut-linux
	rm -rf fyne-cross

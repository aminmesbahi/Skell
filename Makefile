# Skell Makefile
# Usage:
#   make build VERSION=v0.1.0   # build all platform CLI binaries
#   make build                   # build with version=dev
#   make gui-build               # build desktop GUI (requires wails CLI + bun)
#   make gui-dev                 # start GUI in live-reload dev mode
#   make clean                   # remove dist/
#   make test                    # run all tests
#   make lint                    # run golangci-lint

VERSION ?= dev
MODULE   = github.com/aminmesbahi/skell/internal/version
LDFLAGS  = -s -w -X $(MODULE).Version=$(VERSION)
DIST     = dist

PLATFORMS = \
	windows/amd64 \
	windows/arm64 \
	linux/amd64 \
	linux/arm64 \
	darwin/amd64 \
	darwin/arm64

.PHONY: build gui-build gui-dev clean test lint

build: clean
	@echo "Building Skell $(VERSION)"
	@mkdir -p $(DIST)
	@$(foreach PLATFORM,$(PLATFORMS), \
		$(eval GOOS   := $(word 1,$(subst /, ,$(PLATFORM)))) \
		$(eval GOARCH := $(word 2,$(subst /, ,$(PLATFORM)))) \
		$(eval EXT    := $(if $(filter windows,$(GOOS)),.exe,)) \
		$(eval OUT    := $(DIST)/skell_$(GOOS)_$(GOARCH)$(EXT)) \
		printf "  %-35s" "skell_$(GOOS)_$(GOARCH)$(EXT)" ; \
		GOOS=$(GOOS) GOARCH=$(GOARCH) go build -trimpath -ldflags "$(LDFLAGS)" -o $(OUT) . && echo "OK" || echo "FAILED" ; \
	)
	@echo ""
	@echo "Done. Artifacts in ./$(DIST)/"
	@ls -lh $(DIST)/

gui-build:
	@echo "Building Skell Desktop GUI..."
	cd gui && wails build

gui-dev:
	@echo "Starting Skell Desktop GUI in dev mode..."
	cd gui && wails dev

clean:
	@rm -rf $(DIST)

test:
	go test ./...

lint:
	golangci-lint run ./...

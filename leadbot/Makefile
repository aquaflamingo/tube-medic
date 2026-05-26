BIN  := bin/ytleadbot
EMBED := internal/utils/ytdlp_static/yt-dlp

.PHONY: build build-darwin-arm64 test clean distclean run-* lint download-yt-dlp

build: $(EMBED)
	go build -o $(BIN) ./cmd/ytleadbot

build-darwin-arm64: EMBED_URL=https://github.com/yt-dlp/yt-dlp/releases/latest/download/yt-dlp_macos
build-darwin-arm64:
	@mkdir -p internal/utils/ytdlp_static
	curl -sL -o $(EMBED) "$(EMBED_URL)" && chmod +x $(EMBED)
	GOOS=darwin GOARCH=arm64 go build -o bin/ytleadbot-darwin-arm64 ./cmd/ytleadbot

test:
	go test ./... -v -count=1

clean:
	rm -rf $(BIN) exports/*.json

distclean: clean
	rm -f $(EMBED)
	go clean -cache

lint:
	go vet ./...

download-yt-dlp: $(EMBED)

$(EMBED):
	@mkdir -p internal/utils/ytdlp_static
	@case "$$(uname -s)-$$(uname -m)" in \
		Linux-x86_64)  url="https://github.com/yt-dlp/yt-dlp/releases/latest/download/yt-dlp_linux" ;; \
		Linux-aarch64) url="https://github.com/yt-dlp/yt-dlp/releases/latest/download/yt-dlp_linux_aarch64" ;; \
		Darwin-arm64|Darwin-x86_64) url="https://github.com/yt-dlp/yt-dlp/releases/latest/download/yt-dlp_macos" ;; \
		*) echo "unsupported platform: $$(uname -s)-$$(uname -m)"; exit 1 ;; \
	esac; \
	echo "downloading yt-dlp for $$(uname -s)-$$(uname -m) ..."; \
	curl -sL -o "$@" "$$url" && chmod +x "$@"

run-discover: build
	$(BIN) --job discover

run-enrich: build
	$(BIN) --job enrich

run-export: build
	$(BIN) --job export

run-agency-scan: build
	$(BIN) --job agency_scan

run-reenrich: build
	$(BIN) --job reenrich_stale

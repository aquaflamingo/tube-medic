package utils

import (
	_ "embed"
	"fmt"
	"os"
	"path/filepath"
	"sync"
)

//go:embed ytdlp_static/yt-dlp
var ytdlpBinary []byte

var (
	ytdlpExtractedPath string
	ytdlpExtractErr    error
	ytdlpExtractOnce   sync.Once
)

func ensureYTDLPExtracted() (string, error) {
	ytdlpExtractOnce.Do(func() {
		cacheDir, err := os.UserCacheDir()
		if err != nil {
			ytdlpExtractErr = fmt.Errorf("get cache dir: %w", err)
			return
		}
		dir := filepath.Join(cacheDir, "ytleadbot")
		if err := os.MkdirAll(dir, 0755); err != nil {
			ytdlpExtractErr = fmt.Errorf("create cache dir: %w", err)
			return
		}
		path := filepath.Join(dir, "yt-dlp")
		if fi, err := os.Stat(path); err == nil {
			if fi.Size() == int64(len(ytdlpBinary)) {
				ytdlpExtractedPath = path
				return
			}
		}
		if err := os.WriteFile(path, ytdlpBinary, 0755); err != nil {
			ytdlpExtractErr = fmt.Errorf("write yt-dlp: %w", err)
			return
		}
		ytdlpExtractedPath = path
	})
	return ytdlpExtractedPath, ytdlpExtractErr
}

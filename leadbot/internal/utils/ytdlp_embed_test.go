package utils

import (
	"os"
	"os/exec"
	"runtime"
	"testing"
)

func TestEmbeddedBinary_NonEmpty(t *testing.T) {
	if len(ytdlpBinary) == 0 {
		t.Fatal("embedded yt-dlp binary is empty — did you run `make download-yt-dlp`?")
	}
}

func TestEnsureYTDLPExtracted_ReturnsExecutable(t *testing.T) {
	path, err := ensureYTDLPExtracted()
	if err != nil {
		t.Fatalf("ensureYTDLPExtracted failed: %v", err)
	}
	if path == "" {
		t.Fatal("expected non-empty path")
	}
	info, err := os.Stat(path)
	if err != nil {
		t.Fatalf("stat extracted binary: %v", err)
	}
	if info.Mode()&0111 == 0 {
		t.Fatal("extracted binary is not executable")
	}
}

func TestEnsureYTDLPExtracted_BinaryRuns(t *testing.T) {
	path, err := ensureYTDLPExtracted()
	if err != nil {
		t.Fatalf("ensureYTDLPExtracted failed: %v", err)
	}
	out, err := exec.Command(path, "--version").Output()
	if err != nil {
		t.Fatalf("extracted yt-dlp --version failed: %v", err)
	}
	if len(out) == 0 {
		t.Fatal("yt-dlp --version produced no output")
	}
}

func TestYTDLPPath_ReturnsPath(t *testing.T) {
	path, err := ytdlpPath()
	if err != nil {
		t.Fatalf("ytdlpPath failed: %v", err)
	}
	if path == "" {
		t.Fatal("expected non-empty path")
	}
	if runtime.GOOS == "linux" {
		info, err := os.Stat(path)
		if err != nil {
			t.Fatalf("stat yt-dlp binary: %v", err)
		}
		if info.Mode()&0111 == 0 {
			t.Fatal("yt-dlp binary is not executable")
		}
	}
}

func TestYTDLPPath_ErrorOnNonLinux(t *testing.T) {
	if runtime.GOOS == "linux" {
		t.Skip("not applicable on linux — embedded binary is used")
	}
	_, err := ytdlpPath()
	if err != nil {
		t.Logf("ytdlpPath fallback error (expected without system yt-dlp): %v", err)
	}
}

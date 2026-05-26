package handlers_test

import (
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"strings"
	"testing"

	"github.com/aqfl/tubemedic-web/internal/handlers"
)

func TestHandleScan_ValidationErrors(t *testing.T) {
	// Back up the current environment API key
	originalKey := os.Getenv("YT_API_KEY")
	defer func() {
		os.Setenv("YT_API_KEY", originalKey)
	}()

	// 1. Missing Channel URL
	req := httptest.NewRequest(http.MethodPost, "/scan", nil)
	w := httptest.NewRecorder()
	handlers.HandleScan(w, req)

	body := w.Body.String()
	if !strings.Contains(body, "Validation Error") {
		t.Errorf("expected Validation Error, got: %s", body)
	}

	// 2. Missing API Key
	os.Setenv("YT_API_KEY", "")
	form := url.Values{}
	form.Set("channel", "https://youtube.com/@mkbhd")
	req = httptest.NewRequest(http.MethodPost, "/scan", strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	w = httptest.NewRecorder()
	handlers.HandleScan(w, req)

	body = w.Body.String()
	if !strings.Contains(body, "Server Configuration Error") {
		t.Errorf("expected Server Configuration Error, got: %s", body)
	}
}

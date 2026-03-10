package main

import (
	"bytes"
	"github.com/mkailbowdy/internal/assert"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestPing(t *testing.T) {
	// Create a Response Recorder
	rr := httptest.NewRecorder()

	req, err := http.NewRequest(http.MethodGet, "/", nil)
	if err != nil {
		t.Fatal(err)
	}

	ping(rr, req)

	res := rr.Result()
	defer res.Body.Close()

	// Check if the status code is 200
	assert.Equal(t, res.StatusCode, http.StatusOK)

	body, err := io.ReadAll(res.Body)
	if err != nil {
		t.Fatal(err)
	}
	body = bytes.TrimSpace(body)
	// Check if the body of the request is OK
	assert.Equal(t, string(body), "OK")
}

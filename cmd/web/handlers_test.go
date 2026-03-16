package main

import (
	"net/http"
	"testing"

	"github.com/mkailbowdy/internal/assert"
)

func TestPing(t *testing.T) {
	app := newTestApplication(t)

	ts := newTestServer(t, app.routes())
	defer ts.Close()

	res := ts.get(t, "/ping")
	assert.Equal(t, res.status, http.StatusOK)
	assert.Equal(t, res.body, "OK")
}
func TestUserSignup(t *testing.T) {
	// Create the application struct containing our mocked dependencies and
	// establish a new test server.
	app := newTestApplication(t)
	ts := newTestServer(t, app.routes())
	defer ts.Close()

	// Make a GET /user/signup request.
	res := ts.get(t, "/user/signup")

	// Extract the CSRF token from the response body and print it out using the
	// t.Logf() function. This works in exactly the same way as fmt.Printf(),
	// but writes the provided message to the test output.
	t.Logf("CSRF token is: %q", extractCSRFToken(t, res.body))

	// And also log the response cookies in the test output too.
	t.Logf("cookies are: %v", res.cookies)
}

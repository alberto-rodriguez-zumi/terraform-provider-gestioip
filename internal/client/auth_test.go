package client

import (
	"context"
	"encoding/base64"
	"fmt"
	"io"
	"net/http"
	"strings"
	"sync"
	"testing"
)

type roundTripFunc func(*http.Request) (*http.Response, error)

func (f roundTripFunc) RoundTrip(req *http.Request) (*http.Response, error) {
	return f(req)
}

func TestDoFormRequestToEndpointSendsBasicAuthToOfficialAPI(t *testing.T) {
	t.Parallel()

	httpClient := &http.Client{
		Transport: roundTripFunc(func(req *http.Request) (*http.Response, error) {
			assertBasicAuthHeader(t, req, "user", "pass")
			return jsonResponse(req, http.StatusOK, `{"ok":true}`), nil
		}),
	}

	c, err := New(Config{
		BaseURL:    "http://gestioip.example.test",
		Username:   "user",
		Password:   "pass",
		HTTPClient: httpClient,
	})
	if err != nil {
		t.Fatalf("create client: %v", err)
	}

	if _, err := c.doFormRequestToEndpoint(context.Background(), "http://gestioip.example.test/gestioip/api/api.cgi", nil); err != nil {
		t.Fatalf("do form request: %v", err)
	}
}

func TestEnsureInternalSessionSendsBasicAuth(t *testing.T) {
	t.Parallel()

	httpClient := &http.Client{
		Transport: roundTripFunc(func(req *http.Request) (*http.Response, error) {
			assertBasicAuthHeader(t, req, "user", "pass")
			return textResponse(req, http.StatusOK, ""), nil
		}),
	}

	c, err := New(Config{
		BaseURL:    "http://gestioip.example.test",
		Username:   "user",
		Password:   "pass",
		HTTPClient: httpClient,
	})
	if err != nil {
		t.Fatalf("create client: %v", err)
	}

	if err := c.ensureInternalSession(context.Background(), "http://gestioip.example.test/gestioip/intapi.cgi"); err != nil {
		t.Fatalf("ensure internal session: %v", err)
	}
}

func TestEnsureInternalSessionBypassesUnsupportedSessionLogin(t *testing.T) {
	t.Parallel()

	httpClient := &http.Client{
		Transport: roundTripFunc(func(req *http.Request) (*http.Response, error) {
			assertBasicAuthHeader(t, req, "user", "pass")
			return textResponse(req, http.StatusUnauthorized, "unauthorized"), nil
		}),
	}

	c, err := New(Config{
		BaseURL:    "http://gestioip.example.test",
		Username:   "user",
		Password:   "pass",
		HTTPClient: httpClient,
	})
	if err != nil {
		t.Fatalf("create client: %v", err)
	}

	if err := c.ensureInternalSession(context.Background(), "http://gestioip.example.test/gestioip/intapi.cgi"); err != nil {
		t.Fatalf("ensure internal session: %v", err)
	}

	if c.internalSessionMode != internalSessionBypassed {
		t.Fatalf("expected internal session mode %v, got %v", internalSessionBypassed, c.internalSessionMode)
	}
}

func TestPostFrontendFormSendsBasicAuth(t *testing.T) {
	t.Parallel()

	requests := 0
	httpClient := &http.Client{
		Transport: roundTripFunc(func(req *http.Request) (*http.Response, error) {
			assertBasicAuthHeader(t, req, "user", "pass")
			requests++
			if strings.HasSuffix(req.URL.Path, "/gestioip/intapi.cgi") {
				return textResponse(req, http.StatusOK, ""), nil
			}

			return textResponse(req, http.StatusOK, "<html><body>ok</body></html>"), nil
		}),
	}

	c, err := New(Config{
		BaseURL:    "http://gestioip.example.test",
		Username:   "user",
		Password:   "pass",
		HTTPClient: httpClient,
	})
	if err != nil {
		t.Fatalf("create client: %v", err)
	}

	if _, err := c.postFrontendForm(context.Background(), "/gestioip/ip_show_clients.cgi", nil); err != nil {
		t.Fatalf("post frontend form: %v", err)
	}

	if requests != 2 {
		t.Fatalf("expected 2 authenticated requests, got %d", requests)
	}
}

func TestPostFrontendFormContinuesWhenSessionLoginIsBypassed(t *testing.T) {
	t.Parallel()

	requests := 0
	httpClient := &http.Client{
		Transport: roundTripFunc(func(req *http.Request) (*http.Response, error) {
			assertBasicAuthHeader(t, req, "user", "pass")
			requests++
			if strings.HasSuffix(req.URL.Path, "/gestioip/intapi.cgi") {
				return textResponse(req, http.StatusUnauthorized, "unauthorized"), nil
			}

			return textResponse(req, http.StatusOK, "<html><body>ok</body></html>"), nil
		}),
	}

	c, err := New(Config{
		BaseURL:    "http://gestioip.example.test",
		Username:   "user",
		Password:   "pass",
		HTTPClient: httpClient,
	})
	if err != nil {
		t.Fatalf("create client: %v", err)
	}

	if _, err := c.postFrontendForm(context.Background(), "/gestioip/ip_show_clients.cgi", nil); err != nil {
		t.Fatalf("post frontend form: %v", err)
	}

	if requests != 2 {
		t.Fatalf("expected 2 authenticated requests, got %d", requests)
	}

	if c.internalSessionMode != internalSessionBypassed {
		t.Fatalf("expected internal session mode %v, got %v", internalSessionBypassed, c.internalSessionMode)
	}
}

func TestPingSendsBasicAuth(t *testing.T) {
	t.Parallel()

	httpClient := &http.Client{
		Transport: roundTripFunc(func(req *http.Request) (*http.Response, error) {
			assertBasicAuthHeader(t, req, "user", "pass")
			return textResponse(req, http.StatusOK, ""), nil
		}),
	}

	c, err := New(Config{
		BaseURL:    "http://gestioip.example.test",
		Username:   "user",
		Password:   "pass",
		HTTPClient: httpClient,
	})
	if err != nil {
		t.Fatalf("create client: %v", err)
	}

	c.apiURL = "http://gestioip.example.test/gestioip/api/api.cgi"

	if err := c.Ping(context.Background()); err != nil {
		t.Fatalf("ping: %v", err)
	}
}

func TestResolveClientIDIsSafeForConcurrentReads(t *testing.T) {
	t.Parallel()

	page := `<tr><td>DEFAULT</td><td></td><td><input name="id" type="hidden" value="1"></td></tr>`
	httpClient := &http.Client{
		Transport: roundTripFunc(func(req *http.Request) (*http.Response, error) {
			if strings.HasSuffix(req.URL.Path, "/gestioip/intapi.cgi") {
				return textResponse(req, http.StatusUnauthorized, "unauthorized"), nil
			}

			return textResponse(req, http.StatusOK, page), nil
		}),
	}

	c, err := New(Config{
		BaseURL:    "http://gestioip.example.test",
		Username:   "user",
		Password:   "pass",
		HTTPClient: httpClient,
	})
	if err != nil {
		t.Fatalf("create client: %v", err)
	}

	var wg sync.WaitGroup
	errs := make(chan error, 16)

	for i := 0; i < 16; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			clientID, err := c.ResolveClientID(context.Background(), "DEFAULT")
			if err != nil {
				errs <- err
				return
			}
			if clientID != "1" {
				errs <- fmt.Errorf("expected client id 1, got %s", clientID)
			}
		}()
	}

	wg.Wait()
	close(errs)

	for err := range errs {
		if err != nil {
			t.Fatal(err)
		}
	}
}

func assertBasicAuthHeader(t *testing.T, req *http.Request, username, password string) {
	t.Helper()

	got := req.Header.Get("Authorization")
	want := "Basic " + base64.StdEncoding.EncodeToString([]byte(fmt.Sprintf("%s:%s", username, password)))
	if got != want {
		t.Fatalf("expected Authorization header %q, got %q", want, got)
	}
}

func textResponse(req *http.Request, status int, body string) *http.Response {
	return &http.Response{
		StatusCode: status,
		Header:     make(http.Header),
		Body:       io.NopCloser(strings.NewReader(body)),
		Request:    req,
	}
}

func jsonResponse(req *http.Request, status int, body string) *http.Response {
	resp := textResponse(req, status, body)
	resp.Header.Set("Content-Type", "application/json")
	return resp
}

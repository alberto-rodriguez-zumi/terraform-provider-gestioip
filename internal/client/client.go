package client

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/http/cookiejar"
	"net/url"
	"strings"
	"sync"
	"time"
)

type Config struct {
	BaseURL    string
	ClientName string
	Username   string
	Password   string
	HTTPClient *http.Client
}

type Client struct {
	mu                   sync.RWMutex
	baseURL              string
	rootURL              string
	apiURL               string
	apiCandidates        []string
	clientName           string
	clientIDs            map[string]string
	username             string
	password             string
	httpClient           *http.Client
	internalSessionReady bool
	internalSessionMode  internalSessionMode
}

type internalSessionMode int

const (
	internalSessionUnknown internalSessionMode = iota
	internalSessionEstablished
	internalSessionBypassed
)

type apiErrorResponse struct {
	Error string `json:"error"`
}

type APIError struct {
	Message string
}

func (e *APIError) Error() string {
	return e.Message
}

func IsNotFoundError(err error) bool {
	var apiErr *APIError
	if !errors.As(err, &apiErr) {
		return false
	}

	return strings.Contains(strings.ToLower(apiErr.Message), "not found")
}

func New(config Config) (*Client, error) {
	baseURL := strings.TrimRight(config.BaseURL, "/")
	if baseURL == "" {
		return nil, fmt.Errorf("base URL is required")
	}

	httpClient := config.HTTPClient
	if httpClient == nil {
		jar, err := cookiejar.New(nil)
		if err != nil {
			return nil, fmt.Errorf("create cookie jar: %w", err)
		}

		httpClient = &http.Client{
			Timeout: 30 * time.Second,
			Jar:     jar,
		}
	}

	apiCandidates := buildAPIURLs(baseURL)

	return &Client{
		baseURL:       baseURL,
		rootURL:       deriveRootURL(baseURL),
		apiURL:        "",
		apiCandidates: apiCandidates,
		clientName:    config.ClientName,
		clientIDs:     map[string]string{},
		username:      config.Username,
		password:      config.Password,
		httpClient:    httpClient,
	}, nil
}

func (c *Client) BaseURL() string {
	return c.baseURL
}

func (c *Client) APIURL() string {
	c.mu.RLock()
	apiURL := c.apiURL
	c.mu.RUnlock()

	if apiURL != "" {
		return apiURL
	}

	if len(c.apiCandidates) > 0 {
		return c.apiCandidates[0]
	}

	return c.apiURL
}

func (c *Client) RootURL() string {
	return c.rootURL
}

func (c *Client) ClientName() string {
	return c.clientName
}

func (c *Client) Ping(ctx context.Context) error {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, c.APIURL(), nil)
	if err != nil {
		return fmt.Errorf("create request: %w", err)
	}

	c.applyBasicAuth(req)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("perform request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode >= http.StatusBadRequest {
		return fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	return nil
}

func (c *Client) UsesOfficialNetworkAPI(ctx context.Context, clientName string) (bool, error) {
	c.mu.RLock()
	apiURL := c.apiURL
	c.mu.RUnlock()

	if apiURL == "" {
		_, err := c.ListNetworks(ctx, clientName)
		if err != nil {
			return false, err
		}
	}

	c.mu.RLock()
	apiURL = c.apiURL
	c.mu.RUnlock()

	return !isInternalAPIEndpoint(apiURL), nil
}

func (c *Client) doFormRequest(ctx context.Context, values url.Values, target any) error {
	if values == nil {
		values = url.Values{}
	}

	if values.Get("output_type") == "" {
		values.Set("output_type", "json")
	}

	c.mu.RLock()
	apiURL := c.apiURL
	c.mu.RUnlock()

	candidates := c.apiCandidates
	if apiURL != "" {
		candidates = []string{apiURL}
	}

	var lastErr error

	for _, candidate := range candidates {
		body, err := c.doFormRequestToEndpoint(ctx, candidate, values)
		if err != nil {
			var endpointErr *endpointError
			if ok := errorAsEndpoint(err, &endpointErr); ok && endpointErr.Recoverable && apiURL == "" {
				lastErr = err
				continue
			}

			return err
		}

		c.mu.Lock()
		c.apiURL = candidate
		c.mu.Unlock()
		lastErr = nil

		var apiErr apiErrorResponse
		if err := json.Unmarshal(body, &apiErr); err == nil && strings.TrimSpace(apiErr.Error) != "" {
			return &APIError{Message: strings.TrimSpace(apiErr.Error)}
		}

		if target == nil {
			return nil
		}

		if err := json.Unmarshal(body, target); err != nil {
			return fmt.Errorf("decode response from %s: %w", candidate, err)
		}

		return nil
	}

	if lastErr != nil {
		return lastErr
	}

	return fmt.Errorf("unable to resolve GestioIP API endpoint from base URL %q", c.baseURL)
}

func (c *Client) doFormRequestToEndpoint(ctx context.Context, endpoint string, values url.Values) ([]byte, error) {
	if isInternalAPIEndpoint(endpoint) {
		if err := c.ensureInternalSession(ctx, endpoint); err != nil {
			return nil, err
		}
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, endpoint, strings.NewReader(values.Encode()))
	if err != nil {
		return nil, fmt.Errorf("create request: %w", err)
	}

	c.applyBasicAuth(req)

	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("perform request: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("read response body: %w", err)
	}

	if resp.StatusCode == http.StatusNotFound && c.apiURL == "" {
		return nil, &endpointError{
			Recoverable: true,
			Message:     fmt.Sprintf("GestioIP endpoint not found at %s", endpoint),
		}
	}

	if resp.StatusCode == http.StatusUnauthorized && c.apiURL == "" && !isInternalAPIEndpoint(endpoint) {
		return nil, &endpointError{
			Recoverable: true,
			Message:     fmt.Sprintf("GestioIP endpoint at %s requires a different authentication flow", endpoint),
		}
	}

	if resp.StatusCode >= http.StatusBadRequest {
		return nil, fmt.Errorf("unexpected status code from %s: %d", endpoint, resp.StatusCode)
	}

	if !isInternalAPIEndpoint(endpoint) && looksLikeHTMLLogin(body) && c.apiURL == "" {
		return nil, &endpointError{
			Recoverable: true,
			Message:     fmt.Sprintf("GestioIP endpoint at %s returned the login page instead of the API response", endpoint),
		}
	}

	return body, nil
}

func (c *Client) ensureInternalSession(ctx context.Context, endpoint string) error {
	c.mu.RLock()
	internalSessionReady := c.internalSessionReady
	c.mu.RUnlock()
	if internalSessionReady {
		return nil
	}

	values := url.Values{}
	values.Set("httpd_username", c.username)
	values.Set("httpd_password", c.password)
	values.Set("login", "login")
	values.Set("create_csrf_token", "true")

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, endpoint, strings.NewReader(values.Encode()))
	if err != nil {
		return fmt.Errorf("create internal API session request: %w", err)
	}

	c.applyBasicAuth(req)

	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("perform internal API session request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusUnauthorized || resp.StatusCode == http.StatusNotFound {
		c.mu.Lock()
		c.internalSessionReady = true
		c.internalSessionMode = internalSessionBypassed
		c.mu.Unlock()
		return nil
	}

	if resp.StatusCode >= http.StatusBadRequest {
		return fmt.Errorf("unexpected status code from %s during login: %d", endpoint, resp.StatusCode)
	}

	c.mu.Lock()
	c.internalSessionReady = true
	c.internalSessionMode = internalSessionEstablished
	c.mu.Unlock()

	return nil
}

func buildAPIURLs(baseURL string) []string {
	switch {
	case strings.HasSuffix(baseURL, "/gestioip/api/api.cgi"), strings.HasSuffix(baseURL, "/gestioip/intapi.cgi"), strings.HasSuffix(baseURL, "/api.cgi"):
		return []string{baseURL}
	default:
		return []string{
			baseURL + "/gestioip/api/api.cgi",
			baseURL + "/gestioip/intapi.cgi",
		}
	}
}

func deriveRootURL(baseURL string) string {
	switch {
	case strings.HasSuffix(baseURL, "/gestioip/api/api.cgi"):
		return strings.TrimSuffix(baseURL, "/gestioip/api/api.cgi")
	case strings.HasSuffix(baseURL, "/gestioip/intapi.cgi"):
		return strings.TrimSuffix(baseURL, "/gestioip/intapi.cgi")
	case strings.HasSuffix(baseURL, "/api.cgi"):
		return strings.TrimSuffix(baseURL, "/api.cgi")
	default:
		return baseURL
	}
}

func isInternalAPIEndpoint(endpoint string) bool {
	return strings.HasSuffix(endpoint, "/gestioip/intapi.cgi")
}

func looksLikeHTMLLogin(body []byte) bool {
	bodyText := string(body)
	return strings.Contains(bodyText, "<html") && strings.Contains(bodyText, "Sign In")
}

func (c *Client) applyBasicAuth(req *http.Request) {
	if req == nil {
		return
	}

	if c.username == "" && c.password == "" {
		return
	}

	req.SetBasicAuth(c.username, c.password)
}

type endpointError struct {
	Recoverable bool
	Message     string
}

func (e *endpointError) Error() string {
	return e.Message
}

func errorAsEndpoint(err error, target **endpointError) bool {
	endpointErr, ok := err.(*endpointError)
	if !ok {
		return false
	}

	*target = endpointErr
	return true
}

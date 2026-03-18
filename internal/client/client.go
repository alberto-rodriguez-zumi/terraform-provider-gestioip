package client

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
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
	baseURL    string
	apiURL     string
	clientName string
	username   string
	password   string
	httpClient *http.Client
}

type apiErrorResponse struct {
	Error string `json:"error"`
}

type APIError struct {
	Message string
}

func (e *APIError) Error() string {
	return e.Message
}

func New(config Config) (*Client, error) {
	baseURL := strings.TrimRight(config.BaseURL, "/")
	if baseURL == "" {
		return nil, fmt.Errorf("base URL is required")
	}

	httpClient := config.HTTPClient
	if httpClient == nil {
		httpClient = &http.Client{
			Timeout: 30 * time.Second,
		}
	}

	return &Client{
		baseURL:    baseURL,
		apiURL:     buildAPIURL(baseURL),
		clientName: config.ClientName,
		username:   config.Username,
		password:   config.Password,
		httpClient: httpClient,
	}, nil
}

func (c *Client) BaseURL() string {
	return c.baseURL
}

func (c *Client) APIURL() string {
	return c.apiURL
}

func (c *Client) ClientName() string {
	return c.clientName
}

func (c *Client) Ping(ctx context.Context) error {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, c.apiURL, nil)
	if err != nil {
		return fmt.Errorf("create request: %w", err)
	}

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

func (c *Client) doFormRequest(ctx context.Context, values url.Values, target any) error {
	if values == nil {
		values = url.Values{}
	}

	if values.Get("output_type") == "" {
		values.Set("output_type", "json")
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.apiURL, strings.NewReader(values.Encode()))
	if err != nil {
		return fmt.Errorf("create request: %w", err)
	}

	req.SetBasicAuth(c.username, c.password)
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("perform request: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("read response body: %w", err)
	}

	if resp.StatusCode >= http.StatusBadRequest {
		return fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	var apiErr apiErrorResponse
	if err := json.Unmarshal(body, &apiErr); err == nil && strings.TrimSpace(apiErr.Error) != "" {
		return &APIError{Message: strings.TrimSpace(apiErr.Error)}
	}

	if target == nil {
		return nil
	}

	if err := json.Unmarshal(body, target); err != nil {
		return fmt.Errorf("decode response: %w", err)
	}

	return nil
}

func buildAPIURL(baseURL string) string {
	if strings.HasSuffix(baseURL, "/gestioip/api/api.cgi") || strings.HasSuffix(baseURL, "/api.cgi") {
		return baseURL
	}

	return baseURL + "/gestioip/api/api.cgi"
}

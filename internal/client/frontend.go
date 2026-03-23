package client

import (
	"context"
	"fmt"
	"html"
	"io"
	"net/http"
	"net/url"
	"regexp"
	"strings"
)

var (
	clientRowPattern = regexp.MustCompile(`(?s)<tr[^>]*>\s*<td>([^<]+)</td>.*?name="id"\s+type="hidden"\s+value="(\d+)"`)
	errorPattern     = regexp.MustCompile(`(?is)<h3>\s*ERROR\s*</h3>\s*([^<]+)`)
)

func (c *Client) ResolveClientID(ctx context.Context, clientName string) (string, error) {
	if clientName == "" {
		return "", fmt.Errorf("client_name is required")
	}

	if isNumeric(clientName) {
		return clientName, nil
	}

	c.mu.RLock()
	clientID, ok := c.clientIDs[clientName]
	c.mu.RUnlock()
	if ok {
		return clientID, nil
	}

	body, err := c.postFrontendForm(ctx, "/gestioip/ip_show_clients.cgi", url.Values{})
	if err != nil {
		return "", err
	}

	matches := clientRowPattern.FindAllStringSubmatch(string(body), -1)
	for _, match := range matches {
		if len(match) < 3 {
			continue
		}

		name := strings.TrimSpace(html.UnescapeString(match[1]))
		id := strings.TrimSpace(match[2])
		if name == clientName {
			c.mu.Lock()
			c.clientIDs[clientName] = id
			c.mu.Unlock()
			return id, nil
		}
	}

	return "", fmt.Errorf("unable to resolve client_name %q to a GestioIP client_id", clientName)
}

func (c *Client) postFrontendForm(ctx context.Context, path string, values url.Values) ([]byte, error) {
	if err := c.ensureInternalSession(ctx, c.RootURL()+"/gestioip/intapi.cgi"); err != nil {
		return nil, err
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.RootURL()+path, strings.NewReader(values.Encode()))
	if err != nil {
		return nil, fmt.Errorf("create frontend request: %w", err)
	}

	c.applyBasicAuth(req)

	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("perform frontend request: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("read frontend response body: %w", err)
	}

	if resp.StatusCode >= http.StatusBadRequest {
		return nil, fmt.Errorf("unexpected frontend status code from %s: %d", path, resp.StatusCode)
	}

	if looksLikeHTMLLogin(body) {
		return nil, fmt.Errorf("frontend request to %s returned the login page", path)
	}

	return body, nil
}

func extractFrontendError(body []byte) string {
	matches := errorPattern.FindStringSubmatch(string(body))
	if len(matches) < 2 {
		return ""
	}

	return strings.TrimSpace(html.UnescapeString(matches[1]))
}

func isNumeric(value string) bool {
	for _, char := range value {
		if char < '0' || char > '9' {
			return false
		}
	}

	return value != ""
}

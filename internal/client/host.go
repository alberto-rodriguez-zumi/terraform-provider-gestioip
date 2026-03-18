package client

import (
	"context"
	"fmt"
	"html"
	"net/netip"
	"net/url"
	"regexp"
	"strconv"
	"strings"
)

type Host struct {
	ID          string
	IPInt       string
	NetworkID   string
	ClientName  string
	IP          string
	Hostname    string
	Description string
	Site        string
	Category    string
	Comment     string
	IPVersion   string
}

type CreateHostInput struct {
	ClientName  string
	IP          string
	Hostname    string
	Description string
	Site        string
	Category    string
	Comment     string
}

type UpdateHostInput struct {
	ID          string
	IPInt       string
	NetworkID   string
	ClientName  string
	IP          string
	Hostname    string
	Description string
	Site        string
	Category    string
	Comment     string
	IPVersion   string
}

type DeleteHostInput struct {
	IPInt      string
	NetworkID  string
	ClientName string
	IP         string
	IPVersion  string
}

var (
	hostRowPattern    = regexp.MustCompile(`(?s)<tr[^>]*class="show_detail_hosts"[^>]*>.*?</tr>`)
	hostCellPattern   = regexp.MustCompile(`(?s)<td[^>]*>(.*?)</td>`)
	selectPatternTmpl = `(?s)<select[^>]*name="%s"[^>]*>.*?<option[^>]*value="([^"]*)"[^>]*selected[^>]*>`
	textareaPattern   = regexp.MustCompile(`(?s)<textarea[^>]*name="comentario"[^>]*>(.*?)</textarea>`)
	tagPattern        = regexp.MustCompile(`(?s)<[^>]+>`)
	spacePattern      = regexp.MustCompile(`\s+`)
)

func (c *Client) ReadHost(ctx context.Context, clientName, ip string) (*Host, error) {
	network, err := c.findNetworkForIP(ctx, clientName, ip)
	if err != nil {
		return nil, err
	}

	host, err := c.findHostInNetwork(ctx, clientName, network.ID, firstNonEmpty(network.IPVersion, inferIPVersion(ip)), ip)
	if err != nil {
		return nil, err
	}

	host.ClientName = clientName
	host.NetworkID = network.ID
	host.IPVersion = firstNonEmpty(host.IPVersion, network.IPVersion, inferIPVersion(ip))

	return host, nil
}

func (c *Client) CreateHost(ctx context.Context, input CreateHostInput) (*Host, error) {
	network, err := c.findNetworkForIP(ctx, input.ClientName, input.IP)
	if err != nil {
		return nil, err
	}

	clientID, err := c.ResolveClientID(ctx, input.ClientName)
	if err != nil {
		return nil, err
	}

	values := url.Values{}
	values.Set("hostname", input.Hostname)
	values.Set("host_descr", input.Description)
	values.Set("loc", input.Site)
	values.Set("cat", input.Category)
	values.Set("int_admin", "n")
	values.Set("comentario", input.Comment)
	values.Set("update_type", "man")
	values.Set("client_id", clientID)
	values.Set("ip", input.IP)
	values.Set("red", network.IP)
	values.Set("BM", strconv.FormatInt(network.Bitmask, 10))
	values.Set("host_id", "")
	values.Set("host_order_by", "IP_auf")
	values.Set("ip_version", firstNonEmpty(network.IPVersion, inferIPVersion(input.IP)))
	values.Set("from_line", "")
	values.Set("search_index", "")
	values.Set("search_hostname", "")
	values.Set("match", "")
	values.Set("red_num", network.ID)

	body, err := c.postFrontendForm(ctx, "/gestioip/res/ip_modip.cgi", values)
	if err != nil {
		return nil, err
	}

	if frontendErr := extractFrontendError(body); frontendErr != "" {
		return nil, &APIError{Message: frontendErr}
	}

	return c.ReadHost(ctx, input.ClientName, input.IP)
}

func (c *Client) UpdateHost(ctx context.Context, input UpdateHostInput) (*Host, error) {
	clientID, err := c.ResolveClientID(ctx, input.ClientName)
	if err != nil {
		return nil, err
	}

	networkID := input.NetworkID
	ipVersion := firstNonEmpty(input.IPVersion, inferIPVersion(input.IP))
	networkIP := ""
	networkBitmask := int64(0)

	if networkID == "" {
		network, err := c.findNetworkForIP(ctx, input.ClientName, input.IP)
		if err != nil {
			return nil, err
		}

		networkID = network.ID
		networkIP = network.IP
		networkBitmask = network.Bitmask
		ipVersion = firstNonEmpty(ipVersion, network.IPVersion)
	} else {
		network, err := c.findNetworkByID(ctx, input.ClientName, networkID)
		if err != nil {
			return nil, err
		}

		networkIP = network.IP
		networkBitmask = network.Bitmask
		ipVersion = firstNonEmpty(ipVersion, network.IPVersion)
	}

	values := url.Values{}
	values.Set("hostname", input.Hostname)
	values.Set("host_descr", input.Description)
	values.Set("loc", input.Site)
	values.Set("cat", input.Category)
	values.Set("int_admin", "n")
	values.Set("comentario", input.Comment)
	values.Set("update_type", "man")
	values.Set("client_id", clientID)
	values.Set("ip", input.IP)
	values.Set("red", networkIP)
	values.Set("BM", strconv.FormatInt(networkBitmask, 10))
	values.Set("host_id", input.ID)
	values.Set("host_order_by", "IP_auf")
	values.Set("ip_version", ipVersion)
	values.Set("from_line", "")
	values.Set("search_index", "")
	values.Set("search_hostname", "")
	values.Set("match", "")
	values.Set("red_num", networkID)

	body, err := c.postFrontendForm(ctx, "/gestioip/res/ip_modip.cgi", values)
	if err != nil {
		return nil, err
	}

	if frontendErr := extractFrontendError(body); frontendErr != "" {
		return nil, &APIError{Message: frontendErr}
	}

	return c.ReadHost(ctx, input.ClientName, input.IP)
}

func (c *Client) DeleteHost(ctx context.Context, input DeleteHostInput) error {
	clientID, err := c.ResolveClientID(ctx, input.ClientName)
	if err != nil {
		return err
	}

	ipVersion := firstNonEmpty(input.IPVersion, inferIPVersion(input.IP))
	networkID := input.NetworkID
	ipInt := input.IPInt

	if networkID == "" || ipInt == "" {
		host, err := c.ReadHost(ctx, input.ClientName, input.IP)
		if err != nil {
			return err
		}

		if networkID == "" {
			networkID = host.NetworkID
		}
		if ipInt == "" {
			ipInt = host.IPInt
		}
		ipVersion = firstNonEmpty(ipVersion, host.IPVersion)
	}

	values := url.Values{}
	values.Set("ip_int", ipInt)
	values.Set("red_num", networkID)
	values.Set("entries_per_page_hosts", "254")
	values.Set("start_entry_hosts", "0")
	values.Set("host_order_by", "IP_auf")
	values.Set("knownhosts", "hosts")
	values.Set("anz_values_hosts", "1")
	values.Set("client_id", clientID)
	values.Set("ip_version", ipVersion)

	body, err := c.postFrontendForm(ctx, "/gestioip/res/ip_deleteip.cgi", values)
	if err != nil {
		return err
	}

	if frontendErr := extractFrontendError(body); frontendErr != "" {
		return &APIError{Message: frontendErr}
	}

	_, err = c.ReadHost(ctx, input.ClientName, input.IP)
	if err == nil {
		return &APIError{Message: fmt.Sprintf("host %s still exists after delete", input.IP)}
	}

	var apiErr *APIError
	if errorsAsAPI(err, &apiErr) && strings.Contains(apiErr.Message, "not found") {
		return nil
	}

	return err
}

func (c *Client) findHostInNetwork(ctx context.Context, clientName, networkID, ipVersion, ip string) (*Host, error) {
	clientID, err := c.ResolveClientID(ctx, clientName)
	if err != nil {
		return nil, err
	}

	values := url.Values{}
	values.Set("client_id", clientID)
	values.Set("red_num", networkID)
	values.Set("ip_version", ipVersion)
	values.Set("knownhosts", "hosts")

	body, err := c.postFrontendForm(ctx, "/gestioip/ip_show.cgi", values)
	if err != nil {
		return nil, err
	}

	hosts, err := parseHosts(body, clientName)
	if err != nil {
		return nil, err
	}

	for _, host := range hosts {
		if host.IP == ip {
			host.NetworkID = networkID
			host.IPVersion = firstNonEmpty(host.IPVersion, ipVersion)
			return &host, nil
		}
	}

	return nil, &APIError{Message: fmt.Sprintf("host %s not found", ip)}
}

func (c *Client) findNetworkForIP(ctx context.Context, clientName, ip string) (*Network, error) {
	networks, err := c.ListNetworks(ctx, clientName)
	if err != nil {
		return nil, err
	}

	ipAddr, err := netip.ParseAddr(ip)
	if err != nil {
		return nil, fmt.Errorf("parse host ip %q: %w", ip, err)
	}

	var best *Network

	for _, network := range networks {
		networkAddr, err := netip.ParseAddr(network.IP)
		if err != nil {
			continue
		}

		prefix := netip.PrefixFrom(networkAddr, int(network.Bitmask)).Masked()
		if !prefix.Contains(ipAddr) {
			continue
		}

		networkCopy := network
		if best == nil || networkCopy.Bitmask > best.Bitmask {
			best = &networkCopy
		}
	}

	if best == nil {
		return nil, &APIError{Message: fmt.Sprintf("no network contains host ip %s", ip)}
	}

	return best, nil
}

func (c *Client) findNetworkByID(ctx context.Context, clientName, networkID string) (*Network, error) {
	networks, err := c.ListNetworks(ctx, clientName)
	if err != nil {
		return nil, err
	}

	for _, network := range networks {
		if network.ID == networkID {
			networkCopy := network
			return &networkCopy, nil
		}
	}

	return nil, &APIError{Message: fmt.Sprintf("network %s not found", networkID)}
}

func parseHosts(body []byte, clientName string) ([]Host, error) {
	rows := hostRowPattern.FindAllString(string(body), -1)
	hosts := make([]Host, 0, len(rows))

	for _, row := range rows {
		cells := hostCellPattern.FindAllStringSubmatch(row, -1)
		if len(cells) < 9 {
			continue
		}

		ip := cleanCellText(cells[2][1])
		if ip == "" {
			continue
		}

		host := Host{
			ID:          extractHiddenValue(row, "host_id"),
			IPInt:       extractFirstNonEmpty(row, "ip_int", "ip"),
			ClientName:  clientName,
			IP:          ip,
			Hostname:    cleanCellText(cells[3][1]),
			Description: cleanCellText(cells[4][1]),
			Site:        cleanCellText(cells[5][1]),
			Category:    cleanCellText(cells[6][1]),
			Comment:     cleanCellText(cells[8][1]),
			IPVersion:   extractHiddenValue(row, "ip_version"),
			NetworkID:   extractHiddenValue(row, "red_num"),
		}

		if host.IPInt != "" && strings.Contains(host.IPInt, ".") {
			host.IPInt = ""
		}

		hosts = append(hosts, host)
	}

	return hosts, nil
}

func cleanCellText(value string) string {
	text := html.UnescapeString(tagPattern.ReplaceAllString(value, " "))
	text = strings.ReplaceAll(text, "\u00a0", " ")
	text = spacePattern.ReplaceAllString(text, " ")
	return strings.TrimSpace(text)
}

func extractHiddenValue(snippet, name string) string {
	pattern := regexp.MustCompile(fmt.Sprintf(`name="%s"\s+type="hidden"\s+value="([^"]*)"`, regexp.QuoteMeta(name)))
	matches := pattern.FindStringSubmatch(snippet)
	if len(matches) < 2 {
		return ""
	}

	return html.UnescapeString(strings.TrimSpace(matches[1]))
}

func extractFirstNonEmpty(snippet string, names ...string) string {
	for _, name := range names {
		value := extractHiddenValue(snippet, name)
		if value != "" {
			return value
		}
	}

	return ""
}

func errorsAsAPI(err error, target **APIError) bool {
	apiErr, ok := err.(*APIError)
	if !ok {
		return false
	}

	*target = apiErr
	return true
}

package client

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/url"
	"regexp"
	"strconv"
	"strings"
)

type Network struct {
	ID          string
	IP          string
	Bitmask     int64
	Description string
	Site        string
	Category    string
	Comment     string
	Sync        bool
	IPVersion   string
	ClientName  string
}

type CreateNetworkInput struct {
	ClientName  string
	IP          string
	Bitmask     int64
	Description string
	Site        string
	Category    string
	Comment     string
	Sync        bool
}

type UpdateNetworkInput struct {
	ID          string
	ClientName  string
	IP          string
	Bitmask     int64
	Description string
	Site        string
	Category    string
	Comment     string
	Sync        bool
	IPVersion   string
}

type DeleteNetworkInput struct {
	ID         string
	ClientName string
	IP         string
	Bitmask    int64
	IPVersion  string
}

type readNetworkResponse struct {
	Error   string         `json:"error"`
	Network networkPayload `json:"Network"`
}

type createNetworkResponse struct {
	Error   string         `json:"error"`
	Network networkPayload `json:"Network"`
}

type updateNetworkResponse struct {
	Error   string         `json:"error"`
	Network networkPayload `json:"Network"`
}

type listNetworksResponse struct {
	ListNetworksResult struct {
		Error       string          `json:"error"`
		NetworkList json.RawMessage `json:"NetworkList"`
	} `json:"listNetworksResult"`
}

type listNetworksEnvelope struct {
	Network json.RawMessage `json:"Network"`
}

type networkPayload struct {
	ID          string `json:"id"`
	IP          string `json:"IP"`
	Bitmask     string `json:"BM"`
	Description string `json:"descr"`
	Site        string `json:"site"`
	Category    string `json:"cat"`
	Comment     string `json:"comment"`
	Sync        string `json:"sync"`
	IPVersion   string `json:"ip_version"`
}

var networkRowPattern = regexp.MustCompile(`(?s)<tr[^>]*class="show_detail"[^>]*>.*?</tr>`)

func (c *Client) ReadNetwork(ctx context.Context, clientName, ip string, bitmask int64) (*Network, error) {
	if supports, err := c.SupportsNetworkCRUD(ctx, clientName); err == nil && !supports {
		networks, err := c.ListNetworks(ctx, clientName)
		if err != nil {
			return nil, err
		}

		for _, network := range networks {
			if network.IP == ip && network.Bitmask == bitmask {
				return &network, nil
			}
		}

		return nil, &APIError{
			Message: fmt.Sprintf("network %s/%d not found", ip, bitmask),
		}
	}

	values := url.Values{}
	values.Set("request_type", "readNetwork")
	values.Set("client_name", clientName)
	values.Set("ip", ip)
	values.Set("BM", strconv.FormatInt(bitmask, 10))

	var response readNetworkResponse
	if err := c.doFormRequest(ctx, values, &response); err != nil {
		return nil, err
	}

	return response.Network.toNetwork(clientName)
}

func (c *Client) ListNetworks(ctx context.Context, clientName string) ([]Network, error) {
	requestClient := clientName
	if isInternalAPIEndpoint(c.APIURL()) || c.apiURL == "" {
		clientID, err := c.ResolveClientID(ctx, clientName)
		if err == nil {
			requestClient = clientID
		}
	}

	values := url.Values{}
	values.Set("request_type", "listNetworks")
	values.Set("client_name", requestClient)
	values.Set("include_id", "yes")
	values.Set("no_csv", "yes")

	var response listNetworksResponse
	if err := c.doFormRequest(ctx, values, &response); err != nil {
		networks, frontendErr := c.listNetworksViaFrontend(ctx, clientName)
		if frontendErr == nil {
			return networks, nil
		}

		return nil, err
	}

	networks, err := decodeNetworkList(response.ListNetworksResult.NetworkList, clientName)
	if err != nil {
		frontendNetworks, frontendErr := c.listNetworksViaFrontend(ctx, clientName)
		if frontendErr == nil {
			return frontendNetworks, nil
		}

		return nil, err
	}

	if len(networks) == 0 && requestClient == clientName && !isNumeric(clientName) {
		clientID, err := c.ResolveClientID(ctx, clientName)
		if err == nil && clientID != clientName {
			values.Set("client_name", clientID)
			if err := c.doFormRequest(ctx, values, &response); err != nil {
				return nil, err
			}

			return decodeNetworkList(response.ListNetworksResult.NetworkList, clientName)
		}
	}

	return networks, nil
}

func (c *Client) listNetworksViaFrontend(ctx context.Context, clientName string) ([]Network, error) {
	clientID, err := c.ResolveClientID(ctx, clientName)
	if err != nil {
		return nil, err
	}

	ipVersions := []string{"v4", "v6"}
	networks := make([]Network, 0)
	seen := map[string]struct{}{}

	for _, ipVersion := range ipVersions {
		body, err := c.fetchNetworkPage(ctx, clientID, ipVersion)
		if err != nil {
			if ipVersion == "v6" {
				continue
			}
			return nil, err
		}

		parsed, err := parseNetworksFromFrontend(body, clientName)
		if err != nil {
			return nil, err
		}

		for _, network := range parsed {
			key := firstNonEmpty(network.ID, network.IP+"/"+strconv.FormatInt(network.Bitmask, 10))
			if _, ok := seen[key]; ok {
				continue
			}

			seen[key] = struct{}{}
			networks = append(networks, network)
		}
	}

	return networks, nil
}

func (c *Client) fetchNetworkPage(ctx context.Context, clientID, ipVersion string) ([]byte, error) {
	values := url.Values{}
	values.Set("client_id", clientID)
	values.Set("knownhosts", "all")
	values.Set("entries_per_page", "5000")
	values.Set("start_entry", "0")
	values.Set("order_by", "red_ab")
	values.Set("tipo_ele", "NULL")
	values.Set("loc_ele", "NULL")
	values.Set("ip_version_ele", firstNonEmpty(ipVersion, "v4"))

	return c.postFrontendForm(ctx, "/gestioip/index.cgi", values)
}

func (c *Client) SupportsNetworkCRUD(ctx context.Context, clientName string) (bool, error) {
	return c.UsesOfficialNetworkAPI(ctx, clientName)
}

func (c *Client) CreateNetwork(ctx context.Context, input CreateNetworkInput) (*Network, error) {
	usesOfficialAPI, err := c.UsesOfficialNetworkAPI(ctx, input.ClientName)
	if err != nil {
		return nil, err
	}

	if !usesOfficialAPI {
		return c.createNetworkViaFrontend(ctx, input)
	}

	values := url.Values{}
	values.Set("request_type", "createNetwork")
	values.Set("client_name", input.ClientName)
	values.Set("new_ip", input.IP)
	values.Set("new_BM", strconv.FormatInt(input.Bitmask, 10))
	values.Set("new_descr", input.Description)
	values.Set("new_site", input.Site)
	values.Set("new_cat", input.Category)
	values.Set("new_comment", input.Comment)
	values.Set("new_sync", boolToYN(input.Sync))

	var response createNetworkResponse
	if err := c.doFormRequest(ctx, values, &response); err != nil {
		return nil, err
	}

	return response.Network.toNetwork(input.ClientName)
}

func (c *Client) UpdateNetwork(ctx context.Context, input UpdateNetworkInput) (*Network, error) {
	usesOfficialAPI, err := c.UsesOfficialNetworkAPI(ctx, input.ClientName)
	if err != nil {
		return nil, err
	}

	if !usesOfficialAPI {
		return c.updateNetworkViaFrontend(ctx, input)
	}

	values := url.Values{}
	values.Set("request_type", "updateNetwork")
	values.Set("client_name", input.ClientName)
	values.Set("ip", input.IP)
	values.Set("BM", strconv.FormatInt(input.Bitmask, 10))
	values.Set("new_descr", input.Description)
	values.Set("new_site", input.Site)
	values.Set("new_cat", input.Category)
	values.Set("new_comment", input.Comment)
	values.Set("new_sync", boolToYN(input.Sync))

	var response updateNetworkResponse
	if err := c.doFormRequest(ctx, values, &response); err != nil {
		return nil, err
	}

	return response.Network.toNetwork(input.ClientName)
}

func (c *Client) DeleteNetwork(ctx context.Context, input DeleteNetworkInput) error {
	usesOfficialAPI, err := c.UsesOfficialNetworkAPI(ctx, input.ClientName)
	if err != nil {
		return err
	}

	if !usesOfficialAPI {
		return c.deleteNetworkViaFrontend(ctx, input)
	}

	values := url.Values{}
	values.Set("request_type", "deleteNetwork")
	values.Set("client_name", input.ClientName)
	values.Set("ip", input.IP)
	values.Set("BM", strconv.FormatInt(input.Bitmask, 10))

	return c.doFormRequest(ctx, values, nil)
}

func (c *Client) createNetworkViaFrontend(ctx context.Context, input CreateNetworkInput) (*Network, error) {
	clientID, err := c.ResolveClientID(ctx, input.ClientName)
	if err != nil {
		return nil, err
	}

	values := url.Values{}
	values.Set("client_id", clientID)
	values.Set("ip_version", inferIPVersion(input.IP))
	values.Set("red", input.IP)
	values.Set("BM", strconv.FormatInt(input.Bitmask, 10))
	values.Set("descr", input.Description)
	values.Set("loc", input.Site)
	values.Set("cat_red", input.Category)
	values.Set("comentario", input.Comment)
	values.Set("vigilada", boolToYN(input.Sync))

	body, err := c.postFrontendForm(ctx, "/gestioip/res/ip_insertred.cgi", values)
	if err != nil {
		return nil, err
	}

	if frontendErr := extractFrontendError(body); frontendErr != "" {
		return nil, &APIError{Message: frontendErr}
	}

	return c.ReadNetwork(ctx, input.ClientName, input.IP, input.Bitmask)
}

func (c *Client) updateNetworkViaFrontend(ctx context.Context, input UpdateNetworkInput) (*Network, error) {
	clientID, err := c.ResolveClientID(ctx, input.ClientName)
	if err != nil {
		return nil, err
	}

	values := url.Values{}
	values.Set("client_id", clientID)
	values.Set("red_num", input.ID)
	values.Set("red", input.IP)
	values.Set("BM", strconv.FormatInt(input.Bitmask, 10))
	values.Set("BM_new", strconv.FormatInt(input.Bitmask, 10))
	values.Set("descr", input.Description)
	values.Set("loc", input.Site)
	values.Set("cat_net", input.Category)
	values.Set("comentario", input.Comment)
	values.Set("vigilada", boolToYN(input.Sync))
	values.Set("referer", "red_view")
	values.Set("ip_version_ele", firstNonEmpty(input.IPVersion, inferIPVersion(input.IP)))
	values.Set("dyn_dns_updates", "1")

	body, err := c.postFrontendForm(ctx, "/gestioip/res/ip_modred.cgi", values)
	if err != nil {
		return nil, err
	}

	if frontendErr := extractFrontendError(body); frontendErr != "" {
		return nil, &APIError{Message: frontendErr}
	}

	return c.ReadNetwork(ctx, input.ClientName, input.IP, input.Bitmask)
}

func (c *Client) deleteNetworkViaFrontend(ctx context.Context, input DeleteNetworkInput) error {
	clientID, err := c.ResolveClientID(ctx, input.ClientName)
	if err != nil {
		return err
	}

	values := url.Values{}
	values.Set("client_id", clientID)
	values.Set("red_num", input.ID)
	values.Set("referer", "red_view")
	values.Set("ip_version_ele", firstNonEmpty(input.IPVersion, inferIPVersion(input.IP)))

	body, err := c.postFrontendForm(ctx, "/gestioip/res/ip_deletered.cgi", values)
	if err != nil {
		return err
	}

	if frontendErr := extractFrontendError(body); frontendErr != "" {
		return &APIError{Message: frontendErr}
	}

	_, err = c.ReadNetwork(ctx, input.ClientName, input.IP, input.Bitmask)
	if err == nil {
		return &APIError{Message: fmt.Sprintf("network %s/%d still exists after delete", input.IP, input.Bitmask)}
	}

	var apiErr *APIError
	if errors.As(err, &apiErr) && strings.Contains(apiErr.Message, "not found") {
		return nil
	}

	return err
}

func (p networkPayload) toNetwork(clientName string) (*Network, error) {
	bitmask, err := strconv.ParseInt(p.Bitmask, 10, 64)
	if err != nil {
		return nil, fmt.Errorf("parse network bitmask %q: %w", p.Bitmask, err)
	}

	return &Network{
		ID:          p.ID,
		IP:          p.IP,
		Bitmask:     bitmask,
		Description: p.Description,
		Site:        p.Site,
		Category:    p.Category,
		Comment:     p.Comment,
		Sync:        ynToBool(p.Sync),
		IPVersion:   p.IPVersion,
		ClientName:  clientName,
	}, nil
}

func boolToYN(value bool) string {
	if value {
		return "y"
	}

	return "n"
}

func ynToBool(value string) bool {
	return value == "y" || value == "Y"
}

func inferIPVersion(ip string) string {
	if strings.Contains(ip, ":") {
		return "v6"
	}

	return "v4"
}

func firstNonEmpty(values ...string) string {
	for _, value := range values {
		if value != "" {
			return value
		}
	}

	return ""
}

func decodeNetworkList(raw json.RawMessage, clientName string) ([]Network, error) {
	trimmed := strings.TrimSpace(string(raw))
	if trimmed == "" || trimmed == `""` || trimmed == "null" {
		return nil, nil
	}

	var envelope listNetworksEnvelope
	if err := json.Unmarshal(raw, &envelope); err != nil {
		return nil, fmt.Errorf("decode network list envelope: %w", err)
	}

	trimmedNetworks := strings.TrimSpace(string(envelope.Network))
	if trimmedNetworks == "" || trimmedNetworks == `""` || trimmedNetworks == "null" {
		return nil, nil
	}

	var payloads []networkPayload
	if err := json.Unmarshal(envelope.Network, &payloads); err != nil {
		var single networkPayload
		if singleErr := json.Unmarshal(envelope.Network, &single); singleErr != nil {
			return nil, fmt.Errorf("decode network list: %w", err)
		}

		payloads = []networkPayload{single}
	}

	networks := make([]Network, 0, len(payloads))
	for _, payload := range payloads {
		network, err := payload.toNetwork(clientName)
		if err != nil {
			return nil, err
		}

		networks = append(networks, *network)
	}

	return networks, nil
}

func parseNetworksFromFrontend(body []byte, clientName string) ([]Network, error) {
	rows := networkRowPattern.FindAllString(string(body), -1)
	networks := make([]Network, 0, len(rows))

	for _, row := range rows {
		if !strings.Contains(row, `name="red_num"`) {
			continue
		}

		cells := hostCellPattern.FindAllStringSubmatch(row, -1)
		if len(cells) < 7 {
			continue
		}

		ip := cleanCellText(cells[0][1])
		if ip == "" {
			continue
		}

		bitmask, err := strconv.ParseInt(cleanCellText(cells[1][1]), 10, 64)
		if err != nil {
			continue
		}

		networks = append(networks, Network{
			ID:          extractHiddenValue(row, "red_num"),
			IP:          ip,
			Bitmask:     bitmask,
			Description: cleanCellText(cells[2][1]),
			Site:        cleanCellText(cells[3][1]),
			Category:    cleanCellText(cells[4][1]),
			Comment:     cleanCellText(cells[5][1]),
			Sync:        cleanCellText(cells[6][1]) != "",
			IPVersion:   firstNonEmpty(extractHiddenValue(row, "ip_version"), inferIPVersion(ip)),
			ClientName:  clientName,
		})
	}

	return networks, nil
}

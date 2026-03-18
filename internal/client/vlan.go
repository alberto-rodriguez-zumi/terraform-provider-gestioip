package client

import (
	"context"
	"fmt"
	"html"
	"net/url"
	"regexp"
	"strings"
)

type VLAN struct {
	ID          string
	ClientName  string
	Number      string
	Name        string
	Description string
	BGColor     string
	FontColor   string
}

type CreateVLANInput struct {
	ClientName  string
	Number      string
	Name        string
	Description string
	BGColor     string
	FontColor   string
}

type UpdateVLANInput struct {
	ID          string
	ClientName  string
	Number      string
	Name        string
	Description string
	BGColor     string
	FontColor   string
}

type DeleteVLANInput struct {
	ID          string
	ClientName  string
	Number      string
	Name        string
	Description string
	BGColor     string
}

var (
	vlanRowPattern   = regexp.MustCompile(`(?s)<tr[^>]*bgcolor="([^"]+)"[^>]*class="show_detail[^"]*"[^>]*>.*?</tr>`)
	fontColorPattern = regexp.MustCompile(`style="color:\s*([^;"]+)`)
)

func (c *Client) ReadVLAN(ctx context.Context, clientName, number string) (*VLAN, error) {
	clientID, err := c.ResolveClientID(ctx, clientName)
	if err != nil {
		return nil, err
	}

	body, err := c.fetchVLANPage(ctx, clientID)
	if err != nil {
		return nil, err
	}

	vlans, err := parseVLANs(body, clientName)
	if err != nil {
		return nil, err
	}

	for _, vlan := range vlans {
		if vlan.Number == number {
			return &vlan, nil
		}
	}

	return nil, &APIError{Message: fmt.Sprintf("vlan %s not found", number)}
}

func (c *Client) CreateVLAN(ctx context.Context, input CreateVLANInput) (*VLAN, error) {
	clientID, err := c.ResolveClientID(ctx, input.ClientName)
	if err != nil {
		return nil, err
	}

	values := url.Values{}
	values.Set("vlan_num", input.Number)
	values.Set("vlan_name", input.Name)
	values.Set("comment", input.Description)
	values.Set("bg_color", firstNonEmpty(input.BGColor, "blue"))
	values.Set("font_color", firstNonEmpty(input.FontColor, "white"))
	values.Set("client_id", clientID)

	body, err := c.postFrontendForm(ctx, "/gestioip/res/ip_insertvlan.cgi", values)
	if err != nil {
		return nil, err
	}

	if frontendErr := extractFrontendError(body); frontendErr != "" {
		return nil, &APIError{Message: frontendErr}
	}

	return c.ReadVLAN(ctx, input.ClientName, input.Number)
}

func (c *Client) UpdateVLAN(ctx context.Context, input UpdateVLANInput) (*VLAN, error) {
	clientID, err := c.ResolveClientID(ctx, input.ClientName)
	if err != nil {
		return nil, err
	}

	values := url.Values{}
	values.Set("vlan_num", input.Number)
	values.Set("vlan_name", input.Name)
	values.Set("comment", input.Description)
	values.Set("bg_color", firstNonEmpty(input.BGColor, "blue"))
	values.Set("font_color", firstNonEmpty(input.FontColor, "white"))
	values.Set("vlan_id", input.ID)
	values.Set("client_id", clientID)

	body, err := c.postFrontendForm(ctx, "/gestioip/res/ip_modvlan.cgi", values)
	if err != nil {
		return nil, err
	}

	if frontendErr := extractFrontendError(body); frontendErr != "" {
		return nil, &APIError{Message: frontendErr}
	}

	return c.ReadVLAN(ctx, input.ClientName, input.Number)
}

func (c *Client) DeleteVLAN(ctx context.Context, input DeleteVLANInput) error {
	clientID, err := c.ResolveClientID(ctx, input.ClientName)
	if err != nil {
		return err
	}

	values := url.Values{}
	values.Set("vlan_id", input.ID)
	values.Set("vlan_num", input.Number)
	values.Set("vlan_name", input.Name)
	values.Set("comment", input.Description)
	values.Set("bg_color", input.BGColor)
	values.Set("client_id", clientID)

	body, err := c.postFrontendForm(ctx, "/gestioip/res/ip_deletevlan.cgi", values)
	if err != nil {
		return err
	}

	if frontendErr := extractFrontendError(body); frontendErr != "" {
		return &APIError{Message: frontendErr}
	}

	_, err = c.ReadVLAN(ctx, input.ClientName, input.Number)
	if err == nil {
		return &APIError{Message: fmt.Sprintf("vlan %s still exists after delete", input.Number)}
	}

	var apiErr *APIError
	if errorsAsAPI(err, &apiErr) && strings.Contains(apiErr.Message, "not found") {
		return nil
	}

	return err
}

func (c *Client) fetchVLANPage(ctx context.Context, clientID string) ([]byte, error) {
	return c.postFrontendForm(ctx, "/gestioip/show_vlans.cgi?mode=show&client_id="+url.QueryEscape(clientID), url.Values{})
}

func parseVLANs(body []byte, clientName string) ([]VLAN, error) {
	rows := vlanRowPattern.FindAllStringSubmatch(string(body), -1)
	vlans := make([]VLAN, 0, len(rows))

	for _, rowMatch := range rows {
		if len(rowMatch) < 2 {
			continue
		}

		row := rowMatch[0]
		bgColor := strings.TrimSpace(rowMatch[1])
		cells := hostCellPattern.FindAllStringSubmatch(row, -1)
		if len(cells) < 6 {
			continue
		}

		number := cleanCellText(cells[1][1])
		if number == "" {
			continue
		}

		fontColor := ""
		fontMatch := fontColorPattern.FindStringSubmatch(row)
		if len(fontMatch) > 1 {
			fontColor = strings.TrimSpace(fontMatch[1])
		}

		vlans = append(vlans, VLAN{
			ID:          extractHiddenValue(row, "vlan_id"),
			ClientName:  clientName,
			Number:      number,
			Name:        cleanCellText(cells[2][1]),
			Description: cleanCellText(cells[3][1]),
			BGColor:     bgColor,
			FontColor:   html.UnescapeString(fontColor),
		})
	}

	return vlans, nil
}

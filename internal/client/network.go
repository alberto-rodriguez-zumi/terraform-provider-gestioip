package client

import (
	"context"
	"fmt"
	"net/url"
	"strconv"
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
	ClientName  string
	IP          string
	Bitmask     int64
	Description string
	Site        string
	Category    string
	Comment     string
	Sync        bool
}

type DeleteNetworkInput struct {
	ClientName string
	IP         string
	Bitmask    int64
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

func (c *Client) ReadNetwork(ctx context.Context, clientName, ip string, bitmask int64) (*Network, error) {
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

func (c *Client) CreateNetwork(ctx context.Context, input CreateNetworkInput) (*Network, error) {
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
	values := url.Values{}
	values.Set("request_type", "deleteNetwork")
	values.Set("client_name", input.ClientName)
	values.Set("ip", input.IP)
	values.Set("BM", strconv.FormatInt(input.Bitmask, 10))

	return c.doFormRequest(ctx, values, nil)
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

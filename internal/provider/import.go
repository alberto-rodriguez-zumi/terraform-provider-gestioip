package provider

import (
	"fmt"
	"strconv"
	"strings"
)

func splitImportIdentifier(identifier string) (string, string) {
	clientName, value, found := strings.Cut(identifier, "|")
	if !found {
		return "", strings.TrimSpace(identifier)
	}

	return strings.TrimSpace(clientName), strings.TrimSpace(value)
}

func parseHostImportID(identifier string) (string, string, error) {
	clientName, ip := splitImportIdentifier(identifier)
	if ip == "" {
		return "", "", fmt.Errorf("expected import ID in the form [client_name|]<ip>")
	}

	return clientName, ip, nil
}

func parseVLANImportID(identifier string) (string, string, error) {
	clientName, number := splitImportIdentifier(identifier)
	if number == "" {
		return "", "", fmt.Errorf("expected import ID in the form [client_name|]<number>")
	}

	return clientName, number, nil
}

func parseNetworkImportID(identifier string) (string, string, int64, error) {
	clientName, value := splitImportIdentifier(identifier)
	ip, bitmaskText, found := strings.Cut(value, "/")
	if !found || strings.TrimSpace(ip) == "" || strings.TrimSpace(bitmaskText) == "" {
		return "", "", 0, fmt.Errorf("expected import ID in the form [client_name|]<ip>/<bitmask>")
	}

	bitmask, err := strconv.ParseInt(strings.TrimSpace(bitmaskText), 10, 64)
	if err != nil {
		return "", "", 0, fmt.Errorf("parse bitmask %q: %w", strings.TrimSpace(bitmaskText), err)
	}

	return clientName, strings.TrimSpace(ip), bitmask, nil
}

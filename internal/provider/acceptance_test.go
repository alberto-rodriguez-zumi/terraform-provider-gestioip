package provider

import (
	"context"
	"fmt"
	"os"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/alberto-rodriguez-zumi/terraform-provider-gestioip/internal/client"
	"github.com/hashicorp/terraform-plugin-framework/providerserver"
	"github.com/hashicorp/terraform-plugin-go/tfprotov6"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
)

var testAccProtoV6ProviderFactories = map[string]func() (tfprotov6.ProviderServer, error){
	"gestioip": providerserver.NewProtocol6WithError(New()),
}

type acceptanceConfig struct {
	Label           string
	BaseURL         string
	Username        string
	Password        string
	ClientName      string
	NetworkSite     string
	NetworkCategory string
	HostSite        string
	HostCategory    string
	NetworkPrefix   string
	NetworkStart    int
	NetworkEnd      int
	VLANStart       int
	VLANEnd         int
}

type acceptanceValues struct {
	NetworkIP string
	HostIP    string
	VLAN      string
	Suffix    string
}

func TestAccGestioIP35Lifecycle(t *testing.T) {
	t.Parallel()

	cfg := loadAcceptanceConfig(t, "GESTIOIP35")
	values := discoverAcceptanceValues(t, cfg)

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccConfig(cfg, values, false),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("gestioip_network.test", "ip", values.NetworkIP),
					resource.TestCheckResourceAttr("gestioip_network.test", "bitmask", "24"),
					resource.TestCheckResourceAttr("gestioip_network.test", "site", cfg.NetworkSite),
					resource.TestCheckResourceAttr("gestioip_network.test", "category", cfg.NetworkCategory),
					resource.TestCheckResourceAttr("gestioip_host.test", "ip", values.HostIP),
					resource.TestCheckResourceAttr("gestioip_host.test", "site", cfg.HostSite),
					resource.TestCheckResourceAttr("gestioip_vlan.test", "number", values.VLAN),
					resource.TestCheckResourceAttr("data.gestioip_network.test", "ip", values.NetworkIP),
					resource.TestCheckResourceAttr("data.gestioip_host.test", "ip", values.HostIP),
					resource.TestCheckResourceAttr("data.gestioip_vlan.test", "number", values.VLAN),
				),
			},
			{
				Config: testAccConfig(cfg, values, true),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("gestioip_network.test", "description", fmt.Sprintf("tf acc network updated %s", values.Suffix)),
					resource.TestCheckResourceAttr("gestioip_host.test", "hostname", fmt.Sprintf("tf-acc-host-upd-%s", values.Suffix)),
					resource.TestCheckResourceAttr("gestioip_vlan.test", "name", fmt.Sprintf("tf-acc-vlan-upd-%s", values.Suffix)),
				),
			},
			{
				ResourceName:            "gestioip_network.test",
				ImportState:             true,
				ImportStateIdFunc:       importStateIDFunc(cfg.ClientName + "|" + values.NetworkIP + "/24"),
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"client_name", "ip_version"},
			},
			{
				ResourceName:            "gestioip_host.test",
				ImportState:             true,
				ImportStateIdFunc:       importStateIDFunc(cfg.ClientName + "|" + values.HostIP),
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"client_name", "category", "comment", "ip_int", "ip_version", "network_id"},
			},
			{
				ResourceName:            "gestioip_vlan.test",
				ImportState:             true,
				ImportStateIdFunc:       importStateIDFunc(cfg.ClientName + "|" + values.VLAN),
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"client_name"},
			},
		},
	})
}

func TestAccGestioIP32Lifecycle(t *testing.T) {
	t.Parallel()

	cfg := loadAcceptanceConfig(t, "GESTIOIP32")
	values := discoverAcceptanceValues(t, cfg)

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccConfig(cfg, values, false),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("gestioip_network.test", "ip", values.NetworkIP),
					resource.TestCheckResourceAttr("gestioip_network.test", "bitmask", "24"),
					resource.TestCheckResourceAttr("gestioip_network.test", "site", cfg.NetworkSite),
					resource.TestCheckResourceAttr("gestioip_network.test", "category", cfg.NetworkCategory),
					resource.TestCheckResourceAttr("gestioip_host.test", "ip", values.HostIP),
					resource.TestCheckResourceAttr("gestioip_host.test", "site", cfg.HostSite),
					resource.TestCheckResourceAttr("gestioip_vlan.test", "number", values.VLAN),
					resource.TestCheckResourceAttr("data.gestioip_network.test", "ip", values.NetworkIP),
					resource.TestCheckResourceAttr("data.gestioip_host.test", "ip", values.HostIP),
					resource.TestCheckResourceAttr("data.gestioip_vlan.test", "number", values.VLAN),
				),
			},
			{
				Config: testAccConfig(cfg, values, true),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("gestioip_network.test", "description", fmt.Sprintf("tf acc network updated %s", values.Suffix)),
					resource.TestCheckResourceAttr("gestioip_host.test", "hostname", fmt.Sprintf("tf-acc-host-upd-%s", values.Suffix)),
					resource.TestCheckResourceAttr("gestioip_vlan.test", "name", fmt.Sprintf("tf-acc-vlan-upd-%s", values.Suffix)),
				),
			},
			{
				ResourceName:            "gestioip_network.test",
				ImportState:             true,
				ImportStateIdFunc:       importStateIDFunc(cfg.ClientName + "|" + values.NetworkIP + "/24"),
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"client_name", "ip_version"},
			},
			{
				ResourceName:            "gestioip_host.test",
				ImportState:             true,
				ImportStateIdFunc:       importStateIDFunc(cfg.ClientName + "|" + values.HostIP),
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"client_name", "category", "comment", "ip_int", "ip_version", "network_id"},
			},
			{
				ResourceName:            "gestioip_vlan.test",
				ImportState:             true,
				ImportStateIdFunc:       importStateIDFunc(cfg.ClientName + "|" + values.VLAN),
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"client_name"},
			},
		},
	})
}

func loadAcceptanceConfig(t *testing.T, prefix string) acceptanceConfig {
	t.Helper()

	if os.Getenv("TF_ACC") == "" {
		t.Skip("set TF_ACC=1 to run acceptance tests")
	}

	getRequired := func(name string) string {
		t.Helper()

		value := strings.TrimSpace(os.Getenv(name))
		if value == "" {
			t.Skipf("set %s to run this acceptance test", name)
		}

		return value
	}

	cfg := acceptanceConfig{
		Label:      prefix,
		BaseURL:    getRequired(prefix + "_BASE_URL"),
		Username:   getRequired(prefix + "_USERNAME"),
		Password:   getRequired(prefix + "_PASSWORD"),
		ClientName: getRequired(prefix + "_CLIENT_NAME"),
	}

	switch prefix {
	case "GESTIOIP32":
		cfg.NetworkSite = firstEnv(prefix+"_NETWORK_SITE", "ALL-DCs")
		cfg.NetworkCategory = firstEnv(prefix+"_NETWORK_CATEGORY", "DEV_TEST")
		cfg.HostSite = firstEnv(prefix+"_HOST_SITE", "ALL-DCs")
		cfg.HostCategory = firstEnv(prefix+"_HOST_CATEGORY", "")
		cfg.NetworkPrefix = firstEnv(prefix+"_NETWORK_PREFIX", "10.254")
		cfg.NetworkStart = firstEnvInt(prefix+"_NETWORK_START", 240)
		cfg.NetworkEnd = firstEnvInt(prefix+"_NETWORK_END", 254)
		cfg.VLANStart = firstEnvInt(prefix+"_VLAN_START", 3900)
		cfg.VLANEnd = firstEnvInt(prefix+"_VLAN_END", 4094)
	default:
		cfg.NetworkSite = firstEnv(prefix+"_NETWORK_SITE", "Lon")
		cfg.NetworkCategory = firstEnv(prefix+"_NETWORK_CATEGORY", "test")
		cfg.HostSite = firstEnv(prefix+"_HOST_SITE", "Lon")
		cfg.HostCategory = firstEnv(prefix+"_HOST_CATEGORY", "server")
		cfg.NetworkPrefix = firstEnv(prefix+"_NETWORK_PREFIX", "10.64")
		cfg.NetworkStart = firstEnvInt(prefix+"_NETWORK_START", 200)
		cfg.NetworkEnd = firstEnvInt(prefix+"_NETWORK_END", 254)
		cfg.VLANStart = firstEnvInt(prefix+"_VLAN_START", 300)
		cfg.VLANEnd = firstEnvInt(prefix+"_VLAN_END", 399)
	}

	return cfg
}

func discoverAcceptanceValues(t *testing.T, cfg acceptanceConfig) acceptanceValues {
	t.Helper()

	apiClient, err := client.New(client.Config{
		BaseURL:    cfg.BaseURL,
		ClientName: cfg.ClientName,
		Username:   cfg.Username,
		Password:   cfg.Password,
	})
	if err != nil {
		t.Fatalf("create client: %v", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()

	networks, err := apiClient.ListNetworks(ctx, cfg.ClientName)
	if err != nil {
		t.Fatalf("list networks: %v", err)
	}

	usedNetworks := make(map[string]struct{}, len(networks))
	for _, network := range networks {
		usedNetworks[fmt.Sprintf("%s/%d", network.IP, network.Bitmask)] = struct{}{}
	}

	var networkIP string
	var hostIP string

	for octet := cfg.NetworkStart; octet <= cfg.NetworkEnd; octet++ {
		candidateNetwork := fmt.Sprintf("%s.%d.0", cfg.NetworkPrefix, octet)
		key := candidateNetwork + "/24"
		if _, ok := usedNetworks[key]; ok {
			continue
		}

		networkIP = candidateNetwork
		hostIP = fmt.Sprintf("%s.%d.10", cfg.NetworkPrefix, octet)
		break
	}

	if networkIP == "" {
		t.Fatalf("unable to find a free /24 network in prefix %s between %d and %d", cfg.NetworkPrefix, cfg.NetworkStart, cfg.NetworkEnd)
	}

	var vlan string
	for number := cfg.VLANStart; number <= cfg.VLANEnd; number++ {
		_, err := apiClient.ReadVLAN(ctx, cfg.ClientName, strconv.Itoa(number))
		if err == nil {
			continue
		}
		if client.IsNotFoundError(err) {
			vlan = strconv.Itoa(number)
			break
		}

		t.Fatalf("read vlan %d: %v", number, err)
	}

	if vlan == "" {
		t.Fatalf("unable to find a free VLAN number between %d and %d", cfg.VLANStart, cfg.VLANEnd)
	}

	suffix := strconv.FormatInt(time.Now().UnixNano()%100000, 10)

	return acceptanceValues{
		NetworkIP: networkIP,
		HostIP:    hostIP,
		VLAN:      vlan,
		Suffix:    suffix,
	}
}

func testAccConfig(cfg acceptanceConfig, values acceptanceValues, updated bool) string {
	networkDescription := fmt.Sprintf("tf acc network %s", values.Suffix)
	networkComment := fmt.Sprintf("tf acc comment %s", values.Suffix)
	hostHostname := fmt.Sprintf("tf-acc-host-%s", values.Suffix)
	hostDescription := fmt.Sprintf("tf acc host %s", values.Suffix)
	vlanName := fmt.Sprintf("tf-acc-vlan-%s", values.Suffix)
	vlanDescription := fmt.Sprintf("tf acc vlan %s", values.Suffix)

	if updated {
		networkDescription = fmt.Sprintf("tf acc network updated %s", values.Suffix)
		networkComment = fmt.Sprintf("tf acc comment updated %s", values.Suffix)
		hostHostname = fmt.Sprintf("tf-acc-host-upd-%s", values.Suffix)
		hostDescription = fmt.Sprintf("tf acc host updated %s", values.Suffix)
		vlanName = fmt.Sprintf("tf-acc-vlan-upd-%s", values.Suffix)
		vlanDescription = fmt.Sprintf("tf acc vlan updated %s", values.Suffix)
	}

	hostCategoryBlock := ""
	if cfg.HostCategory != "" {
		hostCategoryBlock = fmt.Sprintf("\n  category    = %q", cfg.HostCategory)
	}

	return fmt.Sprintf(`
provider "gestioip" {
  base_url        = %q
  client_name     = %q
  username        = %q
  password        = %q
  allow_overwrite = false
}

resource "gestioip_network" "test" {
  ip          = %q
  bitmask     = 24
  description = %q
  site        = %q
  category    = %q
  comment     = %q
  sync        = false
}

resource "gestioip_host" "test" {
  depends_on = [gestioip_network.test]

  ip          = %q
  hostname    = %q
  description = %q
  site        = %q%s
}

resource "gestioip_vlan" "test" {
  number      = %q
  name        = %q
  description = %q
}

data "gestioip_network" "test" {
  ip      = gestioip_network.test.ip
  bitmask = gestioip_network.test.bitmask
}

data "gestioip_host" "test" {
  ip = gestioip_host.test.ip
}

data "gestioip_vlan" "test" {
  number = gestioip_vlan.test.number
}
`,
		cfg.BaseURL,
		cfg.ClientName,
		cfg.Username,
		cfg.Password,
		values.NetworkIP,
		networkDescription,
		cfg.NetworkSite,
		cfg.NetworkCategory,
		networkComment,
		values.HostIP,
		hostHostname,
		hostDescription,
		cfg.HostSite,
		hostCategoryBlock,
		values.VLAN,
		vlanName,
		vlanDescription,
	)
}

func firstEnv(name, fallback string) string {
	if value := strings.TrimSpace(os.Getenv(name)); value != "" {
		return value
	}

	return fallback
}

func firstEnvInt(name string, fallback int) int {
	if value := strings.TrimSpace(os.Getenv(name)); value != "" {
		parsed, err := strconv.Atoi(value)
		if err == nil {
			return parsed
		}
	}

	return fallback
}

func importStateIDFunc(identifier string) resource.ImportStateIdFunc {
	return func(*terraform.State) (string, error) {
		return identifier, nil
	}
}

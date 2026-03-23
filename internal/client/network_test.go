package client

import "testing"

func TestInferIPVersion(t *testing.T) {
	t.Parallel()

	if got := inferIPVersion("192.168.1.0"); got != "v4" {
		t.Fatalf("expected v4, got %q", got)
	}

	if got := inferIPVersion("2001:db8::"); got != "v6" {
		t.Fatalf("expected v6, got %q", got)
	}
}

func TestParseNetworksFromFrontend(t *testing.T) {
	t.Parallel()

	body := []byte(`
<tr height="24px" bgcolor="#efefef" class="show_detail" style="cursor:pointer;" onClick="document.forms.list_host116.submit()" id="tr_116">
  <td><b>192.168.9.0</b></td>
  <td align="center"><acronym title="255.255.255.0 - 254 hosts"><span id='SMbutton_116'>24</span></acronym></td>
  <td>TEST-Backend</td>
  <td align="center" nowrap>ALL-DCs</td>
  <td align="center" nowrap>DEV_TEST</td>
  <td>VLAN 1007<br></td>
  <td align="center"> </td>
  <td onClick="document.forms.list_host116.submit()"></td>
  <td onClick="document.forms.list_host116.submit()"></td>
  <td onClick="document.forms.list_host116.submit()"></td>
  <td>
    <form method="POST" name="list_host116" action="http://localhost:8080/gestioip/ip_show.cgi">
      <input name="bignet" type="hidden" value="0">
      <input name="ip_version" type="hidden" value="v4">
      <input name="client_id" type="hidden" value="1">
      <input name="red_num" type="hidden" value="202">
      <input name="loc" type="hidden" value="ALL-DCs">
    </form>
  </td>
</tr>`)

	networks, err := parseNetworksFromFrontend(body, "Voxel Group")
	if err != nil {
		t.Fatalf("parse networks from frontend: %v", err)
	}

	if len(networks) != 1 {
		t.Fatalf("expected 1 network, got %d", len(networks))
	}

	network := networks[0]
	if network.ID != "202" {
		t.Fatalf("expected id %q, got %q", "202", network.ID)
	}

	if network.IP != "192.168.9.0" {
		t.Fatalf("expected ip %q, got %q", "192.168.9.0", network.IP)
	}

	if network.Bitmask != 24 {
		t.Fatalf("expected bitmask %d, got %d", 24, network.Bitmask)
	}

	if network.Description != "TEST-Backend" {
		t.Fatalf("expected description %q, got %q", "TEST-Backend", network.Description)
	}

	if network.Site != "ALL-DCs" {
		t.Fatalf("expected site %q, got %q", "ALL-DCs", network.Site)
	}

	if network.Category != "DEV_TEST" {
		t.Fatalf("expected category %q, got %q", "DEV_TEST", network.Category)
	}

	if network.Comment != "VLAN 1007" {
		t.Fatalf("expected comment %q, got %q", "VLAN 1007", network.Comment)
	}

	if network.IPVersion != "v4" {
		t.Fatalf("expected ip_version %q, got %q", "v4", network.IPVersion)
	}
}

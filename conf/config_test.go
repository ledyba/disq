package conf

import (
	"testing"
)

func TestEmptyConfig(t *testing.T) {
	actual, err := Load([]byte(`
{
  "dns-servers": [],
  "v4networks": {},
  "machines": {}
}
`))
	if err != nil {
		t.Fatalf("Parse error: %v", err)
	}
	if len(actual.DNSServers) != 0 {
		t.Errorf("We want empty DNS server list.")
	}
	if len(actual.V4Networks) != 0 {
		t.Errorf("We want empty v4 network list.")
	}
	if len(actual.Machines) != 0 {
		t.Errorf("We want empty machine list.")
	}
}

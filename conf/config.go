package conf

import (
	"encoding/json"
)

type Config struct {
	DNS        DNS                  `json:"dns"`
	V4Networks map[string]V4Network `json:"v4networks"`
	Machines   map[string]Machine   `json:"machines"`
}

type DNS struct {
	Listen    string   `json:"listen"`
	Networks  []string `json:"networks"`
	LocalTTL  int      `json:"local-ttl"`
	GlobalTTL int      `json:"global-ttl"`
}

type V4Network struct {
	InterfaceName     string   `json:"interface"`
	Network           string   `json:"network"`
	DHCP4Listen       string   `json:"dhcp4-listen"`
	LeaseDurationDays float64  `json:"lease-duration-days"`
	NameServerAddrs   []string `json:"nameserver-address,omitempty"`
	GatewayAddr       string   `json:"gateway-address,omitempty"`
}

type Machine []Interface

type Interface struct {
	HardwareAddr string `json:"hardware-address"`
	IPv4Addr     string `json:"ipv4-address"`
	Fqdn         string `json:"fqdn,omitempty"` /* (ex) zoi.eaglejump.jp. */
}

func Load(data []byte) (*Config, error) {
	var conf Config
	err := json.Unmarshal(data, &conf)
	return &conf, err
}

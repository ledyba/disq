package conf

import (
	"encoding/json"
)

type Config struct {
	DNSServers []string             `json:"dns-servers"`
	V4Networks map[string]V4Network `json:"v4networks"`
	Machines   map[string]Machine   `json:"machines"`
}

type V4Network struct {
	InterfaceName     string   `json:"interface"`
	InterfaceIPAddr   string   `json:"interface-address"`
	DNSListen         string   `json:"dns-listen"`
	DHCP4Listen       string   `json:"dhcp4-listen"`
	LeaseDurationDays float64  `json:"lease-duration-days"`
	NameServerAddrs   []string `json:"nameserver-address,omitempty"`
	GatewayAddr       string   `json:"gateway-address,omitempty"`
}

type Machine []Interface

type Interface struct {
	HardwareAddr string `json:"hardware-address"`
	IPAddr       string `json:"ip-address"`
	Fqdn         string `json:"fqdn,omitempty"` /* (ex) zoi.eaglejump.jp. */
}

func Load(data []byte) (*Config, error) {
	var conf Config
	err := json.Unmarshal(data, &conf)
	return &conf, err
}

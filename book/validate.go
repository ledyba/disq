package book

import (
	"fmt"

	log "github.com/Sirupsen/logrus"
)

func (b *Book) Validate() error {
	var err error
	//
	for name, n := range b.V4Networks {
		if n.GatewayAddr != nil && !n.Network.Contains(n.GatewayAddr) {
			return fmt.Errorf("gateway addr %s (for %s) is not in the network(%s)", n.GatewayAddr.String(), name, n.Network.String())
		}
	}

	err = b.validateV4()
	if err != nil {
		return err
	}
	return nil
}

func (b *Book) validateV4() error {
	// Checking v4 network
	ip2hw := make(map[string]*Machine)
	for name, m := range b.Machines {
		for _, nic := range m.Interfaces {
			// Checking IPv4Addr
			ipv4addr := nic.IPv4Addr
			ipv4addrStr := ipv4addr.String()
			if another, ok := ip2hw[ipv4addrStr]; ok {
				return fmt.Errorf("IPv4Addr %s (assigned to %s) is also assigned to %s", ipv4addrStr, name, another.Name)
			}
			{
				// disqで管理してないマシンを追加してDNSとして使ってもよいので、warnを出すだけ。
				found := false
				for _, n := range b.V4Networks {
					if n.Network.Contains(ipv4addr) {
						found = true
						break
					}
				}
				if !found {
					log.Warnf("IPv4Addr %s (assigned to %s) is not in all networks managed by disq.", ipv4addrStr, name)
				}
			}
			ip2hw[ipv4addrStr] = m
		}
	}
	hw2ip := make(map[string]*Machine)
	for name, m := range b.Machines {
		for _, nic := range m.Interfaces {
			hwaddr := nic.HardwareAddr
			hwaddrStr := hwaddr.String()
			if another, ok := hw2ip[hwaddrStr]; ok {
				return fmt.Errorf("HardwareAddr %s (assigned to %s) is also assigned to %s", hwaddrStr, name, another.Name)
			}
			hw2ip[hwaddrStr] = m
		}
	}

	return nil
}

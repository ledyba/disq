package book

import (
	"errors"
	"net"

	"fmt"

	log "github.com/Sirupsen/logrus"
	"github.com/ledyba/disq/conf"
)

var (
	ErrAddressIsNotAssigned = errors.New("specified address is not assigned to the interface")
)

func FromConfig(conf *conf.Config) (*Book, error) {
	var err error
	b := &Book{}

	//DNS
	b.DNS.Listen = conf.DNS.Listen
	b.DNS.Networks = conf.DNS.Networks
	b.DNS.LocalTTL = conf.DNS.LocalTTL
	if b.DNS.LocalTTL < 0 {
		b.DNS.LocalTTL = 0
	}
	b.DNS.GlobalTTL = conf.DNS.GlobalTTL
	if b.DNS.GlobalTTL < 0 {
		b.DNS.GlobalTTL = 0
	}
	for _, network := range b.DNS.Networks {
		_, ok := conf.V4Networks[network]
		if !ok {
			return nil, fmt.Errorf("network [%s] (allowed for serving DNS) not found", network)
		}
	}

	// V4Netrowks
	b.V4Networks = make(map[string]*V4Network)
	for name, netConf := range conf.V4Networks {
		network, err := compileNetwork(&netConf)
		if err != nil {
			return nil, err
		}
		b.V4Networks[name] = network
	}

	// Machines
	b.Machines = make(map[string]*Machine)
	for name, mc := range conf.Machines {
		m, err := compileMachine(name, &mc)
		if err != nil {
			return nil, err
		}
		b.Machines[name] = m
	}

	err = b.Validate()
	if err != nil {
		return nil, err
	}

	return b, nil
}

func compileMachine(name string, c *conf.Machine) (*Machine, error) {
	infs := make([]Interface, len(*c))
	for i, inf := range *c {
		hwaddr, err := net.ParseMAC(inf.HardwareAddr)
		if err != nil {
			return nil, err
		}
		ipv4addr := net.ParseIP(inf.IPv4Addr)
		if ipv4addr != nil {
			ipv4addr = ipv4addr.To4()
		}
		if ipv4addr == nil {
			return nil, &net.ParseError{
				Type: "IP address",
				Text: inf.IPv4Addr,
			}
		}
		infs[i] = Interface{
			HardwareAddr: hwaddr,
			IPv4Addr:     ipv4addr,
			Fqdn:         inf.Fqdn,
		}
	}
	return &Machine{
		Name:       name,
		Interfaces: infs,
	}, nil
}

func compileNetwork(netConf *conf.V4Network) (*V4Network, error) {
	var err error
	nif, err := net.InterfaceByName(netConf.InterfaceName)
	if err != nil {
		log.Errorf("Interface %s not found", netConf.InterfaceName)
		log.Errorf("  All Interfaces:")
		nics, err2 := net.Interfaces()
		if err2 != nil {
			log.Errorf("  Error on listing interfaces: %v", err2)
			return nil, err2
		}
		if len(nics) == 0 {
			log.Error("  <<Not Found>>")
		} else {
			for _, nic := range nics {
				log.Errorf("  [%02d] %s", nic.Index, nic.Name)
				log.Errorf("    -  HW: %s", nic.HardwareAddr)
				log.Errorf("    - MTU: %d", nic.MTU)
				addrs, err3 := nic.Addrs()
				if err3 != nil {
					log.Errorf("       - Addr: error=%v", err3)
					return nil, err3
				}
				for _, addr := range addrs {
					log.Errorf("       - Addr: %s (%s)", addr.String(), addr.Network())
				}
			}
		}
		return nil, err
	}
	addrs, err := nif.Addrs()
	if err != nil {
		log.Errorf("Error on guessing addresses of %s", netConf.InterfaceName)
		return nil, err
	}

	_, network, err := net.ParseCIDR(netConf.Network)
	if err != nil {
		log.Errorf("NetworkAddress %s (configured for %s) is not a valid ipv4 network.", netConf.Network, netConf.InterfaceName)
		return nil, err
	}

	var addr net.IP
	for _, a := range addrs {
		ip, _, err := net.ParseCIDR(a.String())
		if err != nil {
			return nil, err
		}
		if network.Contains(ip) {
			addr = ip
			break
		}
	}

	if addr == nil {
		log.Errorf("Network %s is not assigned to %s", network.String(), netConf.InterfaceName)
		log.Errorf("  Addresses assigned to %s:", netConf.InterfaceName)
		if len(addrs) == 0 {
			log.Error("   <<Not Found>>")
		} else {
			for _, a := range addrs {
				log.Errorf("    - %s (%s)", a.String(), a.Network())
			}
		}
		return nil, ErrAddressIsNotAssigned
	}

	var nameServerAddrs []net.IP
	if len(netConf.NameServerAddrs) == 0 {
		log.Warnf("NameServerAddress is not configured for %s", netConf.InterfaceName)
	} else {
		for _, addr := range netConf.NameServerAddrs {
			ip := net.ParseIP(addr)
			if ip == nil {
				log.Errorf("NameServer %s (configured for %s) is not a valid ipv4 address.", addr, netConf.InterfaceName)
				return nil, &net.ParseError{
					Type: "IP address",
					Text: addr,
				}
			}
			nameServerAddrs = append(nameServerAddrs, ip)
		}
	}

	var gatewayAddress net.IP
	if len(netConf.GatewayAddr) == 0 {
		log.Warnf("GatewayAddress is not configured for %s", netConf.InterfaceName)
	} else {
		gatewayAddress = net.ParseIP(netConf.GatewayAddr)
		if gatewayAddress == nil {
			log.Errorf("NameServer %s (configured for %s) is not a valid ipv4 address.", netConf.GatewayAddr, netConf.InterfaceName)
			return nil, &net.ParseError{
				Type: "IP address",
				Text: netConf.GatewayAddr,
			}
		}
	}
	return &V4Network{
		Interface:         nif,
		MyAddress:         addr,
		Network:           network,
		DHCP4Listen:       netConf.DHCP4Listen,
		NameServerAddrs:   nameServerAddrs,
		GatewayAddr:       gatewayAddress,
		LeaseDurationDays: netConf.LeaseDurationDays,
	}, nil

}

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
	book := &Book{}

	//DNS
	book.DNS.Listen = conf.DNS.Listen
	book.DNS.Networks = conf.DNS.Networks
	for _, network := range book.DNS.Networks {
		_, ok := conf.V4Networks[network]
		if !ok {
			return nil, fmt.Errorf("network [%s] (allowed for serving DNS) not found", network)
		}
	}

	// V4Netrowks
	book.V4Networks = make(map[string]*V4Network)
	for name, netConf := range conf.V4Networks {
		network, err := compileNetwork(&netConf)
		if err != nil {
			return nil, err
		}
		book.V4Networks[name] = network
	}

	// Machines
	book.Machines = make(map[string]*Machine)
	for name, mc := range conf.Machines {
		m, err := compileMachine(name, &mc)
		if err != nil {
			return nil, err
		}
		book.Machines[name] = m
	}

	return book, nil
}

func compileMachine(name string, c *conf.Machine) (*Machine, error) {
	infs := make([]Interface, len(*c))
	for i, inf := range *c {
		hwaddr, err := net.ParseMAC(inf.HardwareAddr)
		if err != nil {
			return nil, err
		}
		ipaddr := net.ParseIP(inf.IPAddr)
		if ipaddr == nil {
			return nil, &net.ParseError{
				Type: "IP address",
				Text: inf.IPAddr,
			}
		}
		infs[i] = Interface{
			HardwareAddr: hwaddr,
			IPAddr:       ipaddr,
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
			log.Errorf("  NotFound: %v", err2)
			return nil, err2
		}
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
		return nil, err
	}
	addrs, err := nif.Addrs()
	if err != nil {
		log.Errorf("Failed to guess addresses of %s", netConf.InterfaceName)
		return nil, err
	}
	interfaceAddress, network, err := net.ParseCIDR(netConf.InterfaceIPAddr)
	if err != nil {
		log.Errorf("InterfaceAddress %s (configured for %s) is not a valid ipv4 address.", netConf.InterfaceIPAddr, netConf.InterfaceName)
		return nil, err
	}

	var addr net.Addr
	for _, a := range addrs {
		ip, _, err := net.ParseCIDR(a.String())
		if err != nil {
			return nil, err
		}
		if ip.Equal(interfaceAddress) {
			addr = a
			break
		}
	}

	if addr == nil {
		log.Errorf("Address %s is not assigned to %s", interfaceAddress.String(), netConf.InterfaceName)
		log.Errorf("  Address assigned to %s:", netConf.InterfaceName)
		for _, a := range addrs {
			log.Errorf("    - %s (%s)", a.String(), a.Network())
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
			log.Errorf("NameServer %s (configured for %s) is not a valid ipv4 address.", netConf.GatewayAddr, netConf.InterfaceIPAddr)
			return nil, &net.ParseError{
				Type: "IP address",
				Text: netConf.GatewayAddr,
			}
		}
	}
	return &V4Network{
		Interface:         nif,
		InterfaceIPAddr:   interfaceAddress,
		DHCP4Listen:       netConf.DHCP4Listen,
		Network:           network,
		NameServerAddrs:   nameServerAddrs,
		GatewayAddr:       gatewayAddress,
		LeaseDurationDays: netConf.LeaseDurationDays,
	}, nil

}

package book

import (
	"bytes"
	"net"
)

// Immutable!!
type Book struct {
	DNS        DNS
	V4Networks map[string]*V4Network
	Machines   map[string]*Machine
}

type DNS struct {
	Listen   string
	Networks []string
}

func (b *Book) LookupIPForHardwareAddr(hwaddr net.HardwareAddr) net.IP {
	for _, machine := range b.Machines {
		for _, nic := range machine.Interfaces {
			if bytes.Compare(nic.HardwareAddr, hwaddr) == 0 {
				return nic.IPv4Addr
			}
		}

	}
	return nil
}
func (b *Book) LookupIPForFQDN(fqdn string) net.IP {
	for _, machine := range b.Machines {
		for _, nic := range machine.Interfaces {
			if nic.Fqdn == fqdn {
				return nic.IPv4Addr
			}
		}

	}
	return nil
}

type V4Network struct {
	Name              string
	Interface         *net.Interface
	MyAddress         net.IP
	Network           *net.IPNet
	DHCP4Listen       string
	NameServerAddrs   []net.IP
	GatewayAddr       net.IP
	LeaseDurationDays float64
}

type Machine struct {
	Name       string
	Interfaces []Interface
}

type Interface struct {
	HardwareAddr net.HardwareAddr
	IPv4Addr     net.IP
	Fqdn         string
}

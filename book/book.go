package book

import (
	"bytes"
	"net"
)

// Immutable!!
type Book struct {
	V4Networks map[string]*V4Network
	Machines   map[string]*Machine
	ARecords   map[string]*Interface
	PTRRecords map[string]*Interface
}

func (book *Book) LookupIPForHardwareAddr(hwaddr net.HardwareAddr) net.IP {
	for _, machine := range book.Machines {
		for _, nic := range machine.Interfaces {
			if bytes.Compare(nic.HardwareAddr, hwaddr) == 0 {
				return nic.IPAddr
			}
		}

	}
	return nil
}
func (book *Book) LookupIPForFQDN(fqdn string) net.IP {
	for _, machine := range book.Machines {
		for _, nic := range machine.Interfaces {
			if nic.Fqdn == fqdn {
				return nic.IPAddr
			}
		}

	}
	return nil
}

type V4Network struct {
	Name              string
	Interface         *net.Interface
	InterfaceIPAddr   net.IP
	DNSListen         string
	DHCP4Listen       string
	Network           *net.IPNet
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
	IPAddr       net.IP
	Fqdn         string
}

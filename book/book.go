package book

import (
	"bytes"
	"net"

	"github.com/krolaw/dhcp4"
)

// Immutable!!
type Book struct {
	DNSListens []string
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
	DHCP4Options      dhcp4.Options
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

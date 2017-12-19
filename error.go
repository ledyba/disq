package disq

import (
	"fmt"
	"net"
)

type DHCP4Error struct {
	Err     error
	Network string
}

func (e *DHCP4Error) Error() string {
	return fmt.Sprintf("DHCP4Error: network=%s err=%s", e.Network, e.Err)
}

type DHCP4WrongAddressRequestedError struct {
	SName        string
	HardwareAddr net.HardwareAddr
	Requested    net.IP
	Expected     net.IP
}

func (e *DHCP4WrongAddressRequestedError) Error() string {
	return fmt.Sprintf("request packet received from %s(%s) for %s, but we expect that the address is %s", e.SName, e.HardwareAddr.String(), e.Requested.String(), e.Expected.String())
}

type DNSError struct {
	Err     error
	Network string
}

func (e *DNSError) Error() string {
	return fmt.Sprintf("DNS Error: network=%s err=%s", e.Network, e.Err)
}

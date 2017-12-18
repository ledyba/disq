package disq

import "fmt"

type DHCP4Error struct {
	Err     error
	Network string
}

func (e *DHCP4Error) Error() string {
	return fmt.Sprintf("DHCP4Error: network=%s err=%s", e.Network, e.Err)
}

type DNSError struct {
	Err    error
	Listen string
}

func (e *DNSError) Error() string {
	return fmt.Sprintf("DNS Error: listen=%s err=%s", e.Listen, e.Err)
}

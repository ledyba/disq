package disq

import (
	"net"
	"testing"

	"bytes"
)

func TestJoinIP(t *testing.T) {
	ips := []net.IP{
		net.IPv4(1, 2, 3, 4),
		net.IPv4(5, 6, 7, 8),
	}
	joinned := joinIPv4(ips)
	expected := []byte{1, 2, 3, 4, 5, 6, 7, 8}
	if bytes.Compare(joinned, expected) != 0 {
		t.Errorf("Expected %v, got %v", expected, joinned)
	}
}

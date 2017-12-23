package disq

import (
	"net"
	"time"

	"fmt"

	"io"

	log "github.com/Sirupsen/logrus"
	dhcp "github.com/krolaw/dhcp4"
	"github.com/krolaw/dhcp4/conn"
)

type dhcp4Server struct {
	parent  *Server
	network string
	conn    dhcp4conn
}

type dhcp4conn interface {
	dhcp.ServeConn
	io.Closer
}

func newDHCP4Server(parent *Server, network string) *dhcp4Server {
	s := &dhcp4Server{
		parent:  parent,
		network: network,
		conn:    nil,
	}
	return s
}

// Network for the clients
func (s *dhcp4Server) Serve() error {
	book := s.parent.book()
	network, ok := book.V4Networks[s.network]

	if !ok {
		return fmt.Errorf("[DHCP][BUG] network not found: %s", s.network)
	}
	var err error
	c, err := conn.NewUDP4FilterListener(network.Interface.Name, network.DHCP4Listen)
	if err != nil {
		return err
	}
	s.conn = c
	defer func() {
		if s.conn != nil {
			s.Shutdown()
		}
	}()
	return dhcp.Serve(c, s)
}

func (s *dhcp4Server) Shutdown() error {
	var err error
	err = s.conn.Close()
	s.conn = nil
	return err
}

func (s *dhcp4Server) log() *log.Entry {
	return log.
		WithField("Module", "DHCP4").
		WithField("Network", s.network)
}

func joinIPv4(ips []net.IP) []byte {
	if len(ips) == 1 {
		return ips[0]
	}
	sum := make([]byte, len(ips)*4)
	off := 0
	for _, ip := range ips {
		copy(sum[off:off+4], ip)
		off += 4
	}
	return sum
}

func (s *dhcp4Server) ServeDHCP(p dhcp.Packet, msgType dhcp.MessageType, options dhcp.Options) dhcp.Packet {
	errorStream := s.parent.ErrorStream
	book := s.parent.book()
	network := book.V4Networks[s.network]
	servOptions := dhcp.Options{
		dhcp.OptionSubnetMask:       []byte(network.Network.Mask),
		dhcp.OptionRouter:           []byte(network.GatewayAddr),
		dhcp.OptionDomainNameServer: joinIPv4(network.NameServerAddrs),
	}

	var err error
	sname := string(p.SName())
	hwaddr := p.CHAddr()
	s.log().Infof("Message from %s (%s)", sname, hwaddr.String())
	ipaddr := book.LookupIPForHardwareAddr(hwaddr)
	leaseDuration := time.Duration(float64(time.Hour) * 24 * network.LeaseDurationDays)
	switch msgType {
	case dhcp.Discover:
		if ipaddr == nil {
			s.log().WithError(err).Error("Could not parse mac addr: %s", hwaddr.String())
			return nil
		}
		//TODO: wait
		return dhcp.ReplyPacket(
			p, dhcp.Offer,
			network.InterfaceIPAddr,
			ipaddr,
			leaseDuration,
			servOptions.SelectOrderOrAll(options[dhcp.OptionParameterRequestList]))

	case dhcp.Request:
		if server, ok := options[dhcp.OptionServerIdentifier]; ok && !net.IP(server).Equal(network.InterfaceIPAddr) {
			return nil // Message not for this dhcp server
		}
		reqIP := net.IP(options[dhcp.OptionRequestedIPAddress])
		if reqIP == nil {
			reqIP = net.IP(p.CIAddr())
		}

		if reqIP.Equal(ipaddr) {
			return dhcp.ReplyPacket(p, dhcp.ACK,
				network.InterfaceIPAddr, reqIP,
				leaseDuration,
				servOptions.SelectOrderOrAll(options[dhcp.OptionParameterRequestList]))
		}
		// Whats wrong?
		err = &DHCP4WrongAddressRequestedError{
			SName:        sname,
			HardwareAddr: hwaddr,
			Requested:    reqIP,
			Expected:     ipaddr,
		}
		errorStream <- err
		s.log().WithError(err).Error("Invalid request received. We sent NAK back.")
		return dhcp.ReplyPacket(p, dhcp.NAK,
			network.InterfaceIPAddr, nil, 0, nil)

	case dhcp.Release:
		// Nothing to do, but log.
		s.log().Infof("Release from %s (assigned to %v)", hwaddr.String(), ipaddr)
		return nil
	case dhcp.Decline:
		// Nothing to do, but log.
		s.log().Infof("Decline from %s (assigned to %v)", hwaddr.String(), ipaddr)
		return nil
	case dhcp.Inform:
		// Nothing to do, but log.
		s.log().Infof("Inform from %s (assigned to %v)", hwaddr.String(), ipaddr)
		return nil
	default:
		s.log().Error("Unknown Message: %d", msgType)
	}
	return nil
}

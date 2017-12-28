package disq

import (
	"net"
	"time"

	"fmt"

	"io"

	"math/rand"

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
	s.log().Infof("Serving @ %s", network.DHCP4Listen)
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
	sum := make([]byte, 0, len(ips)*net.IPv4len)
	off := 0
	for _, ip := range ips {
		ip = ip.To4()
		if ip == nil {
			continue
		}
		for i := range ip {
			sum = append(sum, ip[i])
		}
		off += len(ip)
	}
	return sum
}

func shuffleIP(ips []net.IP) []net.IP {
	shuffled := make([]net.IP, len(ips))
	for i, j := range rand.Perm(len(ips)) {
		shuffled[i] = ips[j]
	}
	return shuffled
}

func (s *dhcp4Server) ServeDHCP(p dhcp.Packet, msgType dhcp.MessageType, options dhcp.Options) dhcp.Packet {
	errorStream := s.parent.ErrorStream
	book := s.parent.book()
	network := book.V4Networks[s.network]

	// Setup options
	servOptions := dhcp.Options{
		dhcp.OptionSubnetMask: []byte(network.Network.Mask),
	}
	var nsList []net.IP
	if len(network.NameServerAddrs) > 0 {
		for _, ip := range network.NameServerAddrs {
			ip = ip.To4()
			if ip == nil {
				continue
			}
			nsList = append(nsList, ip)
		}
		nsList = shuffleIP(nsList)
		servOptions[dhcp.OptionDomainNameServer] = joinIPv4(nsList)
	}
	if len(network.GatewayAddr) > 0 {
		servOptions[dhcp.OptionRouter] = []byte(network.GatewayAddr)
	}

	var err error
	sname := string(p.SName())
	hwaddr := p.CHAddr()
	ipaddr := book.LookupIPForHardwareAddr(hwaddr)
	leaseDuration := time.Duration(float64(time.Hour) * 24 * network.LeaseDurationDays)
	switch msgType {
	case dhcp.Discover:
		if ipaddr == nil {
			s.log().WithError(err).Errorf("Could not find address for %s", hwaddr.String())
			return nil
		}
		s.log().Infof(`Discover from "%s" (%s)
Replying Offer:
  ServerIP: %v
  IPAddr: %v
  LeaseDuration: %v
  Options:
    Netmask: %v
    Nameservers: %v
    Router: %v`,
			sname, hwaddr.String(),
			network.MyAddress,
			ipaddr, leaseDuration,
			network.Network.Mask,
			nsList,
			network.GatewayAddr)
		//TODO: wait
		return dhcp.ReplyPacket(
			p, dhcp.Offer,
			network.MyAddress,
			ipaddr,
			leaseDuration,
			servOptions.SelectOrderOrAll(options[dhcp.OptionParameterRequestList]))

	case dhcp.Request:
		if server, ok := options[dhcp.OptionServerIdentifier]; ok && !net.IP(server).Equal(network.MyAddress) {
			return nil // Message not for this dhcp server
		}
		reqIP := net.IP(options[dhcp.OptionRequestedIPAddress])
		if reqIP == nil {
			reqIP = net.IP(p.CIAddr())
		}
		if !reqIP.Equal(ipaddr) {
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
				network.MyAddress, nil, 0, nil)
		}
		s.log().Infof(`Request from "%s" (%s)
Replying ACK:
  ServerIP: %v
  IPAddr: %v
  LeaseDuration: %v
  Options:
    Netmask: %v
    Nameservers: %v
    Router: %v`,
			sname, hwaddr.String(),
			network.MyAddress,
			ipaddr, leaseDuration,
			network.Network.Mask,
			nsList,
			network.GatewayAddr)
		return dhcp.ReplyPacket(p, dhcp.ACK,
			network.MyAddress, reqIP,
			leaseDuration,
			servOptions.SelectOrderOrAll(options[dhcp.OptionParameterRequestList]))

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

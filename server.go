package disq

import (
	"sync/atomic"

	"sync"

	"fmt"

	log "github.com/Sirupsen/logrus"
	"github.com/ledyba/disq/book"
	"github.com/miekg/dns"
)

type Server struct {
	bookPtr atomic.Value

	dns   map[string]*dns.Server
	dhcp4 map[string]*dhcp4Server

	ErrorStream chan error

	done   int32
	doneWg sync.WaitGroup
}

func (s *Server) storeBook(b *book.Book) {
	s.bookPtr.Store(b)
}
func (s *Server) book() *book.Book {
	return s.bookPtr.Load().(*book.Book)
}

func FromBook(book *book.Book) *Server {
	s := &Server{}
	s.storeBook(book)
	s.ErrorStream = make(chan error, 1)
	s.dns = make(map[string]*dns.Server)
	s.dhcp4 = make(map[string]*dhcp4Server)
	for networkName, network := range book.V4Networks {
		// Nameserver
		dnsListen := network.DNSListen
		if len(dnsListen) > 0 {
			ns := &dns.Server{}
			ns.Handler = s
			ns.Addr = dnsListen
			ns.Net = "udp"
			s.dns[networkName] = ns
		}

		dhcp4Listen := network.DHCP4Listen
		if len(dhcp4Listen) > 0 {
			ds := newDHCP4Server(s, networkName)
			s.dhcp4[networkName] = ds
		}

	}
	return s
}

func (s *Server) Start() {
	for networkName, ns := range s.dns {
		go func(networkName string) {
			s.doneWg.Add(1)
			defer s.doneWg.Done()
			for atomic.LoadInt32(&s.done) == 0 {
				log.
					WithField("Module", "DNS").
					WithField("Network", networkName).
					Info("Started")
				err := ns.ListenAndServe()
				if err != nil {
					err = &DNSError{
						Network: networkName,
						Err:     err,
					}
					s.ErrorStream <- err
				}
			}
			log.
				WithField("Module", "DNS").
				WithField("Network", networkName).
				Info("Stopped")
		}(networkName)
	}
	for networkName, ds := range s.dhcp4 {
		go func(networkName string) {
			s.doneWg.Add(1)
			defer s.doneWg.Done()
			for atomic.LoadInt32(&s.done) == 0 {
				ds.log().Info("Started")
				err := ds.Serve()
				if err != nil {
					err = &DHCP4Error{
						Network: networkName,
						Err:     err,
					}
					s.ErrorStream <- err
				}
			}
			ds.log().Info("Stopped")
		}(networkName)
	}
}

// Graceful shutdown
func (s *Server) Stop() {
	var err error
	atomic.StoreInt32(&s.done, 1)
	for listen, ns := range s.dns {
		log.
			WithField("Module", "DNS").
			WithField("Network", listen).
			Info("Shutdown requested")
		err = ns.Shutdown()
		if err != nil {
			err = &DNSError{
				Network: listen,
				Err:     err,
			}
			s.ErrorStream <- err
		}
		log.
			WithField("Module", "DNS").
			WithField("Network", listen).
			Info("Shutdown succeeded")
	}
	for network, ds := range s.dhcp4 {
		ds.log().Info("shutdown requested")
		err = ds.Shutdown()
		if err != nil {
			err = &DHCP4Error{
				Network: network,
				Err:     err,
			}
			s.ErrorStream <- err
		}
		ds.log().Info("shutdown succeeded")
	}
	log.WithField("Module", "Server").Info("Waiting for shutting down all servers.")
	s.doneWg.Wait()
}

// Reload book.
// But do not stop server.
func (s *Server) Reload(b *book.Book) error {
	dnsCnt := 0
	dhcp4Cnt := 0
	for name, network := range b.V4Networks {
		if len(network.DNSListen) > 0 {
			dnsCnt++
			if _, ok := s.dns[name]; !ok {
				return fmt.Errorf("can't add new DNS servers at this version: %s, %s", name, network.DNSListen)
			}
		}
		if len(network.DHCP4Listen) > 0 {
			dhcp4Cnt++
			if _, ok := s.dhcp4[name]; !ok {
				return fmt.Errorf("can't add new DHCP servers at this version: %s, %s", name, network.DHCP4Listen)
			}
		}
	}
	if dnsCnt != len(s.dns) {
		return fmt.Errorf("can't remove DNS servers at this version: %d -> %d", len(s.dns), dnsCnt)
	}
	if dhcp4Cnt != len(s.dhcp4) {
		return fmt.Errorf("can't remove DHCP4 servers at this version: %d -> %d", len(s.dhcp4), dhcp4Cnt)
	}
	s.storeBook(b)
	return nil
}

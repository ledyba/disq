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

	dns   *dns.Server
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
	// DNS
	if len(book.DNS.Listen) > 0 {
		s.dns = &dns.Server{
			Handler: s,
			Addr:    book.DNS.Listen,
			Net:     "udp",
		}
	}
	// DHCP
	s.dhcp4 = make(map[string]*dhcp4Server)
	for networkName, network := range book.V4Networks {
		dhcp4Listen := network.DHCP4Listen
		if len(dhcp4Listen) > 0 {
			ds := newDHCP4Server(s, networkName)
			s.dhcp4[networkName] = ds
		}
	}
	return s
}

func (s *Server) Start() {
	if s.dns != nil {
		go func() {
			s.doneWg.Add(1)
			defer s.doneWg.Done()
			for atomic.LoadInt32(&s.done) == 0 {
				log.
					WithField("Module", "DNS").
					Infof("Seaving @ %s", s.dns.Addr)
				err := s.dns.ListenAndServe()
				if err != nil {
					err = &DNSError{
						Err: err,
					}
					s.ErrorStream <- err
				}
			}
			log.
				WithField("Module", "DNS").
				Info("Stopped")
		}()
	}
	for networkName, ds := range s.dhcp4 {
		go func(networkName string) {
			s.doneWg.Add(1)
			defer s.doneWg.Done()
			for atomic.LoadInt32(&s.done) == 0 {
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
	if s.dns != nil {
		log.
			WithField("Module", "DNS").
			Info("Shutdown requested")
		err = s.dns.Shutdown()
		if err != nil {
			err = &DNSError{
				Err: err,
			}
			s.ErrorStream <- err
		}
		log.
			WithField("Module", "DNS").
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
	// DNS
	if b.DNS.Listen != s.dns.Addr {
		return fmt.Errorf("can't change DNS listening address at this version: %s, %s", b.DNS.Listen, s.dns.Addr)
	}
	// DHCP
	dhcp4Cnt := 0
	for name, network := range b.V4Networks {
		if len(network.DHCP4Listen) > 0 {
			dhcp4Cnt++
			if _, ok := s.dhcp4[name]; !ok {
				return fmt.Errorf("can't add new DHCP servers at this version: %s, %s", name, network.DHCP4Listen)
			}
		}
	}
	if dhcp4Cnt != len(s.dhcp4) {
		return fmt.Errorf("can't remove DHCP4 servers at this version: %d -> %d", len(s.dhcp4), dhcp4Cnt)
	}
	s.storeBook(b)
	return nil
}

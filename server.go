package disq

import (
	"sync/atomic"

	"sync"

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
					Info("started")
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
				Info("stopped")
		}(networkName)
	}
	for networkName, ds := range s.dhcp4 {
		go func(networkName string) {
			s.doneWg.Add(1)
			defer s.doneWg.Done()
			for atomic.LoadInt32(&s.done) == 0 {
				ds.log().Info("started")
				err := ds.Serve()
				if err != nil {
					err = &DHCP4Error{
						Network: networkName,
						Err:     err,
					}
					s.ErrorStream <- err
				}
			}
			ds.log().Info("stopped")
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
			Info("shutdown requested")
		err = ns.Shutdown()
		if err != nil {
			if err != nil {
				err = &DNSError{
					Network: listen,
					Err:     err,
				}
				s.ErrorStream <- err
			}
		}
		log.
			WithField("Module", "DNS").
			WithField("Network", listen).
			Info("shutdown succeeded")
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
	log.Info("Waiting")
	s.doneWg.Wait()
}

// Reload book.
// But do not stop server.
func (s *Server) Reload(book *book.Book) error {
	return nil
}

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
	for _, dnsListen := range book.DNSListens {
		ns := &dns.Server{}
		ns.Handler = s
		ns.Addr = dnsListen
		s.dns[dnsListen] = ns
	}
	s.dhcp4 = make(map[string]*dhcp4Server)
	for network := range book.V4Networks {
		ds := newDHCP4Server(s, network)
		s.dhcp4[network] = ds
	}
	return s
}

func (s *Server) Start() {
	for listen, ns := range s.dns {
		go func(listen string) {
			s.doneWg.Add(1)
			defer s.doneWg.Done()
			for atomic.LoadInt32(&s.done) == 0 {
				log.
					WithField("Module", "DNS").
					WithField("Listen", listen).
					Info("started")
				err := ns.ListenAndServe()
				if err != nil {
					err = &DNSError{
						Listen: listen,
						Err:    err,
					}
					s.ErrorStream <- err
				}
			}
		}(listen)
	}
	for network, ds := range s.dhcp4 {
		go func(network string) {
			s.doneWg.Add(1)
			defer s.doneWg.Done()
			for atomic.LoadInt32(&s.done) == 0 {
				ds.log().Info("started")
				err := ds.Serve()
				if err != nil {
					err = &DHCP4Error{
						Network: network,
						Err:     err,
					}
					s.ErrorStream <- err
				}
			}
		}(network)
	}
}

// Graceful shutdown
func (s *Server) Stop() {
	var err error
	atomic.StoreInt32(&s.done, 1)
	for listen, ns := range s.dns {
		err = ns.Shutdown()
		log.
			WithField("Module", "DNS").
			WithField("Listen", listen).
			Info("stopped")
		if err != nil {
			if err != nil {
				err = &DNSError{
					Listen: listen,
					Err:    err,
				}
				s.ErrorStream <- err
			}
		}
	}
	for network, ds := range s.dhcp4 {
		err = ds.Shutdown()
		ds.log().WithError(err).Info("stopped")
		if err != nil {
			err = &DHCP4Error{
				Network: network,
				Err:     err,
			}
			s.ErrorStream <- err
		}
	}
	s.doneWg.Wait()
}

// Reload book.
// But do not stop server.
func (s *Server) Reload(book *book.Book) error {
	return nil
}

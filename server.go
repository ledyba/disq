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

func (s *Server) storeBook(book *book.Book) {
	s.bookPtr.Store(s)
}
func (s *Server) book() *book.Book {
	return s.bookPtr.Load().(*book.Book)
}

func FromBook(book *book.Book) *Server {
	s := &Server{}
	s.storeBook(book)
	s.ErrorStream = make(chan error, 1)
	for _, dnsListen := range book.DNSListens {
		ns := &dns.Server{}
		ns.Handler = s
		ns.Addr = dnsListen
		s.dns[dnsListen] = ns
	}
	for network := range book.V4Networks {
		ds := newDHCP4Server(s, network)
		s.dhcp4[network] = ds
	}
	return s
}

func (s *Server) Start() {
	for listen, ns := range s.dns {
		s.doneWg.Add(1)
		func(listen string) {
			s.doneWg.Done()
			for atomic.LoadInt32(&s.done) == 0 {
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
		s.doneWg.Add(1)
		func(network string) {
			for atomic.LoadInt32(&s.done) == 0 {
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
	s.doneWg.Wait()
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
}

// Reload book.
// But do not stop server.
func (s *Server) Reload(book *book.Book) error {
	return nil
}

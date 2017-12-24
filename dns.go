package disq

import (
	"fmt"

	"net"

	log "github.com/Sirupsen/logrus"
	"github.com/miekg/dns"
)

func newReply(r *dns.Msg) *dns.Msg {
	m := new(dns.Msg)
	m.SetReply(r)
	return m
}

func (s *Server) ServeDNS(w dns.ResponseWriter, r *dns.Msg) {
	b := s.book()
	switch r.Opcode {
	case dns.OpcodeQuery:
		m := newReply(r)
		for _, q := range r.Question {
			switch q.Qtype {
			case dns.TypeA:
				if ipaddr := b.LookupIPForFQDN(q.Name); ipaddr != nil {
					// This host is in our datacenter.
					ans := ipaddr.String()
					log.WithField("Module", "DNS").Debugf("%s A %s", q.Name, ans)
					rr, err := dns.NewRR(fmt.Sprintf("%s A %s", q.Name, ans))
					if err != nil {
						log.WithField("Module", "DNS").WithError(err).Error("[BUG] Error when creating DNS response")
						return
					}
					m.Answer = append(m.Answer, rr)
				} else {
					// Host in the outside.
					addrs, err := net.LookupIP(q.Name)
					if err != nil {
						log.WithField("Module", "DNS").WithError(err).Warnf("Not found: %s", q.Name)
						break
					}
					for _, ipaddr = range addrs {
						// LookupIP returns both v4 and v6 addrs.
						ipaddr = ipaddr.To4()
						if ipaddr == nil {
							continue
						}
						resp := fmt.Sprintf("%s A %s", q.Name, ipaddr.String())
						log.WithField("Module", "DNS").Debug(resp)
						rr, err := dns.NewRR(resp)
						m.Answer = append(m.Answer, rr)
						if err != nil {
							log.WithField("Module", "DNS").WithError(err).Error("[BUG] Error when creating DNS response")
							return
						}
					}
				}
			default:
				log.Warn("Unsupported query: ", q.String())
			}
		}
		w.WriteMsg(m)
	case dns.OpcodeIQuery:
		log.WithField("Module", "DNS").Warn("IQuery was questioned, but it is obsoleted. See: https://tools.ietf.org/rfc/rfc3425.txt")
	case dns.OpcodeStatus:
		log.WithField("Module", "DNS").Info("Can't answer Status questions.")
	case dns.OpcodeNotify:
		log.WithField("Module", "DNS").Info("Can't answer Notify questions.")
	case dns.OpcodeUpdate:
		log.WithField("Module", "DNS").Info("Can't answer Update questions.")
	}
}

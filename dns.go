package disq

import (
	"fmt"

	log "github.com/Sirupsen/logrus"
	"github.com/miekg/dns"
)

func newReply(r *dns.Msg) *dns.Msg {
	m := new(dns.Msg)
	m.SetReply(r)
	return m
}

func (s *Server) ServeDNS(w dns.ResponseWriter, r *dns.Msg) {
	switch r.Opcode {
	case dns.OpcodeQuery:
		m := newReply(r)
		for _, q := range r.Question {
			switch q.Qtype {
			case dns.TypeA:
				if ip := s.book().LookupIPForFQDN(q.Name); ip != nil {
					ans := ip.String()
					log.WithField("Module", "DNS").Debugf("%s A %s", q.Name, ans)
					rr, err := dns.NewRR(fmt.Sprintf("%s A %s", q.Name, ans))
					if err != nil {
						log.WithField("Module", "DNS").WithError(err).Error("[BUG] Error when creating DNS response")
						return
					}
					m.Answer = append(m.Answer, rr)
				} else {
					log.WithField("Module", "DNS").Debugf("%s A 0.0.0.0", q.Name)
					rr, err := dns.NewRR(fmt.Sprintf("%s A %s", q.Name, "0.0.0.0"))
					if err != nil {
						log.WithField("Module", "DNS").WithError(err).Error("[BUG] Error when creating DNS response")
						return
					}
					m.Answer = append(m.Answer, rr)
				}
			default:
				log.Info("Unknown name: ", q.Name)
			}
		}
		w.WriteMsg(m)
	case dns.OpcodeIQuery:
	case dns.OpcodeStatus:
	case dns.OpcodeNotify:
	case dns.OpcodeUpdate:
	}
}

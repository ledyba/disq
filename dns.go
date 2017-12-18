package disq

import (
	"github.com/miekg/dns"
)

func (s *Server) ServeDNS(w dns.ResponseWriter, r *dns.Msg) {
	switch r.Opcode {
	case dns.OpcodeQuery:
	case dns.OpcodeIQuery:
	case dns.OpcodeStatus:
	case dns.OpcodeNotify:
	case dns.OpcodeUpdate:
	}
	m := new(dns.Msg)
	m.SetReply(r)
	w.WriteMsg(m)
}

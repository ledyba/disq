package disq

import (
	"testing"

	"time"

	"github.com/Sirupsen/logrus"
	"github.com/ledyba/disq/book-test"
	"github.com/miekg/dns"
)

func init() {
	logrus.SetLevel(logrus.FatalLevel)
}

func BenchmarkDNS(b *testing.B) {
	var err error
	//最初に長さを決める
	b.ResetTimer()
	s := FromBook(book_test.ReadBook(b, "./config-sample.json"))
	s.Start()
	defer s.Stop()

	m := new(dns.Msg)
	c := new(dns.Client)
	for i := 0; i < 100; i++ {
		<-time.After(1 * time.Microsecond)
		m.SetQuestion("aoba.eagle-jump.", dns.TypeA)
		_, _, err = c.Exchange(m, "127.0.0.1:20000")
		if err != nil {
			break
		}
	}
	if err != nil {
		b.Fatal(err)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		m.SetQuestion("google.com.", dns.TypeA)
		r, _, err := c.Exchange(m, "127.0.0.1:20000")
		if err != nil {
			b.Fatal(err)
		}
		if r.Rcode != dns.RcodeSuccess {
			b.Fatal(r)
		}
	}

}

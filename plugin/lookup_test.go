package plugin

import (
	"context"
	"fmt"
	"github.com/coredns/coredns/plugin/pkg/dnstest"
	"github.com/coredns/coredns/plugin/test"
	"github.com/miekg/dns"
	"github.com/rverst/coredns-redis"
	"github.com/rverst/coredns-redis/record"
	"net"
	"testing"
	"time"
)

//todo: mock redis for testing

const (
	prefix, suffix = "lookup-test_", "_lookup-test"
	defaultTtl     = 500
	testTtl        = 4242
	txt            = "Lamas, seekers, and great monkeys will always protect them."
	wcTxt          = "This is a wildcard TXT record"
)

var zones = []string{"example.net", "example.org"}

type testRecord struct {
	l string
	r record.Record
}

var testRecords = []testRecord{
	{"@", record.A{Ttl: testTtl, Ip: net.ParseIP("93.184.216.34")}},
	{"www", record.A{Ttl: testTtl, Ip: net.ParseIP("93.184.216.34")}},
	{"@", record.AAAA{Ttl: testTtl, Ip: net.ParseIP("2606:2800:220:1:248:1893:25c8:1946")}},
	{"www", record.AAAA{Ttl: testTtl, Ip: net.ParseIP("2606:2800:220:1:248:1893:25c8:1946")}},
	{"@", record.TXT{Ttl: testTtl, Text: txt}},
	{"@", record.NS{Ttl: testTtl, Host: "ns1.example.org."}},
	{"@", record.NS{Ttl: testTtl, Host: "ns2.example.org."}},
	{"@", record.NS{Ttl: testTtl, Host: "ns3"}},
	{"@", record.MX{Ttl: testTtl, Host: "mail.example.org.", Preference: 10}},
	{"wwx", record.CNAME{Ttl: testTtl, Host: "www.example.org."}},
	{"_autodiscover._tcp", record.SRV{Ttl: testTtl, Priority: 10, Weight: 50, Port: 443, Target: "mail.example.org."}},
	{"ns1", record.A{Ttl: testTtl, Ip: net.ParseIP("93.184.216.36")}},
	{"ns2", record.A{Ttl: testTtl, Ip: net.ParseIP("93.184.216.37")}},
	{"mail", record.A{Ttl: testTtl, Ip: net.ParseIP("93.184.216.38")}},
	{"mail", record.AAAA{Ttl: testTtl, Ip: net.ParseIP("2606:2800:220:1:248:1893:25c8:1947")}},
	{"@", record.CAA{Ttl: testTtl, Flag: 0, Tag: "issue", Value: "letsencrypt.org"}},
	{"lb", record.A{Ttl: testTtl, Ip: net.ParseIP("93.184.216.39")}},
	{"lb", record.A{Ttl: testTtl, Ip: net.ParseIP("93.184.216.40")}},
	{"lb", record.A{Ttl: testTtl, Ip: net.ParseIP("93.184.216.41")}},
	{"lb", record.A{Ttl: testTtl, Ip: net.ParseIP("93.184.216.42")}},
	{"*", record.A{Ttl: testTtl, Ip: net.ParseIP("93.184.216.43")}},
	{"*", record.TXT{Ttl: testTtl, Text: wcTxt}},
}

func newRedisPlugin() (*Plugin, error) {
	ctxt = context.TODO()

	p := new(Plugin)
	p.Redis = redis.New()
	p.Redis.SetKeyPrefix(prefix)
	p.Redis.SetKeySuffix(suffix)
	p.Redis.SetDefaultTtl(defaultTtl)
	p.Redis.SetAddress("192.168.0.100:6379")
	err := p.Redis.Connect()
	return p, err
}

func TestPlugin_Lookup(t *testing.T) {

	plug, err := newRedisPlugin()
	if err != nil {
		t.Fatal(err)
	}

	for _, z := range zones {
		zone := record.NewZone(z, record.SOA{
			Ttl:     testTtl,
			MName:   "ns1." + z + ".",
			RName:   "hostmaster",
			Serial:  2006010201,
			Refresh: 3600,
			Retry:   1800,
			Expire:  10000,
			MinTtl:  300,
		})

		for _, tr := range testRecords {
			zone.Add(tr.l, tr.r)
		}

		err := plug.Redis.SaveZone(*zone)
		if err != nil {
			t.Error(err)
		}
	}

	tests := []struct {
		name string
		tc   test.Case
	}{
		{name: "example.net. IN SOA", tc: test.Case{Qname: "example.net.", Qtype: dns.TypeSOA,
			Answer: []dns.RR{test.SOA("example.net. 4242 IN SOA ns1.example.net. hostmaster.example.net 2006010201 3600 1800 10000 300")},
		}},
		{name: "example.net. IN A", tc: test.Case{Qname: "example.net.", Qtype: dns.TypeA,
			Answer: []dns.RR{test.A("example.net. 4242 IN A 93.184.216.34")},
		}},
		{name: "www.example.net. IN A", tc: test.Case{Qname: "www.example.net.", Qtype: dns.TypeA,
			Answer: []dns.RR{test.A("www.example.net. 4242 IN A 93.184.216.34")},
		}},
		{name: "example.net. IN AAAA", tc: test.Case{Qname: "example.net.", Qtype: dns.TypeAAAA,
			Answer: []dns.RR{test.AAAA("example.net. 4242 IN AAAA 2606:2800:220:1:248:1893:25c8:1946")},
		}},
		{name: "www.example.net. IN AAAA", tc: test.Case{Qname: "www.example.net.", Qtype: dns.TypeAAAA,
			Answer: []dns.RR{test.AAAA("www.example.net. 4242 IN AAAA 2606:2800:220:1:248:1893:25c8:1946")},
		}},
		{name: "example.net. IN TXT", tc: test.Case{Qname: "example.net.", Qtype: dns.TypeTXT,
			Answer: []dns.RR{test.TXT("example.net. 4242 IN TXT \"" + txt + "\"")},
		}},
		{name: "wwx.example.net. IN CNAME", tc: test.Case{Qname: "wwx.example.net.", Qtype: dns.TypeCNAME,
			Answer: []dns.RR{test.CNAME("wwx.example.net. 4242 IN CNAME www.example.org.")},
		}},
		{name: "example.net. IN NS", tc: test.Case{Qname: "example.net.", Qtype: dns.TypeNS,
			Answer: []dns.RR{
				test.NS("example.net. 4242 IN NS ns1.example.org."),
				test.NS("example.net. 4242 IN NS ns2.example.org."),
				test.NS("example.net. 4242 IN NS ns3.example.net.")},
			Extra: []dns.RR{
				test.A("ns1.example.org. 4242 IN A 93.184.216.36"),
				test.A("ns2.example.org. 4242 IN A 93.184.216.37"),
				test.A("ns3.example.net. 4242 IN A 93.184.216.43")},
		}},
		{
			name:
			"example.net. IN MX", tc: test.Case{Qname: "example.net.", Qtype: dns.TypeMX,
			Answer: []dns.RR{test.MX("example.net. 4242 IN MX 10 mail.example.org. ")},
			Extra: []dns.RR{test.A("mail.example.org. 4242 IN A 93.184.216.38"),
				test.AAAA("mail.example.org. 4242 IN AAAA 2606:2800:220:1:248:1893:25c8:1947")},
		},
		},
		{
			name:
			"_autodiscover._tcp.example.net. IN SRV", tc: test.Case{Qname: "_autodiscover._tcp.example.net.", Qtype: dns.TypeSRV,
			Answer: []dns.RR{test.SRV("_autodiscover._tcp.example.net. 4242 IN SRV 10 50 443 mail.example.org. ")},
			Extra: []dns.RR{test.A("mail.example.org. 4242 IN A 93.184.216.38"),
				test.AAAA("mail.example.org. 4242 IN AAAA 2606:2800:220:1:248:1893:25c8:1947")},
		},
		},
		{
			name:
			"example.net. IN CAA", tc: test.Case{Qname: "example.net.", Qtype: dns.TypeCAA,
			Answer: []dns.RR{&dns.CAA{
				Hdr:  dns.RR_Header{Name: "example.net.", Rrtype: dns.TypeCAA},
				Flag: 0, Tag: "issue", Value: "letsencrypt.org",
			}},
		},
		},
		{
			name:
			"lb.example.net. IN A", tc: test.Case{Qname: "lb.example.net.", Qtype: dns.TypeA,
			Answer: []dns.RR{test.A("lb.example.net. 4242 IN A 93.184.216.39"),
				test.A("lb.example.net. 4242 IN A 93.184.216.40"),
				test.A("lb.example.net. 4242 IN A 93.184.216.41"),
				test.A("lb.example.net. 4242 IN A 93.184.216.42")},
		},
		},
		{
			name:
			"wildcard wc1.example.net. IN A", tc: test.Case{Qname: "wc1.example.net.", Qtype: dns.TypeA,
			Answer: []dns.RR{test.A("wc1.example.net. 4242 IN A 93.184.216.43")},
		},
		},
		{
			name:
			"wildcard wc1.example.net. IN TXT", tc: test.Case{Qname: "wc1.example.net.", Qtype: dns.TypeTXT,
			Answer: []dns.RR{test.TXT("wc1.example.net. 4242 IN TXT \"" + wcTxt + "\"")},
		},
		},
		{
			name:
			"wildcard a.b.c.d.example.net. IN A", tc: test.Case{Qname: "a.b.c.d.example.net.", Qtype: dns.TypeA,
			Answer: []dns.RR{test.A("a.b.c.d.example.net. 4242 IN A 93.184.216.43")},
		},
		},
		{
			name:
			"wildcard x.y.z.example.net. IN TXT", tc: test.Case{Qname: "x.y.z.example.net.", Qtype: dns.TypeTXT,
			Answer: []dns.RR{test.TXT("x.y.z.example.net. 4242 IN TXT \"" + wcTxt + "\"")},
		},
		},
		{
			name:
			"not existing not.*.example.net. IN A", tc: test.Case{Qname: "not.*.example.net.", Qtype: dns.TypeA,
			Rcode: dns.RcodeNameError},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			m := tt.tc.Msg()
			recorder := dnstest.NewRecorder(&test.ResponseWriter{})
			_, _ = plug.ServeDNS(ctxt, recorder, m)

			res := recorder.Msg
			// todo: FIX, should not happen
			if res == nil {
				res = new(dns.Msg)
			}

			err := test.SortAndCheck(res, tt.tc)
			if err != nil {
				t.Error(err)
			}
		})
	}

}

func TestPlugin_Lookup2(t *testing.T) {

	plug, err := newRedisPlugin()
	if err != nil {
		t.Fatal(err)
	}

	for i:=0;i<5;i++ {
		fmt.Println("shutdown redis backend for log testing")
		time.Sleep(time.Second)
	}


	for _, z := range zones {
		zone := record.NewZone(z, record.SOA{
			Ttl:     testTtl,
			MName:   "ns1." + z + ".",
			RName:   "hostmaster",
			Serial:  2006010201,
			Refresh: 3600,
			Retry:   1800,
			Expire:  10000,
			MinTtl:  300,
		})

		for _, tr := range testRecords {
			zone.Add(tr.l, tr.r)
		}

		err := plug.Redis.SaveZone(*zone)
		if err != nil {
			t.Error(err)
		}
	}

	tests := []struct {
		name string
		tc   test.Case
	}{
		{name: "example.net. IN SOA 1", tc: test.Case{Qname: "example.net.", Qtype: dns.TypeSOA,
			Answer: []dns.RR{test.SOA("example.net. 4242 IN SOA ns1.example.net. hostmaster.example.net 2006010201 3600 1800 10000 300")},
		}},
		{name: "example.net. IN SOA 2", tc: test.Case{Qname: "example.net.", Qtype: dns.TypeSOA,
			Answer: []dns.RR{test.SOA("example.net. 4242 IN SOA ns1.example.net. hostmaster.example.net 2006010201 3600 1800 10000 300")},
		}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			m := tt.tc.Msg()
			recorder := dnstest.NewRecorder(&test.ResponseWriter{})
			_, _ = plug.ServeDNS(ctxt, recorder, m)

			res := recorder.Msg
			// todo: FIX, should not happen
			if res == nil {
				res = new(dns.Msg)
			}

			err := test.SortAndCheck(res, tt.tc)
			if err != nil {
				t.Error(err)
			}
			for i:=0;i<5;i++ {
				fmt.Println("shutdown redis backend for log testing")
				time.Sleep(time.Second)
			}
		})
	}

}



var ctxt context.Context

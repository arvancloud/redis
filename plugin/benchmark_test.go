package plugin

import (
	"github.com/rverst/coredns-redis/record"
	"math/rand"
	"net"
	"testing"

	"github.com/coredns/coredns/plugin/pkg/dnstest"
	"github.com/coredns/coredns/plugin/test"

	"github.com/miekg/dns"
)

var benchmarkRecordsA = []testRecord{
	{"@", record.A{Ttl: testTtl, Ip: net.ParseIP("1.1.1.1")}},
	{"x", record.A{Ttl: testTtl, Ip: net.ParseIP("2.2.2.2")}},
	{"y", record.A{Ttl: testTtl, Ip: net.ParseIP("3.3.3.3")}},
	{"z", record.A{Ttl: testTtl, Ip: net.ParseIP("4.4.4.4")}},
}

var benchmarkRecordsMX = []testRecord{
	{"mail1", record.MX{Ttl: testTtl, Host: "x.example.net.", Preference: 10}},
	{"mail1", record.MX{Ttl: testTtl, Host: "y.example.net.", Preference: 20}},
	{"mail2", record.MX{Ttl: testTtl, Host: "x.example.net.", Preference: 10}},
	{"mail2", record.MX{Ttl: testTtl, Host: "z.example.net.", Preference: 20}},
}

func BenchmarkPlugin_HitA(b *testing.B) {
	plug, err := newRedisPlugin()
	if err != nil {
		b.Fatal(err)
	}

	zone := record.NewZone(zones[0], record.SOA{
		Ttl:     testTtl,
		MName:   "ns1." + zones[0] + ".",
		RName:   "hostmaster." + zones[0],
		Serial:  2006010201,
		Refresh: 3600,
		Retry:   1800,
		Expire:  10000,
		MinTtl:  300,
	})

	for _, tr := range benchmarkRecordsA {
		zone.Add(tr.l, tr.r)
	}

	err = plug.Redis.SaveZone(*zone)
	if err != nil {
		b.Error(err)
	}

	tests := []struct {
		name string
		tc   test.Case
	}{
		{name: "example.net. IN A", tc: test.Case{Qname: "example.net.", Qtype: dns.TypeA,
			Answer: []dns.RR{test.A("example.net. 4242 IN A 1.1.1.1")},
		}},
		{name: "x.example.net. IN A", tc: test.Case{Qname: "x.example.net.", Qtype: dns.TypeA,
			Answer: []dns.RR{test.A("x.example.net. 4242 IN A 2.2.2.2")},
		}},
		{name: "y.example.net. IN A", tc: test.Case{Qname: "y.example.net.", Qtype: dns.TypeA,
			Answer: []dns.RR{test.A("y.example.net. 4242 IN A 3.3.3.3")},
		}},
		{name: "z.example.net. IN A", tc: test.Case{Qname: "z.example.net.", Qtype: dns.TypeA,
			Answer: []dns.RR{test.A("z.example.net. 4242 IN A 4.4.4.4")},
		}},
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		j := rand.Intn(len(tests))
		m := tests[j].tc.Msg()
		rec := dnstest.NewRecorder(&test.ResponseWriter{})
		_, _ = plug.ServeDNS(ctxt, rec, m)
	}
}

func BenchmarkPlugin_MissA(b *testing.B) {

	plug, err := newRedisPlugin()
	if err != nil {
		b.Fatal(err)
	}

	zone := record.NewZone(zones[0], record.SOA{
		Ttl:     testTtl,
		MName:   "ns1." + zones[0] + ".",
		RName:   "hostmaster." + zones[0],
		Serial:  2006010201,
		Refresh: 3600,
		Retry:   1800,
		Expire:  10000,
		MinTtl:  300,
	})

	for _, tr := range benchmarkRecordsA {
		zone.Add(tr.l, tr.r)
	}

	err = plug.Redis.SaveZone(*zone)
	if err != nil {
		b.Error(err)
	}

	tests := []struct {
		name string
		tc   test.Case
	}{
		{name: "h.example.net. IN A", tc: test.Case{Qname: "h.example.net.", Qtype: dns.TypeA, Rcode: dns.RcodeNameError}},
		{name: "j.example.net. IN A", tc: test.Case{Qname: "j.example.net.", Qtype: dns.TypeA, Rcode: dns.RcodeNameError}},
		{name: "k.example.net. IN A", tc: test.Case{Qname: "k.example.net.", Qtype: dns.TypeA, Rcode: dns.RcodeNameError}},
		{name: "l.example.net. IN A", tc: test.Case{Qname: "l.example.net.", Qtype: dns.TypeA, Rcode: dns.RcodeNameError}},
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		j := rand.Intn(len(tests))
		m := tests[j].tc.Msg()
		rec := dnstest.NewRecorder(&test.ResponseWriter{})
		_, _ = plug.ServeDNS(ctxt, rec, m)
	}
}

func BenchmarkPlugin_HitMX(b *testing.B) {

	plug, err := newRedisPlugin()
	if err != nil {
		b.Fatal(err)
	}

	zone := record.NewZone(zones[0], record.SOA{
		Ttl:     testTtl,
		MName:   "ns1." + zones[0] + ".",
		RName:   "hostmaster." + zones[0],
		Serial:  2006010201,
		Refresh: 3600,
		Retry:   1800,
		Expire:  10000,
		MinTtl:  300,
	})

	for _, tr := range benchmarkRecordsA {
		zone.Add(tr.l, tr.r)
	}
	for _, tr := range benchmarkRecordsMX {
		zone.Add(tr.l, tr.r)
	}

	err = plug.Redis.SaveZone(*zone)
	if err != nil {
		b.Error(err)
	}

	tests := []struct {
		name string
		tc   test.Case
	}{
		{name: "mail1.example.net. IN MX", tc: test.Case{Qname: "mail1.example.net.", Qtype: dns.TypeMX,
			Answer: []dns.RR{test.MX("mail1.example.net. 4242 IN MX 10 x.example.net. ")},
			Extra: []dns.RR{
				test.A("x.example.net. 4242 IN A 2.2.2.2"),
				test.A("y.example.net. 4242 IN A 3.3.3.3")}}},
		{name: "mail2.example.net. IN MX", tc: test.Case{Qname: "mail2.example.net.", Qtype: dns.TypeMX,
			Answer: []dns.RR{test.MX("mail1.example.net. 4242 IN MX 10 x.example.net. ")},
			Extra: []dns.RR{
				test.A("x.example.net. 4242 IN A 2.2.2.2"),
				test.A("z.example.net. 4242 IN A 4.4.4.4")}}},
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		j := rand.Intn(len(tests))
		m := tests[j].tc.Msg()
		rec := dnstest.NewRecorder(&test.ResponseWriter{})
		_, _ = plug.ServeDNS(ctxt, rec, m)
	}
}

func BenchmarkPlugin_MissMX(b *testing.B) {

	plug, err := newRedisPlugin()
	if err != nil {
		b.Fatal(err)
	}

	zone := record.NewZone(zones[0], record.SOA{
		Ttl:     testTtl,
		MName:   "ns1." + zones[0] + ".",
		RName:   "hostmaster." + zones[0],
		Serial:  2006010201,
		Refresh: 3600,
		Retry:   1800,
		Expire:  10000,
		MinTtl:  300,
	})

	for _, tr := range benchmarkRecordsA {
		zone.Add(tr.l, tr.r)
	}
	for _, tr := range benchmarkRecordsMX {
		zone.Add(tr.l, tr.r)
	}

	err = plug.Redis.SaveZone(*zone)
	if err != nil {
		b.Error(err)
	}

	tests := []struct {
		name string
		tc   test.Case
	}{
		{name: "mail1.example.net. IN MX", tc: test.Case{Qname: "mail1.example.net.", Qtype: dns.TypeA, Rcode: dns.RcodeNameError}},
		{name: "mail2.example.net. IN MX", tc: test.Case{Qname: "mail2.example.net.", Qtype: dns.TypeA, Rcode: dns.RcodeNameError}},
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		m := tests[0].tc.Msg()
		rec := dnstest.NewRecorder(&test.ResponseWriter{})
		_, _ = plug.ServeDNS(ctxt, rec, m)
	}
}

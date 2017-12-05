package redis

import (
	"context"
	"testing"
	"fmt"

	"github.com/coredns/coredns/plugin/pkg/dnstest"
	"github.com/coredns/coredns/plugin/test"

	redisCon "github.com/garyburd/redigo/redis"
	"github.com/miekg/dns"
)

var entries = [][][]string {
	// basic tests
	{
		{"example.com.",
			"SOA", "[{\"ttl\":100, \"mbox\":\"hostmaster.example.com.\",\"ns\":\"ns1.example.com.\",\"refresh\":44,\"retry\":55,\"expire\":66}]",
		},
		{"x.example.com.",
			"A", "[{\"ip4\":\"1.2.3.4\"},{\"ip4\":\"5.6.7.8\"}]",
			"AAAA", "[{\"ip6\":\"::1\"}]",
			"TXT", "[{\"text\":\"foo\"},{\"text\":\"bar\"}]",
			"NS", "[{\"host\":\"ns1.example.com.\"},{\"host\":\"ns2.example.com.\"}]",
			"MX", "[{\"host\":\"mx1.example.com.\", \"priority\":10},{\"host\":\"mx2.example.com.\", \"priority\":10}]"},
		{"y.example.com.",
			"CNAME", "[{\"host\":\"x.example.com.\"}]",
		},
		{"ns1.example.com.",
			"A", "[{\"ip4\":\"2.2.2.2\"}]"},
		{"ns2.example.com.",
			"A", "[{\"ip4\":\"3.3.3.3\"}]"},
		{"_sip._tcp.x.example.com.",
			"SRV", "[{\"host\":\"sip.example.com.\",\"port\":555,\"priority\":10,\"weight\":100}]",
		},
		{"sip.example.com.",
			"A", "[{\"ip4\":\"7.7.7.7\"}]",
			"AAAA", "[{\"ip6\":\"::1\"}]",
		},
	},
	// wildcard tests
	{
		{"example.com.",
			"SOA", "[{\"ttl\":100, \"mbox\":\"hostmaster.example.com.\",\"ns\":\"ns1.example.com.\",\"refresh\":44,\"retry\":55,\"expire\":66}]",
			"NS", "[{\"host\":\"ns1.example.com.\"},{\"host\":\"ns2.example.com.\"}]"},
		{"sub.*.example.com.",
			"TXT", "[{\"text\":\"this is not a wildcard\"}]"},
		{"host1.example.com.",
			"A", "[{\"ip4\":\"5.5.5.5\"}]"},
		{"subdel.example.com.",
			"NS", "[{\"host\":\"ns1.subdel.example.com.\"},{\"host\":\"ns2.subdel.example.com.\"}]"},
		{"*.example.com.",
			"TXT", "[{\"text\":\"this is a wildcard\"}]",
			"MX", "[{\"host\":\"host1.example.com.\",\"priority\": 10}]"},
	},
}

var testCases = [][]test.Case{
	// basic tests
	{
		// A Test
		{
			Qname: "x.example.com.", Qtype: dns.TypeA,
			Answer: []dns.RR{
				test.A("x.example.com. 300 IN A 1.2.3.4"),
				test.A("x.example.com. 300 IN A 5.6.7.8"),
			},
		},
		// AAAA Test
		{
			Qname: "x.example.com.", Qtype: dns.TypeAAAA,
			Answer: []dns.RR{
				test.AAAA("x.example.com. 300 IN AAAA ::1"),
			},
		},
		// TXT Test
		{
			Qname: "x.example.com.", Qtype: dns.TypeTXT,
			Answer: []dns.RR{
				test.TXT("x.example.com. 300 IN TXT bar"),
				test.TXT("x.example.com. 300 IN TXT foo"),
			},
		},
		// CNAME Test
		{
			Qname: "y.example.com.", Qtype: dns.TypeCNAME,
			Answer: []dns.RR{
				test.CNAME("y.example.com. 300 IN CNAME x.example.com."),
			},
		},
		// NS Test
		{
			Qname: "x.example.com.", Qtype: dns.TypeNS,
			Answer: []dns.RR{
				test.NS("x.example.com. 300 IN NS ns1.example.com."),
				test.NS("x.example.com. 300 IN NS ns2.example.com."),
			},
			Extra: []dns.RR{
				test.A("ns1.example.com. 300 IN A 2.2.2.2"),
				test.A("ns2.example.com. 300 IN A 3.3.3.3"),
			},
		},
		// MX Test
		{
			Qname: "x.example.com.", Qtype: dns.TypeMX,
			Answer: []dns.RR{
				test.MX("x.example.com. 300 IN MX 10 mx1.example.com."),
				test.MX("x.example.com. 300 IN MX 10 mx2.example.com."),
			},
		},
		// SRV Test
		{
			Qname: "_sip._tcp.x.example.com.", Qtype: dns.TypeSRV,
			Answer: []dns.RR{
				test.SRV("_sip._tcp.x.example.com. 300 IN SRV 10 100 555 sip.example.com."),
			},
			Extra: []dns.RR{
				test.A("sip.example.com. 300 IN A 7.7.7.7"),
				test.AAAA("sip.example.com 300 IN AAAA ::1"),
			},
		},
		// NXDOMAIN Test
		{
			Qname: "notexists.example.com.", Qtype: dns.TypeA,
			Rcode: dns.RcodeNameError,
		},
		// SOA Test
		{
			Qname: "example.com.", Qtype: dns.TypeSOA,
			Answer: []dns.RR{
				test.SOA("example.com. 100 IN SOA ns1.example.com. hostmaster.example.com. 1460498836 44 55 66 100"),
			},
		},
	},
	// Wildcard Tests
	{
		{
			Qname: "host3.example.com.", Qtype: dns.TypeMX,
			Answer: []dns.RR{
				test.MX("host3.example.com. 300 IN MX 10 host1.example.com."),
			},
			Extra: []dns.RR{
				test.A("host1.example.com. 300 IN A 5.5.5.5"),
			},
		},
		{
			Qname: "host3.example.com.", Qtype: dns.TypeA,
		},
		{
			Qname: "foo.bar.example.com.", Qtype: dns.TypeTXT,
			Answer: []dns.RR{
				test.TXT("foo.bar.example.com. 300 IN TXT \"this is a wildcard\""),
			},
		},
		{
			Qname: "host1.example.com.", Qtype: dns.TypeMX,
		},
		{
			Qname: "sub.*.example.com.", Qtype: dns.TypeMX,
		},
		{
			Qname: "host.subdel.example.com.", Qtype: dns.TypeA,
			Rcode: dns.RcodeNameError,
		},
		{
			Qname: "ghost.*.example.com.", Qtype: dns.TypeMX,
			Rcode: dns.RcodeNameError,
		},
	},
}

func newRedisPlugin() *Redis {
	ctxt = context.TODO()

	opts := []redisCon.DialOption{}
	opts = append(opts, redisCon.DialPassword("foobared"))
	client, _ := redisCon.Dial("tcp", "localhost:6379", opts...)
	return &Redis {
		Zones: []string{"example.com."},
		redisc: client,
		Ttl: 300,
	}
}

func TestAnswer(t *testing.T) {
	r := newRedisPlugin()

	for i, _ := range entries {
		r.redisc.Do("EVAL", "return redis.call('del', unpack(redis.call('keys', ARGV[1])))", 0, "*.example.com.")
		for _, cmd := range entries[i] {
			err := r.set(cmd)
			if err != nil {
				fmt.Println("error in redis", err)
				t.Fail()
			}
		}
		for _, tc := range testCases[i] {
			m := tc.Msg()

			rec := dnstest.NewRecorder(&test.ResponseWriter{})
			r.ServeDNS(ctxt, rec, m)

			resp := rec.Msg

			// TODO(arash): this shouldn't happen, check plugin's empty response
			if resp == nil {
				resp = new(dns.Msg)
			}
			test.SortAndCheck(t, resp, tc)
		}
	}
}

var ctxt context.Context

package redis

import (
	"fmt"
	"math/rand"
	"testing"

	"github.com/coredns/coredns/plugin/pkg/dnstest"
	"github.com/coredns/coredns/plugin/test"

	"github.com/miekg/dns"
)

var zone string = "example.com."

var benchmarkEntries = [][]string{
	{"@",
		"{\"a\":[{\"ttl\":300, \"ip\":\"2.2.2.2\"}]}",
	},
	{"x",
		"{\"a\":[{\"ttl\":300, \"ip\":\"3.3.3.3\"}]}",
	},
	{"y",
		"{\"a\":[{\"ttl\":300, \"ip\":\"4.4.4.4\"}]}",
	},
	{"z",
		"{\"a\":[{\"ttl\":300, \"ip\":\"5.5.5.5\"}]}",
	},
}

var testCasesHit = []test.Case{
	{
		Qname: "example.com.", Qtype: dns.TypeA,
		Answer: []dns.RR{
			test.A("example.com. 300 IN A 2.2.2.2"),
		},
	},
	{
		Qname: "x.example.com.", Qtype: dns.TypeA,
		Answer: []dns.RR{
			test.A("x.example.com. 300 IN A 3.3.3.3"),
		},
	},
	{
		Qname: "y.example.com.", Qtype: dns.TypeA,
		Answer: []dns.RR{
			test.A("y.example.com. 300 IN A 4.4.4.4"),
		},
	},
	{
		Qname: "z.example.com.", Qtype: dns.TypeA,
		Answer: []dns.RR{
			test.A("z.example.com. 300 IN A 5.5.5.5"),
		},
	},
}

var testCasesMiss = []test.Case{
	{
		Qname: "q.example.com.", Qtype: dns.TypeA,
		Rcode: dns.RcodeNameError,
	},
	{
		Qname: "w.example.com.", Qtype: dns.TypeA,
		Rcode: dns.RcodeNameError,
	},
	{
		Qname: "e.example.com.", Qtype: dns.TypeA,
		Rcode: dns.RcodeNameError,
	},
	{
		Qname: "r.example.com.", Qtype: dns.TypeA,
		Rcode: dns.RcodeNameError,
	},
}

func BenchmarkHit(b *testing.B) {
	fmt.Println("benchmark test")
	r := newRedisPlugin()
	conn := r.Pool.Get()
	defer conn.Close()
	conn.Do("EVAL", "return redis.call('del', unpack(redis.call('keys', ARGV[1])))", 0, r.keyPrefix+"*"+r.keySuffix)
	for _, cmd := range benchmarkEntries {
		err := r.save(zone, cmd[0], cmd[1])
		if err != nil {
			fmt.Println("error in redis", err)
		}
	}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		j := rand.Intn(len(testCasesHit))
		m := testCasesHit[j].Msg()
		rec := dnstest.NewRecorder(&test.ResponseWriter{})
		r.ServeDNS(ctxt, rec, m)
	}
}

func BenchmarkMiss(b *testing.B) {
	fmt.Println("benchmark test")
	r := newRedisPlugin()
	conn := r.Pool.Get()
	defer conn.Close()
	conn.Do("EVAL", "return redis.call('del', unpack(redis.call('keys', ARGV[1])))", 0, r.keyPrefix+"*"+r.keySuffix)
	for _, cmd := range benchmarkEntries {
		err := r.save(zone, cmd[0], cmd[1])
		if err != nil {
			fmt.Println("error in redis", err)
		}
	}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		j := rand.Intn(len(testCasesMiss))
		m := testCasesMiss[j].Msg()
		rec := dnstest.NewRecorder(&test.ResponseWriter{})
		r.ServeDNS(ctxt, rec, m)
	}
}

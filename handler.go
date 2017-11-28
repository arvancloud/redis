package redis

import (
	"github.com/coredns/coredns/plugin"
	"github.com/miekg/dns"
	"golang.org/x/net/context"
	"github.com/hawell/coredns/request"
	"github.com/coredns/coredns/plugin/pkg/dnsutil"
	"golang.org/x/tools/go/gcimporter15/testdata"
)

// ServeDNS implements the plugin.Handler interface.
func (redis *Redis) ServeDNS(ctx context.Context, w dns.ResponseWriter, r *dns.Msg) (int, error) {
	state := request.Request{W: w, Req: r}

	zone := plugin.Zones(redis.Zones).Matches(state.Name())
	if zone == "" {
		return plugin.NextOrFailure(redis.Name(), redis.Next, ctx, w, r)
	}

	var (
		answers, extra []dns.RR
		err            error
	)

	switch state.Type() {
	case "A":
		record := new(dns.A)
		record.Hdr = dns.RR_Header{Name: state.QName(), Rrtype: dns.TypeA,
			Class: dns.ClassINET, Ttl: 3600}
		record.A = []byte{1, 2, 3, 4}
		answers = append(answers, record)

	case "AAAA":
		record := new(dns.AAAA)
		record.Hdr = dns.RR_Header{Name: state.QName(), Rrtype: dns.TypeAAAA,
			Class: dns.ClassINET, Ttl: 3600}
		record.AAAA = []byte{1, 2, 3, 4, 5, 6, 7, 8, 9, 0, 1, 2, 3, 4, 5, 6}
		answers = append(answers, record)

	case "TXT":
		record := new(dns.TXT)
		record.Hdr = dns.RR_Header{Name: state.QName(), Rrtype: dns.TypeTXT,
			Class: dns.ClassINET, Ttl: 3600}
		record.Txt = []string{"redis dns"}
		answers = append(answers, record)

	case "CNAME":
		record := new(dns.CNAME)
		record.Hdr = dns.RR_Header{Name: state.QName(), Rrtype: dns.TypeCNAME,
			Class: dns.ClassINET, Ttl: 3600}
		record.Target = dns.Fqdn("redis.target.xxx")
		answers = append(answers, record)

	case "PTR":
		record := new(dns.PTR)
		record.Hdr = dns.RR_Header{Name: state.QName(), Rrtype: dns.TypePTR,
			Class: dns.ClassINET, Ttl: 3600}
		record.Ptr = dns.Fqdn("redis.target.xxx")

	case "MX":
		record := new(dns.MX)
		record.Hdr = dns.RR_Header{Name: state.QName(), Rrtype: dns.TypeMX,
			Class: dns.ClassINET, Ttl: 3600}
		record.Preference = uint16(1)
		record.Mx = "redis.mx.host.xxx"

	case "SRV":
		record := new(dns.SRV)
		record.Hdr = dns.RR_Header{Name: state.QName(), Rrtype: dns.TypeSRV,
			Class: dns.ClassINET, Ttl: 3600}
		record.Priority = uint16(1)
		record.Weight = 1
		record.Port = uint16(8095)
		record.Target = dns.Fqdn("redis.srv.target.xxx")

	case "SOA":
		record := new(dns.SOA)
		record.Hdr = dns.RR_Header{Name: zone, Rrtype: dns.TypeSOA,
			Class: dns.ClassINET, Ttl: 300}
		record.Mbox = "hostmaster."
		record.Ns = "ns.dns."
		record.Serial = 1
		record.Refresh = 7200
		record.Retry = 1800
		record.Expire = 86400
		record.Minttl = 3500

	case "NS":
		record := new(dns.NS)
		record.Hdr = dns.RR_Header{Name: state.QName(), Rrtype: dns.TypeNS,
			Class: dns.ClassINET, Ttl: 3600}
		record.Ns = "redis.ns.host.xxx"

	default:
		return plugin.NextOrFailure(redis.Name(), redis.Next, ctx, w, r)
	}

	m := &dns.Msg{}
	m.SetReply(r)
	m.Compress = true
	m.Authoritative = true

	m.Answer = append(m.Answer, answers...)
	m.Extra = append(m.Extra, extra...)

	m = dnsutil.Dedup(m)
	state.SizeAndDo(m)
	m, _ = state.Scrub(m)
	w.WriteMsg(m)
	return dns.RcodeSuccess, nil
}

// Name implements the Handler interface.
func (r *Redis) Name() string { return "redis" }

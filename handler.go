package redis

import (
	"github.com/coredns/coredns/plugin"
	"github.com/miekg/dns"
	"golang.org/x/net/context"
	"github.com/coredns/coredns/request"
	"github.com/coredns/coredns/plugin/pkg/dnsutil"
)

// ServeDNS implements the plugin.Handler interface.
func (redis *Redis) ServeDNS(ctx context.Context, w dns.ResponseWriter, r *dns.Msg) (int, error) {
	// fmt.Println("serveDNS")
	state := request.Request{W: w, Req: r}

	qname := state.Name()
	qtype := state.Type()

	// fmt.Println(state.Name())
	// fmt.Println(redis.Zones)
	zone := plugin.Zones(redis.Zones).Matches(qname)
	if zone == "" {
		return plugin.NextOrFailure(qname, redis.Next, ctx, w, r)
	}

	key := redis.findKey(qname, zone)
	if key == "" { // empty, no results
		return redis.errorResponse(state, zone, dns.RcodeNameError, nil)
	}

	answers := make([]dns.RR, 0, 10)
	extras := make([]dns.RR, 0, 10)

	records := redis.get(key, qtype)
	if len(records) == 0 {
		cnameRecords := redis.get(key, "CNAME")
		for _, cnameRecord := range cnameRecords {
			if dns.IsSubDomain(zone, cnameRecord.Host) {
				records = append(records,redis.get(cnameRecord.Host, qtype)...)
			}
		}
	}

	if len(records) == 0 {
		return redis.errorResponse(state, zone, dns.RcodeSuccess, nil)
	}

	switch qtype {
	case "A":
		answers, extras = redis.A(qname, zone, records)
	case "AAAA":
		answers, extras = redis.AAAA(qname, zone, records)
	case "CNAME":
		answers, extras = redis.CNAME(qname, zone, records)
	case "TXT":
		answers, extras = redis.TXT(qname, zone, records)
	case "NS":
		answers, extras = redis.NS(qname, zone, records)
	case "MX":
		answers, extras = redis.MX(qname, zone, records)
	case "SRV":
		answers, extras = redis.SRV(qname, zone, records)
	case "SOA":
		answers, extras = redis.SOA(qname, zone, records)
	default:
		return redis.errorResponse(state, zone, dns.RcodeNotImplemented, nil)
	}

	if len(answers) == 0 {
		return redis.errorResponse(state, zone, dns.RcodeSuccess, nil)
	}

	m := new(dns.Msg)
	m.SetReply(r)
	m.Authoritative, m.RecursionAvailable, m.Compress = true, false, true

	m.Answer = append(m.Answer, answers...)
	m.Extra = append(m.Extra, extras...)

	m = dnsutil.Dedup(m)
	state.SizeAndDo(m)
	m, _ = state.Scrub(m)
	w.WriteMsg(m)
	return dns.RcodeSuccess, nil
}

// Name implements the Handler interface.
func (redis *Redis) Name() string { return "redis" }

func (redis *Redis) errorResponse(state request.Request, zone string, rcode int, err error) (int, error) {
	m := new(dns.Msg)
	m.SetRcode(state.Req, rcode)
	m.Authoritative, m.RecursionAvailable, m.Compress = true, false, true

	// m.Ns, _ = redis.SOA(state.Name(), zone, nil)

	state.SizeAndDo(m)
	state.W.WriteMsg(m)
	// Return success as the rcode to signal we have written to the client.
	return dns.RcodeSuccess, err
}
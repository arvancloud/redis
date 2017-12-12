package redis

import (
	"time"
	// "fmt"

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

	// fmt.Println("name : ", qname)
	// fmt.Println("type : ", qtype)
	zone := plugin.Zones(redis.GetZones()).Matches(qname)
	// fmt.Println("zone : ", zone)
	if zone == "" {
		return plugin.NextOrFailure(qname, redis.Next, ctx, w, r)
	}

	if time.Now().Sub(redis.LastUpdate) > time.Duration(redis.Ttl) {
		redis.load()
	}

	location := redis.findLocation(qname, zone)
	if len(location) == 0 { // empty, no results
		return redis.errorResponse(state, zone, dns.RcodeNameError, nil)
	}
	// fmt.Println("location : ", location)

	answers := make([]dns.RR, 0, 10)
	extras := make([]dns.RR, 0, 10)

	record := redis.get(location, zone)

	switch qtype {
	case "A":
		answers, extras = redis.A(qname, zone, record)
	case "AAAA":
		answers, extras = redis.AAAA(qname, zone, record)
	case "CNAME":
		answers, extras = redis.CNAME(qname, zone, record)
	case "TXT":
		answers, extras = redis.TXT(qname, zone, record)
	case "NS":
		answers, extras = redis.NS(qname, zone, record)
	case "MX":
		answers, extras = redis.MX(qname, zone, record)
	case "SRV":
		answers, extras = redis.SRV(qname, zone, record)
	case "SOA":
		answers, extras = redis.SOA(qname, zone, record)
	default:
		return redis.errorResponse(state, zone, dns.RcodeNotImplemented, nil)
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
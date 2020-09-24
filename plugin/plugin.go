package plugin

import (
	"context"
	"fmt"
	"github.com/coredns/coredns/plugin"
	"github.com/coredns/coredns/request"
	"github.com/miekg/dns"
	redis "github.com/rverst/coredns-redis"
	"github.com/rverst/coredns-redis/record"
)

const name = "redis"

type Plugin struct {
	Redis *redis.Redis
	Next  plugin.Handler
}

func (p *Plugin) Name() string {
	return name
}

func (p *Plugin) ServeDNS(ctx context.Context, w dns.ResponseWriter, r *dns.Msg) (int, error) {
	state := request.Request{Req: r, W: w}
	qName := state.Name()
	qType := state.QType()


	if qName == "" || qType == dns.TypeNone {
		return plugin.NextOrFailure(qName, p.Next, ctx, w, r)
	}

	z, err := p.Redis.LoadZones(qName)
	if err != nil {
		return plugin.NextOrFailure(qName, p.Next, ctx, w, r)
	}
	zoneName := plugin.Zones(z).Matches(qName)
	if zoneName == "" {
		return plugin.NextOrFailure(qName, p.Next, ctx, w, r)
	}

	zone := p.Redis.Load(zoneName)
	if zone == nil {
		return p.Redis.ErrorResponse(state, zoneName, dns.RcodeServerFailure, nil)
	}

	if qType == dns.TypeAXFR {
		return p.handleZoneTransfer(zone, w, r)
	}

	location := p.Redis.FindLocation(qName, zone)
	if location == "" {
		return p.Redis.ErrorResponse(state, zoneName, dns.RcodeNameError, nil)
	}

	answers := make([]dns.RR, 0, 0)
	extras := make([]dns.RR, 0, 10)
	record := p.Redis.GetZoneRecords(location, zone)

	switch qType {
	case dns.TypeA:
		answers, extras = p.Redis.A(qName, zone, record)
	case dns.TypeAAAA:
		answers, extras = p.Redis.AAAA(qName, zone, record)
	case dns.TypeCNAME:
		answers, extras = p.Redis.CNAME(qName, zone, record)
	case dns.TypeTXT:
		answers, extras = p.Redis.TXT(qName, zone, record)
	case dns.TypeNS:
		answers, extras = p.Redis.NS(qName, zone, record)
	case dns.TypeMX:
		answers, extras = p.Redis.MX(qName, zone, record)
	case dns.TypeSRV:
		answers, extras = p.Redis.SRV(qName, zone, record)
	case dns.TypeSOA:
		answers, extras = p.Redis.SOA(qName, zone, record)
	case dns.TypeCAA:
		answers, extras = p.Redis.CAA(qName, zone, record)
	default:
		return p.Redis.ErrorResponse(state, zoneName, dns.RcodeNotImplemented, nil)
	}

	m := new(dns.Msg)
	m.SetReply(r)
	m.Authoritative, m.RecursionAvailable, m.Compress = true, false, true
	m.Answer = append(m.Answer, answers...)
	m.Extra = append(m.Extra, extras...)

	state.SizeAndDo(m)
	m = state.Scrub(m)
	_ = w.WriteMsg(m)
	return dns.RcodeSuccess, nil
}

func (p *Plugin) handleZoneTransfer(zone *record.Zone, w dns.ResponseWriter, r *dns.Msg) (int, error) {
	//todo: check and test zone transfer, implement ip-range check
	records := p.Redis.AXFR(zone)
	ch := make(chan *dns.Envelope)
	tr := new(dns.Transfer)
	tr.TsigSecret = nil
	go func(ch chan *dns.Envelope) {
		j, l := 0, 0

		for i, r := range records {
			l += dns.Len(r)
			if l > redis.MaxTransferLength {
				ch <- &dns.Envelope{RR: records[j:i]}
				l = 0
				j = i
			}
		}
		if j < len(records) {
			ch <- &dns.Envelope{RR: records[j:]}
		}
		close(ch)
	}(ch)

	err := tr.Out(w, r, ch)
	if err != nil {
		fmt.Println(err)
	}
	w.Hijack()
	return dns.RcodeSuccess, nil
}


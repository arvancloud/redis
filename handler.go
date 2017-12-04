package redis

import (
	"fmt"

	"github.com/coredns/coredns/plugin"
	"github.com/miekg/dns"
	"golang.org/x/net/context"
	"github.com/coredns/coredns/request"
	"github.com/coredns/coredns/plugin/pkg/dnsutil"
)

// ServeDNS implements the plugin.Handler interface.
func (redis *Redis) ServeDNS(ctx context.Context, w dns.ResponseWriter, r *dns.Msg) (int, error) {
	fmt.Println("serveDNS")
	state := request.Request{W: w, Req: r}

	fmt.Println(state.Name())
	fmt.Println(redis.Zones)
	zone := plugin.Zones(redis.Zones).Matches(state.Name())
	if zone == "" {
		return plugin.NextOrFailure(redis.Name(), redis.Next, ctx, w, r)
	}

	answers, extras := redis.Answer(state.Name(), state.Type(), zone)

	if len(answers) == 0 {
		return plugin.NextOrFailure(redis.Name(), redis.Next, ctx, w, r)
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
func (r *Redis) Name() string { return "redis" }

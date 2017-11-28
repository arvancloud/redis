package redis

import (
	"github.com/mholt/caddy"
	"github.com/coredns/coredns/core/dnsserver"
	"github.com/coredns/coredns/plugin"
)

func init() {
	caddy.RegisterPlugin("redis", caddy.Plugin{
		ServerType: "dns",
		Action:     setup,
	})
}

func setup(c *caddy.Controller) error {
	r, err := redisParse(c)
	if err != nil {
		return plugin.Error("redis", err)
	}

	dnsserver.GetConfig(c).AddPlugin(func(next plugin.Handler) plugin.Handler {
		r.Next = next
		return r
	})

	return nil
}

func redisParse(c *caddy.Controller) (*Redis, error) {
	c.Next()
	if c.neNextArg() {
		return &Redis{}, c.ArgErr()
	}
	redis := Redis{
		Next:        nil,
		Fallthrough: false,
		Zones:       nil,
	}
	return &redis, nil
}
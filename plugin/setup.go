package plugin

import (
	"errors"
	"github.com/caddyserver/caddy"
	"github.com/coredns/coredns/core/dnsserver"
	"github.com/coredns/coredns/plugin"
	"github.com/rverst/coredns-redis"
	"strconv"
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

	}

	if ok, err := r.Ping(); err != nil {
	if ok, err := r.Ping(); err != nil || !ok {
		return plugin.Error("redis", err)
	} else if ok {
		log.Infof("ping to redis ok")
	}

	p := &Plugin{
		Redis: r,
	}

	dnsserver.GetConfig(c).AddPlugin(func(next plugin.Handler) plugin.Handler {
		p.Next = next
		return p
	})

	return nil
}

func redisParse(c *caddy.Controller) (*redis.Redis, error) {
	r := redis.New()

	for c.Next() {
		if c.NextBlock() {
			for {
				switch c.Val() {
				case "address":
					if !c.NextArg() {
						return redis.New(), c.ArgErr()
					}
					r.SetAddress(c.Val())
				case "username":
					if !c.NextArg() {
						return redis.New(), c.ArgErr()
					}
					r.SetUsername(c.Val())
				case "password":
					if !c.NextArg() {
						return redis.New(), c.ArgErr()
					}
					r.SetPassword(c.Val())
				case "prefix":
					if !c.NextArg() {
						return redis.New(), c.ArgErr()
					}
					r.SetKeyPrefix(c.Val())
				case "suffix":
					if !c.NextArg() {
						return redis.New(), c.ArgErr()
					}
					r.SetKeySuffix(c.Val())
				case "connect_timeout":
					if !c.NextArg() {
						return redis.New(), c.ArgErr()
					}
					t, err := strconv.Atoi(c.Val())
					if err == nil {
						r.SetConnectTimeout(t)
					}
				case "read_timeout":
					if !c.NextArg() {
						return redis.New(), c.ArgErr()
					}
					t, err := strconv.Atoi(c.Val())
					if err != nil {
						r.SetReadTimeout(t)
					}
				case "ttl":
					if !c.NextArg() {
						return redis.New(), c.ArgErr()
					}
					t, err := strconv.Atoi(c.Val())
					if err != nil {
						r.SetDefaultTtl(redis.DefaultTtl)
					} else {
						r.SetDefaultTtl(t)
					}
				default:
					if c.Val() != "}" {
						return redis.New(), c.Errf("unknown property '%s'", c.Val())
					}
				}

				if !c.Next() {
					break
				}
			}

		}

		err := r.Connect()
		return r, err
	}

	return nil, errors.New("no configuration found")
}

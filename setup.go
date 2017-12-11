package redis

import (
	"time"
	"strconv"

	"github.com/mholt/caddy"
	"github.com/coredns/coredns/core/dnsserver"
	"github.com/coredns/coredns/plugin"

	redisCon "github.com/garyburd/redigo/redis"
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
	redis := Redis {
		keyPrefix:"",
		keySuffix:"",
		Ttl:300,
	}
	var (
		redisAddress   string
		redisPassword  string
		connectTimeout int
		readTimeout    int
		err            error
	)

	for c.Next() {
		if c.NextBlock() {
			for {
				switch c.Val() {
				case "address":
					if !c.NextArg() {
						return &Redis{}, c.ArgErr()
					}
					redisAddress = c.Val()
				case "password":
					if !c.NextArg() {
						return &Redis{}, c.ArgErr()
					}
					redisPassword = c.Val()
				case "prefix":
					if !c.NextArg() {
						return &Redis{}, c.ArgErr()
					}
					redis.keyPrefix = c.Val()
				case "suffix":
					if !c.NextArg() {
						return &Redis{}, c.ArgErr()
					}
					redis.keySuffix = c.Val()
				case "connect_timeout":
					if !c.NextArg() {
						return &Redis{}, c.ArgErr()
					}
					connectTimeout, err = strconv.Atoi(c.Val())
					if err != nil {
						connectTimeout = 0
					}
				case "read_timeout":
					if !c.NextArg() {
						return &Redis{}, c.ArgErr()
					}
					readTimeout, err = strconv.Atoi(c.Val())
					if err != nil {
						readTimeout = 0;
					}
				case "ttl":
					if !c.NextArg() {
						return &Redis{}, c.ArgErr()
					}
					var val int
					val, err = strconv.Atoi(c.Val())
					if err != nil {
						val = defaultTtl
					}
					redis.Ttl = uint32(val)
				default:
					if c.Val() != "}" {
						return &Redis{}, c.Errf("unknown property '%s'", c.Val())
					}
				}

				if !c.Next() {
					break
				}
			}

		}
		opts := []redisCon.DialOption{}
		if redisPassword != "" {
			opts = append(opts, redisCon.DialPassword(redisPassword))
		}
		if connectTimeout != 0 {
			opts = append(opts, redisCon.DialConnectTimeout(time.Duration(connectTimeout)*time.Millisecond))
		}
		if readTimeout != 0 {
			opts = append(opts, redisCon.DialReadTimeout(time.Duration(readTimeout)*time.Millisecond))
		}

		redis.redisc, err = redisCon.Dial("tcp", redisAddress, opts...)
		if err != nil {
			return &Redis{}, err
		}

		redis.load()

		return &redis, nil
	}
	return &Redis{}, nil
}
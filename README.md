# coredns-redis

use [redis](https://redis.io/) as a backend for [coredns](https://coredns.io)  
this plugin should be located right next to *etcd* in *plugins.cfg*:  

```
...
secondary:secondary
etcd:etcd
redis:github.com/rverst/coredns-redis/plugin
loop:loop
forward:forward
grpc:grpc
...
```

## configuration

corefile
```
{
  .{
    redis {
      address localhost:6379
      username redis_user
      password super_secret
      connect_timeout 2500
      read_timeout 2000
      ttl 300
      prefix PREFX_
      suffix _SUFFX
    }
  }
}
```

- `address` is the address of the redis backend in form of *host:port*, defaults to `localhost:6379`
- `username` is the username for connectiong to the redis backend (optional)
- `password` is the redis password (optional)
- `connect_timeout` maximum time to establish a connection to the redis backend (in ms)
- `read_timeout` maximum time to wait for the redis backend to respond (in ms)
- `ttl` default ttl for dns records which have no ttl set (in seconds)
- `prefix` a prefix added to all redis keys
- `suffix` a suffix added to all redis keys

## reverse zones

not yet supported


## proxy

not yet supported

## credits

this plugin started as a fork of [github.com/arvancloud/redis](https://github.com/arvancloud/redis).


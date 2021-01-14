# coredns-redis

*coredns-redis* uses [redis](https://redis.io/) as a backend for [coredns](https://coredns.io)  
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

```
{
  redis {
    address HOST:PORT
    username USER
    password PASSWORD
    connect_timeout TIME_MS
    read_timeout TIME_MS
    ttl TIME_S
    prefix PREFIX
    suffix SUFFIX
  }
}
```

- `address` is the address of the redis backend in form of *host:port* (defaults to `localhost:6379`)
- `username` is the username for connectiong to the redis backend (optional)
- `password` is the redis password (optional)
- `connect_timeout` maximum time to establish a connection to the redis backend (in ms, optional)
- `read_timeout` maximum time to wait for the redis backend to respond (in ms, optional)
- `ttl` default ttl for dns records which have no ttl set (in seconds, default 3600)
- `prefix` a prefix added to all redis keys
- `suffix` a suffix added to all redis keys

### example

corefile:
```
{
  .{
    redis {
      address localhost:6379
      username redis_user
      password super_secret
      connect_timeout 2000
      read_timeout 2000
      ttl 300
      prefix DNS_
      suffix _DNS
    }
  }
}
```

## reverse zones

not yet supported


## proxy

not yet supported

## API

Package `redis` provides functions to manipulate (get, add, edit, delete) the data in the redis backend.
The DNS zones are saved as hashmaps with the zone-name as key in the backend.
While the data format is JSON at the moment, but I am considering switching to 
*protobuf* for performance reasons later. 

## credits

this plugin started as a fork of [github.com/arvancloud/redis](https://github.com/arvancloud/redis).


# redis

*redis* enables reading zone data from redis database.

## syntax

~~~
redis [ZONES...]
~~~

* **ZONES**  zones redis should be authoritative for.

if no zones are specified redis tries to load zones from redis server, if
no zones could be found the block's zone will be used

~~~
redis [ZONES...] {
    address ADDR
    password PWD
    prefix PREFIX
    connect_timeout TIMEOUT
    read_timeout TIMEOUT
    ttl TTL
}
~~~

* `address` is redis server address to connect in the form of *host:port* or *ip:port*.
* `password` is redis server *auth* key
* `connect_timeout` time in ms to wait for redis server to connect
* `read_timeout` time in ms to wait for redis server to respond
* `ttl` default ttl for dns records, 300 if not provided
* `prefix` add PREFIX to all redis keys

### examples

~~~ corefile
. {
    redis example.com {
        address localhost:6379
        password foobared
        connect_timeout 100
        read_timeout 100
        ttl 360
    }
}
~~~

## reverse zones

reverse zones is not supported yet

## proxy

proxy is not supported yet

## zone format in redis db

### zones

zones are stored in redis as a list of strings with *zones* as key

~~~
redis-cli>LRANGE zones 0 -1
1) "example.com."
redis-cli>
~~~

### dns RRs 

dns RRs are stored in redis as json strings inside a hash map using address as key and RR type as field label

#### A

~~~json
{
    "ip4" : "1.2.3.4",
    "ttl" : 360
}
~~~

#### AAAA

~~~json
{
    "ip6" : "::1",
    "ttl" : 360
}
~~~

#### CNAME

~~~json
{
    "host" : "x.example.com.",
    "ttl" : 360
}
~~~

#### TXT

~~~json
{
    "text" : "this is a text",
    "ttl" : 360
}
~~~

#### NS

~~~json
{
    "host" : "ns1.example.com.",
    "ttl" : 360
}
~~~

#### MX

~~~json
{
    "host" : "mx1.example.com",
    "priority" : 10,
    "ttl" : 360
}
~~~

#### SRV

~~~json
{
    "host" : "sip.example.com.",
    "port" : 555,
    "priority" : 10,
    "weight" : 100,
    "ttl" : 360
}
~~~

#### SOA

~~~json
{
    "ttl" : 100,
    "mbox" : "hostmaster.example.com.",
    "ns" : "ns1.example.com.",
    "refresh" : 44,
    "retry" : 55,
    "expire" : 66
}
~~~

#### example

~~~
redis-cli>hgetall x.example.com.
1) "A"
2) "[{\"ip4\":\"1.2.3.4\"},{\"ip4\":\"5.6.7.8\"}]"
3) "AAAA"
4) "[{\"ip6\":\"::1\"}]"
5) "TXT"
6) "[{\"text\":\"foo\"},{\"text\":\"bar\"}]"
7) "NS"
8) "[{\"host\":\"ns1.example.com.\"},{\"host\":\"ns2.example.com.\"}]"
9) "MX"
10) "[{\"host\":\"mx1.example.com.\", \"priority\":10},{\"host\":\"mx2.example.com.\", \"priority\":10}]"}
~~~

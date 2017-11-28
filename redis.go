package redis

import (
	"github.com/coredns/coredns/plugin"
	"github.com/coredns/coredns/request"
	"github.com/coredns/coredns/plugin/etcd/msg"
	"github.com/miekg/dns"
)

type Redis struct {
	Next        plugin.Handler
}


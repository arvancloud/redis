package redis

import (
	"github.com/rverst/coredns-redis/record"
	"log"
	"net"
	"testing"
)

const prefix, suffix = "", ""
const minTtl = 300

const (
	testTtl    = 4242
	testDomain = "example.net"
	testSub1   = "www"
	testSub2   = "subhost"
	nsHost     = "ns1.example.org"
	soaRname   = "hostmaster.example.org"
	soaSerial  = 2006010201
	soaRefresh = 3600
	soaRetry   = 1800
	soaExpire  = 10000
	soaMinTtl  = minTtl
	ip4        = "93.184.216.34"
	ip6        = "2606:2800:220:1:248:1893:25c8:1946"
	txt        = "Lamas, seekers, and great monkeys will always protect them."
	cSub       = "superservice"
	cHost      = "cname.example.org."
	mxHost     = "mail.example.org."
	mxPref     = 10
	srvService  = "_autodiscover._tcp"
	srvPrio     = 10
	srvWeight   = 80
	srvPort     = 443
	caaFlag     = 0
	caaTag      = "issue"
	caaValue    = "letsencrypt.org"
)

var (
	recA     = record.A{Ttl: testTtl, Ip: net.ParseIP(ip4)}
	recAAAA  = record.AAAA{Ttl: testTtl, Ip: net.ParseIP(ip6)}
	recTXT   = record.TXT{Ttl: testTtl, Text: txt}
	recCNAME = record.CNAME{Ttl: testTtl, Host: cHost}
	recNS    = record.NS{Ttl: testTtl, Host: nsHost}
	recMX    = record.MX{Ttl: testTtl, Host: mxHost, Preference: mxPref}
	recSRV   = record.SRV{Ttl: testTtl, Priority: srvPrio, Weight: srvWeight, Port: srvPort, Target: mxHost}
	recCAA   = record.CAA{Ttl: testTtl, Flag: caaFlag, Tag: caaTag, Value: caaValue}
)

func newRedis() *Redis {

	r := New()
	r.SetKeyPrefix(prefix)
	r.SetKeySuffix(suffix)
	r.SetDefaultTtl(minTtl)
	r.SetAddress("192.168.0.100:6379")
	if err := r.Connect(); err != nil {
		log.Fatal(err)
	}
	return r
}

func TestRedis_SaveZone(t *testing.T) {

	z := record.NewZone(testDomain, record.SOA{
		Ttl:     testTtl,
		MName:   nsHost,
		RName:   soaRname,
		Serial:  soaSerial,
		Refresh: soaRefresh,
		Retry:   soaRetry,
		Expire:  soaExpire,
		MinTtl:  soaMinTtl,
	})

	z.Add("@", recA)
	z.Add("@", recAAAA)
	z.Add("@", recTXT)
	z.Add("@", recMX)
	z.Add("@", recCAA)
	z.Add(cSub, recCNAME)
	z.Add(srvService, recSRV)
	z.Add(testSub2, recNS)

	z.Add(testSub1, recA)
	z.Add(testSub1, recAAAA)

	tests := []struct {
		name    string
		zone    record.Zone
		wantErr bool
	}{
		{name: "Test1", zone: *z, wantErr: false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			redis := newRedis()
			if err := redis.SaveZone(tt.zone); (err != nil) != tt.wantErr {
				t.Errorf("SaveZone() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

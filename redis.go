package redis

import (
	"time"
	"encoding/json"
	"strings"
	"fmt"
	"net"

	"github.com/miekg/dns"

	"github.com/coredns/coredns/plugin"

	redisCon "github.com/garyburd/redigo/redis"
)

type Redis struct {
	Next       plugin.Handler
	Zones      map[string]*Zone
	redisc     redisCon.Conn
	keyPrefix  string
	keySuffix  string
	Ttl        uint32
	LastUpdate time.Time
}

type Zone struct {
	Locations map[string]*Record
	Value     *Record
}

type Record struct {
	A     []A_Record `json:"A",omitempty`
	AAAA  []AAAA_Record `json:"AAAA",omitempty`
	TXT   []TXT_Record `json:"TXT",omitempty`
	CNAME []CNAME_Record `json:"CNAME",omitempty`
	NS    []NS_Record `json:"NS",omitempty`
	MX    []MX_Record `json:"MX",omitempty`
	SRV   []SRV_Record `json:"SRV",omitempty`
	SOA   SOA_Record `json:"SOA",omitempty`
}

type A_Record struct {
	Ttl uint32 `json"ttl",omitempty`
	Ip  net.IP `json:"ip"`
}

type AAAA_Record struct {
	Ttl uint32 `json"ttl",omitempty`
	Ip  net.IP `json:"ip"`
}

type TXT_Record struct {
	Ttl  uint32 `json:"ttl",omitempty`
	Text string `json:"text"`
}

type CNAME_Record struct {
	Ttl  uint32 `json"ttl",omitempty`
	Host string `json:"host"`
}

type NS_Record struct {
	Ttl  uint32 `json"ttl",omitempty`
	Host string `json:"host"`
}

type MX_Record struct {
	Ttl        uint32 `json"ttl",omitempty`
	Host       string `json:"host"`
	Preference uint16 `json:"preference"`
}

type SRV_Record struct {
	Ttl      uint32 `json"ttl",omitempty`
	Priority uint16 `json:"priority"`
	Weight   uint16 `json:"weight"`
	Port     uint16 `json:"port""`
	Target   string `json:"target"`
}

type SOA_Record struct {
	Ttl     uint32 `json"ttl",omitempty`
	Ns      string `json:"ns"`
	MBox    string `json:"MBox"`
	Refresh uint32 `json:"refresh"`
	Retry   uint32 `json:"retry"`
	Expire  uint32 `json:"expire"`
	MinTtl  uint32 `json:"minttl"`
}

func (redis *Redis) GetZones() (zones []string) {
	for zone := range redis.Zones {
		zones = append(zones, zone)
	}
	return
}

func (redis *Redis) A(name string, zone string, record *Record) (answers, extras []dns.RR) {
	for _, a := range record.A {
		if a.Ip == nil {
			continue
		}
		r := new(dns.A)
		r.Hdr = dns.RR_Header{Name: name, Rrtype: dns.TypeA,
			Class: dns.ClassINET, Ttl: redis.minTtl(a.Ttl)}
		r.A = a.Ip
		answers = append(answers, r)
	}
	return
}

func (redis Redis) AAAA(name string, zone string, record *Record) (answers, extras []dns.RR) {
	for _, aaaa := range record.AAAA {
		if aaaa.Ip == nil {
			continue
		}
		r := new(dns.AAAA)
		r.Hdr = dns.RR_Header{Name: name, Rrtype: dns.TypeAAAA,
			Class: dns.ClassINET, Ttl: redis.minTtl(aaaa.Ttl)}
		r.AAAA = aaaa.Ip
		answers = append(answers, r)
	}
	return
}

func (redis *Redis) CNAME(name string, zone string, record *Record) (answers, extras []dns.RR) {
	for _, cname := range record.CNAME {
		if len(cname.Host) == 0 {
			continue
		}
		r := new(dns.CNAME)
		r.Hdr = dns.RR_Header{Name: name, Rrtype: dns.TypeCNAME,
			Class: dns.ClassINET, Ttl: redis.minTtl(cname.Ttl)}
		r.Target = dns.Fqdn(cname.Host)
		answers = append(answers, r)
	}
	return
}

func (redis *Redis) TXT(name string, zone string, record *Record) (answers, extras []dns.RR) {
	for _, txt := range record.TXT {
		if len(txt.Text) == 0 {
			continue
		}
		r:= new(dns.TXT)
		r.Hdr = dns.RR_Header{Name: name, Rrtype: dns.TypeTXT,
			Class: dns.ClassINET, Ttl: redis.minTtl(txt.Ttl)}
		r.Txt = split255(txt.Text)
		answers = append(answers, r)
	}
	return
}

func (redis *Redis) NS(name string, zone string, record *Record) (answers, extras []dns.RR) {
	for _, ns := range record.NS {
		if len(ns.Host) == 0 {
			continue
		}
		r := new(dns.NS)
		r.Hdr = dns.RR_Header{Name: name, Rrtype: dns.TypeNS,
			Class: dns.ClassINET, Ttl: redis.minTtl(ns.Ttl)}
		r.Ns = ns.Host
		answers = append(answers, r)
		extras = append(extras, redis.hosts(ns.Host, zone)...)
	}
	return
}

func (redis *Redis) MX(name string, zone string, record *Record) (answers, extras []dns.RR) {
	for _, mx := range record.MX {
		if len(mx.Host) == 0 {
			continue
		}
		r := new(dns.MX)
		r.Hdr = dns.RR_Header{Name: name, Rrtype: dns.TypeMX,
			Class: dns.ClassINET, Ttl: redis.minTtl(mx.Ttl)}
		r.Mx = mx.Host
		r.Preference = mx.Preference
		answers = append(answers, r)
		extras = append(extras, redis.hosts(mx.Host, zone)...)
	}
	return
}

func (redis *Redis) SRV(name string, zone string, record *Record) (answers, extras []dns.RR) {
	for _, srv := range record.SRV {
		if len(srv.Target) == 0 {
			continue
		}
		r := new(dns.SRV)
		r.Hdr = dns.RR_Header{Name: name, Rrtype: dns.TypeSRV,
			Class: dns.ClassINET, Ttl: redis.minTtl(srv.Ttl)}
		r.Target = srv.Target
		r.Weight = srv.Weight
		r.Port = srv.Port
		r.Priority = srv.Priority
		answers = append(answers, r)
		extras = append(extras, redis.hosts(srv.Target, zone)...)
	}
	return
}

func (redis *Redis) SOA(name string, zone string, record *Record) (answers, extras []dns.RR) {
	r := new(dns.SOA)
	if record.SOA.Ns == "" {
		r.Hdr = dns.RR_Header{Name: name, Rrtype: dns.TypeSOA,
			Class: dns.ClassINET, Ttl: redis.Ttl}
		r.Ns = "ns1." + name
		r.Mbox = "hostmaster." + name
		r.Refresh = 86400
		r.Retry = 7200
		r.Expire = 3600
		r.Minttl = redis.Ttl
	} else {
		r.Hdr = dns.RR_Header{Name: zone, Rrtype: dns.TypeSOA,
			Class: dns.ClassINET, Ttl: redis.minTtl(record.SOA.Ttl)}
		r.Ns = record.SOA.Ns
		r.Mbox = record.SOA.MBox
		r.Refresh = record.SOA.Refresh
		r.Retry = record.SOA.Retry
		r.Expire = record.SOA.Expire
		r.Minttl = record.SOA.MinTtl
	}
	r.Serial = redis.serial()
	answers = append(answers, r)
	return
}

func (redis *Redis) hosts(name string, zone string) []dns.RR {
	var (
		record *Record
		answers []dns.RR
	)
	location := redis.findLocation(name, zone)
	if location == "" {
		return nil
	}
	record = redis.get(location, zone)
	a, _ := redis.A(name, zone, record)
	answers = append(answers, a...)
	aaaa, _ := redis.AAAA(name, zone, record)
	answers = append(answers, aaaa...)
	cname, _ := redis.CNAME(name, zone, record)
	answers = append(answers, cname...)
	return answers
}

func (redis *Redis) serial() uint32 {
	return uint32(time.Now().Unix())
}

func (redis *Redis) minTtl(ttl uint32) uint32 {
	if redis.Ttl == 0 && ttl == 0 {
		return defaultTtl
	}
	if redis.Ttl == 0 {
		return ttl
	}
	if ttl == 0 {
		return redis.Ttl
	}
	if redis.Ttl < ttl {
		return redis.Ttl
	}
	return  ttl
}

func (redis *Redis) findLocation(query string, zone string) string {
	var (
		z *Zone
		ok bool
		closestEncloser, sourceOfSynthesis string
	)

	// no matching zone
	if z, ok = redis.Zones[zone]; !ok {
		return ""
	}

	// request for zone records
	if query == zone {
		return zone
	}

	query = strings.TrimSuffix(query, "." + zone)

	if _, ok = z.Locations[query]; ok {
		return query
	}

	closestEncloser, sourceOfSynthesis, ok = splitQuery(query)
	for ok {
		ceExists := keyMatches(closestEncloser, redis.Zones[zone]) || keyExists(closestEncloser, redis.Zones[zone])
		ssExists := keyExists(sourceOfSynthesis, redis.Zones[zone])
		if ceExists {
			if ssExists {
				return sourceOfSynthesis
			} else {
				return ""
			}
		} else {
			closestEncloser, sourceOfSynthesis, ok = splitQuery(closestEncloser)
		}
	}
	return ""
}

func (redis *Redis) get(key string, zone string) *Record {
	if key == zone {
		return redis.Zones[zone].Value
	}
	return redis.Zones[zone].Locations[key]
}

func keyExists(key string, zone *Zone) bool {
	_, ok := zone.Locations[key]
	return ok
}

func keyMatches(key string, zone *Zone) bool {
	for value := range zone.Locations {
		if strings.HasSuffix(value, key) {
			return true
		}
	}
	return false
}

func splitQuery(query string) (string, string, bool) {
	if query == "" {
		return "", "", false
	}
	var (
		splits []string
		closestEncloser string
		sourceOfSynthesis string
	)
	splits = strings.SplitAfterN(query, ".", 2)
	if len(splits) == 2 {
		closestEncloser = splits[1]
		sourceOfSynthesis = "*." + closestEncloser
	} else {
		closestEncloser = ""
		sourceOfSynthesis = "*"
	}
	return closestEncloser, sourceOfSynthesis, true
}

func (redis *Redis) save(zone string, subdomain string, value string) error {
	var err error
	_, err = redis.redisc.Do("HSET", redis.keyPrefix + zone + redis.keySuffix, subdomain, value)
	return err
}

func (redis *Redis) load() error {
	fmt.Println("load")
	var (
		reply interface{}
		err error
		vals []string
		zones []string
	)
	reply, err = redis.redisc.Do("KEYS", redis.keyPrefix + "*" + redis.keySuffix)
	if err != nil {
		return err
	}
	zones, err = redisCon.Strings(reply, nil)
	if err != nil {
		return err
	}
	redis.Zones = make(map[string]*Zone, len(zones))
	for _, zone := range zones {
		reply, err = redis.redisc.Do("HGETALL", zone)
		if err != nil {
			return err
		}
		zone = strings.TrimPrefix(zone, redis.keyPrefix)
		zone = strings.TrimSuffix(zone, redis.keySuffix)
		redis.Zones[zone] = new(Zone)
		redis.Zones[zone].Locations = make(map[string]*Record)
		vals, err = redisCon.Strings(reply, nil)
		if err != nil {
			return err
		}
		for i := 0; i < len(vals); i += 2 {
			r := new(Record)
			err = json.Unmarshal([]byte(vals[i+1]), r)
			if err != nil {
				fmt.Println("parse error : ", vals[i+1], err)
				continue
			}
			if vals[i] == "@" {
				redis.Zones[zone].Value = r
			} else {
				redis.Zones[zone].Locations[vals[i]] = r
			}
		}
	}
	redis.LastUpdate = time.Now()
	return nil
}

func split255(s string) []string {
	if len(s) < 255 {
		return []string{s}
	}
	sx := []string{}
	p, i := 0, 255
	for {
		if i <= len(s) {
			sx = append(sx, s[p:i])
		} else {
			sx = append(sx, s[p:])
			break

		}
		p, i = p+255, i+255
	}

	return sx
}

const (
	defaultTtl = 360
	hostmaster = "hostmaster"
)
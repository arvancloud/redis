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
	Next        plugin.Handler
	Zones       []string
	redisc      redisCon.Conn
	Ttl         uint32
}

type Record struct {
	Host     string `json:"host,omitempty"`
	Ip4      net.IP `json:"ip4,omitempty"`
	Ip6      net.IP `json:"ip6,omitempty"`
	Port     uint16 `json:"port,omitempty"`
	Priority uint16 `json:"priority,omitempty"`
	Weight   uint16 `json:"weight,omitempty"`
	Text     string `json:"text,omitempty"`
	Ttl      uint32 `json:"ttl,omitempty"`
	MBox     string `json:"mbox,omitempty"`
	Ns       string `json:"ns,omitempty"`
	Refresh  uint32 `json:"refresh,omitempty"`
	Retry    uint32 `json:"retry,omitempty"`
	Expire   uint32 `json:"expire,omitempty"`
}

func (redis *Redis) Answer(qname string, qtype string, zone string) ([]dns.RR, []dns.RR) {
	fmt.Println("looking for ", qname)
	key := redis.findKey(qname, zone)
	if key == "" { // empty, no results
		return nil, nil
	}

	answers := make([]dns.RR, 0, 10)
	extras := make([]dns.RR, 0, 10)

	records := redis.get(key, qtype)
	if len(records) == 0 {
		cnameRecords := redis.get(key, "CNAME")
		for _, cnameRecord := range cnameRecords {
			if dns.IsSubDomain(zone, cnameRecord.Host) {
				records = append(records,redis.get(cnameRecord.Host, qtype)...)
			}
		}
	}

	if len(records) == 0 {
		return nil, nil
	}

	switch qtype {
	case "A":
		answers, extras = redis.A(qname, zone, records)
	case "AAAA":
		answers, extras = redis.AAAA(qname, zone, records)
	case "CNAME":
		answers, extras = redis.CNAME(qname, zone, records)
	case "TXT":
		answers, extras = redis.TXT(qname, zone, records)
	case "NS":
		answers, extras = redis.NS(qname, zone, records)
	case "MX":
		answers, extras = redis.MX(qname, zone, records)
	case "SRV":
		answers, extras = redis.SRV(qname, zone, records)
	case "SOA":
		answers, extras = redis.SOA(qname, zone, records)
	default:

	}
	return answers, extras
}

func (redis *Redis) A(name string, zone string, records []Record) (answers, extras []dns.RR) {
	for _, record := range records {
		if record.Ip4 == nil {
			continue
		}
		r := new(dns.A)
		r.Hdr = dns.RR_Header{Name: name, Rrtype: dns.TypeA,
			Class: dns.ClassINET, Ttl: redis.minTtl(&record)}
		r.A = record.Ip4
		answers = append(answers, r)
	}
	return
}

func (redis Redis) AAAA(name string, zone string, records []Record) (answers, extras []dns.RR) {
	for _, record := range records {
		if record.Ip6 == nil {
			continue
		}
		r := new(dns.AAAA)
		r.Hdr = dns.RR_Header{Name: name, Rrtype: dns.TypeAAAA,
			Class: dns.ClassINET, Ttl: redis.minTtl(&record)}
		r.AAAA = record.Ip6
		answers = append(answers, r)
	}
	return
}

func (redis *Redis) CNAME(name string, zone string, records []Record) (answers, extras []dns.RR) {
	for _, record := range records {
		if len(record.Host) == 0 {
			continue
		}
		r := new(dns.CNAME)
		r.Hdr = dns.RR_Header{Name: name, Rrtype: dns.TypeCNAME,
			Class: dns.ClassINET, Ttl: redis.minTtl(&record)}
		r.Target = dns.Fqdn(record.Host)
		answers = append(answers, r)
	}
	return
}

func (redis *Redis) TXT(name string, zone string, records []Record) (answers, extras []dns.RR) {
	for _, record := range records {
		if len(record.Text) == 0 {
			continue
		}
		r:= new(dns.TXT)
		r.Hdr = dns.RR_Header{Name: name, Rrtype: dns.TypeTXT,
			Class: dns.ClassINET, Ttl: redis.minTtl(&record)}
		r.Txt = split255(record.Text)
		answers = append(answers, r)
	}
	return
}

func (redis *Redis) NS(name string, zone string, records []Record) (answers, extras []dns.RR) {
	for _, record := range records {
		if len(record.Host) == 0 {
			continue
		}
		r := new(dns.NS)
		r.Hdr = dns.RR_Header{Name: name, Rrtype: dns.TypeNS,
			Class: dns.ClassINET, Ttl: redis.minTtl(&record)}
		r.Ns = record.Host
		answers = append(answers, r)
		extras = append(extras, redis.hosts(record.Host, zone)...)
	}
	return
}

func (redis *Redis) MX(name string, zone string, records []Record) (answers, extras []dns.RR) {
	for _, record := range records {
		if len(record.Host) == 0 {
			continue
		}
		r := new(dns.MX)
		r.Hdr = dns.RR_Header{Name: name, Rrtype: dns.TypeMX,
			Class: dns.ClassINET, Ttl: redis.minTtl(&record)}
		r.Mx = record.Host
		r.Preference = record.Priority
		answers = append(answers, r)
		extras = append(extras, redis.hosts(record.Host, zone)...)
	}
	return
}

func (redis *Redis) SRV(name string, zone string, records []Record) (answers, extras []dns.RR) {
	for _, record := range records {
		if len(record.Host) == 0 {
			continue
		}
		r := new(dns.SRV)
		r.Hdr = dns.RR_Header{Name: name, Rrtype: dns.TypeSRV,
			Class: dns.ClassINET, Ttl: redis.minTtl(&record)}
		r.Target = record.Host
		r.Weight = record.Weight
		r.Port = record.Port
		answers = append(answers, r)
	}
	return
}

func (redis *Redis) SOA(name string, zone string, records []Record) (answers, extras []dns.RR) {
	r := new(dns.SOA)
	if records == nil {
		r.Hdr = dns.RR_Header{Name: name, Rrtype: dns.TypeSOA,
			Class: dns.ClassINET, Ttl: redis.Ttl}
		r.Ns = "ns1" + name
		r.Mbox = "hostmaster" + name
		r.Refresh = 86400
		r.Retry = 7200
		r.Expire = 3600
		r.Minttl = redis.Ttl
	} else {
		r.Hdr = dns.RR_Header{Name: zone, Rrtype: dns.TypeSOA,
			Class: dns.ClassINET, Ttl: redis.minTtl(&records[0])}
		r.Ns = records[0].Ns
		r.Mbox = records[0].MBox
		r.Refresh = records[0].Refresh
		r.Retry = records[0].Retry
		r.Expire = records[0].Expire
		r.Minttl = redis.minTtl(&records[0])
	}
	r.Serial = redis.serial()
	answers = append(answers, r)
	return
}

func (redis *Redis) hosts(name string, zone string) []dns.RR {
	var (
		records []Record
		answers []dns.RR
	)

	records = redis.get(name, "A")
	a, _ := redis.A(name, zone, records)
	answers = append(answers, a...)
	records = redis.get(name, "AAAA")
	aaaa, _ := redis.AAAA(name, zone, records)
	answers = append(answers, aaaa...)
	records = redis.get(name, "CNAME")
	cname, _ := redis.CNAME(name, zone, records)
	answers = append(answers, cname...)
	return answers
}

func (redis *Redis) serial() uint32 {
	return uint32(time.Now().Unix())
}

func (redis *Redis) minTtl(record *Record) uint32 {
	if redis.Ttl == 0 && record.Ttl == 0 {
		return defaultTtl
	}
	if redis.Ttl == 0 {
		return record.Ttl
	}
	if record.Ttl == 0 {
		return redis.Ttl
	}
	if redis.Ttl < record.Ttl {
		return redis.Ttl
	}
	return  record.Ttl
}


func (redis *Redis) findKey(query string, zone string) string {
	fmt.Println("looking for exact match for ", query)
	if redis.keyExists(query) {
		fmt.Println("exact match found")
		return query
	}
	fmt.Println("exact match not found")
	closestEncloser, sourceOfSynthesis := splitQuery(query)
	for strings.Contains(closestEncloser, zone) {
		fmt.Println("ce: ", closestEncloser, "ss: ", sourceOfSynthesis)
		ceExists := redis.keyMatches(sourceOfSynthesis)
		ssExists := redis.keyExists(sourceOfSynthesis)
		if ceExists {
			fmt.Println("ce exists")
			if ssExists {
				fmt.Println("ss exists")
				return sourceOfSynthesis
			} else {
				fmt.Println("ss not exist")
				return ""
			}
		} else {
			fmt.Println("ce not exist")
			closestEncloser, sourceOfSynthesis = splitQuery(closestEncloser)
		}
	}
	return ""
}

func splitQuery(query string) (string, string) {
	var (
		splits []string
		closestEncloser string
		sourceOfSynthesis string
	)
	splits = strings.SplitAfterN(query, ".", 2)
	if len(splits) == 2 {
		closestEncloser = splits[1]
	}
	sourceOfSynthesis = "*." + closestEncloser
	return closestEncloser, sourceOfSynthesis
}

func (redis *Redis) keyMatches(pattern string) bool {
	var (
		reply, err interface{}
		res []interface{}
		keys []string
	)
	reply, err = redis.redisc.Do("SCAN",0, "match", pattern)
	if err != nil {
		// report error?
		return false
	}
	res, err = redisCon.Values(reply, nil)
	if err != nil {
		return false
	}
	keys, err = redisCon.Strings(res[1], nil)
	if err == nil && len(keys) > 0 {
		return true
	}
	return false
}

func (redis *Redis) keyExists(key string) bool {
	var (
		reply, err interface{}
		res int
	)
	reply, err = redis.redisc.Do("EXISTS", key)
	if err != nil {
		return false
	}
	res, err = redisCon.Int(reply, nil)
	if err != nil {
		return false
	}
	if res == 1 {
		return true
	}
	return false
}

func (redis *Redis) get(qname string, qtype string) []Record {
	reply, err := redis.redisc.Do("HGET", qname, qtype)
	if err != nil {
		fmt.Printf("error in hget : %s\n", err)
		return nil
	}
	value, err := redisCon.String(reply, nil)
	if err != nil || len(value) == 0 {
		fmt.Println("no values")
		return nil
	}
	var res []Record
	err = json.Unmarshal([]byte(value), &res)
	if err != nil {
		fmt.Println("parse error")
	}
	return res
}

func (redis *Redis) set(params []string) error {
	s := make([]interface{}, len(params))
	for i, v := range params {
		s[i] = v
	}
	_, err := redis.redisc.Do("HMSET", s...)
	return err
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
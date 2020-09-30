package redis

import (
	"encoding/json"
	"fmt"
	"github.com/coredns/coredns/plugin"
	"github.com/coredns/coredns/request"
	"github.com/miekg/dns"
	"github.com/rverst/coredns-redis/record"
	"strings"
	"time"

	redisCon "github.com/gomodule/redigo/redis"
)

const (
	DefaultTtl        = 3600
	MaxTransferLength = 1000
)

type Redis struct {
	Pool           *redisCon.Pool
	address        string
	password       string
	connectTimeout int
	readTimeout    int
	keyPrefix      string
	keySuffix      string
	DefaultTtl     int
	Zones          []string
	LastZoneUpdate time.Time
}

func New() *Redis {
	return &Redis{}
}

func (redis *Redis) SetAddress(a string) {
	redis.address = a
}

func (redis *Redis) SetPassword(p string) {
	redis.password = p
}

func (redis *Redis) SetKeyPrefix(p string) {
	redis.keyPrefix = p
}

func (redis *Redis) SetKeySuffix(s string) {
	redis.keySuffix = s
}

func (redis *Redis) SetConnectTimeout(t int) {
	redis.connectTimeout = t
}

func (redis *Redis) SetReadTimeout(t int) {
	redis.readTimeout = t
}

func (redis *Redis) SetDefaultTtl(t int) {
	redis.DefaultTtl = t
}

// Ping sends a "PING" command to the redis server
// and returns (true, nil) if the server response
// is 'PONG'. Otherwise Ping return false and
// an error
func (redis *Redis) Ping() (bool, error) {
	conn := redis.Pool.Get()
	defer conn.Close()

	r, err := conn.Do("PING")
	s, err := redisCon.String(r, err)
	if err != nil {
		return false, err
	}
	if s != "PONG" {
		return false, fmt.Errorf("unexpected response, expected 'PONG', got: %s", s)
	}
	return true, nil
}

func (redis *Redis) ErrorResponse(state request.Request, zone string, rcode int, err error) (int, error) {
	m := new(dns.Msg)
	m.SetRcode(state.Req, rcode)
	m.Authoritative, m.RecursionAvailable, m.Compress = true, false, true

	state.SizeAndDo(m)
	_ = state.W.WriteMsg(m)
	// Return success as the rcode to signal we have written to the client.
	return dns.RcodeSuccess, err
}

func (redis *Redis) SOA(z *record.Zone, rec *record.Records) (answers, extras []dns.RR) {
	soa := new(dns.SOA)

	soa.Hdr = dns.RR_Header{Name: dns.Fqdn(z.Name), Rrtype: dns.TypeSOA,
		Class: dns.ClassINET, Ttl: redis.ttl(rec.SOA.Ttl)}
	soa.Ns = rec.SOA.MName
	soa.Mbox = rec.SOA.RName
	soa.Serial = rec.SOA.Serial
	soa.Refresh = rec.SOA.Refresh
	soa.Retry = rec.SOA.Retry
	soa.Expire = rec.SOA.Expire
	soa.Minttl = rec.SOA.MinTtl
	if soa.Serial == 0 {
		soa.Serial = record.DefaultSerial()
	}
	answers = append(answers, soa)
	return
}

func (redis *Redis) A(name string, _ *record.Zone, record *record.Records) (answers, extras []dns.RR) {
	for _, a := range record.A {
		if a.Ip == nil {
			continue
		}
		r := new(dns.A)
		r.Hdr = dns.RR_Header{Name: dns.Fqdn(name), Rrtype: dns.TypeA,
			Class: dns.ClassINET, Ttl: redis.ttl(a.Ttl)}
		r.A = a.Ip
		answers = append(answers, r)
	}
	return
}

func (redis Redis) AAAA(name string, _ *record.Zone, record *record.Records) (answers, extras []dns.RR) {
	for _, aaaa := range record.AAAA {
		if aaaa.Ip == nil {
			continue
		}
		r := new(dns.AAAA)
		r.Hdr = dns.RR_Header{Name: dns.Fqdn(name), Rrtype: dns.TypeAAAA,
			Class: dns.ClassINET, Ttl: redis.ttl(aaaa.Ttl)}
		r.AAAA = aaaa.Ip
		answers = append(answers, r)
	}
	return
}

func (redis *Redis) CNAME(name string, _ *record.Zone, record *record.Records) (answers, extras []dns.RR) {
	for _, cname := range record.CNAME {
		if len(cname.Host) == 0 {
			continue
		}
		r := new(dns.CNAME)
		r.Hdr = dns.RR_Header{Name: dns.Fqdn(name), Rrtype: dns.TypeCNAME,
			Class: dns.ClassINET, Ttl: redis.ttl(cname.Ttl)}
		r.Target = dns.Fqdn(cname.Host)
		answers = append(answers, r)
	}
	return
}

func (redis *Redis) TXT(name string, _ *record.Zone, record *record.Records) (answers, extras []dns.RR) {
	for _, txt := range record.TXT {
		if len(txt.Text) == 0 {
			continue
		}
		r := new(dns.TXT)
		r.Hdr = dns.RR_Header{Name: dns.Fqdn(name), Rrtype: dns.TypeTXT,
			Class: dns.ClassINET, Ttl: redis.ttl(txt.Ttl)}
		r.Txt = split255(txt.Text)
		answers = append(answers, r)
	}
	return
}

func (redis *Redis) NS(name string, z *record.Zone, record *record.Records, zones []string) (answers, extras []dns.RR) {
	for _, ns := range record.NS {
		if len(ns.Host) == 0 {
			continue
		}
		r := new(dns.NS)
		r.Hdr = dns.RR_Header{Name: dns.Fqdn(name), Rrtype: dns.TypeNS,
			Class: dns.ClassINET, Ttl: redis.ttl(ns.Ttl)}
		r.Ns = ns.Host
		answers = append(answers, r)
		extras = append(extras, redis.getExtras(ns.Host, z, zones)...)
	}
	return
}

func (redis *Redis) MX(name string, z *record.Zone, record *record.Records, zones []string) (answers, extras []dns.RR) {
	for _, mx := range record.MX {
		if len(mx.Host) == 0 {
			continue
		}
		r := new(dns.MX)
		r.Hdr = dns.RR_Header{Name: dns.Fqdn(name), Rrtype: dns.TypeMX,
			Class: dns.ClassINET, Ttl: redis.ttl(mx.Ttl)}
		r.Mx = mx.Host
		r.Preference = mx.Preference
		answers = append(answers, r)
		extras = append(extras, redis.getExtras(mx.Host, z, zones)...)
	}
	return
}

func (redis *Redis) SRV(name string, z *record.Zone, record *record.Records, zones []string) (answers, extras []dns.RR) {
	for _, srv := range record.SRV {
		if len(srv.Target) == 0 {
			continue
		}
		r := new(dns.SRV)
		r.Hdr = dns.RR_Header{Name: dns.Fqdn(name), Rrtype: dns.TypeSRV,
			Class: dns.ClassINET, Ttl: redis.ttl(srv.Ttl)}
		r.Target = srv.Target
		r.Weight = srv.Weight
		r.Port = srv.Port
		r.Priority = srv.Priority
		answers = append(answers, r)
		extras = append(extras, redis.getExtras(srv.Target, z, zones)...)
	}
	return
}

func (redis *Redis) CAA(name string, _ *record.Zone, record *record.Records) (answers, extras []dns.RR) {
	if record == nil {
		return
	}
	for _, caa := range record.CAA {
		if caa.Value == "" || caa.Tag == "" {
			continue
		}
		r := new(dns.CAA)
		r.Hdr = dns.RR_Header{Name: dns.Fqdn(name), Rrtype: dns.TypeCAA, Class: dns.ClassINET}
		r.Flag = caa.Flag
		r.Tag = caa.Tag
		r.Value = caa.Value
		answers = append(answers, r)
	}
	return
}

func (redis *Redis) AXFR(z *record.Zone, zones []string) (records []dns.RR) {
	//soa, _ := redis.SOA(z.Name, z, record)
	soa := make([]dns.RR, 0)
	answers := make([]dns.RR, 0, 10)
	extras := make([]dns.RR, 0, 10)

	// Allocate slices for rr Records
	records = append(records, soa...)
	for key := range z.Locations {
		if key == "@" {
			location := redis.FindLocation(z.Name, z)
			zoneRecords := redis.LoadZoneRecords(location, z)
			soa, _ = redis.SOA(z, zoneRecords)
		} else {
			fqdnKey := dns.Fqdn(key) + z.Name
			var as []dns.RR
			var xs []dns.RR

			location := redis.FindLocation(fqdnKey, z)
			zoneRecords := redis.LoadZoneRecords(location, z)

			// Pull all zone records
			as, xs = redis.A(fqdnKey, z, zoneRecords)
			answers = append(answers, as...)
			extras = append(extras, xs...)

			as, xs = redis.AAAA(fqdnKey, z, zoneRecords)
			answers = append(answers, as...)
			extras = append(extras, xs...)

			as, xs = redis.CNAME(fqdnKey, z, zoneRecords)
			answers = append(answers, as...)
			extras = append(extras, xs...)

			as, xs = redis.MX(fqdnKey, z, zoneRecords, zones)
			answers = append(answers, as...)
			extras = append(extras, xs...)

			as, xs = redis.SRV(fqdnKey, z, zoneRecords, zones)
			answers = append(answers, as...)
			extras = append(extras, xs...)

			as, xs = redis.TXT(fqdnKey, z, zoneRecords)
			answers = append(answers, as...)
			extras = append(extras, xs...)
		}
	}

	records = soa
	records = append(records, answers...)
	records = append(records, extras...)
	records = append(records, soa...)

	fmt.Println(records)
	return
}

func (redis *Redis) getExtras(name string, z *record.Zone, zones []string) []dns.RR {

	location := redis.FindLocation(name, z)
	if location == "" {
		zoneName := plugin.Zones(zones).Matches(name)
		if zoneName == "" {
			zones, err := redis.LoadZones(name)
			if err != nil {
				return nil
			}
			zoneName = plugin.Zones(zones).Matches(name)
			if zoneName == "" {
				return nil
			}
		}

		z2 := redis.LoadZone(zoneName, false)
		location = redis.FindLocation(name, z2)
		if location == "" {
			return nil
		}
		return redis.fillExtras(name, z2, location)
	}
	return redis.fillExtras(name, z, location)
}

func (redis *Redis) fillExtras(name string, z *record.Zone, location string) []dns.RR {
	var (
		zoneRecords *record.Records
		answers     []dns.RR
	)

	zoneRecords = redis.LoadZoneRecords(location, z)
	if zoneRecords == nil {
		return nil
	}
	a, _ := redis.A(name, z, zoneRecords)
	answers = append(answers, a...)
	aaaa, _ := redis.AAAA(name, z, zoneRecords)
	answers = append(answers, aaaa...)
	cname, _ := redis.CNAME(name, z, zoneRecords)
	answers = append(answers, cname...)
	return answers
}

func (redis *Redis) ttl(ttl int) uint32 {
	if ttl >= 0 {
		return uint32(ttl)
	}
	// todo: return SOA minTTL
	if redis.DefaultTtl >= 0 {
		return uint32(redis.DefaultTtl)
	}
	return DefaultTtl
}

func (redis *Redis) FindLocation(query string, z *record.Zone) string {
	var (
		ok                                 bool
		closestEncloser, sourceOfSynthesis string
	)

	// request for zone records
	if query == z.Name {
		return query
	}

	query = strings.TrimSuffix(query, "."+z.Name)

	if _, ok = z.Locations[query]; ok {
		return query
	}

	closestEncloser, sourceOfSynthesis, ok = splitQuery(query)
	for ok {
		ceExists := keyMatches(closestEncloser, z) || keyExists(closestEncloser, z)
		ssExists := keyExists(sourceOfSynthesis, z)
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

func (redis *Redis) Connect() error {
	redis.Pool = &redisCon.Pool{
		Dial: func() (redisCon.Conn, error) {
			opts := []redisCon.DialOption{}
			if redis.password != "" {
				opts = append(opts, redisCon.DialPassword(redis.password))
			}
			if redis.connectTimeout != 0 {
				opts = append(opts, redisCon.DialConnectTimeout(time.Duration(redis.connectTimeout)*time.Millisecond))
			}
			if redis.readTimeout != 0 {
				opts = append(opts, redisCon.DialReadTimeout(time.Duration(redis.readTimeout)*time.Millisecond))
			}

			return redisCon.Dial("tcp", redis.address, opts...)
		},
	}
	c := redis.Pool.Get()
	defer c.Close()

	if c.Err() != nil {
		return c.Err()
	}

	res, err := c.Do("PING")
	pong, err := redisCon.String(res, err)
	if err != nil {
		return err
	}
	if pong != "PONG" {
		return fmt.Errorf("unexpexted result, 'PONG' expected: %s", pong)
	}
	return nil
}

func (redis *Redis) SaveZone(zone record.Zone) error {
	conn := redis.Pool.Get()
	defer conn.Close()
	for k, v := range zone.Locations {
		data, err := json.Marshal(v)
		if err != nil {
			return err
		}
		_, err = conn.Do("HSET", redis.Key(zone.Name), k, data)
		if err != nil {
			return err
		}
	}

	return nil
}

//func (redis *Redis) Save(zone string, subdomain string, value string) error {
//	var err error
//
//	conn := redis.Pool.Get()
//	if conn == nil {
//		fmt.Println("error connecting to redis")
//		return nil
//	}
//	defer conn.Close()
//
//	_, err = conn.Do("HSET", redis.keyPrefix+zone+redis.keySuffix, subdomain, value)
//	return err
//}

func (redis *Redis) LoadZone(zone string, withRecord bool) *record.Zone {
	var (
		reply interface{}
		err   error
		vals  []string
	)

	conn := redis.Pool.Get()
	if conn == nil {
		fmt.Println("error connecting to redis")
		return nil
	}
	defer conn.Close()

	reply, err = conn.Do("HKEYS", redis.keyPrefix+zone+redis.keySuffix)
	vals, err = redisCon.Strings(reply, err)
	if err != nil || len(vals) == 0 {
		return nil
	}

	z := new(record.Zone)
	z.Name = zone
	z.Locations = make(map[string]record.Records)
	for _, val := range vals {
		if withRecord {
			z.Locations[val] = *redis.LoadZoneRecords(val, z)
		} else {
			z.Locations[val] = record.Records{} //struct{}{}
		}
	}

	return z
}

func (redis *Redis) LoadZoneRecords(key string, z *record.Zone) *record.Records {
	var (
		err   error
		reply interface{}
		val   string
	)
	conn := redis.Pool.Get()
	defer conn.Close()

	var label string
	if key == z.Name {
		label = "@"
	} else {
		label = key
	}

	reply, err = conn.Do("HGET", redis.keyPrefix+z.Name+redis.keySuffix, label)
	if err != nil {
		return nil
	}
	val, err = redisCon.String(reply, nil)
	if err != nil {
		return nil
	}
	r := new(record.Records)
	err = json.Unmarshal([]byte(val), r)
	if err != nil {
		fmt.Println("parse error : ", val, err)
		return nil
	}

	if r.SOA != nil {
		if !dns.IsFqdn(r.SOA.MName) {
			r.SOA.MName += "." + z.Name
		}
		if !dns.IsFqdn(r.SOA.RName) {
			r.SOA.RName += "." + z.Name
		}
	}

	for i := range r.CNAME {
		if !dns.IsFqdn(r.CNAME[i].Host) {
			r.CNAME[i].Host += "." + z.Name
		}
	}
	for i := range r.MX {
		if !dns.IsFqdn(r.MX[i].Host) {
			r.MX[i].Host += "." + z.Name
		}
	}
	for i := range r.NS {
		if !dns.IsFqdn(r.NS[i].Host) {
			r.NS[i].Host += "." + z.Name
		}
	}
	for i := range r.SRV {
		if !dns.IsFqdn(r.SRV[i].Target) {
			r.SRV[i].Target += "." + z.Name
		}
	}
	return r
}

func (redis *Redis) LoadZones(name string) ([]string, error) {
	var (
		reply interface{}
		err   error
		zones []string
	)

	query := reduceZoneName(name)
	if query == "" {
		query = name
	}

	conn := redis.Pool.Get()
	defer conn.Close()

	reply, err = conn.Do("KEYS", redis.keyPrefix+"*"+query+redis.keySuffix)
	zones, err = redisCon.Strings(reply, err)
	if err != nil {
		return nil, err
	}

	for i, _ := range zones {
		zones[i] = strings.TrimPrefix(zones[i], redis.keyPrefix)
		zones[i] = strings.TrimSuffix(zones[i], redis.keySuffix)
	}
	return zones, nil
}

// Key returns the given key with prefix and suffix
func (redis *Redis) Key(key string) string {
	return redis.keyPrefix + key + redis.keySuffix
}

func keyExists(key string, z *record.Zone) bool {
	_, ok := z.Locations[key]
	return ok
}

func keyMatches(key string, z *record.Zone) bool {
	for value := range z.Locations {
		if strings.HasSuffix(value, key) {
			return true
		}
	}
	return false
}

// reduceZoneName strips the zone down to top- and second-level
// so we can query the subset from redis. This should give
// no problems unless we want to run a root dns
func reduceZoneName(name string) string {
	name = dns.Fqdn(name)
	split := strings.Split(name[:len(name)-1], ".")
	if len(split) == 0 {
		return ""
	}
	x := len(split) - 2
	if x > 0 {
		name = ""
		for ; x < len(split); x++ {
			name += split[x] + "."
		}
	}
	return name
}

func splitQuery(query string) (string, string, bool) {
	if query == "" {
		return "", "", false
	}
	var (
		splits            []string
		closestEncloser   string
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

func split255(s string) []string {
	if len(s) < 255 {
		return []string{s}
	}
	var sx []string
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

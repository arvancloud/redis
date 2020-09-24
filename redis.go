package redis

import (
	"encoding/json"
	"fmt"
	"github.com/coredns/coredns/request"
	"github.com/miekg/dns"
	"github.com/rverst/coredns-redis/record"
	"math"
	"strconv"
	"strings"
	"time"

	redisCon "github.com/gomodule/redigo/redis"
)

const (
	DefaultTtl        = 3600
	zoneUpdateTime    = 10 * time.Minute
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
	redis.address = p
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

func (redis *Redis) ErrorResponse(state request.Request, zone string, rcode int, err error) (int, error) {
	m := new(dns.Msg)
	m.SetRcode(state.Req, rcode)
	m.Authoritative, m.RecursionAvailable, m.Compress = true, false, true

	state.SizeAndDo(m)
	_ = state.W.WriteMsg(m)
	// Return success as the rcode to signal we have written to the client.
	return dns.RcodeSuccess, err
}

func (redis *Redis) LoadZones(name string) ([]string, error) {
	var (
		reply interface{}
		err   error
		zones []string
	)

	query := reduceZone(name)
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

// reduceZone strips the zone down to top- and second-level
// so we can query the subset from redis. This should give
// no problems unless we want to run a root dns
func reduceZone(name string) string {
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

func (redis *Redis) A(name string, _ *record.Zone, record *record.ZoneRecords) (answers, extras []dns.RR) {
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

func (redis Redis) AAAA(name string, _ *record.Zone, record *record.ZoneRecords) (answers, extras []dns.RR) {
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

func (redis *Redis) CNAME(name string, _ *record.Zone, record *record.ZoneRecords) (answers, extras []dns.RR) {
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

func (redis *Redis) TXT(name string, _ *record.Zone, record *record.ZoneRecords) (answers, extras []dns.RR) {
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

func (redis *Redis) NS(name string, z *record.Zone, record *record.ZoneRecords) (answers, extras []dns.RR) {
	for _, ns := range record.NS {
		if len(ns.Host) == 0 {
			continue
		}
		r := new(dns.NS)
		r.Hdr = dns.RR_Header{Name: dns.Fqdn(name), Rrtype: dns.TypeNS,
			Class: dns.ClassINET, Ttl: redis.ttl(ns.Ttl)}
		r.Ns = ns.Host
		answers = append(answers, r)
		extras = append(extras, redis.hosts(ns.Host, z)...)
	}
	return
}

func (redis *Redis) MX(name string, z *record.Zone, record *record.ZoneRecords) (answers, extras []dns.RR) {
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
		extras = append(extras, redis.hosts(mx.Host, z)...)
	}
	return
}

func (redis *Redis) SRV(name string, z *record.Zone, record *record.ZoneRecords) (answers, extras []dns.RR) {
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
		extras = append(extras, redis.hosts(srv.Target, z)...)
	}
	return
}

func (redis *Redis) SOA(name string, z *record.Zone, record *record.ZoneRecords) (answers, extras []dns.RR) {
	s := new(dns.SOA)
	// default value if no SOA record in backend
	if record.SOA.MName == "" {
		s.Hdr = dns.RR_Header{Name: dns.Fqdn(name), Rrtype: dns.TypeSOA,
			Class: dns.ClassINET, Ttl: redis.ttl(redis.DefaultTtl)}
		s.Ns = "ns1." + name
		s.Mbox = "hostmaster." + name
		s.Refresh = 86400
		s.Retry = 7200
		s.Expire = 3600000
		s.Minttl = 172800
	} else {
		s.Hdr = dns.RR_Header{Name: dns.Fqdn(z.Name), Rrtype: dns.TypeSOA,
			Class: dns.ClassINET, Ttl: redis.ttl(record.SOA.Ttl)}
		s.Ns = record.SOA.MName
		s.Mbox = record.SOA.RName
		s.Serial = record.SOA.Serial
		s.Refresh = record.SOA.Refresh
		s.Retry = record.SOA.Retry
		s.Expire = record.SOA.Expire
		s.Minttl = record.SOA.Minimum
	}
	if s.Serial == 0 {
		s.Serial = redis.soaSerial()
	}
	answers = append(answers, s)
	return
}

func (redis *Redis) CAA(name string, _ *record.Zone, record *record.ZoneRecords) (answers, extras []dns.RR) {
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

func (redis *Redis) AXFR(z *record.Zone) (records []dns.RR) {
	//soa, _ := redis.SOA(z.Name, z, record)
	soa := make([]dns.RR, 0)
	answers := make([]dns.RR, 0, 10)
	extras := make([]dns.RR, 0, 10)

	// Allocate slices for rr Records
	records = append(records, soa...)
	for key := range z.Locations {
		if key == "@" {
			location := redis.FindLocation(z.Name, z)
			record := redis.GetZoneRecords(location, z)
			soa, _ = redis.SOA(z.Name, z, record)
		} else {
			fqdnKey := dns.Fqdn(key) + z.Name
			var as []dns.RR
			var xs []dns.RR

			location := redis.FindLocation(fqdnKey, z)
			record := redis.GetZoneRecords(location, z)

			// Pull all zone records
			as, xs = redis.A(fqdnKey, z, record)
			answers = append(answers, as...)
			extras = append(extras, xs...)

			as, xs = redis.AAAA(fqdnKey, z, record)
			answers = append(answers, as...)
			extras = append(extras, xs...)

			as, xs = redis.CNAME(fqdnKey, z, record)
			answers = append(answers, as...)
			extras = append(extras, xs...)

			as, xs = redis.MX(fqdnKey, z, record)
			answers = append(answers, as...)
			extras = append(extras, xs...)

			as, xs = redis.SRV(fqdnKey, z, record)
			answers = append(answers, as...)
			extras = append(extras, xs...)

			as, xs = redis.TXT(fqdnKey, z, record)
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

func (redis *Redis) hosts(name string, z *record.Zone) []dns.RR {
	var (
		record  *record.ZoneRecords
		answers []dns.RR
	)
	location := redis.FindLocation(name, z)
	if location == "" {
		return nil
	}
	record = redis.GetZoneRecords(location, z)
	a, _ := redis.A(name, z, record)
	answers = append(answers, a...)
	aaaa, _ := redis.AAAA(name, z, record)
	answers = append(answers, aaaa...)
	cname, _ := redis.CNAME(name, z, record)
	answers = append(answers, cname...)
	return answers
}

func (redis *Redis) soaSerial() uint32 {
	n := time.Now().UTC()
	// calculate two digit number (0-99) based on the minute of the day, 1440 / 14.4545 = 99,0003
	c := int(math.Floor(((float64(n.Hour() + 1)) * float64(n.Minute()+1)) / 14.5454))
	ser, err := strconv.ParseUint(fmt.Sprintf("%s%02d", n.Format("20060102"), c), 10, 32)
	if err != nil {
		return uint32(time.Now().Unix())
	}
	return uint32(ser)
}

func (redis *Redis) ttl(ttl int) uint32 {
	if ttl >= 0 {
		return uint32(ttl)
	}
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

func (redis *Redis) GetZoneRecords(key string, z *record.Zone) *record.ZoneRecords {
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
	r := new(record.ZoneRecords)
	err = json.Unmarshal([]byte(val), r)
	if err != nil {
		fmt.Println("parse error : ", val, err)
		return nil
	}
	return r
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

func (redis *Redis) Save(zone string, subdomain string, value string) error {
	var err error

	conn := redis.Pool.Get()
	if conn == nil {
		fmt.Println("error connecting to redis")
		return nil
	}
	defer conn.Close()

	_, err = conn.Do("HSET", redis.keyPrefix+zone+redis.keySuffix, subdomain, value)
	return err
}

func (redis *Redis) Load(zone string) *record.Zone {
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
	if err != nil {
		return nil
	}
	z := new(record.Zone)
	z.Name = zone
	vals, err = redisCon.Strings(reply, nil)
	if err != nil {
		return nil
	}
	z.Locations = make(map[string][]record.DnsRecord)
	for _, val := range vals {
		z.Locations[val] = nil //struct{}{}
	}

	return z
}

// Key returns the given key with prefix and suffix
func (redis *Redis) Key(key string) string {
	return redis.keyPrefix + key + redis.keySuffix
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

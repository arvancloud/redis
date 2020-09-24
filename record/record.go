package record

import (
	"fmt"
	"github.com/miekg/dns"
	"net"
)

type Type string

// a record type tha implements the Equality interface
// can easily been checked for changes in the record
type Equality interface {
	// Equal returns 'false', if the values of both instances differ, otherwise 'true'
	Equal(r1 Equality) bool
}

type Record interface {
	// TTL returns the ttl for the record, if the second return value is 'false', no
	// ttl is set for the record and you should use the default min-ttl value
	TTL() (uint32, bool)
}

// Zone represents a DNS zone
type Zone struct {
	Name      string
	Locations map[string]Records
}

// NewZone return a new, empty zone record with a SOA resource record
func NewZone(name string, soa SOA) *Zone {
	l := make(map[string]Records)
	l["@"] = Records{
		SOA: soa,
	}

	return &Zone{
		Name: dns.Fqdn(name),
		Locations: l,
	}
}

// Equal determines if the zones are equal
func (z Zone) Equal(zone Zone) bool {
	if z.Name != zone.Name {
		return false
	}
	return false
}

func (z Zone) SOA() (*SOA, error) {
	r, ok := z.Locations["@"]
	if !ok {
		return nil, fmt.Errorf("no SOA record found")
	}
	return &r.SOA, nil
}

func (z Zone) Add(loc string, record Record) {

	switch record.(type) {
	case A:
		z.AddA(loc, record.(A))
	}
}

func (z Zone) AddA(loc string, rec A) {
	_, ok := z.Locations[loc]
	if !ok {
		z.Locations[loc] = Records{}
	}
	r := z.Locations[loc]
	if r.A == nil {
		r.A = make([]A, 1)
		r.A[0] = rec
	} else {
		r.A = append(r.A, rec)
	}
	z.Locations[loc] = r
}

// Records holds the location records for a zone
type Records struct {
	// SOA record for the zone, mandatory but only allowed in '@'
	SOA   SOA     `json:"soa,omitempty"`
	A     []A     `json:"a,omitempty"`
	AAAA  []AAAA  `json:"aaaa,omitempty"`
	TXT   []TXT   `json:"txt,omitempty"`
	CNAME []CNAME `json:"cname,omitempty"`
	NS    []NS    `json:"ns,omitempty"`
	MX    []MX    `json:"mx,omitempty"`
	SRV   []SRV   `json:"srv,omitempty"`
	CAA   []CAA   `json:"caa,omitempty"`
}

type A struct {
	Ttl int    `json:"ttl,omitempty"`
	Ip  net.IP `json:"ip"`
}

// Equal reports if the records are equal
func (a A) Equal(b A) bool {
	return a.Ttl == b.Ttl && a.Ip.Equal(b.Ip)
}

func (a A) TTL() (uint32, bool) {
	if a.Ttl >= 0 {
		return uint32(a.Ttl), true
	}
	return 0, false
}

type AAAA struct {
	Ttl int    `json:"ttl"`
	Ip  net.IP `json:"ip"`
}

// Equal determines if the record is equal
func (a AAAA) Equal(b AAAA) bool {
	return a.Ttl == b.Ttl && a.Ip.Equal(b.Ip)
}

func (a AAAA) TTL() (uint32, bool) {
	if a.Ttl >= 0 {
		return uint32(a.Ttl), true
	}
	return 0, false
}

type TXT struct {
	Ttl  int    `json:"ttl"`
	Text string `json:"text"`
}

// Equal determines if the record is equal
func (a TXT) Equal(b TXT) bool {
	return a.Ttl == b.Ttl && a.Text == b.Text
}

func (a TXT) TTL() (uint32, bool) {
	if a.Ttl >= 0 {
		return uint32(a.Ttl), true
	}
	return 0, false
}

type CNAME struct {
	Ttl  int    `json:"ttl"`
	Host string `json:"host"`
}

// Equal determines if the record is equal
func (a CNAME) Equal(b CNAME) bool {
	return a.Ttl == b.Ttl && a.Host == b.Host
}

func (a CNAME) TTL() (uint32, bool) {
	if a.Ttl >= 0 {
		return uint32(a.Ttl), true
	}
	return 0, false
}

type NS struct {
	Ttl  int    `json:"ttl"`
	Host string `json:"host"`
}

// Equal determines if the record is equal
func (a NS) Equal(b NS) bool {
	return a.Ttl == b.Ttl && a.Host == b.Host
}

func (a NS) TTL() (uint32, bool) {
	if a.Ttl >= 0 {
		return uint32(a.Ttl), true
	}
	return 0, false
}

type MX struct {
	Ttl        int    `json:"ttl"`
	Host       string `json:"host"`
	Preference uint16 `json:"preference"`
}

// Equal determines if the record is equal
func (a MX) Equal(b MX) bool {
	return a.Ttl == b.Ttl && a.Host == b.Host && a.Preference == b.Preference
}

func (a MX) TTL() (uint32, bool) {
	if a.Ttl >= 0 {
		return uint32(a.Ttl), true
	}
	return 0, false
}

type SRV struct {
	Ttl      int    `json:"ttl"`
	Priority uint16 `json:"priority"`
	Weight   uint16 `json:"weight"`
	Port     uint16 `json:"port"`
	Target   string `json:"target"`
}

// Equal determines if the record is equal
func (a SRV) Equal(b SRV) bool {
	return a.Ttl == b.Ttl && a.Priority == b.Priority && a.Weight == b.Weight &&
		a.Port == b.Port && a.Target == b.Target
}

func (a SRV) TTL() (uint32, bool) {
	if a.Ttl >= 0 {
		return uint32(a.Ttl), true
	}
	return 0, false
}

// SOA RDATA (https://tools.ietf.org/html/rfc1035#section-3.3.13)
type SOA struct {
	Ttl     int    `json:"ttl"`
	MName   string `json:"mname"`
	RName   string `json:"rname"`
	Serial  uint32 `json:"serial"`
	Refresh uint32 `json:"refresh"`
	Retry   uint32 `json:"retry"`
	Expire  uint32 `json:"expire"`
	MinTtl  uint32 `json:"min_ttl"`
}

func (a SOA) TTL() (uint32, bool) {
	if a.Ttl >= 0 {
		return uint32(a.Ttl), true
	}
	return 0, false
}

// Equal determines if the record is equal
func (a SOA) Equal(b SOA) bool {
	return a.Ttl == b.Ttl && a.MName == b.MName && a.RName == b.RName &&
		a.Serial == b.Serial && a.Refresh == b.Refresh && a.Retry == b.Retry &&
		a.Expire == b.Expire && a.MinTtl == b.MinTtl
}

type CAA struct {
	Ttl   int    `json:"ttl"`
	Flag  uint8  `json:"flag"`
	Tag   string `json:"tag"`
	Value string `json:"value"`
}

func (a CAA) TTL() (uint32, bool) {
	if a.Ttl >= 0 {
		return uint32(a.Ttl), true
	}
	return 0, false
}

// Equal determines if the record is equal
func (a CAA) Equal(b CAA) bool {
	return a.Ttl == b.Ttl && a.Flag == b.Flag && a.Tag == b.Tag && a.Value == b.Value
}

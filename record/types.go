package record

import (
	"net"
)

type Type string

const (
	TypeSOA   = Type("SOA")
	TypeA     = Type("A")
	TypeAAAA  = Type("AAAA")
	TypeTXT   = Type("TXT")
	TypeCNAME = Type("CNAME")
	TypeNS    = Type("NS")
	TypeMX    = Type("MX")
	TypeSRV   = Type("SRV")
	TypeCAA   = Type("CAA")
)

type DnsRecord interface {
	Equal(r1 DnsRecord) bool
}

type Zone struct {
	Name      string
	Locations map[string][]DnsRecord
}

// Equal determines if the zones are equal
func (z1 Zone) Equal(z Zone) bool {
	if z1.Name != z.Name {
		return false
	}
	return false
}

type ZoneRecords struct {
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
	Ttl uint32 `json:"ttl,omitempty"`
	Ip  net.IP `json:"ip"`
}

// Equal determines if the record is equal
func (a A) Equal(b A) bool {
	return a.Ttl == b.Ttl && a.Ip.Equal(b.Ip)
}

type AAAA struct {
	Ttl uint32 `json:"ttl,omitempty"`
	Ip  net.IP `json:"ip"`
}

// Equal determines if the record is equal
func (a AAAA) Equal(b AAAA) bool {
	return a.Ttl == b.Ttl && a.Ip.Equal(b.Ip)
}

type TXT struct {
	Ttl  uint32 `json:"ttl,omitempty"`
	Text string `json:"text"`
}

// Equal determines if the record is equal
func (a TXT) Equal(b TXT) bool {
	return a.Ttl == b.Ttl && a.Text == b.Text
}

type CNAME struct {
	Ttl  uint32 `json:"ttl,omitempty"`
	Host string `json:"host"`
}

// Equal determines if the record is equal
func (a CNAME) Equal(b CNAME) bool {
	return a.Ttl == b.Ttl && a.Host == b.Host
}

type NS struct {
	Ttl  uint32 `json:"ttl,omitempty"`
	Host string `json:"host"`
}

// Equal determines if the record is equal
func (a NS) Equal(b NS) bool {
	return a.Ttl == b.Ttl && a.Host == b.Host
}

type MX struct {
	Ttl        uint32 `json:"ttl,omitempty"`
	Host       string `json:"host"`
	Preference uint16 `json:"preference"`
}

// Equal determines if the record is equal
func (a MX) Equal(b MX) bool {
	return a.Ttl == b.Ttl && a.Host == b.Host && a.Preference == b.Preference
}

type SRV struct {
	Ttl      uint32 `json:"ttl,omitempty"`
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

// SOA RDATA (https://tools.ietf.org/html/rfc1035#section-3.3.13)
type SOA struct {
	Ttl     uint32 `json:"ttl,omitempty"`
	MName   string `json:"mname"`
	RName   string `json:"rname"`
	Serial  uint32 `json:"serial"`
	Refresh uint32 `json:"refresh"`
	Retry   uint32 `json:"retry"`
	Expire  uint32 `json:"expire"`
	Minimum uint32 `json:"minimum"`
}

// Equal determines if the record is equal
func (a SOA) Equal(b SOA) bool {
	return a.Ttl == b.Ttl && a.MName == b.MName && a.RName == b.RName &&
		a.Serial == b.Serial && a.Refresh == b.Refresh && a.Retry == b.Retry &&
		a.Expire == b.Expire && a.Minimum == b.Minimum
}

type CAA struct {
	Ttl   uint32 `json:"ttl,omitempty"`
	Flag  uint8  `json:"flag"`
	Tag   string `json:"tag"`
	Value string `json:"value"`
}

// Equal determines if the record is equal
func (a CAA) Equal(b CAA) bool {
	return a.Ttl == b.Ttl && a.Flag == b.Flag && a.Tag == b.Tag && a.Value == b.Value
}

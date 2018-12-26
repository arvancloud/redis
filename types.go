package redis

import "net"

type Zone struct {
	Name      string
	Locations map[string]struct{}
}

type Record struct {
	A     []A_Record `json:"a,omitempty"`
	AAAA  []AAAA_Record `json:"aaaa,omitempty"`
	TXT   []TXT_Record `json:"txt,omitempty"`
	CNAME []CNAME_Record `json:"cname,omitempty"`
	NS    []NS_Record `json:"ns,omitempty"`
	MX    []MX_Record `json:"mx,omitempty"`
	SRV   []SRV_Record `json:"srv,omitempty"`
	CAA   []CAA_Record `json:"caa,omitempty"`
	SOA   SOA_Record `json:"soa,omitempty"`
}

type A_Record struct {
	Ttl uint32 `json:"ttl,omitempty"`
	Ip  net.IP `json:"ip"`
}

type AAAA_Record struct {
	Ttl uint32 `json:"ttl,omitempty"`
	Ip  net.IP `json:"ip"`
}

type TXT_Record struct {
	Ttl  uint32 `json:"ttl,omitempty"`
	Text string `json:"text"`
}

type CNAME_Record struct {
	Ttl  uint32 `json:"ttl,omitempty"`
	Host string `json:"host"`
}

type NS_Record struct {
	Ttl  uint32 `json:"ttl,omitempty"`
	Host string `json:"host"`
}

type MX_Record struct {
	Ttl        uint32 `json:"ttl,omitempty"`
	Host       string `json:"host"`
	Preference uint16 `json:"preference"`
}

type SRV_Record struct {
	Ttl      uint32 `json:"ttl,omitempty"`
	Priority uint16 `json:"priority"`
	Weight   uint16 `json:"weight"`
	Port     uint16 `json:"port"`
	Target   string `json:"target"`
}

type SOA_Record struct {
	Ttl     uint32 `json:"ttl,omitempty"`
	Ns      string `json:"ns"`
	MBox    string `json:"MBox"`
	Refresh uint32 `json:"refresh"`
	Retry   uint32 `json:"retry"`
	Expire  uint32 `json:"expire"`
	MinTtl  uint32 `json:"minttl"`
}

type CAA_Record struct {
	Flag  uint8 `json:"flag"`
	Tag   string `json:"tag"`
	Value string `json:"value"`
}
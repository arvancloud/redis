package record

import "net"

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

// Equal determines if the record is equal, the serial number is ignored
func (a SOA) Equal(b SOA) bool {
	return a.Ttl == b.Ttl && a.MName == b.MName && a.RName == b.RName &&
		a.Refresh == b.Refresh && a.Retry == b.Retry &&
		a.Expire == b.Expire && a.MinTtl == b.MinTtl
}

func (a *SOA) IncreaseSerial() {
	if a.Serial == 0 {
		a.Serial = DefaultSerial()
	} else {
		if s, err := IncrementSerial(a.Serial); err == nil {
			a.Serial = s
		}
	}
}

type A struct {
	Ttl int    `json:"ttl"`
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

func aEqual(a, b []A) bool {
	if len(a) != len(b) {
		return false
	}
	c := 0
	for _, ax := range a {
		for _, bx := range b {
			if ax.Ip.Equal(bx.Ip) {
				if !ax.Equal(bx) {
					return false
				}
				c++
			}
		}
	}
	if len(a) != c {
		return false
	}
	return true
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

func aaaaEqual(a, b []AAAA) bool {
	if len(a) != len(b) {
		return false
	}
	c := 0
	for _, ax := range a {
		for _, bx := range b {
			if ax.Ip.Equal(bx.Ip) {
				if !ax.Equal(bx) {
					return false
				}
				c++
			}
		}
	}
	if len(a) != c {
		return false
	}
	return true
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

func txtEqual(a, b []TXT) bool {
	if len(a) != len(b) {
		return false
	}
	c := 0
	for _, ax := range a {
		for _, bx := range b {
			if ax.Text == bx.Text {
				if !ax.Equal(bx) {
					return false
				}
				c++
			}
		}
	}
	if len(a) != c {
		return false
	}
	return true
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

func cnameEqual(a, b []CNAME) bool {
	if len(a) != len(b) {
		return false
	}
	c := 0
	for _, ax := range a {
		for _, bx := range b {
			if ax.Host == bx.Host {
				if !ax.Equal(bx) {
					return false
				}
				c++
			}
		}
	}
	if len(a) != c {
		return false
	}
	return true
}

// NS RDATA (https://tools.ietf.org/html/rfc1035#section-3.3.11)
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

func nsEqual(a, b []NS) bool {
	if len(a) != len(b) {
		return false
	}
	c := 0
	for _, ax := range a {
		for _, bx := range b {
			if ax.Host == bx.Host {
				if !ax.Equal(bx) {
					return false
				}
				c++
			}
		}
	}
	if len(a) != c {
		return false
	}
	return true
}

// MX RDATA (https://tools.ietf.org/html/rfc1035#section-3.3.9)
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

func mxEqual(a, b []MX) bool {
	if len(a) != len(b) {
		return false
	}
	c := 0
	for _, ax := range a {
		for _, bx := range b {
			if ax.Host == bx.Host {
				if !ax.Equal(bx) {
					return false
				}
				c++
			}
		}
	}
	if len(a) != c {
		return false
	}
	return true
}

// SRV RDATA (https://tools.ietf.org/html/rfc2782)
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

func srvEqual(a, b []SRV) bool {
	if len(a) != len(b) {
		return false
	}
	c := 0
	for _, ax := range a {
		for _, bx := range b {
			if ax.Target == bx.Target {
				if !ax.Equal(bx) {
					return false
				}
				c++
			}
		}
	}
	if len(a) != c {
		return false
	}
	return true
}

type PTR struct {
	Ttl  int    `json:"ttl"`
	Name string `json:"name"`
}

// Equal determines if the record is equal
func (a PTR) Equal(b PTR) bool {
	return a.Ttl == b.Ttl && a.Name == b.Name
}

func (a PTR) TTL() (uint32, bool) {
	if a.Ttl >= 0 {
		return uint32(a.Ttl), true
	}
	return 0, false
}

func ptrEqual(a, b []PTR) bool {
	if len(a) != len(b) {
		return false
	}
	c := 0
	for _, ax := range a {
		for _, bx := range b {
			if ax.Name == bx.Name {
				if !ax.Equal(bx) {
					return false
				}
				c++
			}
		}
	}
	if len(a) != c {
		return false
	}
	return true
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

func caaEqual(a, b []CAA) bool {
	if len(a) != len(b) {
		return false
	}
	c := 0
	for _, ax := range a {
		for _, bx := range b {
			if ax.Value == bx.Value {
				if !ax.Equal(bx) {
					return false
				}
				c++
			}
		}
	}
	if len(a) != c {
		return false
	}
	return true
}

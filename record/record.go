package record

import (
	"fmt"
	"github.com/miekg/dns"
	"log"
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

// Records holds the location records for a zone
type Records struct {
	// SOA record for the zone, mandatory but only allowed in '@'
	SOA   *SOA    `json:"SOA,omitempty"`
	A     []A     `json:"A,omitempty"`
	AAAA  []AAAA  `json:"AAAA,omitempty"`
	TXT   []TXT   `json:"TXT,omitempty"`
	CNAME []CNAME `json:"CNAME,omitempty"`
	NS    []NS    `json:"NS,omitempty"`
	MX    []MX    `json:"MX,omitempty"`
	SRV   []SRV   `json:"SRV,omitempty"`
	CAA   []CAA   `json:"CAA,omitempty"`
}

// Zone represents a DNS zone
type Zone struct {
	Name      string
	Locations map[string]Records
}

// NewZone return a new, empty zone with a SOA resource record
func NewZone(name string, soa SOA) *Zone {
	z := Zone{
		Name: dns.Fqdn(name),
	}
	z.addSoa(soa)
	return &z
}

// Equal determines if the zones are equal, the function call
// .Equal() on every record in the zone
func (z Zone) Equal(zone Zone) bool {
	if z.Name != zone.Name {
		return false
	}
	for loc, rec := range z.Locations {
		r2, ok := zone.Locations[loc]
		if !ok {
			return false
		}

		if loc == "@" && rec.SOA != nil {
			if !rec.SOA.Equal(*r2.SOA) {
				return false
			}
		}

		if !aEqual(rec.A, r2.A) {
			return false
		}

		if !aaaaEqual(rec.AAAA, r2.AAAA) {
			return false
		}

		if !txtEqual(rec.TXT, r2.TXT) {
			return false
		}

		if !nsEqual(rec.NS, r2.NS) {
			return false
		}

		if !mxEqual(rec.MX, r2.MX) {
			return false
		}

		if !cnameEqual(rec.CNAME, r2.CNAME) {
			return false
		}

		if !srvEqual(rec.SRV, r2.SRV) {
			return false
		}

		if !caaEqual(rec.CAA, r2.CAA) {
			return false
		}

	}
	return true
}

func (z Zone) SOA() (*SOA, error) {
	r, ok := z.Locations["@"]
	if !ok {
		return nil, fmt.Errorf("no SOA record found")
	}
	return r.SOA, nil
}

func (z *Zone) Add(loc string, record Record) {
	switch record.(type) {
	case SOA:
		z.addSoa(record.(SOA))
	case A:
		z.addA4(loc, record.(A))
	case AAAA:
		z.addA6(loc, record.(AAAA))
	case TXT:
		z.addTxt(loc, record.(TXT))
	case CNAME:
		z.addCname(loc, record.(CNAME))
	case MX:
		z.addMx(loc, record.(MX))
	case NS:
		z.addNs(loc, record.(NS))
	case SRV:
		z.addSrv(loc, record.(SRV))
	case CAA:
		z.addCaa(loc, record.(CAA))
	default:
		log.Fatal(fmt.Errorf("record type: %T is not supported", record))
	}
}

func (z *Zone) addSoa(rec SOA) {
	r := z.getOrAddLocation("@")
	r.SOA = &rec
	z.Locations["@"] = r
}

func (z *Zone) addA4(loc string, rec A) {
	r := z.getOrAddLocation(loc)
	if r.A == nil {
		r.A = make([]A, 1)
		r.A[0] = rec
	} else {
		r.A = append(r.A, rec)
	}
	z.Locations[loc] = r
}

func (z *Zone) addA6(loc string, rec AAAA) {
	r := z.getOrAddLocation(loc)
	if r.AAAA == nil {
		r.AAAA = make([]AAAA, 1)
		r.AAAA[0] = rec
	} else {
		r.AAAA = append(r.AAAA, rec)
	}
	z.Locations[loc] = r
}

func (z *Zone) addTxt(loc string, rec TXT) {
	r := z.getOrAddLocation(loc)
	if r.TXT == nil {
		r.TXT = make([]TXT, 1)
		r.TXT[0] = rec
	} else {
		r.TXT = append(r.TXT, rec)
	}
	z.Locations[loc] = r
}

func (z *Zone) addCname(loc string, rec CNAME) {
	r := z.getOrAddLocation(loc)
	if r.CNAME == nil {
		r.CNAME = make([]CNAME, 1)
		r.CNAME[0] = rec
	} else {
		r.CNAME = append(r.CNAME, rec)
	}
	z.Locations[loc] = r
}

func (z *Zone) addMx(loc string, rec MX) {
	r := z.getOrAddLocation(loc)
	if r.MX == nil {
		r.MX = make([]MX, 1)
		r.MX[0] = rec
	} else {
		r.MX = append(r.MX, rec)
	}
	z.Locations[loc] = r
}

func (z *Zone) addNs(loc string, rec NS) {
	r := z.getOrAddLocation(loc)
	if r.NS == nil {
		r.NS = make([]NS, 1)
		r.NS[0] = rec
	} else {
		r.NS = append(r.NS, rec)
	}
	z.Locations[loc] = r
}

func (z *Zone) addSrv(loc string, rec SRV) {
	r := z.getOrAddLocation(loc)
	if r.SRV == nil {
		r.SRV = make([]SRV, 1)
		r.SRV[0] = rec
	} else {
		r.SRV = append(r.SRV, rec)
	}
	z.Locations[loc] = r
}

func (z *Zone) addCaa(loc string, rec CAA) {
	r := z.getOrAddLocation(loc)
	if r.CAA == nil {
		r.CAA = make([]CAA, 1)
		r.CAA[0] = rec
	} else {
		r.CAA = append(r.CAA, rec)
	}
	z.Locations[loc] = r
}

func (z *Zone) getOrAddLocation(loc string) Records {
	if z.Locations == nil {
		z.Locations = make(map[string]Records)
	}
	_, ok := z.Locations[loc]
	if !ok {
		z.Locations[loc] = Records{}
	}
	r := z.Locations[loc]
	return r
}

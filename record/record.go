package record

import (
	"fmt"
	"github.com/miekg/dns"
	"log"
	"sort"
	"strings"
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
	PTR   []PTR   `json:"PTR,omitempty"`
	CAA   []CAA   `json:"CAA,omitempty"`
}

// appends the zoneName to values which are not fully qualified domain names
func (r *Records) MakeFqdn(zoneName string) {

	if len(zoneName) == 0 {
		return
	} else if zoneName[0] != '.' {
		zoneName = "." + zoneName
	}
	zoneName = dns.Fqdn(zoneName)

	if r.SOA != nil {
		if !dns.IsFqdn(r.SOA.MName) {
			r.SOA.MName += zoneName
		}
		if !dns.IsFqdn(r.SOA.RName) {
			r.SOA.RName += zoneName
		}
	}

	for i := range r.CNAME {
		if !dns.IsFqdn(r.CNAME[i].Host) {
			r.CNAME[i].Host += zoneName
		}
	}
	for i := range r.MX {
		if !dns.IsFqdn(r.MX[i].Host) {
			r.MX[i].Host += zoneName
		}
	}
	for i := range r.NS {
		if !dns.IsFqdn(r.NS[i].Host) {
			r.NS[i].Host += zoneName
		}
	}
	for i := range r.SRV {
		if !dns.IsFqdn(r.SRV[i].Target) {
			r.SRV[i].Target += zoneName
		}
	}
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

		if !ptrEqual(rec.PTR, r2.PTR) {
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
	case PTR:
		z.addPtr(loc, record.(PTR))
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

func (z Zone) addPtr(loc string, rec PTR) {
	r := z.getOrAddLocation(loc)
	if r.PTR == nil {
		r.PTR = make([]PTR, 1)
		r.PTR[0] = rec
	} else {
		r.PTR = append(r.PTR, rec)
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

func (z Zone) String() (str string) {

	defer func() {
		if r := recover(); r != nil {
			str = fmt.Sprintf("ERROR: %s", r)
		}
	}()
	var (
		i   int
		err error
		sb  = strings.Builder{}
	)

	keys := make([]string, 0, len(z.Locations))
	maxL := 0
	for k := range z.Locations {
		keys = append(keys, k)
		if len(k) > maxL {
			maxL = len(k)
		}
	}
	sort.Strings(keys)
	i, err = sb.WriteString(fmt.Sprintf("$ORIGIN  %s\n\n", z.Name))
	checkError(i, err)

	if s, err := z.SOA(); err == nil {
		i, err = sb.WriteString(fmt.Sprintf("%s%d     IN     SOA     %s %s %10d %d %d %d %d\n",
			spacedLoc("@", maxL), s.Ttl, s.MName, s.RName, s.Serial, s.Refresh, s.Expire, s.Retry, s.MinTtl))
		checkError(i, err)

	}

	for _, k := range keys {
		loc := k
		if loc == "@" {
			loc = " "
		}
		for _, r := range z.Locations[k].A {
			i, err = sb.WriteString(fmt.Sprintf("%s%d     IN     A       %s\n",
				spacedLoc(loc, maxL), r.Ttl, r.Ip))
			checkError(i, err)
		}
		for _, r := range z.Locations[k].AAAA {
			i, err = sb.WriteString(fmt.Sprintf("%s%d     IN     AAAA    %s\n",
				spacedLoc(loc, maxL), r.Ttl, r.Ip))
			checkError(i, err)
		}
		for _, r := range z.Locations[k].TXT {
			i, err = sb.WriteString(fmt.Sprintf("%s%d     IN     TXT     %s\n",
				spacedLoc(loc, maxL), r.Ttl, r.Text))
			checkError(i, err)
		}
		for _, r := range z.Locations[k].CNAME {
			i, err = sb.WriteString(fmt.Sprintf("%s%d     IN     CNAME    %s\n",
				spacedLoc(loc, maxL), r.Ttl, r.Host))
			checkError(i, err)
		}
		for _, r := range z.Locations[k].MX {
			i, err = sb.WriteString(fmt.Sprintf("%s%d     IN     MX      %d %s\n",
				spacedLoc(loc, maxL), r.Ttl, r.Preference, r.Host))
			checkError(i, err)
		}
		for _, r := range z.Locations[k].NS {
			i, err = sb.WriteString(fmt.Sprintf("%s%d     IN     NS      %s\n",
				spacedLoc(loc, maxL), r.Ttl, r.Host))
			checkError(i, err)
		}
		for _, r := range z.Locations[k].SRV {
			i, err = sb.WriteString(fmt.Sprintf("%s%d     IN     SRV     %d %d %d %s\n",
				spacedLoc(loc, maxL), r.Ttl, r.Priority, r.Weight, r.Port, r.Target))
			checkError(i, err)
		}
		for _, r := range z.Locations[k].PTR {
			i, err = sb.WriteString(fmt.Sprintf("%s%d     IN     SRV     %s\n",
				spacedLoc(loc, maxL), r.Ttl, r.Name))
			checkError(i, err)
		}
		for _, r := range z.Locations[k].CAA {
			i, err = sb.WriteString(fmt.Sprintf("%s%d     IN     CAA     %d %s %s\n",
				spacedLoc(loc, maxL), r.Ttl, r.Flag, r.Tag, r.Value))
			checkError(i, err)
		}
		if k == "@" {
			i, err = sb.WriteString("\n")
			checkError(i, err)
		}
	}

	i, err = sb.WriteString("\n")
	checkError(i, err)

	return sb.String()
}

func spacedLoc(k string, l int) string {
	l -= len(k)
	if l < 0 {
		l = 5
	} else if l > 10 {
		l += 1
	} else {
		l += 3
	}
	return fmt.Sprintf("%s%s", k, strings.Repeat(" ", l))
}

func checkError(i int, err error) {
	if err != nil {
		panic(err)
	}
	if i == 0 {
		panic("nothing written")
	}
}

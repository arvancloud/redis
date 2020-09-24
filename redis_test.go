package redis

import (
	"github.com/rverst/coredns-redis/record"
	"log"
	"net"
	"testing"
)

const prefix, suffix = "", ""
const minTtl = 300

func newRedis() *Redis{

	r := New()
	r.SetKeyPrefix(prefix)
	r.SetKeySuffix(suffix)
	r.SetDefaultTtl(minTtl)
	r.SetAddress("192.168.0.100:6379")
	if err := r.Connect(); err != nil {
		log.Fatal(err)
	}
	return r
}


func TestRedis_SaveZone(t *testing.T) {

	z := record.NewZone("example.de", record.SOA{
		Ttl:     1234,
		MName:   "ns1.example.net",
		RName:   "hostmaster.example.net",
		Serial:  202009241,
		Refresh: 3600,
		Retry:   1800,
		Expire:  10000,
		MinTtl:  40,
	})

	z.Add("@", record.A{
		Ttl: 42,
		Ip:  net.ParseIP("1.2.3.4"),
	})
	z.AddA("@", record.A{
		Ttl: 23,
		Ip:  net.ParseIP("2.3.4.5"),
	})

	z.Add("sub", record.A{
		Ttl: 43,
		Ip:  net.ParseIP("1.2.3.4"),
	})
	z.AddA("sub", record.A{
		Ttl: 24,
		Ip:  net.ParseIP("2.3.4.5"),
	})
	tests := []struct {
		name    string
		zone    record.Zone
		wantErr bool
	}{
		{name: "Test1", zone: *z, wantErr: false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			redis := newRedis()
			if err := redis.SaveZone(tt.zone); (err != nil) != tt.wantErr {
				t.Errorf("SaveZone() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

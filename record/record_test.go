package record

import (
	"net"
	"testing"
)

func TestZone_Equal(t *testing.T) {

	soa := &SOA{
		Ttl:     42,
		MName:   "example.net.",
		RName:   "hostmaster.example.net.",
		Serial:  2020022002,
		Refresh: 3600,
		Retry:   1800,
		Expire:  9999,
		MinTtl:  1337,
	}

	type fields struct {
		Name      string
		Locations map[string]Records
	}
	type args struct {
		zone Zone
	}
	tests := []struct {
		name   string
		fields fields
		args   args
		want   bool
	}{
		{name: "SOA equal", fields: fields{
			Name: "example.net.",
			Locations: map[string]Records{
				"@": {
					SOA: &SOA{
						Ttl:     42,
						MName:   "example.net.",
						RName:   "hostmaster.example.net.",
						Serial:  2020022002,
						Refresh: 3600,
						Retry:   1800,
						Expire:  9999,
						MinTtl:  1337,
					},
				},
			},
		}, args: args{Zone{
			Name: "example.net.",
			Locations: map[string]Records{
				"@": {
					SOA: &SOA{
						Ttl:     42,
						MName:   "example.net.",
						RName:   "hostmaster.example.net.",
						Serial:  2020022002,
						Refresh: 3600,
						Retry:   1800,
						Expire:  9999,
						MinTtl:  1337,
					},
				},
			},
		}}, want: true},

		{name: "SOA equal, serial differs", fields: fields{
			Name: "example.net.",
			Locations: map[string]Records{
				"@": {
					SOA: &SOA{
						Ttl:     42,
						MName:   "example.net.",
						RName:   "hostmaster.example.net.",
						Serial:  2020022002,
						Refresh: 3600,
						Retry:   1800,
						Expire:  9999,
						MinTtl:  1337,
					},
				},
			},
		}, args: args{Zone{
			Name: "example.net.",
			Locations: map[string]Records{
				"@": {
					SOA: &SOA{
						Ttl:     42,
						MName:   "example.net.",
						RName:   "hostmaster.example.net.",
						Serial:  2020022003,
						Refresh: 3600,
						Retry:   1800,
						Expire:  9999,
						MinTtl:  1337,
					},
				},
			},
		}}, want: true},

		{name: "SOA not equal", fields: fields{
			Name: "example.net.",
			Locations: map[string]Records{
				"@": {
					SOA: &SOA{
						Ttl:     42,
						MName:   "example.net.",
						RName:   "hostmaster.example.net.",
						Serial:  2020022003,
						Refresh: 3600,
						Retry:   1800,
						Expire:  9999,
						MinTtl:  1337,
					},
				},
			},
		}, args: args{Zone{
			Name: "example.net.",
			Locations: map[string]Records{
				"@": {
					SOA: &SOA{
						Ttl:     42,
						MName:   "example.net.",
						RName:   "webmaster.example.net.",
						Serial:  2020022002,
						Refresh: 3600,
						Retry:   1800,
						Expire:  9999,
						MinTtl:  1337,
					},
				},
			},
		}}, want: false},

		{name: "A equal 1", fields: fields{
			Name: "example.net.",
			Locations: map[string]Records{
				"@": {
					SOA: soa,
					A:   []A{{Ttl: 42, Ip: net.ParseIP("1.1.1.1")}, {Ttl: 42, Ip: net.ParseIP("1.1.1.2")}},
				},
			},
		}, args: args{Zone{
			Name: "example.net.",
			Locations: map[string]Records{
				"@": {
					SOA: soa,
					A:   []A{{Ttl: 42, Ip: net.ParseIP("1.1.1.1")}, {Ttl: 42, Ip: net.ParseIP("1.1.1.2")}},
				},
			},
		}}, want: true},

		{name: "A equal 2", fields: fields{
			Name: "example.net.",
			Locations: map[string]Records{
				"@": {
					SOA: soa,
					A:   []A{{Ttl: 42, Ip: net.ParseIP("1.1.1.2")}, {Ttl: 42, Ip: net.ParseIP("1.1.1.1")}},
				},
			},
		}, args: args{Zone{
			Name: "example.net.",
			Locations: map[string]Records{
				"@": {
					SOA: soa,
					A:   []A{{Ttl: 42, Ip: net.ParseIP("1.1.1.1")}, {Ttl: 42, Ip: net.ParseIP("1.1.1.2")}},
				},
			},
		}}, want: true},

		{name: "A equal differ 1", fields: fields{
			Name: "example.net.",
			Locations: map[string]Records{
				"@": {
					SOA: soa,
					A:   []A{{Ttl: 42, Ip: net.ParseIP("1.1.1.2")}, {Ttl: 42, Ip: net.ParseIP("1.1.1.1")}},
				},
			},
		}, args: args{Zone{
			Name: "example.net.",
			Locations: map[string]Records{
				"@": {
					SOA: soa,
					A:   []A{{Ttl: 42, Ip: net.ParseIP("1.1.1.1")}, {Ttl: 42, Ip: net.ParseIP("1.1.1.3")}},
				},
			},
		}}, want: false},

		{name: "A equal differ 2", fields: fields{
			Name: "example.net.",
			Locations: map[string]Records{
				"@": {
					SOA: soa,
				},
			},
		}, args: args{Zone{
			Name: "example.net.",
			Locations: map[string]Records{
				"@": {
					SOA: soa,
					A:   []A{{Ttl: 42, Ip: net.ParseIP("1.1.1.1")}, {Ttl: 42, Ip: net.ParseIP("1.1.1.3")}},
				},
			},
		}}, want: false},


	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			z := Zone{
				Name:      tt.fields.Name,
				Locations: tt.fields.Locations,
			}
			if got := z.Equal(tt.args.zone); got != tt.want {
				t.Errorf("Equal() = %v, want %v", got, tt.want)
			}
		})
	}
}

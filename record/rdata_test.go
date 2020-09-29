package record

import (
	"net"
	"testing"
)

func Test_aEqual(t *testing.T) {
	type args struct {
		a []A
		b []A
	}
	tests := []struct {
		name string
		args args
		want bool
	}{
		{name: "equal 1", args: args{
			a: []A{{42, net.ParseIP("1.1.1.1")}},
			b: []A{{42, net.ParseIP("1.1.1.1")}},
		}, want: true},

		{name: "equal 2", args: args{
			a: []A{{42, net.ParseIP("1.1.1.1")}, {42, net.ParseIP("1.1.1.2")}},
			b: []A{{42, net.ParseIP("1.1.1.1")}, {42, net.ParseIP("1.1.1.2")}},
		}, want: true},

		{name: "equal 4", args: args{
			a: []A{{42, net.ParseIP("1.1.1.1")}, {42, net.ParseIP("1.1.1.2")},
				{42, net.ParseIP("1.1.1.3")}, {42, net.ParseIP("1.1.1.3")}},
			b: []A{{42, net.ParseIP("1.1.1.1")}, {42, net.ParseIP("1.1.1.2")},
				{42, net.ParseIP("1.1.1.3")}, {42, net.ParseIP("1.1.1.4")}},
		}, want: true},

		{name: "equal 5", args: args{
			a: []A{{42, net.ParseIP("1.1.1.1")}, {42, net.ParseIP("1.1.1.2")},
				{42, net.ParseIP("1.1.1.3")}, {42, net.ParseIP("1.1.1.3")}},
			b: []A{{42, net.ParseIP("1.1.1.3")}, {42, net.ParseIP("1.1.1.4")},
				{42, net.ParseIP("1.1.1.2")}, {42, net.ParseIP("1.1.1.1")}},
		}, want: true},

		{name: "not equal 1", args: args{
			a: []A{{42, net.ParseIP("1.1.1.1")}},
			b: []A{{42, net.ParseIP("1.1.1.2")}},
		}, want: false},

		{name: "not equal 2", args: args{
			a: []A{{42, net.ParseIP("1.1.1.1")}, {42, net.ParseIP("1.1.1.2")}},
			b: []A{{42, net.ParseIP("1.1.1.1")}, {42, net.ParseIP("1.1.1.3")}},
		}, want: false},

		{name: "not equal 3", args: args{
			a: []A{{42, net.ParseIP("1.1.1.1")}},
			b: nil,
		}, want: false},

		{name: "not equal 4", args: args{
			a: []A{{42, net.ParseIP("1.1.1.1")}},
			b: []A{{42, net.ParseIP("1.1.1.1")}, {42, net.ParseIP("1.1.1.3")}},
		}, want: false},

		{name: "not equal 5", args: args{
			a: nil,
			b: []A{{42, net.ParseIP("1.1.1.2")}},
		}, want: false},

		{name: "not equal 6", args: args{
			a: []A{{42, net.ParseIP("1.1.1.1")}, {42, net.ParseIP("1.1.1.2")}},
			b: []A{{42, net.ParseIP("1.1.1.1")}},
		}, want: false},

		{name: "not equal 7", args: args{
			a: []A{},
			b: []A{{42, net.ParseIP("1.1.1.2")}},
		}, want: false},

		{name: "not equal 8", args: args{
			a: []A{{42, net.ParseIP("1.1.1.1")}},
			b: []A{},
		}, want: false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := aEqual(tt.args.a, tt.args.b); got != tt.want {
				t.Errorf("aEqual() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_aaaaEqual(t *testing.T) {
	type args struct {
		a []AAAA
		b []AAAA
	}
	tests := []struct {
		name string
		args args
		want bool
	}{
		{name: "equal 1", args: args{
			a: []AAAA{{42, net.ParseIP("42::1337")}},
			b: []AAAA{{42, net.ParseIP("42::1337")}},
		}, want: true},

		{name: "equal 2", args: args{
			a: []AAAA{{42, net.ParseIP("42::1337")}, {42, net.ParseIP("f42::abcd")}},
			b: []AAAA{{42, net.ParseIP("42::1337")}, {42, net.ParseIP("f42::abcd")}},
		}, want: true},

		{name: "equal 4", args: args{
			a: []AAAA{{42, net.ParseIP("42::1337")}, {42, net.ParseIP("f42::abcd")},
				{42, net.ParseIP("::1")}, {42, net.ParseIP("fe80::1")}},
			b: []AAAA{{42, net.ParseIP("42::1337")}, {42, net.ParseIP("f42::abcd")},
				{42, net.ParseIP("::1")}, {42, net.ParseIP("fe80::1")}},
		}, want: true},

		{name: "equal 5", args: args{
			a: []AAAA{{42, net.ParseIP("42::1337")}, {42, net.ParseIP("f42::abcd")},
				{42, net.ParseIP("::1")}, {42, net.ParseIP("f42::abcf")}},
			b: []AAAA{{42, net.ParseIP("f42::abcf")}, {42, net.ParseIP("::1")},
				{42, net.ParseIP("f42::abcd")}, {42, net.ParseIP("42::1337")}},
		}, want: true},

		{name: "not equal 1", args: args{
			a: []AAAA{{42, net.ParseIP("42::1337")}},
			b: []AAAA{{42, net.ParseIP("f42::abcd")}},
		}, want: false},

		{name: "not equal 2", args: args{
			a: []AAAA{{42, net.ParseIP("42::1337")}, {42, net.ParseIP("f42::abcd")}},
			b: []AAAA{{42, net.ParseIP("42::1337")}, {42, net.ParseIP("f42::abcf")}},
		}, want: false},

		{name: "not equal 3", args: args{
			a: []AAAA{{42, net.ParseIP("42::1337")}},
			b: nil,
		}, want: false},

		{name: "not equal 4", args: args{
			a: []AAAA{{42, net.ParseIP("42::1337")}},
			b: []AAAA{{42, net.ParseIP("42::1337")}, {42, net.ParseIP("42:1338")}},
		}, want: false},

		{name: "not equal 5", args: args{
			a: nil,
			b: []AAAA{{42, net.ParseIP("f42::abcd")}},
		}, want: false},

		{name: "not equal 6", args: args{
			a: []AAAA{{42, net.ParseIP("42::1337")}, {42, net.ParseIP("f42::abcd")}},
			b: []AAAA{{42, net.ParseIP("42::1337")}},
		}, want: false},

		{name: "not equal 7", args: args{
			a: []AAAA{},
			b: []AAAA{{42, net.ParseIP("f42::abcd")}},
		}, want: false},

		{name: "not equal 8", args: args{
			a: []AAAA{{42, net.ParseIP("42::1337")}},
			b: []AAAA{},
		}, want: false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := aaaaEqual(tt.args.a, tt.args.b); got != tt.want {
				t.Errorf("aEqual() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_txtEqual(t *testing.T) {
	type args struct {
		a []TXT
		b []TXT
	}
	tests := []struct {
		name string
		args args
		want bool
	}{
		{name: "equal 1", args: args{
			a: []TXT{{42, "Hello World!"}},
			b: []TXT{{42, "Hello World!"}},
		}, want: true},
		{name: "not equal 1", args: args{
			a: []TXT{{42, "Hello World!"}},
			b: []TXT{{42, "Hello World."}},
		}, want: false},
		{name: "not equal 2", args: args{
			a: nil,
			b: []TXT{{42, "Hello World."}},
		}, want: false},
		{name: "not equal 3", args: args{
			a: []TXT{{42, "Hello World!"}},
			b: []TXT{{42, ""}},
		}, want: false},
		{name: "not equal 4", args: args{
			a: []TXT{{42, "Hello World!"}},
			b: []TXT{{43, "Hello World!"}},
		}, want: false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := txtEqual(tt.args.a, tt.args.b); got != tt.want {
				t.Errorf("txtEqual() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_cnameEqual(t *testing.T) {
	type args struct {
		a []CNAME
		b []CNAME
	}
	tests := []struct {
		name string
		args args
		want bool
	}{
		{name: "equal 1", args: args{
			a: []CNAME{{42, "example.net."}},
			b: []CNAME{{42, "example.net."}},
		}, want: true},
		{name: "not equal 1", args: args{
			a: []CNAME{{42, "example.net."}},
			b: []CNAME{{42, "example.com."}},
		}, want: false},
		{name: "not equal 2", args: args{
			a: []CNAME{{42, "example.com."}},
			b: []CNAME{{42, "example.net."}},
		}, want: false},
		{name: "not equal 3", args: args{
			a: nil,
			b: []CNAME{{42, "example.net."}},
		}, want: false},
		{name: "not equal 4", args: args{
			a: []CNAME{{42, "example.net."}},
			b: []CNAME{{42, ""}},
		}, want: false},
		{name: "not equal 5", args: args{
			a: []CNAME{{42, "example.net."}},
			b: []CNAME{{43, "example.net."}},
		}, want: false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := cnameEqual(tt.args.a, tt.args.b); got != tt.want {
				t.Errorf("cnameEqual() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_nsEqual(t *testing.T) {
	type args struct {
		a []NS
		b []NS
	}
	tests := []struct {
		name string
		args args
		want bool
	}{
		{name: "equal 1", args: args{
			a: []NS{{42, "example.net."}},
			b: []NS{{42, "example.net."}},
		}, want: true},
		{name: "not equal 1", args: args{
			a: []NS{{42, "example.net."}},
			b: []NS{{42, "example.com."}},
		}, want: false},
		{name: "not equal 2", args: args{
			a: []NS{{42, "example.com."}},
			b: []NS{{42, "example.net."}},
		}, want: false},
		{name: "not equal 3", args: args{
			a: nil,
			b: []NS{{42, "example.net."}},
		}, want: false},
		{name: "not equal 4", args: args{
			a: []NS{{42, "example.net."}},
			b: []NS{{42, ""}},
		}, want: false},
		{name: "not equal 5", args: args{
			a: []NS{{42, "example.net."}},
			b: []NS{{43, "example.net."}},
		}, want: false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := nsEqual(tt.args.a, tt.args.b); got != tt.want {
				t.Errorf("nsEqual() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_mxEqual(t *testing.T) {
	type args struct {
		a []MX
		b []MX
	}
	tests := []struct {
		name string
		args args
		want bool
	}{
		{name: "equal 1", args: args{
			a: []MX{{42, "example.net.", 10}},
			b: []MX{{42, "example.net.", 10}},
		}, want: true},
		{name: "not equal 1", args: args{
			a: []MX{{42, "example.net.", 10}},
			b: []MX{{42, "example.com.", 10}},
		}, want: false},
		{name: "not equal 2", args: args{
			a: []MX{{42, "example.com.", 10}},
			b: []MX{{42, "example.com.", 20}},
		}, want: false},
		{name: "not equal 3", args: args{
			a: nil,
			b: []MX{{42, "example.net.", 10}},
		}, want: false},
		{name: "not equal 4", args: args{
			a: []MX{{42, "example.net.", 10}},
			b: []MX{{42, "", 10}},
		}, want: false},
		{name: "not equal 5", args: args{
			a: []MX{{42, "example.net.", 10}},
			b: []MX{{43, "example.net.", 10}},
		}, want: false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := mxEqual(tt.args.a, tt.args.b); got != tt.want {
				t.Errorf("mxEqual() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_srvEqual(t *testing.T) {
	type args struct {
		a []SRV
		b []SRV
	}
	tests := []struct {
		name string
		args args
		want bool
	}{
		{name: "equal 1", args: args{
			a: []SRV{ {42, 10, 20, 80, "example.net"}},
			b: []SRV{ {42, 10, 20, 80, "example.net"}},
		}, want: true },
		{name: "not equal 1", args: args{
			a: []SRV{ {42, 10, 20, 80, "example.net"}},
			b: []SRV{ {43, 10, 20, 80, "example.net"}},
		}, want: false },
		{name: "not equal 2", args: args{
			a: []SRV{ {42, 11, 20, 80, "example.net"}},
			b: []SRV{ {42, 10, 20, 80, "example.net"}},
		}, want: false },
		{name: "not equal 3", args: args{
			a: []SRV{ {42, 10, 20, 80, "example.net"}},
			b: []SRV{ {42, 10, 21, 80, "example.net"}},
		}, want: false },
		{name: "not equal 4", args: args{
			a: []SRV{ {42, 10, 20, 81, "example.net"}},
			b: []SRV{ {42, 10, 20, 80, "example.net"}},
		}, want: false },
		{name: "not equal 5", args: args{
			a: []SRV{ {42, 10, 20, 80, "example.com"}},
			b: []SRV{ {42, 10, 20, 80, "example.net"}},
		}, want: false },
		{name: "not equal 6", args: args{
			a: []SRV{ {42, 10, 20, 80, "example.net"}},
			b: []SRV{ },
		}, want: false },
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := srvEqual(tt.args.a, tt.args.b); got != tt.want {
				t.Errorf("srvEqual() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_caaEqual(t *testing.T) {
	type args struct {
		a []CAA
		b []CAA
	}
	tests := []struct {
		name string
		args args
		want bool
	}{
		{name: "equal 1", args: args{
			a: []CAA{{42,0, "issue", "letsencrypt.org"}},
			b: []CAA{{42,0, "issue", "letsencrypt.org"}},
		}, want: true},
		{name: "not equal 2", args: args{
			a: []CAA{{42,0, "issue", "letsencrypt.org"}},
			b: []CAA{{42,0, "issue", "letsencrypt.net"}},
		}, want: false},
		{name: "not equal 3", args: args{
			a: []CAA{{42,0, "issue", "letsencrypt.org"}},
			b: []CAA{{42,128, "issue", "letsencrypt.org"}},
		}, want: false},
		{name: "not equal 4", args: args{
			a: []CAA{{42,0, "issue", "letsencrypt.org"}},
			b: []CAA{{43,0, "issue", "letsencrypt.org"}},
		}, want: false},
		{name: "not equal 5", args: args{
			a: []CAA{},
			b: []CAA{{42,0, "issue", "letsencrypt.net"}},
		}, want: false},
		{name: "not equal 6", args: args{
			a: []CAA{{42,0, "issue", "letsencrypt.org"}},
			b: nil,
		}, want: false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := caaEqual(tt.args.a, tt.args.b); got != tt.want {
				t.Errorf("caaEqual() = %v, want %v", got, tt.want)
			}
		})
	}
}
package network

import (
	"errors"
	"net"
	"reflect"
	"testing"

	"github.com/coreos/coreos-cloudinit/datasource/metadata/digitalocean"
)

func TestParseNameservers(t *testing.T) {
	for _, tt := range []struct {
		dns digitalocean.DNS
		nss []net.IP
		err error
	}{
		{
			dns: digitalocean.DNS{},
			nss: []net.IP{},
		},
		{
			dns: digitalocean.DNS{[]string{"1.2.3.4"}},
			nss: []net.IP{net.ParseIP("1.2.3.4")},
		},
		{
			dns: digitalocean.DNS{[]string{"bad"}},
			err: errors.New("could not parse \"bad\" as nameserver IP address"),
		},
	} {
		nss, err := parseNameservers(tt.dns)
		if !errorsEqual(tt.err, err) {
			t.Fatalf("bad error (%+v): want %q, got %q", tt.dns, tt.err, err)
		}
		if !reflect.DeepEqual(tt.nss, nss) {
			t.Fatalf("bad nameservers (%+v): want %#v, got %#v", tt.dns, tt.nss, nss)
		}
	}
}

func TestParseInterface(t *testing.T) {
	for _, tt := range []struct {
		cfg      digitalocean.Interface
		nss      []net.IP
		useRoute bool
		iface    *logicalInterface
		err      error
	}{
		{
			cfg: digitalocean.Interface{
				MAC: "bad",
			},
			err: errors.New("invalid MAC address: bad"),
		},
		{
			cfg: digitalocean.Interface{
				MAC: "01:23:45:67:89:AB",
			},
			nss: []net.IP{},
			iface: &logicalInterface{
				hwaddr: net.HardwareAddr([]byte{0x01, 0x23, 0x45, 0x67, 0x89, 0xab}),
				config: configMethodStatic{
					addresses:   []net.IPNet{},
					nameservers: []net.IP{},
					routes:      []route{},
				},
			},
		},
		{
			cfg: digitalocean.Interface{
				MAC: "01:23:45:67:89:AB",
			},
			useRoute: true,
			nss:      []net.IP{net.ParseIP("1.2.3.4")},
			iface: &logicalInterface{
				hwaddr: net.HardwareAddr([]byte{0x01, 0x23, 0x45, 0x67, 0x89, 0xab}),
				config: configMethodStatic{
					addresses:   []net.IPNet{},
					nameservers: []net.IP{net.ParseIP("1.2.3.4")},
					routes:      []route{},
				},
			},
		},
		{
			cfg: digitalocean.Interface{
				MAC: "01:23:45:67:89:AB",
				IPv4: &digitalocean.Address{
					IPAddress: "bad",
					Netmask:   "255.255.0.0",
				},
			},
			nss: []net.IP{},
			err: errors.New("could not parse \"bad\" as IPv4 address"),
		},
		{
			cfg: digitalocean.Interface{
				MAC: "01:23:45:67:89:AB",
				IPv4: &digitalocean.Address{
					IPAddress: "1.2.3.4",
					Netmask:   "bad",
				},
			},
			nss: []net.IP{},
			err: errors.New("could not parse \"bad\" as IPv4 mask"),
		},
		{
			cfg: digitalocean.Interface{
				MAC: "01:23:45:67:89:AB",
				IPv4: &digitalocean.Address{
					IPAddress: "1.2.3.4",
					Netmask:   "255.255.0.0",
					Gateway:   "ignoreme",
				},
			},
			nss: []net.IP{},
			iface: &logicalInterface{
				hwaddr: net.HardwareAddr([]byte{0x01, 0x23, 0x45, 0x67, 0x89, 0xab}),
				config: configMethodStatic{
					addresses:   []net.IPNet{net.IPNet{net.ParseIP("1.2.3.4"), net.IPMask(net.ParseIP("255.255.0.0"))}},
					nameservers: []net.IP{},
					routes:      []route{},
				},
			},
		},
		{
			cfg: digitalocean.Interface{
				MAC: "01:23:45:67:89:AB",
				IPv4: &digitalocean.Address{
					IPAddress: "1.2.3.4",
					Netmask:   "255.255.0.0",
					Gateway:   "bad",
				},
			},
			useRoute: true,
			nss:      []net.IP{},
			err:      errors.New("could not parse \"bad\" as IPv4 gateway"),
		},
		{
			cfg: digitalocean.Interface{
				MAC: "01:23:45:67:89:AB",
				IPv4: &digitalocean.Address{
					IPAddress: "1.2.3.4",
					Netmask:   "255.255.0.0",
					Gateway:   "5.6.7.8",
				},
			},
			useRoute: true,
			nss:      []net.IP{},
			iface: &logicalInterface{
				hwaddr: net.HardwareAddr([]byte{0x01, 0x23, 0x45, 0x67, 0x89, 0xab}),
				config: configMethodStatic{
					addresses:   []net.IPNet{net.IPNet{net.ParseIP("1.2.3.4"), net.IPMask(net.ParseIP("255.255.0.0"))}},
					nameservers: []net.IP{},
					routes:      []route{route{net.IPNet{net.IPv4zero, net.IPMask(net.IPv4zero)}, net.ParseIP("5.6.7.8")}},
				},
			},
		},
		{
			cfg: digitalocean.Interface{
				MAC: "01:23:45:67:89:AB",
				IPv6: &digitalocean.Address{
					IPAddress: "bad",
					Cidr:      16,
				},
			},
			nss: []net.IP{},
			err: errors.New("could not parse \"bad\" as IPv6 address"),
		},
		{
			cfg: digitalocean.Interface{
				MAC: "01:23:45:67:89:AB",
				IPv6: &digitalocean.Address{
					IPAddress: "fe00::",
					Cidr:      16,
					Gateway:   "ignoreme",
				},
			},
			nss: []net.IP{},
			iface: &logicalInterface{
				hwaddr: net.HardwareAddr([]byte{0x01, 0x23, 0x45, 0x67, 0x89, 0xab}),
				config: configMethodStatic{
					addresses:   []net.IPNet{net.IPNet{net.ParseIP("fe00::"), net.IPMask(net.ParseIP("ffff::"))}},
					nameservers: []net.IP{},
					routes:      []route{},
				},
			},
		},
		{
			cfg: digitalocean.Interface{
				MAC: "01:23:45:67:89:AB",
				IPv6: &digitalocean.Address{
					IPAddress: "fe00::",
					Cidr:      16,
					Gateway:   "bad",
				},
			},
			useRoute: true,
			nss:      []net.IP{},
			err:      errors.New("could not parse \"bad\" as IPv6 gateway"),
		},
		{
			cfg: digitalocean.Interface{
				MAC: "01:23:45:67:89:AB",
				IPv6: &digitalocean.Address{
					IPAddress: "fe00::",
					Cidr:      16,
					Gateway:   "fe00:1234::",
				},
			},
			useRoute: true,
			nss:      []net.IP{},
			iface: &logicalInterface{
				hwaddr: net.HardwareAddr([]byte{0x01, 0x23, 0x45, 0x67, 0x89, 0xab}),
				config: configMethodStatic{
					addresses:   []net.IPNet{net.IPNet{net.ParseIP("fe00::"), net.IPMask(net.ParseIP("ffff::"))}},
					nameservers: []net.IP{},
					routes:      []route{route{net.IPNet{net.IPv6zero, net.IPMask(net.IPv6zero)}, net.ParseIP("fe00:1234::")}},
				},
			},
		},
	} {
		iface, err := parseInterface(tt.cfg, tt.nss, tt.useRoute)
		if !errorsEqual(tt.err, err) {
			t.Fatalf("bad error (%+v): want %q, got %q", tt.cfg, tt.err, err)
		}
		if !reflect.DeepEqual(tt.iface, iface) {
			t.Fatalf("bad interface (%+v): want %#v, got %#v", tt.cfg, tt.iface, iface)
		}
	}
}

func TestParseInterfaces(t *testing.T) {
	for _, tt := range []struct {
		cfg    digitalocean.Interfaces
		nss    []net.IP
		ifaces []InterfaceGenerator
		err    error
	}{
		{
			ifaces: []InterfaceGenerator{},
		},
		{
			cfg: digitalocean.Interfaces{
				Public: []digitalocean.Interface{{MAC: "01:23:45:67:89:AB"}},
			},
			ifaces: []InterfaceGenerator{
				&physicalInterface{logicalInterface{
					hwaddr: net.HardwareAddr([]byte{0x01, 0x23, 0x45, 0x67, 0x89, 0xab}),
					config: configMethodStatic{
						addresses:   []net.IPNet{},
						nameservers: []net.IP{},
						routes:      []route{},
					},
				}},
			},
		},
		{
			cfg: digitalocean.Interfaces{
				Private: []digitalocean.Interface{{MAC: "01:23:45:67:89:AB"}},
			},
			ifaces: []InterfaceGenerator{
				&physicalInterface{logicalInterface{
					hwaddr: net.HardwareAddr([]byte{0x01, 0x23, 0x45, 0x67, 0x89, 0xab}),
					config: configMethodStatic{
						addresses:   []net.IPNet{},
						nameservers: []net.IP{},
						routes:      []route{},
					},
				}},
			},
		},
		{
			cfg: digitalocean.Interfaces{
				Public: []digitalocean.Interface{{MAC: "01:23:45:67:89:AB"}},
			},
			nss: []net.IP{net.ParseIP("1.2.3.4")},
			ifaces: []InterfaceGenerator{
				&physicalInterface{logicalInterface{
					hwaddr: net.HardwareAddr([]byte{0x01, 0x23, 0x45, 0x67, 0x89, 0xab}),
					config: configMethodStatic{
						addresses:   []net.IPNet{},
						nameservers: []net.IP{net.ParseIP("1.2.3.4")},
						routes:      []route{},
					},
				}},
			},
		},
		{
			cfg: digitalocean.Interfaces{
				Private: []digitalocean.Interface{{MAC: "01:23:45:67:89:AB"}},
			},
			nss: []net.IP{net.ParseIP("1.2.3.4")},
			ifaces: []InterfaceGenerator{
				&physicalInterface{logicalInterface{
					hwaddr: net.HardwareAddr([]byte{0x01, 0x23, 0x45, 0x67, 0x89, 0xab}),
					config: configMethodStatic{
						addresses:   []net.IPNet{},
						nameservers: []net.IP{},
						routes:      []route{},
					},
				}},
			},
		},
		{
			cfg: digitalocean.Interfaces{
				Public: []digitalocean.Interface{{MAC: "bad"}},
			},
			err: errors.New("invalid MAC address: bad"),
		},
		{
			cfg: digitalocean.Interfaces{
				Private: []digitalocean.Interface{{MAC: "bad"}},
			},
			err: errors.New("invalid MAC address: bad"),
		},
	} {
		ifaces, err := parseInterfaces(tt.cfg, tt.nss)
		if !errorsEqual(tt.err, err) {
			t.Fatalf("bad error (%+v): want %q, got %q", tt.cfg, tt.err, err)
		}
		if !reflect.DeepEqual(tt.ifaces, ifaces) {
			t.Fatalf("bad interfaces (%+v): want %#v, got %#v", tt.cfg, tt.ifaces, ifaces)
		}
	}
}

func TestProcessDigitalOceanNetconf(t *testing.T) {
	for _, tt := range []struct {
		cfg    string
		ifaces []InterfaceGenerator
		err    error
	}{
		{
			cfg: ``,
		},
		{
			cfg: `{"dns":{"nameservers":["bad"]}}`,
			err: errors.New("could not parse \"bad\" as nameserver IP address"),
		},
		{
			cfg: `{"interfaces":{"public":[{"ipv4":{"ip_address":"bad"}}]}}`,
			err: errors.New("could not parse \"bad\" as IPv4 address"),
		},
		{
			cfg:    `{}`,
			ifaces: []InterfaceGenerator{},
		},
	} {
		ifaces, err := ProcessDigitalOceanNetconf(tt.cfg)
		if !errorsEqual(tt.err, err) {
			t.Fatalf("bad error (%q): want %q, got %q", tt.cfg, tt.err, err)
		}
		if !reflect.DeepEqual(tt.ifaces, ifaces) {
			t.Fatalf("bad interfaces (%q): want %#v, got %#v", tt.cfg, tt.ifaces, ifaces)
		}
	}
}

func errorsEqual(a, b error) bool {
	if a == nil && b == nil {
		return true
	}
	if (a != nil && b == nil) || (a == nil && b != nil) {
		return false
	}
	return (a.Error() == b.Error())
}

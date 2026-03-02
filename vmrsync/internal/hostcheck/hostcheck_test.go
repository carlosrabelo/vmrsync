package hostcheck

import (
	"net"
	"strings"
	"testing"
)

func TestParseSSHTargetHost(t *testing.T) {
	tests := []struct {
		in      string
		want    string
		wantErr bool
	}{
		{"", "", true},
		{"user@", "", true},
		{"vm21", "vm21", false},
		{"user@vm21", "vm21", false},
		{"user@192.168.1.10", "192.168.1.10", false},
		{"[2001:db8::1]", "2001:db8::1", false},
		{"user@[2001:db8::1]", "2001:db8::1", false},
		{"192.168.1.1:2222", "192.168.1.1", false},
		{"[::1]:2222", "::1", false},
		{"[bad", "", true},
	}
	for _, tt := range tests {
		got, err := parseSSHTargetHost(tt.in)
		if tt.wantErr {
			if err == nil {
				t.Errorf("parseSSHTargetHost(%q) want error, got (%q, nil)", tt.in, got)
			}
			continue
		}
		if err != nil {
			t.Errorf("parseSSHTargetHost(%q) unexpected err: %v", tt.in, err)
			continue
		}
		if got != tt.want {
			t.Errorf("parseSSHTargetHost(%q) = %q, want %q", tt.in, got, tt.want)
		}
	}
}

func TestEnsureRemoteSSHHost_localhostNames(t *testing.T) {
	saveHostname, saveAddrs, saveLookup := GetHostname, ListLocalAddrs, LookupHostIPs
	t.Cleanup(func() {
		GetHostname, ListLocalAddrs, LookupHostIPs = saveHostname, saveAddrs, saveLookup
	})
	GetHostname = func() (string, error) { return "remotebox", nil }
	ListLocalAddrs = func() ([]net.Addr, error) {
		return []net.Addr{&net.IPNet{IP: net.ParseIP("10.0.0.2"), Mask: net.CIDRMask(24, 32)}}, nil
	}
	LookupHostIPs = func(host string) ([]net.IP, error) { return nil, nil }

	for _, h := range []string{"localhost", "LOCALHOST", "localhost.localdomain", "app.localhost"} {
		if err := EnsureRemoteSSHHost(h); err == nil {
			t.Errorf("EnsureRemoteSSHHost(%q) want error", h)
		}
	}
}

func TestEnsureRemoteSSHHost_loopbackIP(t *testing.T) {
	saveHostname, saveAddrs, saveLookup := GetHostname, ListLocalAddrs, LookupHostIPs
	t.Cleanup(func() {
		GetHostname, ListLocalAddrs, LookupHostIPs = saveHostname, saveAddrs, saveLookup
	})
	GetHostname = func() (string, error) { return "remotebox", nil }
	ListLocalAddrs = func() ([]net.Addr, error) { return nil, nil }

	for _, h := range []string{"127.0.0.1", "::1", "user@127.0.0.1", "[::1]"} {
		if err := EnsureRemoteSSHHost(h); err == nil {
			t.Errorf("EnsureRemoteSSHHost(%q) want error", h)
		}
	}
}

func TestEnsureRemoteSSHHost_matchesHostname(t *testing.T) {
	saveHostname, saveAddrs, saveLookup := GetHostname, ListLocalAddrs, LookupHostIPs
	t.Cleanup(func() {
		GetHostname, ListLocalAddrs, LookupHostIPs = saveHostname, saveAddrs, saveLookup
	})
	GetHostname = func() (string, error) { return "my-laptop", nil }
	ListLocalAddrs = func() ([]net.Addr, error) { return nil, nil }
	LookupHostIPs = func(host string) ([]net.IP, error) { return nil, nil }

	if err := EnsureRemoteSSHHost("my-laptop"); err == nil {
		t.Fatal("want error for hostname match")
	}
	if err := EnsureRemoteSSHHost("MY-LAPTOP"); err == nil {
		t.Fatal("want error for case-insensitive hostname match")
	}
}

func TestEnsureRemoteSSHHost_literalLocalInterfaceIP(t *testing.T) {
	saveHostname, saveAddrs, saveLookup := GetHostname, ListLocalAddrs, LookupHostIPs
	t.Cleanup(func() {
		GetHostname, ListLocalAddrs, LookupHostIPs = saveHostname, saveAddrs, saveLookup
	})
	GetHostname = func() (string, error) { return "other", nil }
	ListLocalAddrs = func() ([]net.Addr, error) {
		return []net.Addr{&net.IPNet{IP: net.ParseIP("192.168.5.10"), Mask: net.CIDRMask(24, 32)}}, nil
	}
	LookupHostIPs = func(host string) ([]net.IP, error) { return nil, nil }

	if err := EnsureRemoteSSHHost("192.168.5.10"); err == nil {
		t.Fatal("want error for local interface IP")
	}
}

func TestEnsureRemoteSSHHost_dnsResolvesToLoopback(t *testing.T) {
	saveHostname, saveAddrs, saveLookup := GetHostname, ListLocalAddrs, LookupHostIPs
	t.Cleanup(func() {
		GetHostname, ListLocalAddrs, LookupHostIPs = saveHostname, saveAddrs, saveLookup
	})
	GetHostname = func() (string, error) { return "workstation", nil }
	ListLocalAddrs = func() ([]net.Addr, error) { return nil, nil }
	LookupHostIPs = func(host string) ([]net.IP, error) {
		if host == "trap" {
			return []net.IP{net.ParseIP("127.0.0.1")}, nil
		}
		return nil, nil
	}

	if err := EnsureRemoteSSHHost("trap"); err == nil {
		t.Fatal("want error when DNS resolves to loopback")
	}
}

func TestEnsureRemoteSSHHost_dnsResolvesToLocalIface(t *testing.T) {
	saveHostname, saveAddrs, saveLookup := GetHostname, ListLocalAddrs, LookupHostIPs
	t.Cleanup(func() {
		GetHostname, ListLocalAddrs, LookupHostIPs = saveHostname, saveAddrs, saveLookup
	})
	GetHostname = func() (string, error) { return "workstation", nil }
	ListLocalAddrs = func() ([]net.Addr, error) {
		return []net.Addr{&net.IPNet{IP: net.ParseIP("10.20.30.40"), Mask: net.CIDRMask(8, 32)}}, nil
	}
	LookupHostIPs = func(host string) ([]net.IP, error) {
		return []net.IP{net.ParseIP("10.20.30.40")}, nil
	}

	if err := EnsureRemoteSSHHost("same"); err == nil {
		t.Fatal("want error when DNS resolves to local interface IP")
	}
}

func TestEnsureRemoteSSHHost_lookupFailsAllows(t *testing.T) {
	saveHostname, saveAddrs, saveLookup := GetHostname, ListLocalAddrs, LookupHostIPs
	t.Cleanup(func() {
		GetHostname, ListLocalAddrs, LookupHostIPs = saveHostname, saveAddrs, saveLookup
	})
	GetHostname = func() (string, error) { return "workstation", nil }
	ListLocalAddrs = func() ([]net.Addr, error) { return nil, nil }
	LookupHostIPs = func(host string) ([]net.IP, error) {
		return nil, &net.DNSError{Err: "no such host", Name: host, IsNotFound: true}
	}

	if err := EnsureRemoteSSHHost("ssh-config-alias"); err != nil {
		t.Fatalf("unresolvable name should pass literal checks: %v", err)
	}
}

func TestEnsureRemoteSSHHost_okRemote(t *testing.T) {
	saveHostname, saveAddrs, saveLookup := GetHostname, ListLocalAddrs, LookupHostIPs
	t.Cleanup(func() {
		GetHostname, ListLocalAddrs, LookupHostIPs = saveHostname, saveAddrs, saveLookup
	})
	GetHostname = func() (string, error) { return "workstation", nil }
	ListLocalAddrs = func() ([]net.Addr, error) {
		return []net.Addr{&net.IPNet{IP: net.ParseIP("10.0.0.1"), Mask: net.CIDRMask(24, 32)}}, nil
	}
	LookupHostIPs = func(host string) ([]net.IP, error) {
		return []net.IP{net.ParseIP("10.0.0.99")}, nil
	}

	if err := EnsureRemoteSSHHost("vm99"); err != nil {
		t.Fatal(err)
	}
}

func TestEnsureRemoteSSHHost_errorMessages(t *testing.T) {
	saveHostname, saveAddrs, saveLookup := GetHostname, ListLocalAddrs, LookupHostIPs
	t.Cleanup(func() {
		GetHostname, ListLocalAddrs, LookupHostIPs = saveHostname, saveAddrs, saveLookup
	})
	GetHostname = func() (string, error) { return "box", nil }
	ListLocalAddrs = func() ([]net.Addr, error) { return nil, nil }
	LookupHostIPs = func(host string) ([]net.IP, error) { return nil, nil }

	err := EnsureRemoteSSHHost("localhost")
	if err == nil || !strings.Contains(err.Error(), "security error") {
		t.Fatalf("want security_error prefix, got %v", err)
	}
}

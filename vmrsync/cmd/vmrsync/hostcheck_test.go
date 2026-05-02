package main

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
	saveHostname, saveAddrs, saveLookup := getHostname, listLocalAddrs, lookupHostIPsFn
	t.Cleanup(func() {
		getHostname, listLocalAddrs, lookupHostIPsFn = saveHostname, saveAddrs, saveLookup
	})
	getHostname = func() (string, error) { return "remotebox", nil }
	listLocalAddrs = func() ([]net.Addr, error) {
		return []net.Addr{&net.IPNet{IP: net.ParseIP("10.0.0.2"), Mask: net.CIDRMask(24, 32)}}, nil
	}
	lookupHostIPsFn = func(host string) ([]net.IP, error) { return nil, nil }

	for _, h := range []string{"localhost", "LOCALHOST", "localhost.localdomain", "app.localhost"} {
		if err := ensureRemoteSSHHost(h); err == nil {
			t.Errorf("ensureRemoteSSHHost(%q) want error", h)
		}
	}
}

func TestEnsureRemoteSSHHost_loopbackIP(t *testing.T) {
	saveHostname, saveAddrs, saveLookup := getHostname, listLocalAddrs, lookupHostIPsFn
	t.Cleanup(func() {
		getHostname, listLocalAddrs, lookupHostIPsFn = saveHostname, saveAddrs, saveLookup
	})
	getHostname = func() (string, error) { return "remotebox", nil }
	listLocalAddrs = func() ([]net.Addr, error) { return nil, nil }

	for _, h := range []string{"127.0.0.1", "::1", "user@127.0.0.1", "[::1]"} {
		if err := ensureRemoteSSHHost(h); err == nil {
			t.Errorf("ensureRemoteSSHHost(%q) want error", h)
		}
	}
}

func TestEnsureRemoteSSHHost_matchesHostname(t *testing.T) {
	saveHostname, saveAddrs, saveLookup := getHostname, listLocalAddrs, lookupHostIPsFn
	t.Cleanup(func() {
		getHostname, listLocalAddrs, lookupHostIPsFn = saveHostname, saveAddrs, saveLookup
	})
	getHostname = func() (string, error) { return "my-laptop", nil }
	listLocalAddrs = func() ([]net.Addr, error) { return nil, nil }
	lookupHostIPsFn = func(host string) ([]net.IP, error) { return nil, nil }

	if err := ensureRemoteSSHHost("my-laptop"); err == nil {
		t.Fatal("want error for hostname match")
	}
	if err := ensureRemoteSSHHost("MY-LAPTOP"); err == nil {
		t.Fatal("want error for case-insensitive hostname match")
	}
}

func TestEnsureRemoteSSHHost_literalLocalInterfaceIP(t *testing.T) {
	saveHostname, saveAddrs, saveLookup := getHostname, listLocalAddrs, lookupHostIPsFn
	t.Cleanup(func() {
		getHostname, listLocalAddrs, lookupHostIPsFn = saveHostname, saveAddrs, saveLookup
	})
	getHostname = func() (string, error) { return "other", nil }
	listLocalAddrs = func() ([]net.Addr, error) {
		return []net.Addr{&net.IPNet{IP: net.ParseIP("192.168.5.10"), Mask: net.CIDRMask(24, 32)}}, nil
	}
	lookupHostIPsFn = func(host string) ([]net.IP, error) { return nil, nil }

	if err := ensureRemoteSSHHost("192.168.5.10"); err == nil {
		t.Fatal("want error for local interface IP")
	}
}

func TestEnsureRemoteSSHHost_dnsResolvesToLoopback(t *testing.T) {
	saveHostname, saveAddrs, saveLookup := getHostname, listLocalAddrs, lookupHostIPsFn
	t.Cleanup(func() {
		getHostname, listLocalAddrs, lookupHostIPsFn = saveHostname, saveAddrs, saveLookup
	})
	getHostname = func() (string, error) { return "workstation", nil }
	listLocalAddrs = func() ([]net.Addr, error) { return nil, nil }
	lookupHostIPsFn = func(host string) ([]net.IP, error) {
		if host == "trap" {
			return []net.IP{net.ParseIP("127.0.0.1")}, nil
		}
		return nil, nil
	}

	if err := ensureRemoteSSHHost("trap"); err == nil {
		t.Fatal("want error when DNS resolves to loopback")
	}
}

func TestEnsureRemoteSSHHost_dnsResolvesToLocalIface(t *testing.T) {
	saveHostname, saveAddrs, saveLookup := getHostname, listLocalAddrs, lookupHostIPsFn
	t.Cleanup(func() {
		getHostname, listLocalAddrs, lookupHostIPsFn = saveHostname, saveAddrs, saveLookup
	})
	getHostname = func() (string, error) { return "workstation", nil }
	listLocalAddrs = func() ([]net.Addr, error) {
		return []net.Addr{&net.IPNet{IP: net.ParseIP("10.20.30.40"), Mask: net.CIDRMask(8, 32)}}, nil
	}
	lookupHostIPsFn = func(host string) ([]net.IP, error) {
		return []net.IP{net.ParseIP("10.20.30.40")}, nil
	}

	if err := ensureRemoteSSHHost("same"); err == nil {
		t.Fatal("want error when DNS resolves to local interface IP")
	}
}

func TestEnsureRemoteSSHHost_lookupFailsAllows(t *testing.T) {
	saveHostname, saveAddrs, saveLookup := getHostname, listLocalAddrs, lookupHostIPsFn
	t.Cleanup(func() {
		getHostname, listLocalAddrs, lookupHostIPsFn = saveHostname, saveAddrs, saveLookup
	})
	getHostname = func() (string, error) { return "workstation", nil }
	listLocalAddrs = func() ([]net.Addr, error) { return nil, nil }
	lookupHostIPsFn = func(host string) ([]net.IP, error) {
		return nil, &net.DNSError{Err: "no such host", Name: host, IsNotFound: true}
	}

	if err := ensureRemoteSSHHost("ssh-config-alias"); err != nil {
		t.Fatalf("unresolvable name should pass literal checks: %v", err)
	}
}

func TestEnsureRemoteSSHHost_okRemote(t *testing.T) {
	saveHostname, saveAddrs, saveLookup := getHostname, listLocalAddrs, lookupHostIPsFn
	t.Cleanup(func() {
		getHostname, listLocalAddrs, lookupHostIPsFn = saveHostname, saveAddrs, saveLookup
	})
	getHostname = func() (string, error) { return "workstation", nil }
	listLocalAddrs = func() ([]net.Addr, error) {
		return []net.Addr{&net.IPNet{IP: net.ParseIP("10.0.0.1"), Mask: net.CIDRMask(24, 32)}}, nil
	}
	lookupHostIPsFn = func(host string) ([]net.IP, error) {
		return []net.IP{net.ParseIP("10.0.0.99")}, nil
	}

	if err := ensureRemoteSSHHost("vm99"); err != nil {
		t.Fatal(err)
	}
}

func TestEnsureRemoteSSHHost_errorMessages(t *testing.T) {
	saveHostname, saveAddrs, saveLookup := getHostname, listLocalAddrs, lookupHostIPsFn
	t.Cleanup(func() {
		getHostname, listLocalAddrs, lookupHostIPsFn = saveHostname, saveAddrs, saveLookup
	})
	getHostname = func() (string, error) { return "box", nil }
	listLocalAddrs = func() ([]net.Addr, error) { return nil, nil }
	lookupHostIPsFn = func(host string) ([]net.IP, error) { return nil, nil }

	err := ensureRemoteSSHHost("localhost")
	if err == nil || !strings.Contains(err.Error(), "security error") {
		t.Fatalf("want security_error prefix, got %v", err)
	}
}

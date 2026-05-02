package main

import (
	"fmt"
	"net"
	"os"
	"strings"
)

// Overridable for tests.
var (
	getHostname     = os.Hostname
	listLocalAddrs  = net.InterfaceAddrs
	lookupHostIPsFn = net.LookupIP
)

func parseSSHTargetHost(machine string) (host string, err error) {
	h := strings.TrimSpace(machine)
	if h == "" {
		return "", fmt.Errorf("empty host")
	}
	if i := strings.LastIndex(h, "@"); i >= 0 {
		h = h[i+1:]
	}
	if h == "" {
		return "", fmt.Errorf("empty host after @")
	}
	if strings.HasPrefix(h, "[") {
		closing := strings.Index(h, "]")
		if closing <= 0 {
			return "", fmt.Errorf("invalid bracketed host in %q", machine)
		}
		return h[1:closing], nil
	}
	// Possible IPv4:port (SSH uses -p for port; reject ambiguous : in IPv6 here).
	if strings.Count(h, ":") == 1 {
		colon := strings.LastIndex(h, ":")
		tail := h[colon+1:]
		if isNumericPort(tail) {
			if ip := net.ParseIP(h[:colon]); ip != nil {
				return h[:colon], nil
			}
		}
	}
	return h, nil
}

func isNumericPort(s string) bool {
	if s == "" {
		return false
	}
	for i := 0; i < len(s); i++ {
		if s[i] < '0' || s[i] > '9' {
			return false
		}
	}
	return true
}

func isLocalHostLabel(host string) bool {
	h := strings.ToLower(strings.TrimSuffix(host, "."))
	if h == "localhost" || h == "localhost.localdomain" {
		return true
	}
	if strings.HasSuffix(h, ".localhost") {
		return true
	}
	return false
}

func parseIPMaybeZone(host string) net.IP {
	if i := strings.IndexByte(host, '%'); i >= 0 {
		host = host[:i]
	}
	return net.ParseIP(host)
}

func collectLocalInterfaceIPs() (map[string]struct{}, error) {
	addrs, err := listLocalAddrs()
	if err != nil {
		return nil, err
	}
	set := make(map[string]struct{})
	for _, a := range addrs {
		var ip net.IP
		switch v := a.(type) {
		case *net.IPNet:
			ip = v.IP
		case *net.IPAddr:
			ip = v.IP
		}
		if ip == nil || ip.IsLoopback() {
			continue
		}
		set[ip.String()] = struct{}{}
	}
	return set, nil
}

func ipIsForbidden(ip net.IP, localIPs map[string]struct{}) bool {
	if ip.IsLoopback() {
		return true
	}
	_, isLocal := localIPs[ip.String()]
	return isLocal
}

// ensureRemoteSSHHost rejects machine targets that refer to this host (loopback,
// local hostname, or addresses assigned to local non-loopback interfaces).
func ensureRemoteSSHHost(machine string) error {
	host, err := parseSSHTargetHost(machine)
	if err != nil {
		return fmt.Errorf("security error: invalid SSH target %q: %w", machine, err)
	}

	if isLocalHostLabel(host) {
		return fmt.Errorf("security error: refusing remote host %q: localhost names are not allowed", machine)
	}

	localName, err := getHostname()
	if err != nil {
		return fmt.Errorf("security error: could not read local hostname: %w", err)
	}
	if strings.EqualFold(strings.TrimSuffix(host, "."), strings.TrimSuffix(localName, ".")) {
		return fmt.Errorf("security error: refusing remote host %q: matches this machine hostname %q", machine, localName)
	}

	if ip := parseIPMaybeZone(host); ip != nil {
		localIPs, err := collectLocalInterfaceIPs()
		if err != nil {
			return fmt.Errorf("security error: could not list local addresses: %w", err)
		}
		if ipIsForbidden(ip, localIPs) {
			return fmt.Errorf("security error: refusing remote host %q: target is loopback or a local interface address", machine)
		}
		return nil
	}

	localIPs, err := collectLocalInterfaceIPs()
	if err != nil {
		return fmt.Errorf("security error: could not list local addresses: %w", err)
	}

	ips, err := lookupHostIPsFn(host)
	if err != nil {
		// Unresolved names may be SSH config aliases; literal checks already passed.
		return nil
	}
	for _, ip := range ips {
		if ipIsForbidden(ip, localIPs) {
			return fmt.Errorf("security error: refusing remote host %q: name resolves to loopback or a local interface address", machine)
		}
	}
	return nil
}

package p_seer_websites

import (
	"context"
	"fmt"
	"html"
	"net"
	"net/url"
	"strings"
)

func isPublicIP(ip net.IP) bool {
	if ip == nil {
		return false
	}
	if ip.IsLoopback() || ip.IsPrivate() || ip.IsMulticast() || ip.IsUnspecified() || ip.IsLinkLocalUnicast() {
		return false
	}
	if ip4 := ip.To4(); ip4 != nil {
		if ip4[0] == 169 && ip4[1] == 254 {
			return false
		}
	}
	return true
}

// urlFailsSSRF reports whether parsed must not be fetched (blocks non-public targets).
func urlFailsSSRF(ctx context.Context, parsed *url.URL) bool {
	host := strings.TrimSpace(strings.ToLower(parsed.Hostname()))
	if host == "" || host == "localhost" {
		return true
	}
	if ip := net.ParseIP(host); ip != nil {
		return !isPublicIP(ip)
	}
	ips, err := net.DefaultResolver.LookupIPAddr(ctx, host)
	if err != nil || len(ips) == 0 {
		return true
	}
	for _, ia := range ips {
		if !isPublicIP(ia.IP) {
			return true
		}
	}
	return false
}

// normalizeWebsiteURL parses raw input and returns a canonical http(s) [*url.URL].
func normalizeWebsiteURL(raw string) (*url.URL, error) {
	raw = strings.TrimSpace(html.UnescapeString(raw))
	if raw == "" {
		return nil, fmt.Errorf("url is required")
	}
	parsed, err := url.Parse(raw)
	if err != nil {
		return nil, fmt.Errorf("invalid url: %w", err)
	}
	if parsed.Host == "" || parsed.Scheme == "" {
		return nil, fmt.Errorf("url must be absolute http(s)")
	}
	out := *parsed
	out.User = nil
	out.Scheme = strings.ToLower(out.Scheme)
	switch out.Scheme {
	case "http", "https":
	default:
		return nil, fmt.Errorf("url must use http or https")
	}
	host := strings.ToLower(out.Hostname())
	if host == "" {
		return nil, fmt.Errorf("url host is required")
	}
	port := out.Port()
	switch {
	case port == "":
		out.Host = host
	case out.Scheme == "http" && port == "80":
		out.Host = host
	case out.Scheme == "https" && port == "443":
		out.Host = host
	default:
		out.Host = net.JoinHostPort(host, port)
	}
	out.Fragment = ""
	if out.Path == "" {
		out.Path = "/"
	}
	query := out.Query()
	out.RawQuery = query.Encode()
	return new(out), nil
}

// fetchableWebsiteURL validates scheme/host and SSRF policy, returning canonical [*url.URL].
func fetchableWebsiteURL(ctx context.Context, raw string) (*url.URL, error) {
	u, err := normalizeWebsiteURL(raw)
	if err != nil {
		return nil, err
	}
	if urlFailsSSRF(ctx, u) {
		return nil, fmt.Errorf("url blocked by ssrf guard")
	}
	return u, nil
}

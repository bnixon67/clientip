package main

import (
	"fmt"
	"net"
	"net/http"
	"strings"

	"github.com/bnixon67/webapp/webhandler"
	"github.com/bnixon67/webapp/webutil"
)

// parseForwardedHeader extracts the "for=" IP from the Forwarded header.
func parseForwardedHeader(headerVal string) string {
	// Split multiple "Forwarded" headers
	entries := strings.Split(headerVal, ",")
	for _, entry := range entries {
		for _, part := range strings.Split(entry, ";") {
			part = strings.TrimSpace(part)
			if strings.HasPrefix(strings.ToLower(part), "for=") {
				ip := strings.TrimPrefix(part, "for=")
				ip = strings.Trim(ip, "\"") // some proxies wrap the IP in quotes
				if parsed := net.ParseIP(ip); parsed != nil {
					return ip
				}
			}
		}
	}
	return ""
}

// GetClientIP returns the real client IP from headers or, if none, RemoteAddr.
//
// WARNING: Ensure your reverse proxy strips/sets these headers to prevent
// spoofing via direct client requests.
func GetClientIP(r *http.Request) string {
	// Ordered by trustworthiness
	headerOrder := []string{
		"Forwarded",
		"X-Forwarded-For",
		"X-Real-IP",
		"Cf-Connecting-Ip",
		"True-Client-Ip",
		"X-Client-Ip",
		"X-ProxyUser-IP",
	}

	for _, header := range headerOrder {
		hv := r.Header.Get(header)
		if hv == "" {
			continue
		}

		if header == "Forwarded" {
			if ip := parseForwardedHeader(hv); ip != "" {
				return ip
			}
		}

		// Handle comma-separated lists of IPs
		for _, ip := range strings.Split(hv, ",") {
			ip = strings.TrimSpace(ip)
			if parsed := net.ParseIP(ip); parsed != nil {
				return ip
			}
		}
	}

	// Fallback to RemoteAddr
	if host, _, err := net.SplitHostPort(r.RemoteAddr); err == nil {
		if parsed := net.ParseIP(host); parsed != nil {
			return host
		}
	}

	return ""
}

func ClientIPGetHandler(w http.ResponseWriter, r *http.Request) {
	logger := webhandler.RequestLoggerWithFuncName(r)

	if !webutil.IsMethodOrError(w, r, http.MethodGet) {
		logger.Error("invalid method")
		return
	}

	webutil.SetContentTypeText(w)
	webutil.SetNoCacheHeaders(w)

	fmt.Fprintln(w, GetClientIP(r))
}

package main

import (
	"encoding/json"
	"net"
	"net/http"
	"strings"
)

func writeJSON(w http.ResponseWriter, s int, v any) error {
	w.WriteHeader(s)
	return json.NewEncoder(w).Encode(v)
}

func getIpAddress(r *http.Request) string {
	xff := r.Header.Get("X-Forwarded-For")
	if xff != "" {
		ips := strings.Split(xff, ",")
		if len(ips) > 0 && ips[0] != "" {
			return ips[0]
		}
	}
	xRealIP := r.Header.Get("X-Real-IP")
	if xRealIP != "" {
		return xRealIP
	}
	remoteAddr := r.RemoteAddr
	// RemoteAddr has the format "IP:port". We only want the IP part.
	ip, _, err := net.SplitHostPort(remoteAddr)
	if err != nil {
		// If splitting the RemoteAddr fails, return the full RemoteAddr string.
		// This could happen if the request comes from a local source (e.g., localhost without a port).
		return remoteAddr
	}
	return ip
}

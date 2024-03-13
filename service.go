package main

import (
	"context"
	"net"
	"net/http"
	"strings"
	"time"

	"github.com/redis/go-redis/v9"
)

type Service interface {
	Serve(w http.ResponseWriter, r *http.Request)
}
type RateLimitedService struct {
	rdsClient   *redis.Client
	ctx         context.Context
	duration    int64
	maxRequests int64
}

func NewRateLimitedService(rdsClient *redis.Client, ctx context.Context, duration int64, maxRequests int64) *RateLimitedService {
	return &RateLimitedService{
		rdsClient:   rdsClient,
		ctx:         ctx,
		duration:    duration,
		maxRequests: maxRequests,
	}
}

func (rls *RateLimitedService) Serve(w http.ResponseWriter, r *http.Request) {
	println("Received a request")
	ip := getIpAddress(r)
	resetKey := "reset_time:" + ip
	tokenKey := "tokens:" + ip

	// Check if the reset key exists to determine if it's a new period or existing user
	exists, err := rls.rdsClient.Exists(rls.ctx, resetKey).Result()
	if err != nil {
		http.Error(w, "Internal error", http.StatusInternalServerError)
		return
	}

	now := time.Now().Unix()
	if exists == 0 {
		// First request from this IP or new period
		rls.rdsClient.Set(rls.ctx, resetKey, now, time.Duration(rls.duration)*time.Second)
		rls.rdsClient.Set(rls.ctx, tokenKey, rls.maxRequests-1, time.Duration(rls.duration)*time.Second) // Decrement on set for first request
	} else {

		// Existing user: fetch prev req time and counter
		prevRequestTime, err := rls.rdsClient.Get(rls.ctx, resetKey).Int64()
		if err != nil {
			http.Error(w, "Internal error", http.StatusInternalServerError)
			return
		}
		prevTokens, err := rls.rdsClient.Get(rls.ctx, tokenKey).Int64()
		if err != nil {
			http.Error(w, "Internal error", http.StatusInternalServerError)
			return
		}
		//Compute token changes
		time_passed := now - prevRequestTime
		tokens_gained := time_passed * rls.maxRequests / rls.duration
		newTokens := prevTokens + tokens_gained
		if newTokens > rls.maxRequests {
			newTokens = rls.maxRequests
		}
		if newTokens <= 0 {
			// Rate limit exceeded
			http.Error(w, "Rate limit exceeded", http.StatusTooManyRequests)
			return
		}
		rls.rdsClient.Set(rls.ctx, resetKey, now, time.Duration(rls.duration)*time.Second)
		rls.rdsClient.Set(rls.ctx, tokenKey, newTokens-1, time.Duration(rls.duration)*time.Second)
	}
	query := r.URL.Query()
	startPath := query.Get("start_path")
	endPath := query.Get("end_path")
	if startPath == "" || endPath == "" {
		http.Error(w, "Missing start_path or end_path parameter", http.StatusBadRequest)
		return
	}
	//forward call to gRPC service
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

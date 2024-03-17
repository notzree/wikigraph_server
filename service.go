package main

import (
	"context"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/redis/go-redis/v9"
)

type Servicer interface {
	Serve(w http.ResponseWriter, r *http.Request)
}
type RateLimitedService struct {
	rdsClient *redis.Client
	ctx       context.Context
	svc       Servicer
	duration  int64
	req_limit int64
}

func NewRateLimitedService(rdsClient *redis.Client, ctx context.Context, svc Servicer, duration int64, req_limit int64) *RateLimitedService {
	return &RateLimitedService{
		rdsClient: rdsClient,
		ctx:       ctx,
		svc:       svc,
		duration:  duration,
		req_limit: req_limit,
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
		log.Println("Error:", err)
		http.Error(w, "Internal error", http.StatusInternalServerError)
		return
	}

	now := time.Now().Unix()
	if exists == 0 {
		// First request from this IP or new period
		rls.rdsClient.Set(rls.ctx, resetKey, now, time.Duration(rls.duration)*time.Hour)
		rls.rdsClient.Set(rls.ctx, tokenKey, rls.req_limit-1, time.Duration(rls.duration)*time.Hour) // Decrement on set for first request
	} else {

		// Existing user: fetch prev req time and counter
		prevRequestTime, err := rls.rdsClient.Get(rls.ctx, resetKey).Int64()
		if err != nil {
			log.Println("Error:", err)
			http.Error(w, "Internal error", http.StatusInternalServerError)
			return
		}
		prevTokens, err := rls.rdsClient.Get(rls.ctx, tokenKey).Int64()
		if err != nil {
			log.Println("Error:", err)
			http.Error(w, "Internal error", http.StatusInternalServerError)
			return
		}
		//Compute token changes
		time_passed := now - prevRequestTime
		tokens_gained := time_passed * rls.req_limit / rls.duration
		newTokens := prevTokens + tokens_gained
		if newTokens > rls.req_limit {
			newTokens = rls.req_limit
		}
		if newTokens <= 0 {
			// Rate limit exceeded
			http.Error(w, "Rate limit exceeded", http.StatusTooManyRequests)
			return
		}
		rls.rdsClient.Set(rls.ctx, resetKey, now, time.Duration(rls.duration)*time.Second)
		rls.rdsClient.Set(rls.ctx, tokenKey, newTokens-1, time.Duration(rls.duration)*time.Second)
	}
	rls.svc.Serve(w, r)
}

// Wrapper for pathfinder client
type PathFinderService struct {
	pf  PathFinder
	ctx context.Context
}

func NewPathFinderService(pf PathFinder, ctx context.Context) *PathFinderService {
	return &PathFinderService{pf: pf, ctx: ctx}
}

func (pfs *PathFinderService) Serve(w http.ResponseWriter, r *http.Request) {
	query := r.URL.Query()
	startPath := strings.TrimSpace(query.Get("start_path"))
	endPath := strings.TrimSpace(query.Get("end_path"))
	if startPath == "" || endPath == "" {
		http.Error(w, "Missing start_path or end_path parameter", http.StatusBadRequest)
		return
	}
	//forward call to gRPC server

	res, err := pfs.pf.FindPath(pfs.ctx, startPath, endPath)
	if err != nil {
		log.Println("Error:", err)
		http.Error(w, "Internal error", http.StatusInternalServerError)
		return
	}
	err = writeJSON(w, http.StatusOK, res)
	if err != nil {
		log.Println("Error:", err)
		http.Error(w, "Internal error", http.StatusInternalServerError)
		return
	}

}

// wrapper for autocomplete client
type AutoCompleteService struct {
	ac  AutoCompleter
	ctx context.Context
}

func NewAutoCompleterService(ac AutoCompleter, ctx context.Context) *AutoCompleteService {
	return &AutoCompleteService{ac: ac, ctx: ctx}
}
func (acs *AutoCompleteService) Serve(w http.ResponseWriter, r *http.Request) {
	query := r.URL.Query()
	prefix := strings.TrimSpace(query.Get("prefix"))
	if prefix == "" {
		http.Error(w, "Missing prefix parameter", http.StatusBadRequest)
		return
	}
	//forward call to gRPC server
	res, err := acs.ac.Complete(acs.ctx, prefix)
	if err != nil {
		log.Println("Error:", err)
		http.Error(w, "Internal error", http.StatusInternalServerError)
		return
	}
	err = writeJSON(w, http.StatusOK, res)
	if err != nil {
		log.Println("Error:", err)
		http.Error(w, "Internal error", http.StatusInternalServerError)
		return
	}

}

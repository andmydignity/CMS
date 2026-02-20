package server

import (
	"fmt"
	"net"
	"net/http"
	"sync"
	"time"

	"golang.org/x/time/rate"
)

func (cms *CmsStruct) uncaughtErrorMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if err := recover(); err != nil {
				w.Header().Set("Connection", "close")
				cms.Logger.Error("CRITICAL: Uncaught error.", "error", fmt.Sprintf("%s", err))
				cms.internalError(w, fmt.Errorf("%s", err))
			}
		}()
		next.ServeHTTP(w, r)
	})
}

func (cms *CmsStruct) rateLimitMiddleware(next http.Handler) http.Handler {
	type client struct {
		limitter  *rate.Limiter
		lastSince time.Time
	}
	clients := map[string]*client{}
	mutexClients := sync.Mutex{}
	go func() {
		time.Sleep(time.Minute)
		mutexClients.Lock()
		for ip, client := range clients {
			if time.Since(client.lastSince) >= time.Duration(float64(time.Second)*float64(cms.Config.RateLimit.Burst)/cms.Config.RateLimit.Rps) {
				delete(clients, ip)
			}
		}
		mutexClients.Unlock()
	}()
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ip, _, err := net.SplitHostPort(r.RemoteAddr)
		if err != nil {
			cms.internalError(w, err)
			return
		}
		mutexClients.Lock()
		if _, found := clients[ip]; !found {
			clients[ip] = &client{rate.NewLimiter(rate.Limit(cms.Config.RateLimit.Rps), cms.Config.RateLimit.Burst), time.Now()}
		} else {
			temp := clients[ip]
			temp.lastSince = time.Now()
		}
		if !clients[ip].limitter.Allow() {
			mutexClients.Unlock()
			cms.tooManyRequests(w)
			return
		}
		mutexClients.Unlock()
		next.ServeHTTP(w, r)
	})
}

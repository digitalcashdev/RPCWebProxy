package rpcproxy

import (
	"net/http"
	"sync"
	"time"
)

const (
	requestLimit = 100
	timeWindow   = 5 * time.Minute
	status420    = 420
)

type requestCounter struct {
	times []time.Time
}

var (
	ipRequests = make(map[string]*requestCounter)
	mu         sync.RWMutex
)

func init() {
	go func() {
		for {
			time.Sleep(3 * timeWindow)
			cleanupAll()
		}
	}()
}

func cleanupOne(now time.Time, counter *requestCounter) {
	validIndex := 0
	for i, t := range counter.times {
		isFresh := now.Sub(t) <= timeWindow
		if isFresh {
			validIndex = i
			break
		}
	}

	counter.times = counter.times[validIndex:]
}

func cleanupAll() {
	now := time.Now()

	mu.Lock()
	for _, counter := range ipRequests {
		cleanupOne(now, counter)
	}
	mu.Unlock()
}

// RateLimitMiddleware checks the IP address of incoming requests
// and limits them to 100 requests per 5 minutes.
func RateLimitMiddleware(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ip := r.RemoteAddr

		mu.Lock()
		counter, exists := ipRequests[ip]
		if !exists {
			counter = &requestCounter{}
			ipRequests[ip] = counter
		}

		now := time.Now()
		counter.times = append(counter.times, now)
		count := len(counter.times)
		last := counter.times[0]
		mu.Unlock()

		isCalm := count < requestLimit
		if isCalm {
			next(w, r)
			return
		}

		isHyper := now.Sub(last) <= timeWindow
		if isHyper {
			http.Error(w, "Enhance Your Calm", status420)
			return
		}

		{
			mu.Lock()
			cleanupOne(now, counter)
			count := len(counter.times)
			mu.Unlock()

			isCalm := count < requestLimit
			if !isCalm {
				http.Error(w, "Enhance Your Calm", status420)
				return
			}
		}

		next(w, r)
	}
}

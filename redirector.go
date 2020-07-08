package main

import (
	"context"
	"fmt"
	"net/http"
	"os"

	"github.com/go-redis/redis/v8"
	_ "github.com/lib/pq"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

var ctx = context.Background()

var (
	redirectRequests = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "redirect_requests",
		Help: "Redirect request counter by domain.",
	}, []string{"domain"})
	redirectRequestsForUnknownDomains = promauto.NewCounter(prometheus.CounterOpts{
		Name: "redirect_requests_unknown_domain",
		Help: "The number of redirect requests for unknown domains",
	})
)

type redirectHandler struct {
	redis *redis.Client
}

func ipAddr(r *http.Request) string {
	forwardedFor := r.Header.Get("X-Forwarded-For")
	if forwardedFor != "" {
		return forwardedFor
	}
	return r.RemoteAddr
}

func (h *redirectHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	fmt.Print(r.Host + " ")
	fmt.Print(r.URL.Path + " ")
	fmt.Print(r.URL.RawQuery)
	site := r.Host
	path := r.URL.Path
	if r.URL.RawQuery != "" {
		path += "?" + r.URL.RawQuery
	}
	redirectDomain, err := h.redis.Get(ctx, "{"+site+"}:redirect").Result()
	if err != nil {
		fmt.Println("Not registered: " + site)
		redirectRequestsForUnknownDomains.Inc()
		w.WriteHeader(http.StatusNotFound)
		return
	}
	redirectRequests.WithLabelValues(site).Inc()

	destination := "https://" + redirectDomain + path
	fmt.Println(destination)
	http.Redirect(w, r, destination, http.StatusTemporaryRedirect)
}

func newRedirectHandler(redis *redis.Client) *redirectHandler {
	a := new(redirectHandler)
	a.redis = redis
	return a
}

func main() {
	fmt.Println("redirector starting on port " + os.Getenv("PORT"))

	client := redis.NewClient(&redis.Options{
		Addr:     os.Getenv("REDIS_HOST") + ":" + os.Getenv("REDIS_PORT"),
		Password: os.Getenv("REDIS_KEY"), // no password set
		DB:       0,                      // use default DB
	})

	_, err := client.Ping(ctx).Result()
	if err != nil {
		panic(err)
	}
	fmt.Println("Successfully connected to redis.")

	redirectHandler := newRedirectHandler(client)

	http.Handle("/", redirectHandler)
	http.ListenAndServe(":"+os.Getenv("PORT"), nil)
}

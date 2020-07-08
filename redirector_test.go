package main

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/alicebob/miniredis"
	"github.com/go-redis/redis/v8"
)

func initFixtures() (*miniredis.Miniredis, *redis.Client) {
	s, err := miniredis.Run()
	if err != nil {
		panic(err)
	}

	s.FlushAll()
	client := redis.NewClient(&redis.Options{
		Addr:     s.Addr(),
		Password: "", // no password set
		DB:       0,  // use default DB
	})
	return s, client
}

func TestRootHandler(t *testing.T) {
	var ctx = context.Background()

	// Create a request to pass to our handler. We don't have any query parameters for now, so we'll
	// pass 'nil' as the third parameter.
	req, err := http.NewRequest("GET", "/", nil)
	if err != nil {
		t.Fatal(err)
	}
	req.Header.Set("Referer", "https://blog.singleton.io/foo/bar/page.html")
	req.Header.Set("X-Forwarded-For", "157.131.203.2")
	req.Header.Set("User-Agent", "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_14_2) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/76.0.3809.100 Safari/537.36")
	req.Host = "blog.davidsingleton.org"

	s, client := initFixtures()
	defer s.Close()

	client.Set(ctx, "{blog.davidsingleton.org}:redirect", "blog.singleton.io", 0)

	// We create a ResponseRecorder (which satisfies http.ResponseWriter) to record the response.
	rr := httptest.NewRecorder()
	handler := newRedirectHandler(client)

	// Our handlers satisfy http.Handler, so we can call their ServeHTTP method
	// directly and pass in our Request and ResponseRecorder.
	handler.ServeHTTP(rr, req)

	// Check the status code is what we expect.
	if status := rr.Code; status != http.StatusTemporaryRedirect {
		t.Errorf("handler returned wrong status code: got %v want %v",
			status, http.StatusOK)
	}

	// Check the response body is what we expect.
	if rr.HeaderMap.Get("Location") != "https://blog.singleton.io/" {
		t.Errorf("handler returned unexpected body: got %v want %v...",
			rr.HeaderMap, "https://blog.singleton.io/")
	}

}

func TestPathHandler(t *testing.T) {
	var ctx = context.Background()

	// Create a request to pass to our handler. We don't have any query parameters for now, so we'll
	// pass 'nil' as the third parameter.
	req, err := http.NewRequest("GET", "/path/to/a/resource.html?foo=bar&baz=1", nil)
	if err != nil {
		t.Fatal(err)
	}
	req.Header.Set("Referer", "https://blog.singleton.io/foo/bar/page.html")
	req.Header.Set("X-Forwarded-For", "157.131.203.2")
	req.Header.Set("User-Agent", "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_14_2) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/76.0.3809.100 Safari/537.36")
	req.Host = "blog.davidsingleton.org"

	s, client := initFixtures()
	defer s.Close()

	client.Set(ctx, "{blog.davidsingleton.org}:redirect", "blog.singleton.io", 0)

	// We create a ResponseRecorder (which satisfies http.ResponseWriter) to record the response.
	rr := httptest.NewRecorder()
	handler := newRedirectHandler(client)

	// Our handlers satisfy http.Handler, so we can call their ServeHTTP method
	// directly and pass in our Request and ResponseRecorder.
	handler.ServeHTTP(rr, req)

	// Check the status code is what we expect.
	if status := rr.Code; status != http.StatusTemporaryRedirect {
		t.Errorf("handler returned wrong status code: got %v want %v",
			status, http.StatusOK)
	}

	expected := "https://blog.singleton.io/path/to/a/resource.html?foo=bar&baz=1"
	// Check the response body is what we expect.
	if rr.HeaderMap.Get("Location") != expected {
		t.Errorf("handler returned unexpected body: got %v want %v...",
			rr.HeaderMap, expected)
	}

}
func TestUnregisteredDomain(t *testing.T) {
	// Create a request to pass to our handler. We don't have any query parameters for now, so we'll
	// pass 'nil' as the third parameter.
	req, err := http.NewRequest("GET", "/pixel.gif?uid=foo&referrer=google.com", nil)
	if err != nil {
		t.Fatal(err)
	}
	req.Header.Set("Referer", "https://blog.singleton.foo/foo/bar/page.html")
	req.Header.Set("X-Forwarded-For", "157.131.203.2")
	req.Header.Set("User-Agent", "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_14_2) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/76.0.3809.100 Safari/537.36")
	req.Host = "notreg.com"

	_, client := initFixtures()
	defer client.Close()

	// We create a ResponseRecorder (which satisfies http.ResponseWriter) to record the response.
	rr := httptest.NewRecorder()
	handler := newRedirectHandler(client)

	// Our handlers satisfy http.Handler, so we can call their ServeHTTP method
	// directly and pass in our Request and ResponseRecorder.
	handler.ServeHTTP(rr, req)

	// Check the status code is what we expect.
	if status := rr.Code; status != http.StatusNotFound {
		t.Errorf("handler returned wrong status code: got %v want %v",
			status, http.StatusNotFound)
	}
}

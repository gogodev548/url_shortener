// comment

package main

import (
	"errors"
	"fmt"
	"log"
	"net/http"
	"net/url"
)

type shortener struct {
	links map[string]string
	next  int
}

func newShortener() *shortener {
	return &shortener{
		links: make(map[string]string),
		next:  1,
	}
}

func (s *shortener) createShortURL(rawURL string) (string, error) {
	if err := validateURL(rawURL); err != nil {
		return "", err
	}

	code := fmt.Sprintf("%d", s.next)
	s.links[code] = rawURL
	s.next++
	return code, nil
}

func (s *shortener) getOriginalURL(code string) (string, bool) {
	originalURL, ok := s.links[code]
	return originalURL, ok
}

func validateURL(rawURL string) error {
	parsed, err := url.ParseRequestURI(rawURL)
	if err != nil {
		return errors.New("invalid url")
	}
	if parsed.Scheme != "http" && parsed.Scheme != "https" {
		return errors.New("url must start with http:// or https://")
	}
	return nil
}

func main() {
	s := newShortener()

	mux := http.NewServeMux()

	mux.HandleFunc("POST /shorten", func(w http.ResponseWriter, r *http.Request) {
		rawURL := r.URL.Query().Get("url")
		if rawURL == "" {
			http.Error(w, "missing query param: url", http.StatusBadRequest)
			return
		}

		code, err := s.createShortURL(rawURL)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		host := r.Host
		if host == "" {
			host = "localhost:8080"
		}

		shortURL := fmt.Sprintf("http://%s/%s", host, code)
		w.WriteHeader(http.StatusCreated)
		if _, err := fmt.Fprintf(w, "short url: %s\n", shortURL); err != nil {
			log.Printf("failed to write response: %v", err)
		}
	})

	mux.HandleFunc("GET /{code}", func(w http.ResponseWriter, r *http.Request) {
		code := r.PathValue("code")
		if code == "" {
			http.NotFound(w, r)
			return
		}

		originalURL, exists := s.getOriginalURL(code)
		if !exists {
			http.NotFound(w, r)
			return
		}

		http.Redirect(w, r, originalURL, http.StatusFound)
	})

	mux.HandleFunc("GET /", func(w http.ResponseWriter, r *http.Request) {
		if _, err := fmt.Fprintln(w, "URL Shortener is running"); err != nil {
			log.Printf("failed to write response: %v", err)
		}
	})

	addr := ":8080"
	log.Printf("Server started at http://localhost%s", addr)
	if err := http.ListenAndServe(addr, mux); err != nil {
		log.Fatalf("server failed: %v", err)
	}
}
qwe
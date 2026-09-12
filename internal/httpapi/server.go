package httpapi

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"net"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"github.com/k11ngp1ng/url-shortener-go/internal/domain"
	"github.com/k11ngp1ng/url-shortener-go/internal/store"
)

type Server struct {
	store   store.URLStore
	limiter *rateLimiter
	metrics metrics
	logger  *slog.Logger
}

func NewServer(urlStore store.URLStore) *Server {
	return &Server{store: urlStore, limiter: newRateLimiter(60, time.Minute), logger: slog.Default()}
}

// WithRateLimit changes the per-client API request allowance. A non-positive
// limit disables limiting, which is useful in focused tests.
func (s *Server) WithRateLimit(limit int, window time.Duration) *Server {
	s.limiter = newRateLimiter(limit, window)
	return s
}

func (s *Server) Routes() http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("GET /health", s.health)
	mux.HandleFunc("GET /metrics", s.prometheusMetrics)
	mux.HandleFunc("POST /api/v1/urls", s.createURL)
	mux.HandleFunc("GET /api/v1/urls/{code}", s.getURL)
	mux.HandleFunc("GET /{code}", s.redirect)
	return s.observe(s.limit(mux))
}

func (s *Server) health(w http.ResponseWriter, _ *http.Request) {
	if checker, ok := s.store.(interface{ Ping(context.Context) error }); ok {
		if err := checker.Ping(context.Background()); err != nil {
			writeError(w, http.StatusServiceUnavailable, "storage unavailable")
			return
		}
	}
	writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}

func (s *Server) createURL(w http.ResponseWriter, r *http.Request) {
	var input struct {
		URL       string     `json:"url"`
		ExpiresAt *time.Time `json:"expires_at"`
	}
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		writeError(w, http.StatusBadRequest, "invalid JSON body")
		return
	}
	if !isValidHTTPURL(input.URL) {
		writeError(w, http.StatusBadRequest, "url must be a valid http or https URL")
		return
	}
	if input.ExpiresAt != nil && !input.ExpiresAt.After(time.Now().UTC()) {
		writeError(w, http.StatusBadRequest, "expires_at must be in the future")
		return
	}

	shortURL, err := s.createWithUniqueCode(input.URL, input.ExpiresAt)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "could not create short URL")
		return
	}

	baseURL := "http://" + r.Host
	writeJSON(w, http.StatusCreated, map[string]any{
		"code":         shortURL.Code,
		"short_url":    baseURL + "/" + shortURL.Code,
		"original_url": shortURL.OriginalURL,
		"clicks":       shortURL.Clicks,
		"expires_at":   shortURL.ExpiresAt,
	})
}

func (s *Server) getURL(w http.ResponseWriter, r *http.Request) {
	url, err := s.store.Get(r.PathValue("code"))
	if errors.Is(err, store.ErrNotFound) {
		writeError(w, http.StatusNotFound, "short URL not found")
		return
	}
	if err != nil {
		writeError(w, http.StatusInternalServerError, "could not retrieve short URL")
		return
	}
	writeJSON(w, http.StatusOK, url)
}

func (s *Server) redirect(w http.ResponseWriter, r *http.Request) {
	url, err := s.store.Get(r.PathValue("code"))
	if errors.Is(err, store.ErrNotFound) {
		writeError(w, http.StatusNotFound, "short URL not found")
		return
	}
	if err != nil {
		writeError(w, http.StatusInternalServerError, "could not retrieve short URL")
		return
	}
	if url.Expired(time.Now().UTC()) {
		writeError(w, http.StatusGone, "short URL has expired")
		return
	}
	url, err = s.store.IncrementClicks(url.Code)
	if errors.Is(err, store.ErrNotFound) {
		writeError(w, http.StatusNotFound, "short URL not found")
		return
	}
	if err != nil {
		writeError(w, http.StatusInternalServerError, "could not redirect")
		return
	}
	http.Redirect(w, r, url.OriginalURL, http.StatusFound)
}

func (s *Server) createWithUniqueCode(originalURL string, expiresAt *time.Time) (domain.URL, error) {
	for range 3 {
		code, err := generateCode()
		if err != nil {
			return domain.URL{}, err
		}
		created, err := s.store.Create(domain.URL{Code: code, OriginalURL: originalURL, ExpiresAt: expiresAt})
		if !errors.Is(err, store.ErrCodeAlreadyExists) {
			return created, err
		}
	}
	return domain.URL{}, errors.New("could not generate unique short code")
}

type statusRecorder struct {
	http.ResponseWriter
	status int
}

func (r *statusRecorder) WriteHeader(status int) {
	r.status = status
	r.ResponseWriter.WriteHeader(status)
}
func (r *statusRecorder) Write(b []byte) (int, error) {
	if r.status == 0 {
		r.WriteHeader(http.StatusOK)
	}
	return r.ResponseWriter.Write(b)
}

func (s *Server) observe(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		recorder := &statusRecorder{ResponseWriter: w}
		next.ServeHTTP(recorder, r)
		status := recorder.status
		if status == 0 {
			status = http.StatusOK
		}
		s.metrics.requests.Add(1)
		s.metrics.durationNS.Add(time.Since(start).Nanoseconds())
		s.logger.Info("http request", "method", r.Method, "path", r.URL.Path, "status", status, "duration", time.Since(start))
	})
}

func (s *Server) limit(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/health" && r.URL.Path != "/metrics" && !s.limiter.allow(clientIP(r.RemoteAddr), time.Now()) {
			w.Header().Set("Retry-After", strconv.Itoa(int(s.limiter.window.Seconds())))
			writeError(w, http.StatusTooManyRequests, "rate limit exceeded")
			return
		}
		next.ServeHTTP(w, r)
	})
}

func (s *Server) prometheusMetrics(w http.ResponseWriter, _ *http.Request) {
	w.Header().Set("Content-Type", "text/plain; version=0.0.4")
	fmt.Fprintf(w, "# TYPE url_shortener_http_requests_total counter\nurl_shortener_http_requests_total %d\n# TYPE url_shortener_http_request_duration_seconds_total counter\nurl_shortener_http_request_duration_seconds_total %.9f\n", s.metrics.requests.Load(), float64(s.metrics.durationNS.Load())/float64(time.Second))
}

type metrics struct {
	requests   atomic.Int64
	durationNS atomic.Int64
}
type visitor struct {
	count int
	since time.Time
}
type rateLimiter struct {
	mu       sync.Mutex
	visitors map[string]visitor
	limit    int
	window   time.Duration
}

func newRateLimiter(limit int, window time.Duration) *rateLimiter {
	return &rateLimiter{visitors: make(map[string]visitor), limit: limit, window: window}
}
func (l *rateLimiter) allow(key string, now time.Time) bool {
	if l.limit <= 0 {
		return true
	}
	l.mu.Lock()
	defer l.mu.Unlock()
	v := l.visitors[key]
	if v.since.IsZero() || now.Sub(v.since) >= l.window {
		l.visitors[key] = visitor{count: 1, since: now}
		return true
	}
	if v.count >= l.limit {
		return false
	}
	v.count++
	l.visitors[key] = v
	return true
}
func clientIP(remote string) string {
	host, _, err := net.SplitHostPort(remote)
	if err == nil {
		return host
	}
	return remote
}

func generateCode() (string, error) {
	bytes := make([]byte, 3)
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}
	return hex.EncodeToString(bytes), nil
}

func isValidHTTPURL(value string) bool {
	parsed, err := url.ParseRequestURI(value)
	return err == nil && parsed.Host != "" && (parsed.Scheme == "http" || parsed.Scheme == "https")
}

func writeJSON(w http.ResponseWriter, status int, value any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(value)
}

func writeError(w http.ResponseWriter, status int, message string) {
	writeJSON(w, status, map[string]string{"error": strings.TrimSpace(message)})
}

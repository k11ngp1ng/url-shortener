package httpapi

import (
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"errors"
	"net/http"
	"net/url"
	"strings"

	"github.com/k11ngp1ng/url-shortener-go/internal/domain"
	"github.com/k11ngp1ng/url-shortener-go/internal/store"
)

type Server struct {
	store store.URLStore
}

func NewServer(urlStore store.URLStore) *Server {
	return &Server{store: urlStore}
}

func (s *Server) Routes() http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("GET /health", s.health)
	mux.HandleFunc("POST /api/v1/urls", s.createURL)
	mux.HandleFunc("GET /api/v1/urls/{code}", s.getURL)
	mux.HandleFunc("GET /{code}", s.redirect)
	return mux
}

func (s *Server) health(w http.ResponseWriter, _ *http.Request) {
	writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}

func (s *Server) createURL(w http.ResponseWriter, r *http.Request) {
	var input struct {
		URL string `json:"url"`
	}
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		writeError(w, http.StatusBadRequest, "invalid JSON body")
		return
	}
	if !isValidHTTPURL(input.URL) {
		writeError(w, http.StatusBadRequest, "url must be a valid http or https URL")
		return
	}

	shortURL, err := s.createWithUniqueCode(input.URL)
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
	})
}

func (s *Server) getURL(w http.ResponseWriter, r *http.Request) {
	url, err := s.store.Get(r.PathValue("code"))
	if errors.Is(err, store.ErrNotFound) {
		writeError(w, http.StatusNotFound, "short URL not found")
		return
	}
	writeJSON(w, http.StatusOK, url)
}

func (s *Server) redirect(w http.ResponseWriter, r *http.Request) {
	url, err := s.store.IncrementClicks(r.PathValue("code"))
	if errors.Is(err, store.ErrNotFound) {
		writeError(w, http.StatusNotFound, "short URL not found")
		return
	}
	http.Redirect(w, r, url.OriginalURL, http.StatusFound)
}

func (s *Server) createWithUniqueCode(originalURL string) (domain.URL, error) {
	for range 3 {
		code, err := generateCode()
		if err != nil {
			return domain.URL{}, err
		}
		created, err := s.store.Create(domain.URL{Code: code, OriginalURL: originalURL})
		if !errors.Is(err, store.ErrCodeAlreadyExists) {
			return created, err
		}
	}
	return domain.URL{}, errors.New("could not generate unique short code")
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

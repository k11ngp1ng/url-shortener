package httpapi_test

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/k11ngp1ng/url-shortener-go/internal/httpapi"
	"github.com/k11ngp1ng/url-shortener-go/internal/store"
)

func newTestServer() http.Handler {
	return httpapi.NewServer(store.NewMemoryURLStore()).Routes()
}

func TestCreateURLAndReadMetrics(t *testing.T) {
	server := newTestServer()
	request := httptest.NewRequest(http.MethodPost, "/api/v1/urls", bytes.NewBufferString(`{"url":"https://go.dev"}`))
	request.Header.Set("Content-Type", "application/json")
	response := httptest.NewRecorder()

	server.ServeHTTP(response, request)
	if response.Code != http.StatusCreated {
		t.Fatalf("expected status %d, got %d", http.StatusCreated, response.Code)
	}

	var created struct {
		Code string `json:"code"`
	}
	if err := json.NewDecoder(response.Body).Decode(&created); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if len(created.Code) != 6 {
		t.Fatalf("expected a 6-character code, got %q", created.Code)
	}

	metricsRequest := httptest.NewRequest(http.MethodGet, "/api/v1/urls/"+created.Code, nil)
	metricsResponse := httptest.NewRecorder()
	server.ServeHTTP(metricsResponse, metricsRequest)
	if metricsResponse.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d", http.StatusOK, metricsResponse.Code)
	}
}

func TestCreateURLRejectsInvalidURL(t *testing.T) {
	server := newTestServer()
	request := httptest.NewRequest(http.MethodPost, "/api/v1/urls", bytes.NewBufferString(`{"url":"not-a-url"}`))
	response := httptest.NewRecorder()

	server.ServeHTTP(response, request)
	if response.Code != http.StatusBadRequest {
		t.Fatalf("expected status %d, got %d", http.StatusBadRequest, response.Code)
	}
}

func TestRedirectIncrementsClicks(t *testing.T) {
	server := newTestServer()
	createRequest := httptest.NewRequest(http.MethodPost, "/api/v1/urls", bytes.NewBufferString(`{"url":"https://go.dev"}`))
	createResponse := httptest.NewRecorder()
	server.ServeHTTP(createResponse, createRequest)

	var created struct {
		Code string `json:"code"`
	}
	_ = json.NewDecoder(createResponse.Body).Decode(&created)

	redirectRequest := httptest.NewRequest(http.MethodGet, "/"+created.Code, nil)
	redirectResponse := httptest.NewRecorder()
	server.ServeHTTP(redirectResponse, redirectRequest)
	if redirectResponse.Code != http.StatusFound {
		t.Fatalf("expected status %d, got %d", http.StatusFound, redirectResponse.Code)
	}

	metricsRequest := httptest.NewRequest(http.MethodGet, "/api/v1/urls/"+created.Code, nil)
	metricsResponse := httptest.NewRecorder()
	server.ServeHTTP(metricsResponse, metricsRequest)
	var metrics struct {
		Clicks int64 `json:"clicks"`
	}
	_ = json.NewDecoder(metricsResponse.Body).Decode(&metrics)
	if metrics.Clicks != 1 {
		t.Fatalf("expected 1 click, got %d", metrics.Clicks)
	}
}

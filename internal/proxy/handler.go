package proxy

import (
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/thomasmarlow/the-trainman/internal/config"
)

type Handler struct {
	configManager *config.Manager
	client        *http.Client
}

func NewHandler(configManager *config.Manager) *Handler {
	return &Handler{
		configManager: configManager,
		client: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

func (h *Handler) HandleProxy(w http.ResponseWriter, r *http.Request) {
	// Extract service name and path from URL
	// Expected format: /api/{service}/{path...}
	serviceName := chi.URLParam(r, "service")
	if serviceName == "" {
		http.Error(w, "Service name is required", http.StatusBadRequest)
		return
	}

	// Check x-request-id header if enforcement is enabled
	if h.configManager.ShouldRequireRequestID(serviceName) {
		requestID := r.Header.Get("x-request-id")
		if requestID == "" {
			// Log detailed rejection information
			log.Printf("request rejected: missing x-request-id header for service '%s' from IP %s",
				serviceName, r.RemoteAddr)

			errorMsg := h.configManager.GetRequestIDErrorMessage()
			http.Error(w, errorMsg, http.StatusBadRequest)
			return
		}
	}

	// Check x-api-key header if enforcement is enabled
	if h.configManager.ShouldRequireAPIKey(serviceName) {
		apiKey := r.Header.Get("x-api-key")
		if apiKey == "" {
			// Log detailed rejection information
			log.Printf("request rejected: missing x-api-key header for service '%s' from IP %s",
				serviceName, r.RemoteAddr)

			errorMsg := h.configManager.GetAPIKeyErrorMessage()
			http.Error(w, errorMsg, http.StatusBadRequest)
			return
		}

		if !h.configManager.IsValidAPIKey(apiKey) {
			// Log detailed rejection information
			log.Printf("request rejected: invalid x-api-key for service '%s' from IP %s",
				serviceName, r.RemoteAddr)

			http.Error(w, "Invalid API key", http.StatusUnauthorized)
			return
		}
	}

	// Get the remaining path after /api/{service}/
	fullPath := r.URL.Path
	prefix := fmt.Sprintf("/api/%s/", serviceName)
	if !strings.HasPrefix(fullPath, prefix) {
		http.Error(w, "Invalid path format", http.StatusBadRequest)
		return
	}

	// Extract the path to forward (everything after /api/{service}/)
	forwardPath := strings.TrimPrefix(fullPath, prefix)
	if forwardPath == "" {
		forwardPath = "/"
	} else {
		forwardPath = "/" + forwardPath
	}

	// Get backend service configuration
	service, exists := h.configManager.GetBackendService(serviceName)
	if !exists {
		http.Error(w, fmt.Sprintf("Service '%s' not found or disabled", serviceName), http.StatusNotFound)
		return
	}

	// Build target URL
	targetURL, err := url.Parse(service.URL)
	if err != nil {
		log.Printf("error parsing service URL %s: %v", service.URL, err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	targetURL.Path = forwardPath
	targetURL.RawQuery = r.URL.RawQuery

	log.Printf("proxying %s %s -> %s", r.Method, r.URL.Path, targetURL.String())

	// Create the proxy request
	proxyReq, err := http.NewRequest(r.Method, targetURL.String(), r.Body)
	if err != nil {
		log.Printf("error creating proxy request: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	// Copy headers from original request
	for name, values := range r.Header {
		// Skip hop-by-hop headers
		if isHopByHopHeader(name) {
			continue
		}
		for _, value := range values {
			proxyReq.Header.Add(name, value)
		}
	}

	// Set X-Forwarded headers
	proxyReq.Header.Set("X-Forwarded-For", r.RemoteAddr)
	proxyReq.Header.Set("X-Forwarded-Proto", "http")
	proxyReq.Header.Set("X-Forwarded-Host", r.Host)

	// Execute the proxy request
	resp, err := h.client.Do(proxyReq)
	if err != nil {
		log.Printf("error executing proxy request to %s: %v", targetURL.String(), err)
		http.Error(w, "Service unavailable", http.StatusServiceUnavailable)
		return
	}
	defer resp.Body.Close()

	// Copy response headers
	for name, values := range resp.Header {
		if isHopByHopHeader(name) {
			continue
		}
		for _, value := range values {
			w.Header().Add(name, value)
		}
	}

	// Set response status code
	w.WriteHeader(resp.StatusCode)

	// Copy response body
	_, err = io.Copy(w, resp.Body)
	if err != nil {
		log.Printf("error copying response body: %v", err)
	}
}

// isHopByHopHeader checks if a header is hop-by-hop and should not be forwarded
func isHopByHopHeader(header string) bool {
	hopByHopHeaders := []string{
		"Connection",
		"Keep-Alive",
		"Proxy-Authenticate",
		"Proxy-Authorization",
		"Te",
		"Trailers",
		"Transfer-Encoding",
		"Upgrade",
	}

	headerLower := strings.ToLower(header)
	for _, hopHeader := range hopByHopHeaders {
		if strings.ToLower(hopHeader) == headerLower {
			return true
		}
	}
	return false
}

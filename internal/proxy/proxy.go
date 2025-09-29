package proxy

import (
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/aykay76/ai-idp/internal/logger"
)

// ProxyConfig holds configuration for service proxying
type ProxyConfig struct {
	ApplicationServiceURL string
	TeamServiceURL        string
	Logger                *logger.Logger
}

// ProxyHandler handles proxying requests to backend services
type ProxyHandler struct {
	config *ProxyConfig
}

// NewProxyHandler creates a new proxy handler
func NewProxyHandler(config *ProxyConfig) *ProxyHandler {
	return &ProxyHandler{
		config: config,
	}
}

// ServeHTTP implements the http.Handler interface for proxying
func (p *ProxyHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	// Determine target service based on path
	var targetURL string
	var serviceName string

	if strings.HasPrefix(r.URL.Path, "/api/v1/teams") {
		targetURL = p.config.TeamServiceURL
		serviceName = "team-service"
	} else if strings.HasPrefix(r.URL.Path, "/api/v1/applications") {
		targetURL = p.config.ApplicationServiceURL
		serviceName = "application-service"
	} else {
		p.config.Logger.WithFields(logger.LogFields{
			logger.FieldHTTPMethod: r.Method,
			logger.FieldHTTPPath:   r.URL.Path,
		}).Warn("No service found for path")
		http.Error(w, "Service not found", http.StatusNotFound)
		return
	}

	// Parse target URL
	target, err := url.Parse(targetURL)
	if err != nil {
		p.config.Logger.WithFields(logger.LogFields{
			logger.FieldError: err.Error(),
			"target_url":      targetURL,
		}).Error("Failed to parse target URL")
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	// Create the proxy request
	proxyURL := &url.URL{
		Scheme:   target.Scheme,
		Host:     target.Host,
		Path:     r.URL.Path,
		RawQuery: r.URL.RawQuery,
	}

	p.config.Logger.WithFields(logger.LogFields{
		logger.FieldHTTPMethod: r.Method,
		logger.FieldHTTPPath:   r.URL.Path,
		"service":              serviceName,
		"proxy_url":            proxyURL.String(),
	}).Debug("Proxying request")

	// Create proxy request
	proxyReq, err := http.NewRequestWithContext(r.Context(), r.Method, proxyURL.String(), r.Body)
	if err != nil {
		p.config.Logger.WithFields(logger.LogFields{
			logger.FieldError: err.Error(),
		}).Error("Failed to create proxy request")
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	// Copy headers from original request
	for name, values := range r.Header {
		// Skip hop-by-hop headers
		if name == "Connection" || name == "Keep-Alive" || name == "Proxy-Authenticate" ||
			name == "Proxy-Authorization" || name == "Te" || name == "Trailers" ||
			name == "Transfer-Encoding" || name == "Upgrade" {
			continue
		}
		for _, value := range values {
			proxyReq.Header.Add(name, value)
		}
	}

	// Add X-Forwarded headers
	proxyReq.Header.Set("X-Forwarded-For", r.RemoteAddr)
	proxyReq.Header.Set("X-Forwarded-Host", r.Host)
	proxyReq.Header.Set("X-Forwarded-Proto", "http") // TODO: detect actual protocol

	// Create HTTP client with timeout
	client := &http.Client{
		Timeout: 30 * time.Second,
	}

	// Make the proxy request
	resp, err := client.Do(proxyReq)
	if err != nil {
		p.config.Logger.WithFields(logger.LogFields{
			logger.FieldError: err.Error(),
			"service":         serviceName,
			"proxy_url":       proxyURL.String(),
		}).Error("Proxy request failed")
		http.Error(w, "Service Unavailable", http.StatusServiceUnavailable)
		return
	}
	defer resp.Body.Close()

	// Copy response headers
	for name, values := range resp.Header {
		// Skip hop-by-hop headers
		if name == "Connection" || name == "Keep-Alive" || name == "Proxy-Authenticate" ||
			name == "Proxy-Authorization" || name == "Te" || name == "Trailers" ||
			name == "Transfer-Encoding" || name == "Upgrade" {
			continue
		}
		for _, value := range values {
			w.Header().Add(name, value)
		}
	}

	// Set status code
	w.WriteHeader(resp.StatusCode)

	// Copy response body
	_, err = io.Copy(w, resp.Body)
	if err != nil {
		p.config.Logger.WithFields(logger.LogFields{
			logger.FieldError: err.Error(),
			"service":         serviceName,
		}).Error("Failed to copy response body")
		return
	}

	p.config.Logger.WithFields(logger.LogFields{
		logger.FieldHTTPMethod: r.Method,
		logger.FieldHTTPPath:   r.URL.Path,
		logger.FieldHTTPStatus: resp.StatusCode,
		"service":              serviceName,
	}).Debug("Request proxied successfully")
}

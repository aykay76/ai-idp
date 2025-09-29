package proxypackage proxy

package proxy

import (

	"net/http"import (

	"net/http/httptest"	"net/http"

	"strings"	"net/http/httptest"

	"testing"	"strings"

	"testing"

	"github.com/aykay76/ai-idp/internal/logger"

	"github.com/stretchr/testify/assert"	"github.com/aykay76/ai-idp/internal/logger"

)	"github.com/stretchr/testify/assert"

)

func TestProxyHandler_ServeHTTP(t *testing.T) {

	// Create mock backend servers for teams and applicationsfunc TestProxyHandler_ServeHTTP(t *testing.T) {

	teamsServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {	// Create mock backend servers for teams and applications

		w.Header().Set("Content-Type", "application/json")	teamsServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		w.WriteHeader(http.StatusOK)		w.Header().Set("Content-Type", "application/json")

		w.Write([]byte(`{"teams":[],"pagination":{"limit":50,"offset":0,"total":0}}`))		w.WriteHeader(http.StatusOK)

	}))		w.Write([]byte(`{"teams":[],"pagination":{"limit":50,"offset":0,"total":0}}`))

	defer teamsServer.Close()	}))

	defer teamsServer.Close()

	appsServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		w.Header().Set("Content-Type", "application/json")	appsServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		w.WriteHeader(http.StatusOK)		w.Header().Set("Content-Type", "application/json")

		w.Write([]byte(`{"applications":[],"pagination":{"limit":50,"offset":0,"total":0}}`))		w.WriteHeader(http.StatusOK)

	}))		w.Write([]byte(`{"applications":[],"pagination":{"limit":50,"offset":0,"total":0}}`))

	defer appsServer.Close()	}))

	defer appsServer.Close()

	// Create proxy config

	config := &ProxyConfig{	// Create proxy config

		TeamServiceURL:        teamsServer.URL,	config := &ProxyConfig{

		ApplicationServiceURL: appsServer.URL,		TeamServiceURL:        teamsServer.URL,

		Logger:               logger.New("debug", "text"),		ApplicationServiceURL: appsServer.URL,

	}		Logger:               logger.New("debug", "text"),

	}

	// Create proxy handler

	proxy := NewProxyHandler(config)	// Create proxy handler

	proxy := NewProxyHandler(config)

	t.Run("proxy teams request", func(t *testing.T) {

		req := httptest.NewRequest(http.MethodGet, "/api/v1/teams", nil)	t.Run("proxy teams request", func(t *testing.T) {

		rr := httptest.NewRecorder()		req := httptest.NewRequest(http.MethodGet, "/api/v1/teams", nil)

		rr := httptest.NewRecorder()

		proxy.ServeHTTP(rr, req)

		proxy.ServeHTTP(rr, req)

		assert.Equal(t, http.StatusOK, rr.Code)

		assert.Equal(t, "application/json", rr.Header().Get("Content-Type"))		assert.Equal(t, http.StatusOK, rr.Code)

				assert.Equal(t, "application/json", rr.Header().Get("Content-Type"))

		body := rr.Body.String()		

		assert.Contains(t, body, "teams")		body := rr.Body.String()

		assert.Contains(t, body, "pagination")		assert.Contains(t, body, "teams")

	})		assert.Contains(t, body, "pagination")

	})

	t.Run("proxy applications request", func(t *testing.T) {

		req := httptest.NewRequest(http.MethodGet, "/api/v1/applications", nil)	t.Run("proxy applications request", func(t *testing.T) {

		rr := httptest.NewRecorder()		req := httptest.NewRequest(http.MethodGet, "/api/v1/applications", nil)

		rr := httptest.NewRecorder()

		proxy.ServeHTTP(rr, req)

		proxy.ServeHTTP(rr, req)

		assert.Equal(t, http.StatusOK, rr.Code)

		assert.Equal(t, "application/json", rr.Header().Get("Content-Type"))		assert.Equal(t, http.StatusOK, rr.Code)

				assert.Equal(t, "application/json", rr.Header().Get("Content-Type"))

		body := rr.Body.String()		

		assert.Contains(t, body, "applications")		body := rr.Body.String()

		assert.Contains(t, body, "pagination")		assert.Contains(t, body, "applications")

	})		assert.Contains(t, body, "pagination")

	})

	t.Run("proxy POST request", func(t *testing.T) {

		// Update teams server to handle POST	t.Run("proxy POST request", func(t *testing.T) {

		teamsServer = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {		// Update teams server to handle POST

			assert.Equal(t, http.MethodPost, r.Method)		teamsServer = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

			assert.Equal(t, "application/json", r.Header.Get("Content-Type"))			assert.Equal(t, http.MethodPost, r.Method)

						assert.Equal(t, "application/json", r.Header.Get("Content-Type"))

			w.Header().Set("Content-Type", "application/json")			

			w.WriteHeader(http.StatusCreated)			w.Header().Set("Content-Type", "application/json")

			w.Write([]byte(`{"id":"123","name":"test-team"}`))			w.WriteHeader(http.StatusCreated)

		}))			w.Write([]byte(`{"id":"123","name":"test-team"}`))

		defer teamsServer.Close()		}))

		defer teamsServer.Close()

		config.TeamServiceURL = teamsServer.URL

		proxy = NewProxyHandler(config)		config.TeamServiceURL = teamsServer.URL

		proxy = NewProxyHandler(config)

		body := strings.NewReader(`{"name":"test-team","lead_email":"test@example.com"}`)

		req := httptest.NewRequest(http.MethodPost, "/api/v1/teams", body)		body := strings.NewReader(`{"name":"test-team","lead_email":"test@example.com"}`)

		req.Header.Set("Content-Type", "application/json")		req := httptest.NewRequest(http.MethodPost, "/api/v1/teams", body)

				req.Header.Set("Content-Type", "application/json")

		rr := httptest.NewRecorder()		

		proxy.ServeHTTP(rr, req)		rr := httptest.NewRecorder()

		proxy.ServeHTTP(rr, req)

		assert.Equal(t, http.StatusCreated, rr.Code)

		assert.Equal(t, "application/json", rr.Header().Get("Content-Type"))		assert.Equal(t, http.StatusCreated, rr.Code)

	})		assert.Equal(t, "application/json", rr.Header().Get("Content-Type"))

	})

	t.Run("unknown path returns 404", func(t *testing.T) {

		req := httptest.NewRequest(http.MethodGet, "/api/v1/unknown", nil)	t.Run("unknown path returns 404", func(t *testing.T) {

		rr := httptest.NewRecorder()		req := httptest.NewRequest(http.MethodGet, "/api/v1/unknown", nil)

		rr := httptest.NewRecorder()

		proxy.ServeHTTP(rr, req)

		proxy.ServeHTTP(rr, req)

		assert.Equal(t, http.StatusNotFound, rr.Code)

	})		assert.Equal(t, http.StatusNotFound, rr.Code)

	})

	t.Run("preserve query parameters", func(t *testing.T) {

		// Update teams server to verify query params	t.Run("preserve query parameters", func(t *testing.T) {

		teamsServer = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {		// Update teams server to verify query params

			assert.Equal(t, "50", r.URL.Query().Get("limit"))		teamsServer = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

			assert.Equal(t, "10", r.URL.Query().Get("offset"))			assert.Equal(t, "50", r.URL.Query().Get("limit"))

						assert.Equal(t, "10", r.URL.Query().Get("offset"))

			w.Header().Set("Content-Type", "application/json")			

			w.WriteHeader(http.StatusOK)			w.Header().Set("Content-Type", "application/json")

			w.Write([]byte(`{"teams":[],"pagination":{"limit":50,"offset":10,"total":0}}`))			w.WriteHeader(http.StatusOK)

		}))			w.Write([]byte(`{"teams":[],"pagination":{"limit":50,"offset":10,"total":0}}`))

		defer teamsServer.Close()		}))

		defer teamsServer.Close()

		config.TeamServiceURL = teamsServer.URL

		proxy = NewProxyHandler(config)		config.TeamServiceURL = teamsServer.URL

		proxy = NewProxyHandler(config)

		req := httptest.NewRequest(http.MethodGet, "/api/v1/teams?limit=50&offset=10", nil)

		rr := httptest.NewRecorder()		req := httptest.NewRequest(http.MethodGet, "/api/v1/teams?limit=50&offset=10", nil)

		rr := httptest.NewRecorder()

		proxy.ServeHTTP(rr, req)

		proxy.ServeHTTP(rr, req)

		assert.Equal(t, http.StatusOK, rr.Code)

	})		assert.Equal(t, http.StatusOK, rr.Code)

	})

	t.Run("preserve headers", func(t *testing.T) {

		// Update teams server to verify headers	t.Run("preserve headers", func(t *testing.T) {

		teamsServer = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {		// Update teams server to verify headers

			assert.Equal(t, "Bearer token123", r.Header.Get("Authorization"))		teamsServer = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

			assert.Equal(t, "application/json", r.Header.Get("Content-Type"))			assert.Equal(t, "Bearer token123", r.Header.Get("Authorization"))

			assert.NotEmpty(t, r.Header.Get("X-Forwarded-For"))			assert.Equal(t, "application/json", r.Header.Get("Content-Type"))

			assert.NotEmpty(t, r.Header.Get("X-Forwarded-Host"))			assert.NotEmpty(t, r.Header.Get("X-Forwarded-For"))

			assert.NotEmpty(t, r.Header.Get("X-Forwarded-Proto"))			assert.NotEmpty(t, r.Header.Get("X-Forwarded-Host"))

						assert.NotEmpty(t, r.Header.Get("X-Forwarded-Proto"))

			w.Header().Set("Content-Type", "application/json")			

			w.WriteHeader(http.StatusOK)			w.Header().Set("Content-Type", "application/json")

			w.Write([]byte(`{"teams":[]}`))			w.WriteHeader(http.StatusOK)

		}))			w.Write([]byte(`{"teams":[]}`))

		defer teamsServer.Close()		}))

		defer teamsServer.Close()

		config.TeamServiceURL = teamsServer.URL

		proxy = NewProxyHandler(config)		config.TeamServiceURL = teamsServer.URL

		proxy = NewProxyHandler(config)

		req := httptest.NewRequest(http.MethodGet, "/api/v1/teams", nil)

		req.Header.Set("Authorization", "Bearer token123")		req := httptest.NewRequest(http.MethodGet, "/api/v1/teams", nil)

		req.Header.Set("Content-Type", "application/json")		req.Header.Set("Authorization", "Bearer token123")

				req.Header.Set("Content-Type", "application/json")

		rr := httptest.NewRecorder()		

		proxy.ServeHTTP(rr, req)		rr := httptest.NewRecorder()

		proxy.ServeHTTP(rr, req)

		assert.Equal(t, http.StatusOK, rr.Code)

	})		assert.Equal(t, http.StatusOK, rr.Code)

}	})

}

func TestProxyHandler_ServiceUnavailable(t *testing.T) {

	config := &ProxyConfig{func TestProxyHandler_ServiceUnavailable(t *testing.T) {

		TeamServiceURL:        "http://localhost:99999", // Invalid port	config := &ProxyConfig{

		ApplicationServiceURL: "http://localhost:99998", // Invalid port		TeamServiceURL:        "http://localhost:99999", // Invalid port

		Logger:               logger.New("debug", "text"),		ApplicationServiceURL: "http://localhost:99998", // Invalid port

	}		Logger:               logger.New("debug", "text"),

	}

	proxy := NewProxyHandler(config)

	proxy := NewProxyHandler(config)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/teams", nil)

	rr := httptest.NewRecorder()	req := httptest.NewRequest(http.MethodGet, "/api/v1/teams", nil)

	rr := httptest.NewRecorder()

	proxy.ServeHTTP(rr, req)

	proxy.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusServiceUnavailable, rr.Code)

}	assert.Equal(t, http.StatusServiceUnavailable, rr.Code)

}

func TestProxyHandler_InvalidTargetURL(t *testing.T) {

	config := &ProxyConfig{func TestProxyHandler_InvalidTargetURL(t *testing.T) {

		TeamServiceURL:        "://invalid-url", // Invalid URL	config := &ProxyConfig{

		ApplicationServiceURL: "http://localhost:8082",		TeamServiceURL:        "://invalid-url", // Invalid URL

		Logger:               logger.New("debug", "text"),		ApplicationServiceURL: "http://localhost:8082",

	}		Logger:               logger.New("debug", "text"),

	}

	proxy := NewProxyHandler(config)

	proxy := NewProxyHandler(config)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/teams", nil)

	rr := httptest.NewRecorder()	req := httptest.NewRequest(http.MethodGet, "/api/v1/teams", nil)

	rr := httptest.NewRecorder()

	proxy.ServeHTTP(rr, req)

	proxy.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusInternalServerError, rr.Code)

}	assert.Equal(t, http.StatusInternalServerError, rr.Code)
}
package server

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/aykay76/ai-idp/internal/database"
)

// Config holds server configuration
type Config struct {
	Port            string
	ServiceName     string
	Environment     string
	Debug           bool
	DatabaseURL     string
	RedisURL        string
	JWTSecret       string
	ShutdownTimeout time.Duration
}

// Server wraps the HTTP server with database and utilities
type Server struct {
	config     *Config
	mux        *http.ServeMux
	database   *database.Pool
	server     *http.Server
	middleware []Middleware
}

// Middleware represents HTTP middleware
type Middleware func(http.Handler) http.Handler

// NewServer creates a new server instance
func NewServer(config *Config) *Server {
	mux := http.NewServeMux()

	server := &Server{
		config:     config,
		mux:        mux,
		middleware: make([]Middleware, 0),
	}

	// Add default middleware
	server.Use(LoggingMiddleware)
	server.Use(RecoveryMiddleware)

	// CORS middleware for development
	if config.Environment != "production" {
		server.Use(CORSMiddleware([]string{
			"http://localhost:3000",
			"http://localhost:5173",
			"http://localhost:8080",
		}))
	}

	// Add default health endpoints
	server.HandleFunc("GET /health", server.healthHandler)
	server.HandleFunc("GET /readiness", server.readinessHandler)

	return server
}

// SetupDatabase initializes the database connection
func (s *Server) SetupDatabase(ctx context.Context) error {
	dbConfig := database.DefaultConfig(s.config.DatabaseURL)

	var err error
	s.database, err = database.NewPool(ctx, dbConfig)
	if err != nil {
		return fmt.Errorf("failed to setup database: %w", err)
	}

	// Update readiness handler to include database health
	s.HandleFunc("GET /readiness", s.readinessWithDBHandler)

	return nil
}

// GetDB returns the database pool
func (s *Server) GetDB() *database.Pool {
	return s.database
}

// Use adds middleware to the server
func (s *Server) Use(middleware Middleware) {
	s.middleware = append(s.middleware, middleware)
}

// HandleFunc registers a handler function for the given pattern
func (s *Server) HandleFunc(pattern string, handler http.HandlerFunc) {
	s.mux.HandleFunc(pattern, handler)
}

// Handle registers a handler for the given pattern
func (s *Server) Handle(pattern string, handler http.Handler) {
	s.mux.Handle(pattern, handler)
}

// Start starts the HTTP server
func (s *Server) Start() error {
	// Apply all middleware to the mux
	var finalHandler http.Handler = s.mux
	for i := len(s.middleware) - 1; i >= 0; i-- {
		finalHandler = s.middleware[i](finalHandler)
	}

	s.server = &http.Server{
		Addr:         ":" + s.config.Port,
		Handler:      finalHandler,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	fmt.Printf("🚀 %s starting on port %s (env: %s)\n",
		s.config.ServiceName, s.config.Port, s.config.Environment)

	return s.server.ListenAndServe()
}

// Stop gracefully stops the server
func (s *Server) Stop(ctx context.Context) error {
	fmt.Println("🛑 Shutting down server...")

	// Close database connections
	if s.database != nil {
		s.database.Close()
	}

	// Shutdown HTTP server
	if s.server != nil {
		return s.server.Shutdown(ctx)
	}

	return nil
}

// Run starts the server and handles graceful shutdown
func (s *Server) Run(ctx context.Context) error {
	// Start server in a goroutine
	errChan := make(chan error, 1)
	go func() {
		if err := s.Start(); err != nil && err != http.ErrServerClosed {
			errChan <- fmt.Errorf("server failed to start: %w", err)
		}
	}()

	// Wait for interrupt signal or server error
	signalChan := make(chan os.Signal, 1)
	signal.Notify(signalChan, syscall.SIGINT, syscall.SIGTERM)

	select {
	case err := <-errChan:
		return err
	case sig := <-signalChan:
		fmt.Printf("📡 Received signal: %v\n", sig)
	case <-ctx.Done():
		fmt.Println("📡 Context cancelled")
	}

	// Graceful shutdown with timeout
	shutdownCtx, cancel := context.WithTimeout(context.Background(), s.config.ShutdownTimeout)
	defer cancel()

	return s.Stop(shutdownCtx)
}

// Handler functions

// healthHandler returns basic health status
func (s *Server) healthHandler(w http.ResponseWriter, r *http.Request) {
	RespondWithJSON(w, http.StatusOK, map[string]interface{}{
		"status":  "healthy",
		"service": s.config.ServiceName,
		"time":    time.Now().UTC().Format(time.RFC3339),
	})
}

// readinessHandler returns readiness status (before database setup)
func (s *Server) readinessHandler(w http.ResponseWriter, r *http.Request) {
	RespondWithJSON(w, http.StatusOK, map[string]interface{}{
		"status": "ready",
		"time":   time.Now().UTC().Format(time.RFC3339),
	})
}

// readinessWithDBHandler returns readiness status including database
func (s *Server) readinessWithDBHandler(w http.ResponseWriter, r *http.Request) {
	// Check database health
	ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
	defer cancel()

	if s.database == nil {
		RespondWithJSON(w, http.StatusServiceUnavailable, map[string]interface{}{
			"status": "not ready",
			"reason": "database not initialized",
			"time":   time.Now().UTC().Format(time.RFC3339),
		})
		return
	}

	if err := s.database.HealthCheck(ctx); err != nil {
		RespondWithJSON(w, http.StatusServiceUnavailable, map[string]interface{}{
			"status": "not ready",
			"reason": fmt.Sprintf("database unhealthy: %v", err),
			"time":   time.Now().UTC().Format(time.RFC3339),
		})
		return
	}

	// Get database stats for additional info
	stats := s.database.Stats()

	RespondWithJSON(w, http.StatusOK, map[string]interface{}{
		"status": "ready",
		"database": map[string]interface{}{
			"status":      "healthy",
			"connections": stats.TotalConnections,
			"idle":        stats.IdleConnections,
			"used":        stats.UsedConnections,
		},
		"time": time.Now().UTC().Format(time.RFC3339),
	})
}

// Middleware implementations

// LoggingMiddleware logs HTTP requests
func LoggingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()

		// Create a custom response writer to capture status code
		rw := &responseWriter{ResponseWriter: w, statusCode: http.StatusOK}

		next.ServeHTTP(rw, r)

		duration := time.Since(start)
		log.Printf("[%s] %s %s %d %v", r.Method, r.URL.Path, r.RemoteAddr, rw.statusCode, duration)
	})
}

// RecoveryMiddleware recovers from panics
func RecoveryMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if err := recover(); err != nil {
				log.Printf("Panic recovered: %v", err)
				RespondWithError(w, http.StatusInternalServerError,
					fmt.Errorf("internal server error"), "Internal server error")
			}
		}()

		next.ServeHTTP(w, r)
	})
}

// CORSMiddleware handles CORS for development
func CORSMiddleware(allowedOrigins []string) Middleware {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			origin := r.Header.Get("Origin")

			allowed := false
			for _, allowedOrigin := range allowedOrigins {
				if origin == allowedOrigin {
					allowed = true
					break
				}
			}

			if allowed {
				w.Header().Set("Access-Control-Allow-Origin", origin)
				w.Header().Set("Access-Control-Allow-Credentials", "true")
				w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization, accept, origin, Cache-Control, X-Requested-With, X-Tenant-ID, X-User-Email")
				w.Header().Set("Access-Control-Allow-Methods", "POST, GET, OPTIONS, PUT, DELETE, PATCH")
			}

			if r.Method == "OPTIONS" {
				w.WriteHeader(http.StatusNoContent)
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}

// Response utilities

// ErrorResponse represents a standard error response
type ErrorResponse struct {
	Error   string                 `json:"error"`
	Message string                 `json:"message"`
	Code    int                    `json:"code"`
	Details map[string]interface{} `json:"details,omitempty"`
	Time    string                 `json:"time"`
}

// SuccessResponse represents a standard success response
type SuccessResponse struct {
	Data    interface{} `json:"data,omitempty"`
	Message string      `json:"message,omitempty"`
	Time    string      `json:"time"`
}

// RespondWithJSON writes a JSON response
func RespondWithJSON(w http.ResponseWriter, code int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)

	if err := json.NewEncoder(w).Encode(data); err != nil {
		log.Printf("Error encoding JSON response: %v", err)
	}
}

// RespondWithError sends an error response
func RespondWithError(w http.ResponseWriter, code int, err error, message string) {
	response := &ErrorResponse{
		Error:   err.Error(),
		Message: message,
		Code:    code,
		Time:    time.Now().UTC().Format(time.RFC3339),
	}
	RespondWithJSON(w, code, response)
}

// RespondWithValidationError sends a validation error response
func RespondWithValidationError(w http.ResponseWriter, err error) {
	response := &ErrorResponse{
		Error:   err.Error(),
		Message: "Validation failed",
		Code:    http.StatusBadRequest,
		Time:    time.Now().UTC().Format(time.RFC3339),
	}
	RespondWithJSON(w, http.StatusBadRequest, response)
}

// RespondWithNotFound sends a not found response
func RespondWithNotFound(w http.ResponseWriter, resource string) {
	err := fmt.Errorf("%s not found", resource)
	response := &ErrorResponse{
		Error:   err.Error(),
		Message: "Resource not found",
		Code:    http.StatusNotFound,
		Time:    time.Now().UTC().Format(time.RFC3339),
	}
	RespondWithJSON(w, http.StatusNotFound, response)
}

// RespondWithInternalError sends an internal server error response
func RespondWithInternalError(w http.ResponseWriter, err error) {
	response := &ErrorResponse{
		Error:   err.Error(),
		Message: "Internal server error",
		Code:    http.StatusInternalServerError,
		Time:    time.Now().UTC().Format(time.RFC3339),
	}
	RespondWithJSON(w, http.StatusInternalServerError, response)
}

// RespondWithData sends a success response with data
func RespondWithData(w http.ResponseWriter, data interface{}) {
	response := &SuccessResponse{
		Data: data,
		Time: time.Now().UTC().Format(time.RFC3339),
	}
	RespondWithJSON(w, http.StatusOK, response)
}

// RespondWithMessage sends a success response with a message
func RespondWithMessage(w http.ResponseWriter, message string) {
	response := &SuccessResponse{
		Message: message,
		Time:    time.Now().UTC().Format(time.RFC3339),
	}
	RespondWithJSON(w, http.StatusOK, response)
}

// RespondCreated sends a 201 Created response with data
func RespondCreated(w http.ResponseWriter, data interface{}) {
	response := &SuccessResponse{
		Data: data,
		Time: time.Now().UTC().Format(time.RFC3339),
	}
	RespondWithJSON(w, http.StatusCreated, response)
}

// responseWriter wraps http.ResponseWriter to capture status code
type responseWriter struct {
	http.ResponseWriter
	statusCode int
}

func (rw *responseWriter) WriteHeader(code int) {
	rw.statusCode = code
	rw.ResponseWriter.WriteHeader(code)
}

// Helper functions for request parsing

// ParseJSONBody parses JSON request body into the provided interface
func ParseJSONBody(r *http.Request, v interface{}) error {
	if r.Body == nil {
		return fmt.Errorf("empty request body")
	}
	defer r.Body.Close()

	if err := json.NewDecoder(r.Body).Decode(v); err != nil {
		return fmt.Errorf("invalid JSON: %w", err)
	}

	return nil
}

// GetHeaderValue gets a header value with validation
func GetHeaderValue(r *http.Request, headerName string, required bool) (string, error) {
	value := r.Header.Get(headerName)
	if required && value == "" {
		return "", fmt.Errorf("missing required header: %s", headerName)
	}
	return value, nil
}

// ExtractPathSegment extracts a path segment by position (0-based from the end)
func ExtractPathSegment(path string, position int) string {
	segments := strings.Split(strings.Trim(path, "/"), "/")
	if position >= len(segments) {
		return ""
	}
	return segments[len(segments)-1-position]
}

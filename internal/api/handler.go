package api

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"cyto-viewer/internal/config"
	"cyto-viewer/internal/scanner"
	"cyto-viewer/internal/tiler"
	"cyto-viewer/pkg/auth"
	"cyto-viewer/pkg/logger"

	"github.com/gorilla/mux"
)

type Handler struct {
	log         *logger.Logger
	tiler       *tiler.GPUTileProcessor
	scanner     *scanner.Interface
	auth        *auth.Manager
	config      *config.Config
}

func NewHandler(log *logger.Logger, tiler *tiler.GPUTileProcessor, 
                scanner *scanner.Interface, auth *auth.Manager, 
                cfg *config.Config) *Handler {
	return &Handler{
		log:     log,
		tiler:   tiler,
		scanner: scanner,
		auth:    auth,
		config:  cfg,
	}
}

func (h *Handler) RegisterRoutes(r *mux.Router) {
	api := r.PathPrefix("/api").Subrouter()

	// Authentication
	api.HandleFunc("/login", h.handleLogin).Methods("POST")
	api.HandleFunc("/logout", h.handleLogout).Methods("POST")

	// Protected routes
	protected := api.PathPrefix("").Subrouter()
	protected.Use(h.authMiddleware)

	// Tile serving - the most critical endpoint
	protected.HandleFunc("/tiles/{slideId}", h.handleGetTile).Methods("GET")
	protected.HandleFunc("/tiles/{slideId}/batch", h.handleBatchTiles).Methods("POST")

	// Slide management
	protected.HandleFunc("/slides", h.handleListSlides).Methods("GET")
	protected.HandleFunc("/slides/{slideId}", h.handleGetSlide).Methods("GET")
	protected.HandleFunc("/slides/{slideId}", h.handleDeleteSlide).Methods("DELETE")

	// Scanner control
	protected.HandleFunc("/scanner/status", h.handleScannerStatus).Methods("GET")
	protected.HandleFunc("/scanner/scan", h.handleStartScan).Methods("POST")
	protected.HandleFunc("/scanner/layers", h.handleGetLayers).Methods("GET")

	// System info
	protected.HandleFunc("/system/stats", h.handleSystemStats).Methods("GET")
}

func (h *Handler) handleGetTile(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	slideId := vars["slideId"]

	// Parse query parameters
	layer, _ := strconv.Atoi(r.URL.Query().Get("layer"))
	x, _ := strconv.Atoi(r.URL.Query().Get("x"))
	y, _ := strconv.Atoi(r.URL.Query().Get("y"))
	z, _ := strconv.Atoi(r.URL.Query().Get("z"))
	format := r.URL.Query().Get("format")
	if format == "" {
		format = "webp" // Default to WebP for better compression
	}
	quality, _ := strconv.Atoi(r.URL.Query().Get("quality"))
	if quality == 0 {
		quality = 85
	}

	// Process tile request
	req := &tiler.TileRequest{
		SlideID:  slideId,
		Layer:    layer,
		X:        x,
		Y:        y,
		Z:        z,
		Width:    512,
		Height:   512,
		Format:   format,
		Quality:  quality,
	}

	start := time.Now()
	resp, err := h.tiler.ProcessTile(r.Context(), req)
	if err != nil {
		h.log.Error("Failed to process tile", "error", err)
		http.Error(w, "Failed to process tile", http.StatusInternalServerError)
		return
	}

	// Set aggressive caching headers
	w.Header().Set("Content-Type", resp.ContentType)
	w.Header().Set("Cache-Control", "public, max-age=31536000, immutable")
	w.Header().Set("ETag", fmt.Sprintf(`"%s"`, resp.CacheKey))
	w.Header().Set("X-Processing-Time", fmt.Sprintf("%dms", time.Since(start).Milliseconds()))

	// Check ETag
	if match := r.Header.Get("If-None-Match"); match != "" {
		if match == fmt.Sprintf(`"%s"`, resp.CacheKey) {
			w.WriteHeader(http.StatusNotModified)
			return
		}
	}

	w.Write(resp.Data)
}

func (h *Handler) handleBatchTiles(w http.ResponseWriter, r *http.Request) {
	var requests []*tiler.TileRequest
	if err := json.NewDecoder(r.Body).Decode(&requests); err != nil {
		http.Error(w, "Invalid request", http.StatusBadRequest)
		return
	}

	// Limit batch size to prevent abuse
	if len(requests) > 100 {
		http.Error(w, "Batch size too large (max 100)", http.StatusBadRequest)
		return
	}

	responses, err := h.tiler.ProcessBatch(r.Context(), requests)
	if err != nil {
		h.log.Error("Batch processing failed", "error", err)
		http.Error(w, "Batch processing failed", http.StatusInternalServerError)
		return
	}

	// Return as multipart response or JSON array
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(responses)
}

func (h *Handler) handleListSlides(w http.ResponseWriter, r *http.Request) {
	// TODO: Implement slide listing from storage
	slides := []map[string]interface{}{
		{
			"id":        "slide-001",
			"name":      "Sample Cytology Slide",
			"created":   time.Now().Add(-24 * time.Hour),
			"width":     102400,
			"height":    102400,
			"layers":    40,
		},
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(slides)
}

func (h *Handler) handleGetSlide(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	slideId := vars["slideId"]

	// TODO: Fetch slide metadata from storage
	slide := map[string]interface{}{
		"id":        slideId,
		"name":      "Sample Cytology Slide",
		"created":   time.Now().Add(-24 * time.Hour),
		"width":     102400,
		"height":    102400,
		"layers":    40,
		"tileSize":  512,
		"format":    "webp",
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(slide)
}

func (h *Handler) handleDeleteSlide(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	slideId := vars["slideId"]

	// TODO: Implement slide deletion
	h.log.Info("Slide deleted", "slideId", slideId)

	w.WriteHeader(http.StatusNoContent)
}

func (h *Handler) handleScannerStatus(w http.ResponseWriter, r *http.Request) {
	status, err := h.scanner.GetStatus()
	if err != nil {
		http.Error(w, "Failed to get scanner status", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(status)
}

func (h *Handler) handleStartScan(w http.ResponseWriter, r *http.Request) {
	var req scanner.ScanRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request", http.StatusBadRequest)
		return
	}

	result, err := h.scanner.StartScan(r.Context(), &req)
	if err != nil {
		http.Error(w, fmt.Sprintf("Scan failed: %v", err), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(result)
}

func (h *Handler) handleGetLayers(w http.ResponseWriter, r *http.Request) {
	layers := h.scanner.GetLayerInfo()

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(layers)
}

func (h *Handler) handleSystemStats(w http.ResponseWriter, r *http.Request) {
	hits, misses, cacheSize, tileCount := h.tiler.(*tiler.GPUTileProcessor).TileCache.Stats()

	stats := map[string]interface{}{
		"cache": map[string]interface{}{
			"hits":    hits,
			"misses":  misses,
			"hitRate": float64(hits) / float64(hits+misses),
			"size":    cacheSize,
			"tiles":   tileCount,
		},
		"uptime": time.Since(h.config.StartTime).String(),
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(stats)
}

func (h *Handler) handleLogin(w http.ResponseWriter, r *http.Request) {
	var credentials struct {
		Username string `json:"username"`
		Password string `json:"password"`
	}

	if err := json.NewDecoder(r.Body).Decode(&credentials); err != nil {
		http.Error(w, "Invalid request", http.StatusBadRequest)
		return
	}

	token, err := h.auth.Authenticate(credentials.Username, credentials.Password)
	if err != nil {
		http.Error(w, "Invalid credentials", http.StatusUnauthorized)
		return
	}

	// Set secure cookie
	http.SetCookie(w, &http.Cookie{
		Name:     "auth_token",
		Value:    token,
		Path:     "/",
		HttpOnly: true,
		Secure:   true,
		SameSite: http.SameSiteStrictMode,
		MaxAge:   86400, // 24 hours
	})

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{
		"token": token,
	})
}

func (h *Handler) handleLogout(w http.ResponseWriter, r *http.Request) {
	http.SetCookie(w, &http.Cookie{
		Name:     "auth_token",
		Value:    "",
		Path:     "/",
		HttpOnly: true,
		MaxAge:   -1,
	})

	w.WriteHeader(http.StatusNoContent)
}

func (h *Handler) authMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		cookie, err := r.Cookie("auth_token")
		if err != nil {
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}

		if !h.auth.ValidateToken(cookie.Value) {
			http.Error(w, "Invalid token", http.StatusUnauthorized)
			return
		}

		next.ServeHTTP(w, r)
	})
}

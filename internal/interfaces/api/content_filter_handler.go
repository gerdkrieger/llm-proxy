// Package api provides content filter HTTP handlers.
package api

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
	"github.com/llm-proxy/llm-proxy/internal/application/filtering"
	"github.com/llm-proxy/llm-proxy/internal/domain/models"
	"github.com/llm-proxy/llm-proxy/internal/infrastructure/database/repositories"
	"github.com/llm-proxy/llm-proxy/pkg/logger"
)

// ContentFilterHandler handles content filter management requests
type ContentFilterHandler struct {
	repo          *repositories.ContentFilterRepository
	filterService *filtering.Service
	logger        *logger.Logger
}

// NewContentFilterHandler creates a new content filter handler
func NewContentFilterHandler(
	repo *repositories.ContentFilterRepository,
	filterService *filtering.Service,
	log *logger.Logger,
) *ContentFilterHandler {
	return &ContentFilterHandler{
		repo:          repo,
		filterService: filterService,
		logger:        log,
	}
}

// ============================================================================
// CRUD OPERATIONS
// ============================================================================

// CreateFilter creates a new content filter
// POST /admin/filters
func (h *ContentFilterHandler) CreateFilter(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	var req models.CreateContentFilterRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.respondError(w, http.StatusBadRequest, "invalid_request", "Invalid request body: "+err.Error())
		return
	}

	// Validate required fields
	if req.Pattern == "" {
		h.respondError(w, http.StatusBadRequest, "invalid_request", "pattern is required")
		return
	}
	if req.Replacement == "" {
		h.respondError(w, http.StatusBadRequest, "invalid_request", "replacement is required")
		return
	}

	// Set defaults
	if req.FilterType == "" {
		req.FilterType = "word"
	}

	// Validate filter type
	validTypes := map[string]bool{
		"word":   true,
		"phrase": true,
		"regex":  true,
	}
	if !validTypes[req.FilterType] {
		h.respondError(w, http.StatusBadRequest, "invalid_request", "filter_type must be 'word', 'phrase', or 'regex'")
		return
	}

	// Validate pattern
	if err := h.filterService.ValidateFilterPattern(req.FilterType, req.Pattern); err != nil {
		h.respondError(w, http.StatusBadRequest, "invalid_pattern", "Invalid pattern: "+err.Error())
		return
	}

	// Create filter model
	filter := &models.ContentFilter{
		Pattern:       req.Pattern,
		Replacement:   req.Replacement,
		Description:   req.Description,
		FilterType:    req.FilterType,
		CaseSensitive: req.CaseSensitive,
		Enabled:       req.Enabled,
		Priority:      req.Priority,
	}

	// Save to database
	if err := h.repo.Create(ctx, filter); err != nil {
		h.logger.Errorf(err, "Failed to create filter")
		h.respondError(w, http.StatusInternalServerError, "internal_error", "Failed to create filter")
		return
	}

	h.logger.Infof("Created content filter: %d (%s)", filter.ID, filter.Pattern)

	// Refresh filter cache
	go h.filterService.RefreshFilters(ctx)

	h.respondJSON(w, http.StatusCreated, filter)
}

// ListFilters lists all content filters
// GET /admin/filters
func (h *ContentFilterHandler) ListFilters(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	// Parse query parameters
	enabledOnlyParam := r.URL.Query().Get("enabled_only")
	enabledOnly := enabledOnlyParam == "true"

	// Get filters from database
	filters, err := h.repo.List(ctx, enabledOnly)
	if err != nil {
		h.logger.Errorf(err, "Failed to list filters")
		h.respondError(w, http.StatusInternalServerError, "internal_error", "Failed to list filters")
		return
	}

	h.respondJSON(w, http.StatusOK, map[string]interface{}{
		"filters": filters,
		"count":   len(filters),
	})
}

// GetFilter retrieves a single content filter
// GET /admin/filters/{id}
func (h *ContentFilterHandler) GetFilter(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	// Parse filter ID
	idStr := chi.URLParam(r, "id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		h.respondError(w, http.StatusBadRequest, "invalid_request", "Invalid filter ID")
		return
	}

	// Get filter from database
	filter, err := h.repo.GetByID(ctx, id)
	if err != nil {
		h.logger.Errorf(err, "Failed to get filter %d", id)
		h.respondError(w, http.StatusNotFound, "not_found", "Filter not found")
		return
	}

	h.respondJSON(w, http.StatusOK, filter)
}

// UpdateFilter updates an existing content filter
// PUT /admin/filters/{id}
func (h *ContentFilterHandler) UpdateFilter(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	// Parse filter ID
	idStr := chi.URLParam(r, "id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		h.respondError(w, http.StatusBadRequest, "invalid_request", "Invalid filter ID")
		return
	}

	// Parse request body
	var req models.UpdateContentFilterRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.respondError(w, http.StatusBadRequest, "invalid_request", "Invalid request body: "+err.Error())
		return
	}

	// Validate pattern if provided
	if req.Pattern != nil && req.FilterType != nil {
		if err := h.filterService.ValidateFilterPattern(*req.FilterType, *req.Pattern); err != nil {
			h.respondError(w, http.StatusBadRequest, "invalid_pattern", "Invalid pattern: "+err.Error())
			return
		}
	}

	// Update in database
	if err := h.repo.Update(ctx, id, &req); err != nil {
		h.logger.Errorf(err, "Failed to update filter %d", id)
		h.respondError(w, http.StatusInternalServerError, "internal_error", "Failed to update filter")
		return
	}

	h.logger.Infof("Updated content filter: %d", id)

	// Refresh filter cache
	go h.filterService.RefreshFilters(ctx)

	// Get updated filter
	filter, err := h.repo.GetByID(ctx, id)
	if err != nil {
		h.logger.Errorf(err, "Failed to get updated filter %d", id)
		h.respondError(w, http.StatusInternalServerError, "internal_error", "Failed to retrieve updated filter")
		return
	}

	h.respondJSON(w, http.StatusOK, filter)
}

// DeleteFilter deletes a content filter
// DELETE /admin/filters/{id}
func (h *ContentFilterHandler) DeleteFilter(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	// Parse filter ID
	idStr := chi.URLParam(r, "id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		h.respondError(w, http.StatusBadRequest, "invalid_request", "Invalid filter ID")
		return
	}

	// Delete from database
	if err := h.repo.Delete(ctx, id); err != nil {
		h.logger.Errorf(err, "Failed to delete filter %d", id)
		h.respondError(w, http.StatusInternalServerError, "internal_error", "Failed to delete filter")
		return
	}

	h.logger.Infof("Deleted content filter: %d", id)

	// Refresh filter cache
	go h.filterService.RefreshFilters(ctx)

	h.respondJSON(w, http.StatusOK, map[string]interface{}{
		"message": "Filter deleted successfully",
		"id":      id,
	})
}

// ============================================================================
// TESTING & UTILITIES
// ============================================================================

// TestFilterRequest represents a filter test request
type TestFilterRequest struct {
	Text          string `json:"text" binding:"required"`
	Pattern       string `json:"pattern" binding:"required"`
	Replacement   string `json:"replacement" binding:"required"`
	FilterType    string `json:"filter_type"`
	CaseSensitive bool   `json:"case_sensitive"`
}

// TestFilter tests a filter against sample text
// POST /admin/filters/{id}/test
// POST /admin/filters/test (without ID, uses request body)
func (h *ContentFilterHandler) TestFilter(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	// Check if testing existing filter or ad-hoc filter
	idStr := chi.URLParam(r, "id")

	if idStr != "" {
		// Test existing filter
		id, err := strconv.Atoi(idStr)
		if err != nil {
			h.respondError(w, http.StatusBadRequest, "invalid_request", "Invalid filter ID")
			return
		}

		// Get filter from database
		filter, err := h.repo.GetByID(ctx, id)
		if err != nil {
			h.respondError(w, http.StatusNotFound, "not_found", "Filter not found")
			return
		}

		// Parse test text
		var req struct {
			Text string `json:"text"`
		}
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			h.respondError(w, http.StatusBadRequest, "invalid_request", "Invalid request body")
			return
		}

		if req.Text == "" {
			h.respondError(w, http.StatusBadRequest, "invalid_request", "text is required")
			return
		}

		// Test filter
		result, err := h.filterService.TestFilter(filter, req.Text)
		if err != nil {
			h.logger.Errorf(err, "Failed to test filter %d", id)
			h.respondError(w, http.StatusInternalServerError, "internal_error", "Failed to test filter")
			return
		}

		h.respondJSON(w, http.StatusOK, result)
	} else {
		// Test ad-hoc filter
		var req TestFilterRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			h.respondError(w, http.StatusBadRequest, "invalid_request", "Invalid request body: "+err.Error())
			return
		}

		// Validate required fields
		if req.Text == "" || req.Pattern == "" || req.Replacement == "" {
			h.respondError(w, http.StatusBadRequest, "invalid_request", "text, pattern, and replacement are required")
			return
		}

		// Set default filter type
		if req.FilterType == "" {
			req.FilterType = "word"
		}

		// Create temporary filter
		filter := &models.ContentFilter{
			Pattern:       req.Pattern,
			Replacement:   req.Replacement,
			FilterType:    req.FilterType,
			CaseSensitive: req.CaseSensitive,
		}

		// Test filter
		result, err := h.filterService.TestFilter(filter, req.Text)
		if err != nil {
			h.logger.Errorf(err, "Failed to test filter")
			h.respondError(w, http.StatusBadRequest, "invalid_pattern", "Invalid pattern: "+err.Error())
			return
		}

		h.respondJSON(w, http.StatusOK, result)
	}
}

// GetFilterStats returns statistics about filter usage
// GET /admin/filters/stats
func (h *ContentFilterHandler) GetFilterStats(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	stats, err := h.filterService.GetFilterStats(ctx)
	if err != nil {
		h.logger.Errorf(err, "Failed to get filter stats")
		h.respondError(w, http.StatusInternalServerError, "internal_error", "Failed to get stats")
		return
	}

	h.respondJSON(w, http.StatusOK, stats)
}

// RefreshFilters forces a refresh of the filter cache
// POST /admin/filters/refresh
func (h *ContentFilterHandler) RefreshFilters(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	if err := h.filterService.RefreshFilters(ctx); err != nil {
		h.logger.Errorf(err, "Failed to refresh filters")
		h.respondError(w, http.StatusInternalServerError, "internal_error", "Failed to refresh filters")
		return
	}

	count := h.filterService.GetCachedFilterCount()
	h.logger.Infof("Refreshed filter cache, loaded %d filters", count)

	h.respondJSON(w, http.StatusOK, map[string]interface{}{
		"message":        "Filters refreshed successfully",
		"cached_filters": count,
	})
}

// BulkImportFilters imports multiple filters at once
// POST /admin/filters/bulk-import
func (h *ContentFilterHandler) BulkImportFilters(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	// Parse request body
	var req struct {
		Filters []models.CreateContentFilterRequest `json:"filters"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.respondError(w, http.StatusBadRequest, "invalid_request", "Invalid request body: "+err.Error())
		return
	}

	if len(req.Filters) == 0 {
		h.respondError(w, http.StatusBadRequest, "invalid_request", "No filters provided")
		return
	}

	results := struct {
		Success []int    `json:"success"`
		Failed  []string `json:"failed"`
		Total   int      `json:"total"`
	}{
		Success: []int{},
		Failed:  []string{},
		Total:   len(req.Filters),
	}

	// Process each filter
	for i, filterReq := range req.Filters {
		// Set defaults
		if filterReq.FilterType == "" {
			filterReq.FilterType = "word"
		}

		// Validate filter type
		if filterReq.FilterType != "word" && filterReq.FilterType != "phrase" && filterReq.FilterType != "regex" {
			results.Failed = append(results.Failed, fmt.Sprintf("Filter %d: invalid filter_type", i+1))
			continue
		}

		// Validate pattern
		if err := h.filterService.ValidateFilterPattern(filterReq.FilterType, filterReq.Pattern); err != nil {
			results.Failed = append(results.Failed, fmt.Sprintf("Filter %d (%s): %v", i+1, filterReq.Pattern, err))
			continue
		}

		// Create filter
		filter := &models.ContentFilter{
			Pattern:       filterReq.Pattern,
			Replacement:   filterReq.Replacement,
			Description:   filterReq.Description,
			FilterType:    filterReq.FilterType,
			CaseSensitive: filterReq.CaseSensitive,
			Enabled:       filterReq.Enabled,
			Priority:      filterReq.Priority,
		}

		if err := h.repo.Create(ctx, filter); err != nil {
			results.Failed = append(results.Failed, fmt.Sprintf("Filter %d (%s): %v", i+1, filterReq.Pattern, err))
			continue
		}

		results.Success = append(results.Success, filter.ID)
	}

	h.logger.Infof("Bulk import: %d succeeded, %d failed", len(results.Success), len(results.Failed))

	// Refresh filter cache
	go h.filterService.RefreshFilters(ctx)

	h.respondJSON(w, http.StatusOK, results)
}

// ============================================================================
// HELPER METHODS
// ============================================================================

func (h *ContentFilterHandler) respondJSON(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if err := json.NewEncoder(w).Encode(data); err != nil {
		h.logger.Errorf(err, "Failed to encode JSON response")
	}
}

func (h *ContentFilterHandler) respondError(w http.ResponseWriter, status int, errorType, message string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"error": map[string]interface{}{
			"type":    errorType,
			"message": message,
		},
	})
}

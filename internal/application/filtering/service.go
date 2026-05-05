// Package filtering provides content filtering services for user prompts.
package filtering

import (
	"context"
	"fmt"
	"regexp"
	"strings"
	"sync"
	"time"

	"github.com/llm-proxy/llm-proxy/internal/domain/models"
	"github.com/llm-proxy/llm-proxy/internal/infrastructure/database/repositories"
	"github.com/llm-proxy/llm-proxy/pkg/logger"
)

const (
	// FilterTypWord matches whole words
	FilterTypeWord = "word"
	// FilterTypePhrase matches exact phrases
	FilterTypePhrase = "phrase"
	// FilterTypeRegex matches using regular expressions
	FilterTypeRegex = "regex"

	// DefaultCacheTTL is the default cache duration for filters
	DefaultCacheTTL = 5 * time.Minute
)

// FilterMatch represents a single filter match
type FilterMatch struct {
	FilterID    int    `json:"filter_id"`
	Pattern     string `json:"pattern"`
	Replacement string `json:"replacement"`
	MatchCount  int    `json:"match_count"`
}

// FilterResult contains the filtered text and match information
type FilterResult struct {
	OriginalText string        `json:"original_text"`
	FilteredText string        `json:"filtered_text"`
	Matches      []FilterMatch `json:"matches"`
	HasMatches   bool          `json:"has_matches"`
}

// compiledFilter represents a filter with pre-compiled regex
type compiledFilter struct {
	filter *models.ContentFilter
	regex  *regexp.Regexp
}

// Service handles content filtering operations
type Service struct {
	repo   *repositories.ContentFilterRepository
	logger *logger.Logger

	// Cache
	mu            sync.RWMutex
	cachedFilters []*compiledFilter
	cacheTime     time.Time
	cacheTTL      time.Duration
	loading       sync.Mutex // Prevents concurrent DB loads (cache stampede/avalanche)
}

// NewService creates a new filtering service
func NewService(
	repo *repositories.ContentFilterRepository,
	log *logger.Logger,
) *Service {
	s := &Service{
		repo:     repo,
		logger:   log,
		cacheTTL: DefaultCacheTTL,
	}

	// Load filters on initialization
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := s.loadFilters(ctx); err != nil {
		log.Warnf("Failed to load initial filters: %v", err)
	}

	return s
}

// ApplyFilters applies content filters to all messages in the request
func (s *Service) ApplyFilters(ctx context.Context, messages []models.OpenAIMessage) ([]models.OpenAIMessage, []FilterMatch, error) {
	// Get active filters
	filters, err := s.getActiveFilters(ctx)
	if err != nil {
		return messages, nil, fmt.Errorf("failed to get filters: %w", err)
	}

	if len(filters) == 0 {
		return messages, nil, nil
	}

	// Track all matches across all messages
	allMatches := make(map[int]*FilterMatch) // filterID -> match

	// Process each message
	filteredMessages := make([]models.OpenAIMessage, len(messages))
	for i, msg := range messages {
		filteredMessages[i] = msg

		// Only filter user messages (not system or assistant)
		if msg.Role != "user" {
			continue
		}

		// Extract content as string
		var content string
		switch v := msg.Content.(type) {
		case string:
			content = v
		default:
			// Skip non-string content (e.g., multi-modal messages)
			continue
		}

		// Apply filters to message content
		result := s.applyFiltersToText(content, filters)

		if result.HasMatches {
			filteredMessages[i].Content = result.FilteredText

			// Aggregate matches
			for _, match := range result.Matches {
				if existing, ok := allMatches[match.FilterID]; ok {
					existing.MatchCount += match.MatchCount
				} else {
					allMatches[match.FilterID] = &FilterMatch{
						FilterID:    match.FilterID,
						Pattern:     match.Pattern,
						Replacement: match.Replacement,
						MatchCount:  match.MatchCount,
					}
				}
			}
		}
	}

	// Convert matches map to slice
	matches := make([]FilterMatch, 0, len(allMatches))
	for _, match := range allMatches {
		matches = append(matches, *match)

		// Record match asynchronously (batch increment by count)
		go func(filterID int, count int) {
			ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
			defer cancel()

			if err := s.repo.IncrementMatchCountBy(ctx, filterID, count); err != nil {
				s.logger.Warnf("Failed to record match for filter %d: %v", filterID, err)
			}
		}(match.FilterID, match.MatchCount)
	}

	return filteredMessages, matches, nil
}

// applyFiltersToText applies all filters to a single text string
func (s *Service) applyFiltersToText(text string, filters []*compiledFilter) FilterResult {
	result := FilterResult{
		OriginalText: text,
		FilteredText: text,
		Matches:      []FilterMatch{},
		HasMatches:   false,
	}

	// Apply each filter in priority order
	for _, cf := range filters {
		if cf.regex == nil {
			continue
		}

		// Find all matches
		matches := cf.regex.FindAllString(result.FilteredText, -1)
		if len(matches) == 0 {
			continue
		}

		// Replace all matches
		result.FilteredText = cf.regex.ReplaceAllString(result.FilteredText, cf.filter.Replacement)
		result.HasMatches = true

		// Record match
		result.Matches = append(result.Matches, FilterMatch{
			FilterID:    cf.filter.ID,
			Pattern:     cf.filter.Pattern,
			Replacement: cf.filter.Replacement,
			MatchCount:  len(matches),
		})
	}

	return result
}

// getActiveFilters returns cached filters or loads from database.
// Uses a lock to prevent cache stampede (multiple goroutines reloading simultaneously).
func (s *Service) getActiveFilters(ctx context.Context) ([]*compiledFilter, error) {
	s.mu.RLock()
	if time.Since(s.cacheTime) < s.cacheTTL && s.cachedFilters != nil {
		filters := s.cachedFilters
		s.mu.RUnlock()
		return filters, nil
	}
	s.mu.RUnlock()

	// Only one goroutine should reload the cache
	s.loading.Lock()
	defer s.loading.Unlock()

	// Double-check after acquiring lock
	s.mu.RLock()
	if time.Since(s.cacheTime) < s.cacheTTL && s.cachedFilters != nil {
		filters := s.cachedFilters
		s.mu.RUnlock()
		return filters, nil
	}
	s.mu.RUnlock()

	// Cache expired or not loaded, reload
	if err := s.loadFilters(ctx); err != nil {
		return nil, err
	}

	s.mu.RLock()
	filters := s.cachedFilters
	s.mu.RUnlock()

	return filters, nil
}

// loadFilters loads and compiles filters from database
func (s *Service) loadFilters(ctx context.Context) error {
	// Load enabled filters from database (sorted by priority DESC)
	filters, err := s.repo.List(ctx, true)
	if err != nil {
		return fmt.Errorf("failed to load filters: %w", err)
	}

	// Compile regex patterns
	compiled := make([]*compiledFilter, 0, len(filters))
	for _, filter := range filters {
		regex, err := s.compileFilter(filter)
		if err != nil {
			s.logger.Warnf("Failed to compile filter %d (%s): %v", filter.ID, filter.Pattern, err)
			continue
		}

		compiled = append(compiled, &compiledFilter{
			filter: filter,
			regex:  regex,
		})
	}

	// Update cache
	s.mu.Lock()
	s.cachedFilters = compiled
	s.cacheTime = time.Now()
	s.mu.Unlock()

	s.logger.Infof("Loaded %d content filters", len(compiled))
	return nil
}

// compileFilter compiles a filter pattern into a regex
func (s *Service) compileFilter(filter *models.ContentFilter) (*regexp.Regexp, error) {
	var pattern string

	switch filter.FilterType {
	case FilterTypeWord:
		// Word boundary matching for whole words
		pattern = fmt.Sprintf(`\b%s\b`, regexp.QuoteMeta(filter.Pattern))

	case FilterTypePhrase:
		// Exact phrase matching (escaped)
		pattern = regexp.QuoteMeta(filter.Pattern)

	case FilterTypeRegex:
		// Direct regex pattern
		pattern = filter.Pattern

	default:
		return nil, fmt.Errorf("unknown filter type: %s", filter.FilterType)
	}

	// Add case-insensitive flag if needed
	if !filter.CaseSensitive {
		pattern = "(?i)" + pattern
	}

	// Compile regex
	regex, err := regexp.Compile(pattern)
	if err != nil {
		return nil, fmt.Errorf("invalid regex pattern: %w", err)
	}

	return regex, nil
}

// RefreshFilters forces a reload of filters from database
func (s *Service) RefreshFilters(ctx context.Context) error {
	s.logger.Info("Refreshing content filters cache")
	return s.loadFilters(ctx)
}

// GetCachedFilterCount returns the number of cached filters
func (s *Service) GetCachedFilterCount() int {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return len(s.cachedFilters)
}

// TestFilter tests a filter pattern against sample text
func (s *Service) TestFilter(filter *models.ContentFilter, text string) (FilterResult, error) {
	// Compile the filter
	regex, err := s.compileFilter(filter)
	if err != nil {
		return FilterResult{}, fmt.Errorf("failed to compile filter: %w", err)
	}

	cf := &compiledFilter{
		filter: filter,
		regex:  regex,
	}

	// Apply filter to text
	result := s.applyFiltersToText(text, []*compiledFilter{cf})
	return result, nil
}

// ValidateFilterPattern validates a filter pattern
func (s *Service) ValidateFilterPattern(filterType, pattern string) error {
	// Create temporary filter for validation
	tempFilter := &models.ContentFilter{
		Pattern:       pattern,
		FilterType:    filterType,
		CaseSensitive: false,
	}

	// Try to compile
	_, err := s.compileFilter(tempFilter)
	if err != nil {
		return fmt.Errorf("invalid pattern: %w", err)
	}

	// Additional validation for regex patterns
	if filterType == FilterTypeRegex {
		patLower := strings.ToLower(pattern)

		// Heuristic 1: Catastrophic backtracking patterns
		dangerous := []string{
			".*.*",
			".+.+",
			"(.*)+",
			"(.+)+",
		}
		for _, d := range dangerous {
			if strings.Contains(patLower, d) {
				return fmt.Errorf("potentially dangerous regex pattern: catastrophic backtracking detected")
			}
		}

		// Heuristic 2: Nested quantifiers (a+)+ or (a*)*
		nestedRe := regexp.MustCompile(`\([^{}]*[+?*]\)[+?*]`)
		if nestedRe.MatchString(patLower) {
			return fmt.Errorf("potentially dangerous regex pattern: nested quantifiers detected")
		}

		// Heuristic 3: Excessive repetition like {1000,} or {0,99999}
		excessiveRe := regexp.MustCompile(`\{\d{4,},?\}|\{\d+,\d{3,}\}`)
		if excessiveRe.MatchString(patLower) {
			return fmt.Errorf("potentially dangerous regex pattern: excessive repetition detected")
		}

		// Heuristic 4: Alternation with overlapping patterns (a|a)*
		overlapRe := regexp.MustCompile(`\([^|]+\|[^)]+\)[*+]`)
		if overlapRe.MatchString(patLower) {
			return fmt.Errorf("potentially dangerous regex pattern: overlapping alternation detected")
		}
	}

	return nil
}

// GetFilterStats returns statistics about filter usage
func (s *Service) GetFilterStats(ctx context.Context) (map[string]interface{}, error) {
	// Get all filters (including disabled)
	allFilters, err := s.repo.List(ctx, false)
	if err != nil {
		return nil, fmt.Errorf("failed to get filters: %w", err)
	}

	// Calculate statistics
	totalFilters := len(allFilters)
	enabledCount := 0
	totalMatches := 0
	var lastMatch time.Time

	byType := make(map[string]int)

	for _, filter := range allFilters {
		if filter.Enabled {
			enabledCount++
		}

		totalMatches += filter.MatchCount
		byType[filter.FilterType]++

		if filter.LastMatchedAt != nil && filter.LastMatchedAt.After(lastMatch) {
			lastMatch = *filter.LastMatchedAt
		}
	}

	stats := map[string]interface{}{
		"total_filters":     totalFilters,
		"enabled_filters":   enabledCount,
		"cached_filters":    s.GetCachedFilterCount(),
		"total_matches":     totalMatches,
		"by_type":           byType,
		"cache_age_seconds": int(time.Since(s.cacheTime).Seconds()),
	}

	if !lastMatch.IsZero() {
		stats["last_match"] = lastMatch
	}

	return stats, nil
}

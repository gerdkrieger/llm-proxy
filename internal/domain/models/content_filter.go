package models

import "time"

// ContentFilter represents a content filtering rule
type ContentFilter struct {
	ID            int        `json:"id" db:"id"`
	Pattern       string     `json:"pattern" db:"pattern"`
	Replacement   string     `json:"replacement" db:"replacement"`
	Description   string     `json:"description" db:"description"`
	FilterType    string     `json:"filter_type" db:"filter_type"` // word, phrase, regex
	CaseSensitive bool       `json:"case_sensitive" db:"case_sensitive"`
	Enabled       bool       `json:"enabled" db:"enabled"`
	Priority      int        `json:"priority" db:"priority"`
	CreatedAt     time.Time  `json:"created_at" db:"created_at"`
	UpdatedAt     time.Time  `json:"updated_at" db:"updated_at"`
	CreatedBy     *string    `json:"created_by,omitempty" db:"created_by"`
	MatchCount    int        `json:"match_count" db:"match_count"`
	LastMatchedAt *time.Time `json:"last_matched_at,omitempty" db:"last_matched_at"`
}

// CreateContentFilterRequest represents the request to create a filter
type CreateContentFilterRequest struct {
	Pattern       string `json:"pattern" binding:"required"`
	Replacement   string `json:"replacement" binding:"required"`
	Description   string `json:"description"`
	FilterType    string `json:"filter_type"` // defaults to "word"
	CaseSensitive bool   `json:"case_sensitive"`
	Enabled       bool   `json:"enabled"`
	Priority      int    `json:"priority"`
}

// UpdateContentFilterRequest represents the request to update a filter
type UpdateContentFilterRequest struct {
	Pattern       *string `json:"pattern"`
	Replacement   *string `json:"replacement"`
	Description   *string `json:"description"`
	FilterType    *string `json:"filter_type"`
	CaseSensitive *bool   `json:"case_sensitive"`
	Enabled       *bool   `json:"enabled"`
	Priority      *int    `json:"priority"`
}

// Package attachment provides attachment content analysis and filtering.
package attachment

import (
	"context"
	"encoding/base64"
	"fmt"
	"strings"

	"github.com/llm-proxy/llm-proxy/internal/application/filtering"
	"github.com/llm-proxy/llm-proxy/internal/domain/models"
	"github.com/llm-proxy/llm-proxy/pkg/logger"
)

// Service handles attachment content analysis
type Service struct {
	filterService    *filtering.Service
	redactionService *RedactionService
	logger           *logger.Logger
}

// NewService creates a new attachment service
func NewService(filterService *filtering.Service, log *logger.Logger) *Service {
	redactionService := NewRedactionService(filterService, log)

	return &Service{
		filterService:    filterService,
		redactionService: redactionService,
		logger:           log,
	}
}

// AttachmentAnalysisResult contains the result of attachment analysis
type AttachmentAnalysisResult struct {
	HasAttachments     bool
	TotalAttachments   int
	BlockedAttachments int
	FilterMatches      []filtering.FilterMatch
	CleanedMessages    []models.OpenAIMessage
}

// AnalyzeAttachments analyzes all attachments in messages for sensitive content
func (s *Service) AnalyzeAttachments(ctx context.Context, messages []models.OpenAIMessage) (*AttachmentAnalysisResult, error) {
	result := &AttachmentAnalysisResult{
		CleanedMessages: make([]models.OpenAIMessage, 0, len(messages)),
	}

	for _, msg := range messages {
		cleanedMsg := msg
		msgHasAttachment := false
		msgBlocked := false

		// Check if message has content array (multimodal format)
		if msg.Content != nil {
			switch content := msg.Content.(type) {
			case []interface{}:
				// Multimodal message with potential attachments
				cleanedContent := make([]interface{}, 0)

				for _, part := range content {
					if partMap, ok := part.(map[string]interface{}); ok {
						partType, _ := partMap["type"].(string)

						switch partType {
						case "text":
							// Text content - no attachment
							cleanedContent = append(cleanedContent, part)

						case "image_url", "image":
							// Image attachment
							msgHasAttachment = true
							result.HasAttachments = true
							result.TotalAttachments++

							// Extract image data if available
							if imageURL, ok := partMap["image_url"].(map[string]interface{}); ok {
								if url, ok := imageURL["url"].(string); ok {
									// Check if it's base64 encoded
									if strings.HasPrefix(url, "data:image/") {
										blocked, matches, redactedURL := s.analyzeImageData(ctx, url)
										if blocked {
											msgBlocked = true
											result.BlockedAttachments++
											result.FilterMatches = append(result.FilterMatches, matches...)
											s.logger.Warnf("Blocked image attachment due to sensitive content detection")
											continue
										}

										// If redacted, replace with redacted version
										if redactedURL != "" {
											s.logger.Infof("Image redacted, replacing with redacted version")
											imageURL["url"] = redactedURL
											partMap["image_url"] = imageURL
											result.FilterMatches = append(result.FilterMatches, matches...)
										}
									}
								}
							}
							cleanedContent = append(cleanedContent, part)

						case "file", "document":
							// Document attachment
							msgHasAttachment = true
							result.HasAttachments = true
							result.TotalAttachments++

							// Try to extract and analyze text content
							blocked, matches := s.analyzeDocumentPart(ctx, partMap)
							if blocked {
								msgBlocked = true
								result.BlockedAttachments++
								result.FilterMatches = append(result.FilterMatches, matches...)
								s.logger.Warnf("Blocked document attachment due to sensitive content detection")
								continue
							}
							cleanedContent = append(cleanedContent, part)

						default:
							// Unknown type - pass through
							cleanedContent = append(cleanedContent, part)
						}
					} else {
						cleanedContent = append(cleanedContent, part)
					}
				}

				cleanedMsg.Content = cleanedContent
			}
		}

		// Only add message if not completely blocked
		if !msgBlocked || !msgHasAttachment {
			result.CleanedMessages = append(result.CleanedMessages, cleanedMsg)
		}
	}

	return result, nil
}

// analyzeImageData analyzes image data for sensitive content and redacts if needed
func (s *Service) analyzeImageData(ctx context.Context, dataURL string) (bool, []filtering.FilterMatch, string) {
	// Extract base64 data and decode
	var imageData []byte
	var filename string
	var mimeType string

	// Parse data URL: data:image/png;base64,iVBORw0KG...
	if strings.HasPrefix(dataURL, "data:") {
		parts := strings.Split(dataURL, ",")
		if len(parts) >= 2 {
			// Extract mime type and filename
			header := parts[0]
			if strings.Contains(header, ";name=") {
				nameParts := strings.Split(header, ";name=")
				if len(nameParts) > 1 {
					filename = strings.Split(nameParts[1], ";")[0]
				}
			} else {
				filename = "image.png" // Default filename
			}

			// Extract mime type
			if strings.Contains(header, "image/") {
				mimeStart := strings.Index(header, "image/")
				mimeEnd := strings.Index(header[mimeStart:], ";")
				if mimeEnd > 0 {
					mimeType = header[mimeStart : mimeStart+mimeEnd]
				} else {
					mimeType = strings.TrimPrefix(header, "data:")
				}
			}

			// Decode base64
			decoded, err := base64.StdEncoding.DecodeString(parts[1])
			if err == nil {
				imageData = decoded
			}
		}
	}

	// Check filename for sensitive patterns first
	if filename != "" {
		blocked, matches := s.checkStringForSensitiveContent(ctx, filename)
		if blocked {
			return true, matches, ""
		}
	}

	// If we have image data, try OCR-based redaction
	if len(imageData) > 0 && s.redactionService != nil {
		result, err := s.redactionService.RedactDocument(ctx, imageData, filename)
		if err != nil {
			s.logger.Warnf("Redaction failed for image: %v, falling back to original", err)
			return false, nil, ""
		}

		if result.Success && result.TotalRedactions > 0 {
			// PII found and redacted - return redacted image
			s.logger.Infof("Redacted %d PII locations in image", result.TotalRedactions)

			// Convert PIIMatches to FilterMatches for logging
			matches := make([]filtering.FilterMatch, len(result.PIIMatches))
			for i, pii := range result.PIIMatches {
				matches[i] = filtering.FilterMatch{
					FilterID:    0, // 0 indicates attachment redaction (not from content_filters)
					Pattern:     pii.Type,
					Replacement: pii.Replacement,
					MatchCount:  1,
				}
			}

			// Reconstruct data URL with redacted image
			redactedDataURL := fmt.Sprintf("data:%s;base64,%s", mimeType, result.Base64Content)
			return false, matches, redactedDataURL
		}
	}

	return false, nil, ""
}

// analyzeDocumentPart analyzes document content and redacts if needed
func (s *Service) analyzeDocumentPart(ctx context.Context, partMap map[string]interface{}) (bool, []filtering.FilterMatch) {
	filename := ""
	if fn, ok := partMap["filename"].(string); ok {
		filename = fn
	}

	// Extract filename and check for sensitive patterns
	if filename != "" {
		blocked, matches := s.checkStringForSensitiveContent(ctx, filename)
		if blocked {
			return true, matches
		}
	}

	// Extract text content if available
	if text, ok := partMap["text"].(string); ok {
		blocked, matches := s.checkStringForSensitiveContent(ctx, text)
		if blocked {
			return true, matches
		}
	}

	// Extract base64 content and try redaction
	if data, ok := partMap["data"].(string); ok {
		decoded, err := base64.StdEncoding.DecodeString(data)
		if err == nil {
			// Try OCR-based redaction
			if s.redactionService != nil && filename != "" {
				result, err := s.redactionService.RedactDocument(ctx, decoded, filename)
				if err != nil {
					s.logger.Warnf("Redaction failed for document %s: %v", filename, err)
					// Fall back to text extraction
					text := string(decoded)
					blocked, matches := s.checkStringForSensitiveContent(ctx, text)
					if blocked {
						return true, matches
					}
				} else if result.Success && result.TotalRedactions > 0 {
					// PII found and redacted - replace data with redacted version
					s.logger.Infof("Redacted %d PII locations in document %s", result.TotalRedactions, filename)
					partMap["data"] = result.Base64Content

					// Convert PIIMatches to FilterMatches for logging
					matches := make([]filtering.FilterMatch, len(result.PIIMatches))
					for i, pii := range result.PIIMatches {
						matches[i] = filtering.FilterMatch{
							FilterID:    0, // 0 indicates attachment redaction (not from content_filters)
							Pattern:     pii.Type,
							Replacement: pii.Replacement,
							MatchCount:  1,
						}
					}

					// Don't block, but return matches for logging
					return false, matches
				}
			} else {
				// No redaction service, fall back to text extraction
				text := string(decoded)
				blocked, matches := s.checkStringForSensitiveContent(ctx, text)
				if blocked {
					return true, matches
				}
			}
		}
	}

	return false, nil
}

// checkStringForSensitiveContent checks a string for sensitive content using filters
func (s *Service) checkStringForSensitiveContent(ctx context.Context, content string) (bool, []filtering.FilterMatch) {
	if s.filterService == nil {
		return false, nil
	}

	// Create a temporary message to use with filter service
	messages := []models.OpenAIMessage{
		{
			Role:    "user",
			Content: content,
		},
	}

	filteredMessages, matches, err := s.filterService.ApplyFilters(ctx, messages)
	if err != nil {
		s.logger.Warnf("Failed to apply filters to attachment content: %v", err)
		return false, nil
	}

	// If filters matched, check if content was modified (indicating sensitive data)
	if len(matches) > 0 {
		return true, matches
	}

	// Also check if content was significantly modified
	if len(filteredMessages) > 0 {
		if filteredContent, ok := filteredMessages[0].Content.(string); ok {
			if filteredContent != content {
				return true, matches
			}
		}
	}

	return false, matches
}

// GetAttachmentStats returns statistics about attachments in messages
func (s *Service) GetAttachmentStats(messages []models.OpenAIMessage) map[string]interface{} {
	stats := map[string]interface{}{
		"total_messages":            len(messages),
		"messages_with_attachments": 0,
		"total_attachments":         0,
		"image_attachments":         0,
		"document_attachments":      0,
	}

	for _, msg := range messages {
		if msg.Content != nil {
			if content, ok := msg.Content.([]interface{}); ok {
				hasAttachment := false
				for _, part := range content {
					if partMap, ok := part.(map[string]interface{}); ok {
						partType, _ := partMap["type"].(string)
						switch partType {
						case "image_url", "image":
							stats["image_attachments"] = stats["image_attachments"].(int) + 1
							stats["total_attachments"] = stats["total_attachments"].(int) + 1
							hasAttachment = true
						case "file", "document":
							stats["document_attachments"] = stats["document_attachments"].(int) + 1
							stats["total_attachments"] = stats["total_attachments"].(int) + 1
							hasAttachment = true
						}
					}
				}
				if hasAttachment {
					stats["messages_with_attachments"] = stats["messages_with_attachments"].(int) + 1
				}
			}
		}
	}

	return stats
}

// ExtractTextFromPDF extracts text from PDF data (placeholder for future implementation)
func (s *Service) ExtractTextFromPDF(pdfData []byte) (string, error) {
	// TODO: Implement PDF text extraction using a library like pdftotext or unidoc
	return "", fmt.Errorf("PDF text extraction not yet implemented")
}

// ExtractTextFromDOCX extracts text from DOCX data (placeholder for future implementation)
func (s *Service) ExtractTextFromDOCX(docxData []byte) (string, error) {
	// TODO: Implement DOCX text extraction
	return "", fmt.Errorf("DOCX text extraction not yet implemented")
}

// PerformOCR performs OCR on image data (placeholder for future implementation)
func (s *Service) PerformOCR(imageData []byte) (string, error) {
	// TODO: Implement OCR using tesseract or cloud OCR service
	return "", fmt.Errorf("OCR not yet implemented")
}

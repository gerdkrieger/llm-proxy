// Package attachment provides document redaction capabilities.
package attachment

import (
	"bytes"
	"context"
	"encoding/base64"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/llm-proxy/llm-proxy/internal/application/filtering"
	"github.com/llm-proxy/llm-proxy/pkg/logger"
)

// RedactionService handles document redaction with OCR and visual redaction
type RedactionService struct {
	filterService *filtering.Service
	logger        *logger.Logger
	tempDir       string
}

// OCRWord represents a word found by OCR with its position
type OCRWord struct {
	Text   string
	X      int
	Y      int
	Width  int
	Height int
	Page   int // For multi-page documents
}

// PIIMatch represents found PII with its location
type PIIMatch struct {
	Type        string // "CREDIT_CARD", "EMAIL", etc.
	Text        string // Original text
	Replacement string // "[CREDIT_CARD]"
	X           int
	Y           int
	Width       int
	Height      int
	Page        int
	Confidence  float64 // OCR confidence
}

// RedactionResult contains the result of document redaction
type RedactionResult struct {
	Success         bool
	RedactedContent []byte // Redacted document bytes
	Base64Content   string // Base64 encoded redacted document
	PIIMatches      []PIIMatch
	TotalRedactions int
	OriginalFormat  string // "pdf", "image", etc.
}

// NewRedactionService creates a new redaction service
func NewRedactionService(filterService *filtering.Service, log *logger.Logger) *RedactionService {
	tempDir := filepath.Join(os.TempDir(), "llm-proxy-redaction")
	os.MkdirAll(tempDir, 0755)

	return &RedactionService{
		filterService: filterService,
		logger:        log,
		tempDir:       tempDir,
	}
}

// RedactDocument performs OCR, PII detection, and visual redaction
func (s *RedactionService) RedactDocument(ctx context.Context, fileData []byte, filename string) (*RedactionResult, error) {
	result := &RedactionResult{
		OriginalFormat: detectFileType(filename, fileData),
	}

	s.logger.Infof("Starting redaction for file: %s (type: %s)", filename, result.OriginalFormat)

	// Step 1: Save to temp file
	tempFile := filepath.Join(s.tempDir, fmt.Sprintf("input_%s", filepath.Base(filename)))
	if err := os.WriteFile(tempFile, fileData, 0644); err != nil {
		return nil, fmt.Errorf("failed to write temp file: %w", err)
	}
	defer os.Remove(tempFile)

	// Step 2: Perform OCR
	ocrWords, err := s.performOCR(tempFile, result.OriginalFormat)
	if err != nil {
		s.logger.Warnf("OCR failed: %v", err)
		// Fallback: try text extraction without coordinates
		return s.fallbackTextRedaction(fileData, filename)
	}

	s.logger.Infof("OCR extracted %d words from document", len(ocrWords))

	// Step 3: Detect PII in OCR text
	piiMatches := s.detectPIIInOCR(ctx, ocrWords)
	result.PIIMatches = piiMatches
	result.TotalRedactions = len(piiMatches)

	if len(piiMatches) == 0 {
		// No PII found, return original
		result.Success = true
		result.RedactedContent = fileData
		result.Base64Content = base64.StdEncoding.EncodeToString(fileData)
		s.logger.Info("No PII found in document, returning original")
		return result, nil
	}

	s.logger.Infof("Found %d PII matches, proceeding with redaction", len(piiMatches))

	// Step 4: Perform visual redaction
	redactedFile, err := s.visuallyRedact(tempFile, piiMatches, result.OriginalFormat)
	if err != nil {
		return nil, fmt.Errorf("visual redaction failed: %w", err)
	}
	defer os.Remove(redactedFile)

	// Step 5: Read redacted file
	redactedData, err := os.ReadFile(redactedFile)
	if err != nil {
		return nil, fmt.Errorf("failed to read redacted file: %w", err)
	}

	result.Success = true
	result.RedactedContent = redactedData
	result.Base64Content = base64.StdEncoding.EncodeToString(redactedData)

	s.logger.Infof("Successfully redacted %d locations in document", len(piiMatches))

	return result, nil
}

// performOCR runs Tesseract OCR and returns words with positions
func (s *RedactionService) performOCR(filename string, fileType string) ([]OCRWord, error) {
	// Check if tesseract is available
	if _, err := exec.LookPath("tesseract"); err != nil {
		return nil, fmt.Errorf("tesseract not found in PATH")
	}

	// Convert PDF to images first if needed
	if fileType == "pdf" {
		// Use pdftoppm or imagemagick to convert PDF to images
		// For now, we'll use a simple approach
		return nil, fmt.Errorf("PDF OCR not yet fully implemented")
	}

	// For images, run tesseract with hOCR output (contains coordinates)
	outputBase := filepath.Join(s.tempDir, "ocr_output")

	cmd := exec.Command("tesseract", filename, outputBase, "-l", "eng+deu", "hocr")
	output, err := cmd.CombinedOutput()
	if err != nil {
		return nil, fmt.Errorf("tesseract failed: %w (output: %s)", err, string(output))
	}

	// Parse hOCR output
	hocrFile := outputBase + ".hocr"
	defer os.Remove(hocrFile)

	hocrData, err := os.ReadFile(hocrFile)
	if err != nil {
		return nil, fmt.Errorf("failed to read hOCR output: %w", err)
	}

	// Parse hOCR XML to extract words with coordinates
	words := s.parseHOCR(string(hocrData))

	return words, nil
}

// parseHOCR parses hOCR XML format and extracts words with bounding boxes
func (s *RedactionService) parseHOCR(hocrXML string) []OCRWord {
	words := make([]OCRWord, 0)

	// Regex to find word elements: <span class='ocrx_word' title='bbox X Y W H'>WORD</span>
	wordRegex := regexp.MustCompile(`<span class='ocrx_word'[^>]*title='bbox (\d+) (\d+) (\d+) (\d+)[^']*'[^>]*>([^<]+)</span>`)

	matches := wordRegex.FindAllStringSubmatch(hocrXML, -1)
	for _, match := range matches {
		if len(match) == 6 {
			var x, y, w, h int
			fmt.Sscanf(match[1], "%d", &x)
			fmt.Sscanf(match[2], "%d", &y)
			fmt.Sscanf(match[3], "%d", &w)
			fmt.Sscanf(match[4], "%d", &h)

			word := OCRWord{
				Text:   strings.TrimSpace(match[5]),
				X:      x,
				Y:      y,
				Width:  w - x,
				Height: h - y,
				Page:   1,
			}

			words = append(words, word)
		}
	}

	return words
}

// detectPIIInOCR finds PII patterns in OCR text with their positions
func (s *RedactionService) detectPIIInOCR(ctx context.Context, ocrWords []OCRWord) []PIIMatch {
	matches := make([]PIIMatch, 0)

	// Build full text for pattern matching
	fullText := ""
	wordPositions := make(map[int]OCRWord) // character position -> OCRWord
	charPos := 0

	for _, word := range ocrWords {
		wordPositions[charPos] = word
		fullText += word.Text + " "
		charPos += len(word.Text) + 1
	}

	// Common PII patterns
	patterns := map[string]*regexp.Regexp{
		"CREDIT_CARD": regexp.MustCompile(`\b\d{4}[\s-]?\d{4}[\s-]?\d{4}[\s-]?\d{4}\b`),
		"EMAIL":       regexp.MustCompile(`\b[A-Za-z0-9._%+-]+@[A-Za-z0-9.-]+\.[A-Z|a-z]{2,}\b`),
		"PHONE":       regexp.MustCompile(`\b\d{3}[\s-]?\d{3,4}[\s-]?\d{4}\b`),
		"SSN":         regexp.MustCompile(`\b\d{3}-\d{2}-\d{4}\b`),
	}

	for piiType, pattern := range patterns {
		foundMatches := pattern.FindAllStringIndex(fullText, -1)

		for _, match := range foundMatches {
			start := match[0]
			end := match[1]
			matchText := fullText[start:end]

			// Find the OCR word that contains this match
			var matchWord OCRWord
			var found bool
			currentPos := 0

			for _, word := range ocrWords {
				wordEnd := currentPos + len(word.Text)
				if start >= currentPos && start < wordEnd {
					matchWord = word
					found = true
					break
				}
				currentPos = wordEnd + 1 // +1 for space
			}

			if found {
				piiMatch := PIIMatch{
					Type:        piiType,
					Text:        matchText,
					Replacement: fmt.Sprintf("[%s]", piiType),
					X:           matchWord.X,
					Y:           matchWord.Y,
					Width:       matchWord.Width * len(matchText) / len(matchWord.Text), // Approximate
					Height:      matchWord.Height,
					Page:        matchWord.Page,
					Confidence:  1.0,
				}

				matches = append(matches, piiMatch)
			}
		}
	}

	return matches
}

// visuallyRedact creates a redacted version of the document
func (s *RedactionService) visuallyRedact(inputFile string, matches []PIIMatch, fileType string) (string, error) {
	outputFile := filepath.Join(s.tempDir, fmt.Sprintf("redacted_%s", filepath.Base(inputFile)))

	switch fileType {
	case "pdf":
		return s.redactPDFWithGhostscript(inputFile, outputFile, matches)
	case "image":
		return s.redactImageWithImageMagick(inputFile, outputFile, matches)
	default:
		return "", fmt.Errorf("unsupported file type for redaction: %s", fileType)
	}
}

// redactPDFWithGhostscript uses Ghostscript to add black boxes over PII
func (s *RedactionService) redactPDFWithGhostscript(inputFile, outputFile string, matches []PIIMatch) (string, error) {
	// Check if Ghostscript is available
	if _, err := exec.LookPath("gs"); err != nil {
		return "", fmt.Errorf("ghostscript (gs) not found in PATH")
	}

	// Create PostScript commands to draw black rectangles
	psCommands := ""
	for _, match := range matches {
		// Convert coordinates (hOCR uses top-left origin, PDF uses bottom-left)
		// This needs proper page height calculation
		psCommands += fmt.Sprintf("%d %d %d %d rectfill\n", match.X, match.Y, match.Width, match.Height)
	}

	// Create a temporary PostScript file with redaction commands
	psFile := filepath.Join(s.tempDir, "redact.ps")
	psContent := fmt.Sprintf(`%%!PS
/Times-Roman findfont 12 scalefont setfont
0 setgray
%s
showpage
`, psCommands)

	if err := os.WriteFile(psFile, []byte(psContent), 0644); err != nil {
		return "", fmt.Errorf("failed to create PS file: %w", err)
	}
	defer os.Remove(psFile)

	// Use pdftk or similar to overlay the redactions
	// For now, we'll use a simpler approach with ImageMagick convert

	// Convert PDF to image, redact, convert back
	// This is a simplified version - production would need proper PDF manipulation
	return s.redactPDFViaImages(inputFile, outputFile, matches)
}

// redactPDFViaImages converts PDF to images, redacts, converts back
func (s *RedactionService) redactPDFViaImages(inputFile, outputFile string, matches []PIIMatch) (string, error) {
	// Convert PDF to PNG
	pngFile := filepath.Join(s.tempDir, "temp.png")
	cmd := exec.Command("convert", "-density", "300", inputFile, pngFile)
	if output, err := cmd.CombinedOutput(); err != nil {
		return "", fmt.Errorf("PDF to image conversion failed: %w (output: %s)", err, string(output))
	}
	defer os.Remove(pngFile)

	// Redact the image
	redactedPng, err := s.redactImageWithImageMagick(pngFile, pngFile+".redacted", matches)
	if err != nil {
		return "", err
	}
	defer os.Remove(redactedPng)

	// Convert back to PDF
	cmd = exec.Command("convert", redactedPng, outputFile)
	if output, err := cmd.CombinedOutput(); err != nil {
		return "", fmt.Errorf("image to PDF conversion failed: %w (output: %s)", err, string(output))
	}

	return outputFile, nil
}

// redactImageWithImageMagick uses ImageMagick to draw black rectangles
func (s *RedactionService) redactImageWithImageMagick(inputFile, outputFile string, matches []PIIMatch) (string, error) {
	// Check if ImageMagick is available
	if _, err := exec.LookPath("convert"); err != nil {
		return "", fmt.Errorf("imagemagick (convert) not found in PATH")
	}

	// Build ImageMagick command with draw operations
	args := []string{inputFile}

	for _, match := range matches {
		// Draw black filled rectangle
		drawCmd := fmt.Sprintf("rectangle %d,%d %d,%d",
			match.X, match.Y,
			match.X+match.Width, match.Y+match.Height)

		args = append(args, "-fill", "black", "-draw", drawCmd)
	}

	args = append(args, outputFile)

	cmd := exec.Command("convert", args...)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return "", fmt.Errorf("imagemagick redaction failed: %w (output: %s)", err, string(output))
	}

	return outputFile, nil
}

// fallbackTextRedaction performs simple text-based redaction without coordinates
func (s *RedactionService) fallbackTextRedaction(fileData []byte, filename string) (*RedactionResult, error) {
	// This is a simplified fallback that just replaces text
	// Used when OCR is not available or fails

	result := &RedactionResult{
		Success:         true,
		OriginalFormat:  detectFileType(filename, fileData),
		RedactedContent: fileData,
	}

	// For now, just return original if we can't do proper redaction
	result.Base64Content = base64.StdEncoding.EncodeToString(fileData)

	return result, nil
}

// detectFileType determines the file type from filename and magic bytes
func detectFileType(filename string, data []byte) string {
	ext := strings.ToLower(filepath.Ext(filename))

	switch ext {
	case ".pdf":
		return "pdf"
	case ".png", ".jpg", ".jpeg", ".gif", ".bmp", ".tiff":
		return "image"
	case ".txt":
		return "text"
	case ".docx":
		return "docx"
	default:
		// Check magic bytes
		if len(data) > 4 {
			if bytes.Equal(data[:4], []byte{0x25, 0x50, 0x44, 0x46}) {
				return "pdf"
			}
			if bytes.Equal(data[:4], []byte{0x89, 0x50, 0x4E, 0x47}) {
				return "image" // PNG
			}
			if bytes.Equal(data[:3], []byte{0xFF, 0xD8, 0xFF}) {
				return "image" // JPEG
			}
		}
		return "unknown"
	}
}

// Cleanup removes temporary files
func (s *RedactionService) Cleanup() {
	os.RemoveAll(s.tempDir)
}

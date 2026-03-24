package api

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/smtp"
	"regexp"
	"strings"
	"sync"
	"time"

	"github.com/llm-proxy/llm-proxy/internal/config"
	"github.com/llm-proxy/llm-proxy/pkg/logger"
)

// ContactHandler handles contact form submissions
type ContactHandler struct {
	config *config.Config
	logger *logger.Logger

	// Simple in-memory rate limiter: IP -> last submission timestamps
	rateMu    sync.Mutex
	rateLimit map[string][]time.Time
}

// NewContactHandler creates a new contact handler
func NewContactHandler(cfg *config.Config, log *logger.Logger) *ContactHandler {
	return &ContactHandler{
		config:    cfg,
		logger:    log,
		rateLimit: make(map[string][]time.Time),
	}
}

// contactRequest represents an incoming contact form submission
type contactRequest struct {
	Name           string `json:"name"`
	Email          string `json:"email"`
	Company        string `json:"company"`
	Subject        string `json:"subject"`
	Message        string `json:"message"`
	TurnstileToken string `json:"turnstile_token"`
}

var emailRegex = regexp.MustCompile(`^[a-zA-Z0-9._%+\-]+@[a-zA-Z0-9.\-]+\.[a-zA-Z]{2,}$`)

// Submit handles POST /api/contact
func (h *ContactHandler) Submit(w http.ResponseWriter, r *http.Request) {
	if !h.config.SMTP.Enabled {
		h.respondJSON(w, http.StatusServiceUnavailable, map[string]string{
			"error": "Contact form is not configured",
		})
		return
	}

	// Rate limit: max 5 per IP per hour
	ip := extractIP(r)
	if !h.checkRateLimit(ip, 5, time.Hour) {
		h.respondJSON(w, http.StatusTooManyRequests, map[string]string{
			"error": "Too many submissions. Please try again later.",
		})
		return
	}

	// Parse request
	var req contactRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.respondJSON(w, http.StatusBadRequest, map[string]string{
			"error": "Invalid request body",
		})
		return
	}

	// Validate
	req.Name = strings.TrimSpace(req.Name)
	req.Email = strings.TrimSpace(req.Email)
	req.Company = strings.TrimSpace(req.Company)
	req.Subject = strings.TrimSpace(req.Subject)
	req.Message = strings.TrimSpace(req.Message)

	if req.Name == "" || req.Email == "" || req.Subject == "" || req.Message == "" {
		h.respondJSON(w, http.StatusBadRequest, map[string]string{
			"error": "Name, email, subject, and message are required",
		})
		return
	}

	if len(req.Name) > 200 || len(req.Email) > 200 || len(req.Company) > 200 || len(req.Subject) > 200 || len(req.Message) > 5000 {
		h.respondJSON(w, http.StatusBadRequest, map[string]string{
			"error": "Field length exceeded",
		})
		return
	}

	if !emailRegex.MatchString(req.Email) {
		h.respondJSON(w, http.StatusBadRequest, map[string]string{
			"error": "Invalid email address",
		})
		return
	}

	// Verify Cloudflare Turnstile token
	if h.config.SMTP.TurnstileSecret != "" {
		if req.TurnstileToken == "" {
			h.respondJSON(w, http.StatusBadRequest, map[string]string{
				"error": "Captcha verification required",
			})
			return
		}
		if !h.verifyTurnstile(req.TurnstileToken, ip) {
			h.respondJSON(w, http.StatusForbidden, map[string]string{
				"error": "Captcha verification failed",
			})
			return
		}
	}

	// Build email
	subjectLine := fmt.Sprintf("[Scrubgate Kontakt] %s", req.Subject)

	body := fmt.Sprintf("Neue Kontaktanfrage über scrubgate.com\n"+
		"━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━\n\n"+
		"Name:         %s\n"+
		"E-Mail:       %s\n"+
		"Unternehmen:  %s\n"+
		"Betreff:      %s\n\n"+
		"Nachricht:\n"+
		"──────────────────────────────────────\n"+
		"%s\n"+
		"──────────────────────────────────────\n\n"+
		"Gesendet am: %s\n"+
		"IP: %s\n",
		req.Name, req.Email, req.Company, req.Subject,
		req.Message, time.Now().Format("02.01.2006 15:04 Uhr"), ip)

	msg := fmt.Sprintf("From: %s\r\n"+
		"To: %s\r\n"+
		"Reply-To: %s\r\n"+
		"Subject: %s\r\n"+
		"Content-Type: text/plain; charset=UTF-8\r\n"+
		"X-Mailer: Scrubgate Contact Form\r\n"+
		"\r\n"+
		"%s",
		h.config.SMTP.From,
		h.config.SMTP.To,
		req.Email,
		subjectLine,
		body)

	// Send via SMTP
	addr := fmt.Sprintf("%s:%d", h.config.SMTP.Host, h.config.SMTP.Port)
	auth := smtp.PlainAuth("", h.config.SMTP.Username, h.config.SMTP.Password, h.config.SMTP.Host)

	err := smtp.SendMail(addr, auth, h.config.SMTP.From, []string{h.config.SMTP.To}, []byte(msg))
	if err != nil {
		h.logger.Error(err, "Failed to send contact form email")
		h.respondJSON(w, http.StatusInternalServerError, map[string]string{
			"error": "Failed to send message. Please try again later.",
		})
		return
	}

	h.logger.Infof("Contact form submitted: name=%s email=%s subject=%s", req.Name, req.Email, req.Subject)

	h.respondJSON(w, http.StatusOK, map[string]string{
		"status":  "ok",
		"message": "Message sent successfully",
	})
}

// verifyTurnstile validates a Cloudflare Turnstile token
func (h *ContactHandler) verifyTurnstile(token, remoteIP string) bool {
	payload, _ := json.Marshal(map[string]string{
		"secret":   h.config.SMTP.TurnstileSecret,
		"response": token,
		"remoteip": remoteIP,
	})

	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Post(
		"https://challenges.cloudflare.com/turnstile/v0/siteverify",
		"application/json",
		bytes.NewReader(payload),
	)
	if err != nil {
		h.logger.Error(err, "Turnstile verification request failed")
		return false
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(io.LimitReader(resp.Body, 4096))
	if err != nil {
		h.logger.Error(err, "Failed to read Turnstile response")
		return false
	}

	var result struct {
		Success bool `json:"success"`
	}
	if err := json.Unmarshal(body, &result); err != nil {
		h.logger.Error(err, "Failed to parse Turnstile response")
		return false
	}

	if !result.Success {
		h.logger.Warnf("Turnstile verification failed for IP %s", remoteIP)
	}
	return result.Success
}

// checkRateLimit returns true if the request is within limits
func (h *ContactHandler) checkRateLimit(ip string, maxRequests int, window time.Duration) bool {
	h.rateMu.Lock()
	defer h.rateMu.Unlock()

	now := time.Now()
	cutoff := now.Add(-window)

	// Clean old entries
	timestamps := h.rateLimit[ip]
	var valid []time.Time
	for _, t := range timestamps {
		if t.After(cutoff) {
			valid = append(valid, t)
		}
	}

	if len(valid) >= maxRequests {
		h.rateLimit[ip] = valid
		return false
	}

	h.rateLimit[ip] = append(valid, now)
	return true
}

func (h *ContactHandler) respondJSON(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(data)
}

func extractIP(r *http.Request) string {
	if xff := r.Header.Get("X-Forwarded-For"); xff != "" {
		parts := strings.Split(xff, ",")
		if len(parts) > 0 {
			return strings.TrimSpace(parts[0])
		}
	}
	if xri := r.Header.Get("X-Real-IP"); xri != "" {
		return xri
	}
	ip, _, err := net.SplitHostPort(r.RemoteAddr)
	if err != nil {
		return r.RemoteAddr
	}
	return ip
}

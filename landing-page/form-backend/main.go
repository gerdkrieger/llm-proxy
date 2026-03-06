package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"net/smtp"
	"os"
	"strings"
	"time"
)

type DemoRequest struct {
	Name    string `json:"name"`
	Email   string `json:"email"`
	Company string `json:"company"`
	Message string `json:"message"`
}

func main() {
	// CORS-enabled handler
	http.HandleFunc("/api/demo-request", corsMiddleware(handleDemoRequest))
	http.HandleFunc("/health", handleHealth)

	port := os.Getenv("PORT")
	if port == "" {
		port = "8081"
	}

	log.Printf("Form backend starting on port %s", port)
	log.Fatal(http.ListenAndServe(":"+port, nil))
}

func corsMiddleware(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "POST, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type")

		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusOK)
			return
		}

		next(w, r)
	}
}

func handleHealth(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"status": "ok"})
}

func handleDemoRequest(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req DemoRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Validate input
	if req.Name == "" || req.Email == "" || req.Company == "" {
		http.Error(w, "Name, email and company are required", http.StatusBadRequest)
		return
	}

	// Basic email validation
	if !strings.Contains(req.Email, "@") {
		http.Error(w, "Invalid email address", http.StatusBadRequest)
		return
	}

	// Send email
	if err := sendEmail(req); err != nil {
		log.Printf("Error sending email: %v", err)
		http.Error(w, "Failed to send email", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{
		"status":  "success",
		"message": "Demo request received. We'll contact you soon!",
	})
}

func sendEmail(req DemoRequest) error {
	// Email configuration from environment variables
	smtpHost := os.Getenv("SMTP_HOST")
	smtpPort := os.Getenv("SMTP_PORT")
	smtpUser := os.Getenv("SMTP_USER")
	smtpPass := os.Getenv("SMTP_PASS")
	toEmail := os.Getenv("TO_EMAIL")

	// Fallback values for testing/development
	if smtpHost == "" {
		smtpHost = "localhost"
	}
	if smtpPort == "" {
		smtpPort = "25"
	}
	if toEmail == "" {
		toEmail = "service@aitrail.ch"
	}

	// Build email body
	subject := fmt.Sprintf("LLM-Proxy Demo Request from %s", req.Company)
	body := fmt.Sprintf(`New Demo Request for LLM-Proxy

Name: %s
Email: %s
Company: %s
Date: %s

Message:
%s

---
Sent from LLM-Proxy Landing Page
https://landing.llmproxy.aitrail.ch/
`, req.Name, req.Email, req.Company, time.Now().Format("2006-01-02 15:04:05"), req.Message)

	// Build email message
	message := []byte(fmt.Sprintf("From: %s\r\n"+
		"To: %s\r\n"+
		"Subject: %s\r\n"+
		"Content-Type: text/plain; charset=UTF-8\r\n"+
		"\r\n"+
		"%s\r\n", smtpUser, toEmail, subject, body))

	// If SMTP credentials are not set, just log (for development)
	if smtpUser == "" || smtpPass == "" {
		log.Printf("SMTP not configured. Would send email to %s:\n%s", toEmail, string(message))
		return nil
	}

	// Send via SMTP
	auth := smtp.PlainAuth("", smtpUser, smtpPass, smtpHost)
	addr := fmt.Sprintf("%s:%s", smtpHost, smtpPort)

	err := smtp.SendMail(addr, auth, smtpUser, []string{toEmail}, message)
	if err != nil {
		return fmt.Errorf("smtp.SendMail failed: %w", err)
	}

	log.Printf("Demo request email sent successfully to %s", toEmail)
	return nil
}

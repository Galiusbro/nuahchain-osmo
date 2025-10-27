package monitoring

import (
	"log"
)

// ConsoleNotifier sends alerts to console
type ConsoleNotifier struct {
	name string
}

// NewConsoleNotifier creates a new console notifier
func NewConsoleNotifier() *ConsoleNotifier {
	return &ConsoleNotifier{
		name: "console",
	}
}

// SendAlert sends an alert to console
func (cn *ConsoleNotifier) SendAlert(alert Alert) error {
	log.Printf("ALERT [%s] %s: %s", alert.Severity, alert.Title, alert.Message)
	if alert.Trader != "" {
		log.Printf("  Trader: %s", alert.Trader)
	}
	if alert.Symbol != "" {
		log.Printf("  Symbol: %s", alert.Symbol)
	}
	if alert.Metadata != nil && len(alert.Metadata) > 0 {
		log.Printf("  Metadata: %+v", alert.Metadata)
	}
	return nil
}

// Name returns the notifier name
func (cn *ConsoleNotifier) Name() string {
	return cn.name
}

// FileNotifier sends alerts to a file
type FileNotifier struct {
	name     string
	filename string
}

// NewFileNotifier creates a new file notifier
func NewFileNotifier(filename string) *FileNotifier {
	return &FileNotifier{
		name:     "file",
		filename: filename,
	}
}

// SendAlert sends an alert to file
func (fn *FileNotifier) SendAlert(alert Alert) error {
	// In a real implementation, would write to file
	// For now, just log
	log.Printf("FILE ALERT [%s] %s: %s -> %s", alert.Severity, alert.Title, alert.Message, fn.filename)
	return nil
}

// Name returns the notifier name
func (fn *FileNotifier) Name() string {
	return fn.name
}

// WebhookNotifier sends alerts via webhook
type WebhookNotifier struct {
	name    string
	url     string
	headers map[string]string
}

// NewWebhookNotifier creates a new webhook notifier
func NewWebhookNotifier(url string, headers map[string]string) *WebhookNotifier {
	return &WebhookNotifier{
		name:    "webhook",
		url:     url,
		headers: headers,
	}
}

// SendAlert sends an alert via webhook
func (wn *WebhookNotifier) SendAlert(alert Alert) error {
	// In a real implementation, would send HTTP POST request
	// For now, just log
	log.Printf("WEBHOOK ALERT [%s] %s: %s -> %s", alert.Severity, alert.Title, alert.Message, wn.url)
	return nil
}

// Name returns the notifier name
func (wn *WebhookNotifier) Name() string {
	return wn.name
}

// EmailNotifier sends alerts via email (placeholder)
type EmailNotifier struct {
	name     string
	smtpHost string
	smtpPort int
	from     string
	to       []string
}

// NewEmailNotifier creates a new email notifier
func NewEmailNotifier(smtpHost string, smtpPort int, from string, to []string) *EmailNotifier {
	return &EmailNotifier{
		name:     "email",
		smtpHost: smtpHost,
		smtpPort: smtpPort,
		from:     from,
		to:       to,
	}
}

// SendAlert sends an alert via email
func (en *EmailNotifier) SendAlert(alert Alert) error {
	// In a real implementation, would send email
	// For now, just log
	log.Printf("EMAIL ALERT [%s] %s: %s -> %s", alert.Severity, alert.Title, alert.Message, en.to)
	return nil
}

// Name returns the notifier name
func (en *EmailNotifier) Name() string {
	return en.name
}

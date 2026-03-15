package email

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/infosec554/clean-archtectura/config"
)

const brevoURL = "https://api.brevo.com/v3/smtp/email"

type Sender struct {
	cfg config.Config
}

func NewSender(cfg config.Config) *Sender {
	return &Sender{cfg: cfg}
}

type brevoRequest struct {
	Sender      brevoContact   `json:"sender"`
	To          []brevoContact `json:"to"`
	Subject     string         `json:"subject"`
	TextContent string         `json:"textContent"`
}

type brevoContact struct {
	Name  string `json:"name,omitempty"`
	Email string `json:"email"`
}

// SendVerificationCode — 2FA login kodi (5 daqiqa)
func (s *Sender) SendVerificationCode(to, code string) error {
	return s.send(to, "Admin Login Code", fmt.Sprintf(
		"Your login verification code is: %s\n\nThis code expires in 5 minutes.",
		code,
	))
}

// SendInviteCode — yangi admin taklif kodi (24 soat)
func (s *Sender) SendInviteCode(to, code string) error {
	return s.send(to, "Admin Panel Invitation", fmt.Sprintf(
		"You have been invited to the admin panel.\n\nYour invite code: %s\n\nThis code expires in 24 hours.",
		code,
	))
}

// send — Brevo API orqali email yuborish
func (s *Sender) send(to, subject, text string) error {
	body := brevoRequest{
		Sender: brevoContact{
			Name:  s.cfg.BrevoSenderName,
			Email: s.cfg.BrevoSenderEmail,
		},
		To:          []brevoContact{{Email: to}},
		Subject:     subject,
		TextContent: text,
	}

	payload, err := json.Marshal(body)
	if err != nil {
		return err
	}

	req, err := http.NewRequest(http.MethodPost, brevoURL, bytes.NewReader(payload))
	if err != nil {
		return err
	}
	req.Header.Set("api-key", s.cfg.BrevoAPIKey)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 300 {
		return fmt.Errorf("brevo: unexpected status %d", resp.StatusCode)
	}

	return nil
}

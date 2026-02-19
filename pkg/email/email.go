package email

import (
	"fmt"
	"net/smtp"

	"github.com/infosec554/clean-archtectura/config"
)

type Sender struct {
	cfg config.Config
}

func NewSender(cfg config.Config) *Sender {
	return &Sender{cfg: cfg}
}

// SendVerificationCode sends a 6-digit OTP code to the given email address.
// Works with MailHog (local) and real SMTP providers.
func (s *Sender) SendVerificationCode(to, code string) error {
	subject := "Email Verification Code"
	body := fmt.Sprintf(
		"Your verification code is: %s\n\nThis code expires in 5 minutes.",
		code,
	)

	msg := []byte(
		"From: " + s.cfg.SMTPFrom + "\r\n" +
			"To: " + to + "\r\n" +
			"Subject: " + subject + "\r\n" +
			"\r\n" +
			body + "\r\n",
	)

	addr := s.cfg.SMTPHost + ":" + s.cfg.SMTPPort

	// MailHog va boshqa auth talab qilmaydigan serverlar uchun
	if s.cfg.SMTPUser == "" {
		return smtp.SendMail(addr, nil, s.cfg.SMTPFrom, []string{to}, msg)
	}

	// Haqiqiy SMTP (Gmail, etc.) uchun auth bilan
	auth := smtp.PlainAuth("", s.cfg.SMTPUser, s.cfg.SMTPPass, s.cfg.SMTPHost)
	return smtp.SendMail(addr, auth, s.cfg.SMTPFrom, []string{to}, msg)
}

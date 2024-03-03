package email

import (
	"context"
	"crypto/tls"

	"github.com/sirupsen/logrus"
	"gopkg.in/gomail.v2"
)

type MailSender struct {
	logger       *logrus.Logger
	dialler      *gomail.Dialer
	emailAddress string
}

type MailSenderConfig struct {
	Password     string `yaml:"email_password" env:"EMAIL_PASSWORD"`
	Port         int    `yaml:"email_port" env:"EMAIL_PORT"`
	Host         string `yaml:"email_host" env:"EMAIL_HOST"`
	EmailAddress string `yaml:"email_address" env:"EMAIL_ADDRESS"`
	EmailLogin   string `yaml:"email_login" env:"EMAIl_LOGIN"`
	EnableTLS    bool   `yaml:"enable_TLS" env:"ENABLE_TLS"`
}

func NewMailSender(cfg MailSenderConfig, logger *logrus.Logger) *MailSender {
	s := MailSender{logger: logger, emailAddress: cfg.EmailAddress}

	s.logger.Infoln("Creating mail dialler.")
	s.dialler = gomail.NewDialer(cfg.Host, cfg.Port, cfg.EmailLogin, cfg.Password)
	s.dialler.TLSConfig = &tls.Config{InsecureSkipVerify: true}
	return &s
}

func (s *MailSender) SendEmail(ctx context.Context, email string, subject string, emailBody, altBody string) error {
	sender, err := s.dialler.Dial()
	if err != nil {
		s.logger.Error(err)
		return err
	}
	defer sender.Close()

	s.logger.Infoln("Creating message.")
	m := gomail.NewMessage()
	m.SetHeader("From", s.emailAddress)
	m.SetHeader("Subject", subject)
	m.AddAlternative("text/plain", altBody)
	m.SetBody("text/html", emailBody)

	s.logger.Infoln("Sending message.")
	if err := sender.Send(s.emailAddress, []string{email}, m); err != nil {
		s.logger.Error(err.Error())
		return nil
	}

	s.logger.Infoln("Message sended.")
	return nil
}

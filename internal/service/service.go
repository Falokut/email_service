package service

import (
	"bytes"
	"context"
	"encoding/base64"
	"fmt"
	"image"
	"image/png"
	"text/template"
	"time"

	"github.com/Falokut/email_service/internal/models"
	"github.com/Falokut/email_service/internal/utils"
	"github.com/boombuler/barcode"
	"github.com/boombuler/barcode/code128"
	"github.com/boombuler/barcode/qr"
	"github.com/k3a/html2text"
)

type TokenTopic int32

const (
	EmailVerificationTopic TokenTopic = iota
	PasswordChangingTopic
)

func (t TokenTopic) MailSubjectType() MailSubjectType {
	switch t {
	case EmailVerificationTopic:
		return EmailVerfication
	case PasswordChangingTopic:
		return PasswordChanging
	}

	return PasswordChanging
}

type MailService interface {
	SendTokenToEmail(ctx context.Context, email, url string, topic TokenTopic, urlTtl time.Duration) error
	SendOrderCreatedNotification(ctx context.Context, email string, order models.Order) error
}

type MailSubjectType string

const (
	EmailVerfication MailSubjectType = "EMAIL_VERIFICATION"
	PasswordChanging MailSubjectType = "CHANGING_PASSWORD"
	OrderCreated     MailSubjectType = "ORDER_CREATED"
)

type MailSender interface {
	SendEmail(ctx context.Context, email string, subject string, emailBody, altBody string) error
}
type ScreeningService interface {
	GetScreeningInfo(ctx context.Context, screeningId int64) (models.Screening, error)
}

type mailService struct {
	mailSender       MailSender
	screeningService ScreeningService
	temp             *template.Template
	Subjects         map[MailSubjectType]string
	TemplatesNames   map[MailSubjectType]string
}

const (
	templatesOrigin = "templates"
)

func NewMailService(
	mailSender MailSender,
	screeningService ScreeningService,
	Subjects map[MailSubjectType]string,
	TemplatesNames map[MailSubjectType]string) (*mailService, error) {
	temp, err := template.ParseGlob(fmt.Sprintf("%s/*.html", templatesOrigin))
	if err != nil {
		return nil, err
	}

	return &mailService{
		mailSender:       mailSender,
		screeningService: screeningService,
		Subjects:         Subjects,
		TemplatesNames:   TemplatesNames,
		temp:             temp,
	}, nil
}
func (s *mailService) SendTokenToEmail(ctx context.Context, email, url string, topic TokenTopic, urlTtl time.Duration) (err error) {

	subject := s.Subjects[topic.MailSubjectType()]
	var body bytes.Buffer

	err = s.temp.ExecuteTemplate(&body, s.TemplatesNames[topic.MailSubjectType()], struct {
		URL string
		TTL string
	}{
		URL: url,
		TTL: utils.ResolveTime(urlTtl.Seconds()),
	})

	if err != nil {
		return
	}

	err = s.mailSender.SendEmail(ctx, email, subject, body.String(), html2text.HTML2Text(body.String()))
	return
}

func GetBarCode(id string) (img image.Image, err error) {
	bc, err := code128.Encode(id)
	if err != nil {
		return nil, err
	}

	scaled, err := barcode.Scale(bc, bc.Bounds().Dx(), 200)
	if err != nil {
		return nil, err
	}

	return scaled, nil
}

func GetQrCode(id string) (img image.Image, err error) {
	bc, err := qr.Encode(id, qr.H, qr.Auto)
	if err != nil {
		return nil, err
	}

	scaled, err := barcode.Scale(bc, 400, 400)
	if err != nil {
		return nil, err
	}

	return scaled, nil
}

func toBase64(img image.Image) string {
	buff := new(bytes.Buffer)
	png.Encode(buff, img)
	return base64.StdEncoding.EncodeToString(buff.Bytes())
}

type orderCreatedNotification struct {
	OrderId   string
	OrderIdQR string
	Screening models.Screening
	Tickets   []models.TicketNotification
}

func (s *mailService) SendOrderCreatedNotification(ctx context.Context,
	email string, order models.Order) (err error) {
	subject := s.Subjects[OrderCreated]

	qrCode, _ := GetQrCode(order.Id)
	var notification orderCreatedNotification = orderCreatedNotification{
		OrderId:   order.Id,
		OrderIdQR: toBase64(qrCode),
	}

	errCh := make(chan error, 1)
	go func() {
		defer close(errCh)
		screening, err := s.screeningService.GetScreeningInfo(ctx, order.ScreeningId)
		if err != nil {
			errCh <- err
			return
		}
		notification.Screening = screening
	}()

	for i := range order.Tickets {
		barcode, _ := GetBarCode(order.Tickets[i].Id)
		notification.Tickets = append(notification.Tickets, models.TicketNotification{
			Id:        order.Tickets[i].Id,
			IdBarCode: toBase64(barcode),
			Row:       order.Tickets[i].Place.Row,
			Seat:      order.Tickets[i].Place.Seat,
			Price:     fmt.Sprintf("%d.%02d", order.Tickets[i].Price/100, order.Tickets[i].Price%100),
		})
	}
	err = <-errCh
	if err != nil {
		return
	}

	var body bytes.Buffer
	err = s.temp.ExecuteTemplate(&body, s.TemplatesNames[OrderCreated], notification)
	if err != nil {
		return
	}

	err = s.mailSender.SendEmail(ctx, email, subject, body.String(), html2text.HTML2Text(body.String()))
	return
}

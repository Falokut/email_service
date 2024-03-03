package events

import (
	"context"
	"encoding/json"
	"errors"
	"time"

	"github.com/Falokut/email_service/internal/models"
	"github.com/Falokut/email_service/internal/service"
	"github.com/segmentio/kafka-go"
	"github.com/sirupsen/logrus"
)

type tokensDeliveryRequests struct {
	reader  *kafka.Reader
	logger  *logrus.Logger
	service service.MailService
}

const (
	emailVerificationTopic = "email_verification_delivery_request"
	passwordChangeTopic    = "password_change_delivery_request"
)

func NewTokensDeliveryRequestsConsumer(
	cfg KafkaReaderConfig,
	logger *logrus.Logger,
	service service.MailService) *tokensDeliveryRequests {
	r := kafka.NewReader(kafka.ReaderConfig{
		Brokers:          cfg.Brokers,
		GroupTopics:      []string{emailVerificationTopic, passwordChangeTopic},
		GroupID:          cfg.GroupID,
		Logger:           logger,
		ReadBatchTimeout: cfg.ReadBatchTimeout,
	})

	return &tokensDeliveryRequests{
		reader:  r,
		logger:  logger,
		service: service,
	}
}

func (c *tokensDeliveryRequests) Run(ctx context.Context) {
	for {
		select {
		default:
			c.Consume(ctx)
		case <-ctx.Done():
			c.logger.Info("tokens delivery consumer shutting down")
			c.reader.Close()
			c.logger.Info("tokens delivery consumer shutted down")
			return
		}
	}
}

func (e *tokensDeliveryRequests) Shutdown() error {
	return e.reader.Close()
}

func (e *tokensDeliveryRequests) handleError(ctx context.Context, err *error) {
	if ctx.Err() != nil {
		var code models.ErrorCode
		switch {
		case errors.Is(ctx.Err(), context.Canceled):
			code = models.Canceled
		case errors.Is(ctx.Err(), context.DeadlineExceeded):
			code = models.DeadlineExceeded
		}
		*err = models.Error(code, ctx.Err().Error())
		return
	}

	if err == nil || *err == nil {
		return
	}

	var serviceErr = &models.ServiceError{}
	if !errors.As(*err, &serviceErr) {
		*err = models.Error(models.Internal, "error while sending event notification")
	}
}

func (e *tokensDeliveryRequests) logError(err error, functionName string) {
	if err == nil {
		return
	}

	var eventsErr = &models.ServiceError{}
	if errors.As(err, &eventsErr) {
		e.logger.WithFields(
			logrus.Fields{
				"error.function.name": functionName,
				"error.msg":           eventsErr.Msg,
				"error.code":          eventsErr.Code,
			},
		).Error("tokens delivery error occurred")
	} else {
		e.logger.WithFields(
			logrus.Fields{
				"error.function.name": functionName,
				"error.msg":           err.Error(),
			},
		).Error("tokens delivery error occurred")
	}
}

type tokenDeviveryRequest struct {
	Email          string        `json:"email"`
	Token          string        `json:"token"`
	CallbackUrl    string        `json:"callback_url"`
	CallbackUrlTtl time.Duration `json:"callback_url_ttl"`
}

func (c *tokensDeliveryRequests) Consume(ctx context.Context) {
	var err error
	defer c.handleError(ctx, &err)

	message, err := c.reader.FetchMessage(ctx)
	if err != nil {
		return
	}

	var tokensDeliveryRequest tokenDeviveryRequest

	err = json.Unmarshal(message.Value, &tokensDeliveryRequest)
	if err != nil {
		// skip messages with invalid structure
		err = c.reader.CommitMessages(ctx, message)
		return
	}

	Expired := time.Since(message.Time) >= tokensDeliveryRequest.CallbackUrlTtl
	if Expired {
		c.logger.Debugf("Message expired, message sended: %s. %s since message sended. linkTTL: %s",
			message.Time, time.Since(message.Time), time.Duration(tokensDeliveryRequest.CallbackUrlTtl))
		err = c.reader.CommitMessages(ctx, message)
		return
	}

	topic := service.EmailVerificationTopic
	if message.Topic == passwordChangeTopic {
		topic = service.PasswordChangingTopic
	}

	err = c.service.SendTokenToEmail(ctx, tokensDeliveryRequest.Email, tokensDeliveryRequest.CallbackUrl+"/"+tokensDeliveryRequest.Token,
		topic, tokensDeliveryRequest.CallbackUrlTtl-time.Since(message.Time))

	if err != nil {
		return
	}

	err = c.reader.CommitMessages(ctx, message)
}

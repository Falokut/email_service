package events

import (
	"context"
	"encoding/json"
	"errors"

	"github.com/Falokut/email_service/internal/models"
	"github.com/Falokut/email_service/internal/service"
	"github.com/segmentio/kafka-go"
	"github.com/sirupsen/logrus"
)

type ordersEventsConsumer struct {
	reader  *kafka.Reader
	logger  *logrus.Logger
	service service.MailService
}

const (
	orderCreatedTopic = "order_created"
)

func NewOrdersEventsConsumer(
	cfg KafkaReaderConfig,
	logger *logrus.Logger,
	service service.MailService) *ordersEventsConsumer {
	r := kafka.NewReader(kafka.ReaderConfig{
		Brokers:          cfg.Brokers,
		GroupTopics:      []string{orderCreatedTopic},
		GroupID:          cfg.GroupID,
		Logger:           logger,
		ReadBatchTimeout: cfg.ReadBatchTimeout,
	})

	return &ordersEventsConsumer{
		reader:  r,
		logger:  logger,
		service: service,
	}
}

func (c *ordersEventsConsumer) Run(ctx context.Context) {
	for {
		select {
		default:
			c.Consume(ctx)
		case <-ctx.Done():
			c.logger.Info("orders events consumer shutting down")
			c.reader.Close()
			c.logger.Info("orders events consumer shutted down")
			return
		}
	}
}

func (e *ordersEventsConsumer) Shutdown() error {
	return e.reader.Close()
}

func (e *ordersEventsConsumer) handleError(ctx context.Context, err *error) {
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

func (e *ordersEventsConsumer) logError(err error, functionName string) {
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

type orderCreated struct {
	Email string       `json:"email"`
	Order models.Order `json:"order"`
}

func (c *ordersEventsConsumer) Consume(ctx context.Context) {
	var err error
	defer c.handleError(ctx, &err)

	message, err := c.reader.FetchMessage(ctx)
	if err != nil {
		return
	}

	var orderCreated orderCreated

	err = json.Unmarshal(message.Value, &orderCreated)
	if err != nil {
		// skip messages with invalid structure
		err = c.reader.CommitMessages(ctx, message)
		return
	}

	err = c.service.SendOrderCreatedNotification(ctx, orderCreated.Email, orderCreated.Order)
	if err != nil {
		return
	}

	err = c.reader.CommitMessages(ctx, message)
}

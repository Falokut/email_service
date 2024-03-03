package main

import (
	"context"
	"os"
	"os/signal"
	"sync"
	"syscall"

	"github.com/Falokut/email_service/internal/config"
	"github.com/Falokut/email_service/internal/email"
	"github.com/Falokut/email_service/internal/events"
	"github.com/Falokut/email_service/internal/screeningsservice"
	"github.com/Falokut/email_service/internal/service"
	"github.com/Falokut/email_service/pkg/logging"
	"github.com/sirupsen/logrus"
)

func main() {
	logging.NewEntry(logging.ConsoleOutput)
	logger := logging.GetLogger()

	cfg := config.GetConfig()
	log_level, err := logrus.ParseLevel(cfg.LogLevel)
	if err != nil {
		logger.Fatal(err)
	}

	logger.Logger.SetLevel(log_level)

	screeningService, err := screeningsservice.NewScreeningsService(
		cfg.CinemaServiceConfig.Addr, cfg.CinemaServiceConfig.SecureConfig,
		cfg.MoviesServiceConfig.Addr, cfg.MoviesServiceConfig.SecureConfig, logger.Logger)

	if err != nil {
		logger.Error(err)
		return
	}
	defer screeningService.Shutdown()

	subjects := map[service.MailSubjectType]string{
		service.EmailVerfication: cfg.EmailVerificationConfig.Subject,
		service.OrderCreated:     cfg.OrderCreatedConfig.Subject,
		service.PasswordChanging: cfg.ChangePasswordConfig.Subject,
	}
	templateNames := map[service.MailSubjectType]string{
		service.EmailVerfication: cfg.EmailVerificationConfig.Template,
		service.OrderCreated:     cfg.OrderCreatedConfig.Template,
		service.PasswordChanging: cfg.ChangePasswordConfig.Template,
	}

	mailSender := email.NewMailSender(cfg.MailSenderCfg, logger.Logger)
	service, err := service.NewMailService(mailSender, screeningService, subjects, templateNames)
	if err != nil {
		logger.Error(err)
		return
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	logger.Infoln("event consumers initializing")
	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		logger.Info("Running orders events consumer")
		ordersEventsConsumer := events.NewOrdersEventsConsumer(getKafkaReaderConfig(cfg.OrdersEventsConfig),
			logger.Logger, service)
		ordersEventsConsumer.Run(ctx)
		wg.Done()
	}()
	wg.Add(1)
	go func() {
		logger.Info("Running tokens delivery request consumer")
		tokensDeliveryRequestsConsumer := events.NewTokensDeliveryRequestsConsumer(getKafkaReaderConfig(cfg.TokensDeliveryRequestsConfig),
			logger.Logger, service)
		tokensDeliveryRequestsConsumer.Run(ctx)
		wg.Done()
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGHUP, syscall.SIGTERM)

	<-quit
	wg.Wait()
	logger.Infoln("Shutted down successfully")
}

func getKafkaReaderConfig(cfg config.KafkaReaderConfig) events.KafkaReaderConfig {
	return events.KafkaReaderConfig{
		Brokers:          cfg.Brokers,
		GroupID:          cfg.GroupID,
		ReadBatchTimeout: cfg.ReadBatchTimeout,
	}
}

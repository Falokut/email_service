package main

import (
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"github.com/Falokut/email_service/internal/config"
	"github.com/Falokut/email_service/internal/email"
	logging "github.com/Falokut/online_cinema_ticket_office.loggerwrapper"
	"github.com/segmentio/kafka-go"
	"github.com/sirupsen/logrus"
)

func main() {
	logging.NewEntry(logging.ConsoleOutput)
	logger := logging.GetLogger()
	termChan := make(chan os.Signal, 1)
	signal.Notify(termChan, syscall.SIGTERM, syscall.SIGINT)

	appCfg := config.GetConfig()
	log_level, err := logrus.ParseLevel(appCfg.LogLevel)
	if err != nil {
		logger.Fatal(err)
	}

	logger.Logger.SetLevel(log_level)

	mailSender := email.NewMailSender(appCfg.MailSenderCfg, logger.Logger)

	logger.Infoln("kafka consumer initializing")
	kafkaReader := NewKafkaReader(*appCfg)

	logger.Infoln("worker initializing")
	mailWorker := email.NewMailWorker(mailSender, logger.Logger, appCfg.MailWorkerCfg, kafkaReader)
	go func() {
		mailWorker.Run()
	}()

	var wg sync.WaitGroup
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGHUP, syscall.SIGTERM)

	<-quit
	kafkaReader.Close()
	wg.Wait()
	logger.Infoln("Shutted down successfully")
}

func NewKafkaReader(appCfg config.Config) *kafka.Reader {
	r := kafka.NewReader(kafka.ReaderConfig{
		Brokers:        appCfg.KafkaConfig.Brokers,
		GroupID:        appCfg.KafkaConfig.GroupID,
		Topic:          appCfg.KafkaConfig.Topic,
		MaxBytes:       appCfg.KafkaConfig.MaxBytes,
		Logger:         logging.GetLogger(),
		MaxAttempts:    2,
		StartOffset:    kafka.LastOffset,
		QueueCapacity:  appCfg.KafkaConfig.QueueCapacity,
		CommitInterval: time.Millisecond * 10,
	})
	return r
}

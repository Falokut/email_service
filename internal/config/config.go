package config

import (
	"crypto/tls"
	"crypto/x509"
	"errors"
	"sync"
	"time"

	"github.com/Falokut/email_service/internal/email"
	"github.com/Falokut/email_service/pkg/logging"
	"github.com/ilyakaznacheev/cleanenv"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/credentials/insecure"
)

type KafkaReaderConfig struct {
	Brokers          []string      `yaml:"brokers"`
	GroupID          string        `yaml:"group_id"`
	ReadBatchTimeout time.Duration `yaml:"read_batch_timeout"`
}

type Config struct {
	LogLevel      string                 `yaml:"log_level" env:"LOG_LEVEL"`
	MailSenderCfg email.MailSenderConfig `yaml:"mail_sender"`

	CinemaServiceConfig struct {
		Addr         string                 `yaml:"addr" env:"CINEMA_SERVICE_ADDRESS"`
		SecureConfig ConnectionSecureConfig `yaml:"secure_config"`
	} `yaml:"cinema_service_config"`

	MoviesServiceConfig struct {
		Addr         string                 `yaml:"addr" env:"MOVIES_SERVICE_ADDRESS"`
		SecureConfig ConnectionSecureConfig `yaml:"secure_config"`
	} `yaml:"movies_service_config"`

	OrdersEventsConfig           KafkaReaderConfig `yaml:"orders_events"`
	TokensDeliveryRequestsConfig KafkaReaderConfig `yaml:"tokens_delivery_requests"`

	EmailVerificationConfig struct {
		Subject  string `yaml:"subject" env:"EMAIL_VERIFICATION_SUBJECT"`
		Template string `yaml:"template" env:"EMAIL_VERIFICATION_TEMPLATE"`
	} `yaml:"email_verification"`

	ChangePasswordConfig struct {
		Subject  string `yaml:"subject" env:"CHANGE_PASSWORD_SUBJECT"`
		Template string `yaml:"template" env:"CHANGE_PASSWORD_TEMPLATE"`
	} `yaml:"change_password"`

	OrderCreatedConfig struct {
		Subject  string `yaml:"subject" env:"ORDER_CREATED_SUBJECT"`
		Template string `yaml:"template" env:"ORDER_CREATED_TEMPLATE"`
	} `yaml:"order_created"`
}

const configsPath string = "configs/"

var instance *Config
var once sync.Once

func GetConfig() *Config {
	once.Do(func() {
		instance = &Config{}

		logger := logging.GetLogger()
		if err := cleanenv.ReadConfig(configsPath+"config.yml", instance); err != nil {
			help, _ := cleanenv.GetDescription(instance, nil)
			logger.Fatal(help, " ", err)
		}
	})
	return instance
}

type DialMethod = string

const (
	Insecure                 DialMethod = "INSECURE"
	InsecureSkipVerify       DialMethod = "INSECURE_SKIP_VERIFY"
	ClientWithSystemCertPool DialMethod = "CLIENT_WITH_SYSTEM_CERT_POOL"
)

type ConnectionSecureConfig struct {
	Method DialMethod `yaml:"dial_method"`
	// Only for client connection with system pool
	ServerName string `yaml:"server_name"`
}

func (c ConnectionSecureConfig) GetGrpcTransportCredentials() (grpc.DialOption, error) {
	if c.Method == Insecure {
		return grpc.WithTransportCredentials(insecure.NewCredentials()), nil
	}

	if c.Method == InsecureSkipVerify {
		return grpc.WithTransportCredentials(credentials.NewTLS(&tls.Config{InsecureSkipVerify: true})), nil
	}

	if c.Method == ClientWithSystemCertPool {
		certPool, err := x509.SystemCertPool()
		if err != nil {
			return grpc.EmptyDialOption{}, err
		}
		return grpc.WithTransportCredentials(credentials.NewClientTLSFromCert(certPool, c.ServerName)), nil
	}

	return nil, errors.ErrUnsupported
}

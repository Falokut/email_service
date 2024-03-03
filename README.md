# Email service #

[![Go Report Card](https://goreportcard.com/badge/github.com/Falokut/email_service)](https://goreportcard.com/report/github.com/Falokut/email_service)
[![go.dev reference](https://img.shields.io/badge/go.dev-reference-007d9c?logo=go&logoColor=white&style=flat-square)](https://pkg.go.dev/github.com/Falokut/email_service)
![GitHub go.mod Go version](https://img.shields.io/github/go-mod/go-version/Falokut/email_service)
[![Go](https://github.com/Falokut/email_service/actions/workflows/go.yml/badge.svg?branch=master)](https://github.com/Falokut/email_service/actions/workflows/go.yml) ![](https://changkun.de/urlstat?mode=github&repo=Falokut/email_service)
[![License](https://img.shields.io/badge/license-MIT-green)](./LICENSE)
---

# Content
+ [Configuration](#configuration)
    + [Params info](#configuration-params-info)
        + [Secure connection config](#secure-connection-config)
        + [Kafka reader config](#kafka-reader-config)
+ [Metrics](#metrics)
+ [Docs](#docs)
+ [Author](#author)
+ [License](#license)
---------


# Configuration

1. Create a configuration file or change the config.yml file in docker\containers-configs.
If you are creating a new configuration file, specify the path to it in docker-compose volume section (your-path/config.yml:configs/)
2. Configure kafka broker [example compose file](kafka-cluster.yml)


## Configuration params info
if supported values is empty, then any type values are supported

| yml name | yml section | env name | param type| description | supported values |
|-|-|-|-|-|-|
| log_level   |      | LOG_LEVEL  |   string   |      logging level        | panic, fatal, error, warning, warn, info, debug, trace|
| email_password   |   mail_sender   | EMAIL_PASSWORD  |   string   |password or api key||
| email_port   |   mail_sender   | EMAIL_PORT  |   int   |smtp server port||
| email_host   |   mail_sender   | EMAIL_PASSWORD  |   string   |smtp server host name||
| email_address   |   mail_sender   | EMAIL_ADDRESS  |   string   |email address from which the emails will be sent ||
| email_login   |   mail_sender   | EMAIl_LOGIN  |   string   |||
| enable_TLS   |   mail_sender   | ENABLE_TLS  |   bool   |enable or disable tls for stmp server connection||
| addr   |   cinema_service_config   | CINEMA_SERVICE_ADDRESS  |   string   | cinema service address|all valid addresses formatted like host:port or ip-address:port|
| secure_config   |  cinema_service_config    |  |  nested yml configuration [secure connection config](#secure-connection-config)||  |
| addr   |   movies_service_config   | MOVIES_SERVICE_ADDRESS  |   string   | movies service address|all valid addresses formatted like host:port or ip-address:port|
| secure_config   |  movies_service_config    |  |  nested yml configuration [secure connection config](#secure-connection-config)||  |
|   subject |    email_verification| EMAIL_VERIFICATION_SUBJECT  |   string   |subject for mail||
|   template |    email_verification| EMAIL_VERIFICATION_TEMPLATE  |   string   |html template name for mail||
|   subject |    change_password| CHANGE_PASSWORD_SUBJECT  |   string   |subject for mail||
|   template |    change_password| CHANGE_PASSWORD_TEMPLATE  |   string   |html template name for mail||
|   subject |    order_created| ORDER_CREATED_SUBJECT  |   string   |subject for mail||
|   template |    order_created| ORDER_CREATED_TEMPLATE  |   string   |html template name for mail||
|orders_events|||nested yml configuration  [kafka reader config](#kafka-reader-config)|configuration for kafka connection ||
|tokens_delivery_requests|||nested yml configuration  [kafka reader config](#kafka-reader-config)|configuration for kafka connection ||


### Secure connection config
|yml name| param type| description | supported values |
|-|-|-|-|
|dial_method|string|dial method|INSECURE,INSECURE_SKIP_VERIFY,CLIENT_WITH_SYSTEM_CERT_POOL,SERVER|
|server_name|string|server name overriding, used when dial_method=CLIENT_WITH_SYSTEM_CERT_POOL||
|cert_name|string|certificate file name, used when dial_method=SERVER||
|key_name|string|key file name, used when dial_method=SERVER||

### Kafka reader config
|yml name| env name|param type| description | supported values |
|-|-|-|-|-|
|brokers||[]string, array of strings|list of all kafka brokers||
|group_id||string|id or name for consumer group||
|read_batch_timeout||time.Duration with positive duration|amount of time to wait to fetch message from kafka messages batch|[supported values](#time.Duration-yaml-supported-values)|

# Author

- [@Falokut](https://github.com/Falokut) - Primary author of the project

# License

This project is licensed under the terms of the [MIT License](https://opensource.org/licenses/MIT).

---
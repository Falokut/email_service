# Content

+ [Configuration](#configuration)
+ [Metrics](#metrics)
+ [Docs](#docs)
+ [Author](#author)
+ [License](#license)
---------


# Configuration

1. Create .env in root dir and provide EMAIL_PASSWORD  
Example .env
```env
EMAIL_PASSWORD="passwordOrAPIKEY_ForEmail"
```
2. Expose this vars inside config.yml in folder  docker/containers-configs/app-configs
``` yaml
mail_sender:
  email_port: 465              # smtp port       
  email_host: "smtp.yandex.ru" # smtp host
  email_address: "Email"       # email of the sender of the emails
  email_login: "YourLogin"     # login for email
```
3. Configure kafka broker [example compose file](kafka-cluster.yml)


# Author

- [@Falokut](https://github.com/Falokut) - Primary author of the project

# License

This project is licensed under the terms of the [MIT License](https://opensource.org/licenses/MIT).

---
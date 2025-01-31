# SMTP Notifier

SMTP Notifier implements notifier for send SMTP notifications.

## Configuration

The Subscription service using SMTP Notifier is configured using the environment variables presented in the
following table. Note that any unset variables will be replaced with their
default values.

| Variable                          | Description                                                             | Default                        |
| --------------------------------- | ----------------------------------------------------------------------- | ------------------------------ |
| MF_SMTP_NOTIFIER_LOG_LEVEL        | Log level for SMT Notifier (debug, info, warn, error)                   | info                           |
| MF_SMTP_NOTIFIER_FROM_ADDRESS     | From address for SMTP notifications                                     |                                |
| MF_SMTP_NOTIFIER_CONFIG_PATH      | Path to the config file with message broker subjects configuration      | disable                        |
| MF_SMTP_NOTIFIER_HTTP_HOST        | SMTP Notifier service HTTP host                                         | localhost                      |
| MF_SMTP_NOTIFIER_HTTP_PORT        | SMTP Notifier service HTTP port                                         | 9015                           |
| MF_SMTP_NOTIFIER_HTTP_SERVER_CERT | SMTP Notifier service HTTP server certificate path                      | ""                             |
| MF_SMTP_NOTIFIER_HTTP_SERVER_KEY  | SMTP Notifier service HTTP server key                                   | ""                             |
| MF_SMTP_NOTIFIER_DB_HOST          | Database host address                                                   | localhost                      |
| MF_SMTP_NOTIFIER_DB_PORT          | Database host port                                                      | 5432                           |
| MF_SMTP_NOTIFIER_DB_USER          | Database user                                                           | mainflux                       |
| MF_SMTP_NOTIFIER_DB_PASS          | Database password                                                       | mainflux                       |
| MF_SMTP_NOTIFIER_DB_NAME          | Name of the database used by the service                                | subscriptions                  |
| MF_SMTP_NOTIFIER_DB_SSL_MODE      | Database connection SSL mode (disable, require, verify-ca, verify-full) | disable                        |
| MF_SMTP_NOTIFIER_DB_SSL_CERT      | Path to the PEM encoded cert file                                       | ""                             |
| MF_SMTP_NOTIFIER_DB_SSL_KEY       | Path to the PEM encoded certificate key                                 | ""                             |
| MF_SMTP_NOTIFIER_DB_SSL_ROOT_CERT | Path to the PEM encoded root certificate file                           | ""                             |
| MF_JAEGER_URL                     | Jaeger server URL                                                       | http://jaeger:14268/api/traces |
| MF_BROKER_URL                     | Message broker URL                                                      | nats://127.0.0.1:4222          |
| MF_EMAIL_HOST                     | Mail server host                                                        | localhost                      |
| MF_EMAIL_PORT                     | Mail server port                                                        | 25                             |
| MF_EMAIL_USERNAME                 | Mail server username                                                    |                                |
| MF_EMAIL_PASSWORD                 | Mail server password                                                    |                                |
| MF_EMAIL_FROM_ADDRESS             | Email "from" address                                                    |                                |
| MF_EMAIL_FROM_NAME                | Email "from" name                                                       |                                |
| MF_EMAIL_TEMPLATE                 | Email template for sending notification emails                          | email.tmpl                     |
| MF_AUTH_GRPC_URL                  | Users service gRPC URL                                                  | localhost:7001                 |
| MF_AUTH_GRPC_TIMEOUT              | Users service gRPC request timeout in seconds                           | 1s                             |
| MF_AUTH_GRPC_CLIENT_TLS           | Users service gRPC TLS flag                                             | false                          |
| MF_AUTH_GRPC_CA_CERT              | Path to Users service CA cert in pem format                             | ""                             |
| MF_AUTH_CLIENT_TLS                | Auth client TLS flag                                                    | false                          |
| MF_AUTH_CA_CERTS                  | Path to Auth client CA certs in pem format                              | ""                             |
| MF_SEND_TELEMETRY                 | Send telemetry to mainflux call home server                             | true                           |
| MF_SMTP_NOTIFIER_INSTANCE_ID      | SMTP Notifier instance ID                                               | ""                             |

## Usage

Starting service will start consuming messages and sending emails when a message is received.

[doc]: https://docs.mainflux.io

# Mail Store Service

Service for persisting mail sending requests and dealing with mail sending business logic, if you want to simply send a email and dont care about storing the request check the mailer service.

---

## Configuration

configuration is set by a yml config file and enviroment variables, each variable on the yaml file can be overwrittern by a env var,
check `config/config.yml` for details.

when developing its easier to use the yml equivalent of those variables on your `config/config.dev.yml` file and running the service
with `make run_dev` or `go run cmd/main.go --config-file="./config/config.dev.yml"`

---

## Rabbitmq

This services consumes a single queue defined by the `RMQ_QUEUE` env var (defaults to `mail_store`) the operation name should be defined on the message `type` property
and the message body should be the payload for said operation

if the amqp delivery `correlation id` and `reply to` properties are set the rpc response is sent to the queue on the `reply to` property, with the same `correlation id`.



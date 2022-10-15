package config

import (
	"flag"

	"github.com/ilyakaznacheev/cleanenv"
)

type AppConfig struct {
	Debug bool `yaml:"debug" env:"APP_DEBUG"`
}

type DbConfig struct {
	Url string `env-required:"true" yaml:"url" env:"DB_URL"`
}

type TracerConfig struct {
	Url         string `env-required:"true" yaml:"url" env:"TRACER_URL"`
	ServiceName string `env-required:"true" yaml:"service_name" env:"TRACER_SERVICE_NAME"`
}

type RmqConfig struct {
	Url                        string `env-required:"true" yaml:"url" env:"RMQ_URL"`
	Queue                      string `env-required:"true" yaml:"queue" env:"RMQ_QUEUE"`
	RpcTimeout                 int    `env-required:"true" yaml:"rpc_timeout" env:"RMQ_RPC_TIMEOUT"`
	ReconnectWaitTime          int    `env-required:"true" yaml:"reconnect_wait_time" env:"RMQ_RECONNECT_WAIT_TIME"`
	MailerServiceQueue         string `env-required:"true" yaml:"mailer_service_queue" env:"RMQ_MAILER_SERVICE_QUEUE"`
	MailerServiceResponseQueue string `env-required:"true" yaml:"mailer_service_response_queue" env:"RMQ_MAILER_SERVICE_RESPONSE_QUEUE"`
}

type Config struct {
	Db     DbConfig     `yaml:"db"`
	App    AppConfig    `yaml:"app"`
	Rmq    RmqConfig    `yaml:"rmq"`
	Tracer TracerConfig `yaml:"tracer"`
}

func Parse() (*Config, error) {
	var cfgFilePath = flag.String("config-file", "/etc/config.yml", "A filepath to the yml file containing the microservice configuration")
	flag.Parse()

	cfg := &Config{}

	if err := cleanenv.ReadConfig(*cfgFilePath, cfg); err != nil {
		return nil, err
	}

	return cfg, nil
}

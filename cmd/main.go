package main

import (
	"context"
	"log"
	"mail-store-ms/config"
	"mail-store-ms/controller"
	"mail-store-ms/db"
	"mail-store-ms/db/repository"
	"mail-store-ms/queue"
	"mail-store-ms/services/mail"
	"mail-store-ms/tracer"
	"os"
	"os/signal"
	"syscall"

	"github.com/go-playground/validator/v10"
)

// The version/build, this gets replaced at build time to the commit SHA
// with the use of linker flags. see ldfflags on the makefile build cmd

var version = "development"
var build = "development"

func init() {
	log.Println("[ GIT ] build:   ", build)
	log.Println("[ GIT ] version: ", version)
}

func main() {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	cfg, err := config.Parse()
	if err != nil {
		log.Fatalf("[CONFIG] failed to parse config: %v", err)
	}

	err = tracer.Start(&cfg.Tracer)
	if err != nil {
		log.Fatalf("[TRACER] failed to init tracer: %v", err)
	}
	defer tracer.Stop(ctx)

	repo := repository.New(db.New(cfg.Db.Url, cfg.App.Debug))

	queue := queue.New(cfg.Rmq)
	mailer := mail.New(cfg, queue, repo)

	queue.DeliveryRouter = controller.NewRouter(queue, mailer, repo, validator.New())

	queue.Start()
	defer queue.Stop()

	exit := make(chan os.Signal, 1)
	signal.Notify(exit, os.Interrupt, syscall.SIGINT, syscall.SIGTERM)

	<-exit
}

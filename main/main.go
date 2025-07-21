package main

import (
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/BorisIosifov/money-transfers-api/controller"
	"github.com/BorisIosifov/money-transfers-api/model"
)

func main() {
	config, err := model.LoadConfig()
	if err != nil {
		log.Fatal(err)
		return
	}
	log.Printf("Config: %+v", config)

	db, err := model.DBConnect(config)
	if err != nil {
		log.Fatal(err)
		return
	}

	signals := make(chan os.Signal, 1)
	signal.Notify(signals, os.Interrupt, syscall.SIGTERM, syscall.SIGHUP, syscall.SIGQUIT, syscall.SIGABRT)

	ctrl := controller.Controller{Config: config, DB: db, NeedToStop: make(chan bool)}

	go func() {
		sig := <-signals
		log.Printf("Exiting program because signal `%v` received", sig)
		ctrl.Destroy()
	}()

	ctrl.Run()
}

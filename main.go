package main

import (
	"os"
	"os/signal"
	"syscall"

	"oauth2.0go/authorizationServer"
	"oauth2.0go/client"
	"oauth2.0go/protectedResource"
)

func main() {
	client.Start()

	authorizationServer.Start()

	protectedResource.Start()

	c := make(chan os.Signal, 1)
	signal.Notify(c, syscall.SIGHUP, syscall.SIGQUIT, syscall.SIGTERM, syscall.SIGINT)
	<-c
}

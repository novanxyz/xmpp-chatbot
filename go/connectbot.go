package main

import (
	"./xmpp"
	"time"
	"os"
	"os/signal"
	"syscall"
)

func main() {
	handler := xmpp.NewXmppHandler()
	handler.Connect(
		"localhost",
		"localhost",
		"me",
		"me")

    exitSignal := make(chan os.Signal)
    signal.Notify(exitSignal, syscall.SIGINT, syscall.SIGTERM)
    <-exitSignal
}

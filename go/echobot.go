package main

import (
	"./xmpp"
	"fmt"
	"os"
	"os/signal"
	"syscall"
)

func main() {
	handler := xmpp.NewXmppHandler()

	HandleMessage := func (element *xmpp.XmppElement) {
		sender := element.Attributes["from"]
		msg := element.Children[0].Text
		fmt.Printf("Got message from %s\n%s\n", sender, msg)
		handler.Message(sender, msg)
	}
	handler.HandleMessage = HandleMessage
	handler.Connect(
		"localhost",
		"localhost",
		"echobot",
		"echobot")

    exitSignal := make(chan os.Signal)
    signal.Notify(exitSignal, syscall.SIGINT, syscall.SIGTERM)
    <-exitSignal
}


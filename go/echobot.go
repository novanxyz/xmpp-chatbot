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
		text := ""
		for _, child := range(element.Children) {
			if child.Tag == "body" {
				text = child.Text
			}
		}

		fmt.Printf("Got message from %s\n%s\n", sender, text)
		handler.Message(sender, text)
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


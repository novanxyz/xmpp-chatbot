package xmpp

import (
	"testing"
)

func TestCanCreate(t *testing.T) {
	var elem = new(XmppElement)
	elem.Tag = "test"
}

func TestXmppElement_Children(t *testing.T) {
	var elem = new(XmppElement)
	var child = new(XmppElement)
	elem.Children = append(elem.Children, child)
}